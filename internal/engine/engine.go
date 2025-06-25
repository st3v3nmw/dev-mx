package engine

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/canonical/starlark/starlark"
	"github.com/st3v3nmw/devd/internal/gatherer"
	"github.com/st3v3nmw/devd/policies"
)

type ValidationResult struct {
	Checked bool   `json:"checked"`
	Error   string `json:"error,omitempty"`

	Compliant  bool     `json:"compliant"`
	Violations []string `json:"violations"`
	Plan       []string `json:"plan"`
}

func (r ValidationResult) String() string {
	out, _ := json.MarshalIndent(r, "", "  ")
	return string(out)
}

// Unsafe & very kumbaya, this is a significant attack vector.
func (r ValidationResult) ExecutePlan() {
	if len(r.Plan) == 0 {
		return
	}

	go func() {
		for _, command := range r.Plan {
			cmd := exec.Command("bash", "-c", command)
			cmd.Run()
		}
	}()
}

func CheckPolicy(policy string) ValidationResult {
	var result ValidationResult

	// Declare context to pass to Starlark
	context, err := gatherer.Get(policy)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Execute Starlark program
	thread := &starlark.Thread{Name: "Check Password Policy"}
	policyFile := fmt.Sprintf("starlark/%s.star", policy)

	src, err := policies.StarlarkPoliciesFS.ReadFile(policyFile)
	if err != nil {
		result.Error = fmt.Sprintf("cannot execute policy file: %v", err)
		return result
	}

	globals, err := starlark.ExecFile(thread, policyFile, src, context)
	if err != nil {
		result.Error = fmt.Sprintf("cannot execute policy file: %v", err)
		return result
	}

	// Check the policy
	checkPolicy := globals["check_policy"]
	v, err := starlark.Call(thread, checkPolicy, nil, nil)
	if err != nil {
		result.Error = fmt.Sprintf("cannot check policy: %v", err)
		return result
	}

	result.Checked = true

	// Extract returned values
	dict, ok := v.(*starlark.Dict)
	if !ok {
		result.Error = "cannot parse result: expected dict return value"
		return result
	}

	compliantVal, found, err := dict.Get(starlark.String("compliant"))
	if !found || err != nil {
		result.Error = fmt.Sprintf("cannot extract .compliant value: %v", err)
		return result
	}
	result.Compliant = bool(compliantVal.(starlark.Bool))

	violationsVal, found, err := dict.Get(starlark.String("violations"))
	if !found || err != nil {
		result.Error = fmt.Sprintf("cannot extract .violations value: %v", err)
		return result
	}

	violationsList := violationsVal.(*starlark.List)
	result.Violations = make([]string, violationsList.Len())
	for i := 0; i < violationsList.Len(); i++ {
		result.Violations[i] = string(violationsList.Index(i).(starlark.String))
	}

	planVal, found, err := dict.Get(starlark.String("plan"))
	if !found || err != nil {
		result.Error = fmt.Sprintf("cannot extract .plan value: %v", err)
		return result
	}

	planList := planVal.(*starlark.List)
	result.Plan = make([]string, planList.Len())
	for i := 0; i < planList.Len(); i++ {
		result.Plan[i] = string(planList.Index(i).(starlark.String))
	}

	// Unsafe & very kumbaya, this is a significant attack vector.
	result.ExecutePlan()

	return result
}
