package policy

import (
	"encoding/json"
	"log"

	"github.com/canonical/starlark/starlark"
	"github.com/st3v3nmw/devd/internal/gatherer"
)

type ValidationResult struct {
	Compliant  bool     `json:"compliant"`
	Violations []string `json:"violations"`
	Plan       []string `json:"plan"`
}

func (r ValidationResult) String() string {
	out, _ := json.MarshalIndent(r, "", "  ")
	return string(out)
}

func Check(policy string, args map[string]string) ValidationResult {
	// Declare context to pass to Starlark
	context := gatherer.Get(policy)
	for k, v := range args {
		context[k] = starlark.String(v)
	}

	// Execute Starlark program
	thread := &starlark.Thread{Name: "Check Password Policy"}
	globals, err := starlark.ExecFile(thread, "policies/starlark/snaps.star", nil, context)
	if err != nil {
		log.Fatalf("Error executing policy file: %v\n", err)
	}

	// Check the policy
	checkPolicy := globals["check_policy"]
	v, err := starlark.Call(thread, checkPolicy, nil, nil)
	if err != nil {
		log.Fatalf("Error checking policy: %v\n", err)
	}

	// Extract returned values
	dict, ok := v.(*starlark.Dict)
	if !ok {
		log.Fatal("Expected dict return value")
	}

	compliantVal, found, err := dict.Get(starlark.String("compliant"))
	if !found || err != nil {
		log.Fatalf("Couldn't extract .compliant value: %v\n", err)
	}
	compliant := bool(compliantVal.(starlark.Bool))

	violationsVal, found, err := dict.Get(starlark.String("violations"))
	if !found || err != nil {
		log.Fatalf("Couldn't extract .violations value: %v\n", err)
	}

	violationsList := violationsVal.(*starlark.List)
	violations := make([]string, violationsList.Len())
	for i := 0; i < violationsList.Len(); i++ {
		violations[i] = string(violationsList.Index(i).(starlark.String))
	}

	planVal, found, err := dict.Get(starlark.String("plan"))
	if !found || err != nil {
		log.Fatalf("Couldn't extract .plan value: %v\n", err)
	}

	planList := planVal.(*starlark.List)
	plan := make([]string, planList.Len())
	for i := 0; i < planList.Len(); i++ {
		plan[i] = string(planList.Index(i).(starlark.String))
	}

	return ValidationResult{
		Compliant:  compliant,
		Violations: violations,
		Plan:       plan,
	}
}
