package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/canonical/starform/starform"
	"github.com/canonical/starlark/starlark"
	"github.com/st3v3nmw/devd/internal/gatherer"
	"github.com/st3v3nmw/devd/policies"
)

var (
	policyNames = []string{
		"snaps",
		"experimental_flags",
	}
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

type Engine struct {
	scriptSet *starform.ScriptSet
}

type ScriptSource struct {
	name string
}

func (ss *ScriptSource) Path() string {
	return fmt.Sprintf("starlark/%s.star", ss.name)
}

func (ss *ScriptSource) Content(ctx context.Context) ([]byte, error) {
	return policies.StarlarkPoliciesFS.ReadFile(ss.Path())
}

func New() (*Engine, error) {
	setResult := starlark.NewBuiltin("set_result", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		event := starform.Event(thread)
		event.Attrs["result"] = args[0]
		return starlark.None, nil
	})

	app := &starform.AppObject{
		Name:    "engine",
		Methods: []*starlark.Builtin{setResult},
	}

	scriptSet, err := starform.NewScriptSet(&starform.ScriptSetOptions{
		App:       app,
		MaxAllocs: 10 * 1024 * 1024,
	})
	if err != nil {
		return nil, err
	}

	sources := []starform.ScriptSource{}
	for _, policy := range policyNames {
		sources = append(sources, &ScriptSource{name: policy})
	}

	err = scriptSet.LoadSources(context.TODO(), sources)
	if err != nil {
		return nil, fmt.Errorf("cannot load sources: %v", err)
	}

	return &Engine{scriptSet: scriptSet}, nil
}

func (e *Engine) CheckPolicy(policy string) ValidationResult {
	var result ValidationResult

	// Declare context to pass to Starlark
	ctx, err := gatherer.Get(policy)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// Handle event
	event := &starform.EventObject{
		Name:  policy,
		Attrs: ctx,
	}
	err = e.scriptSet.Handle(context.TODO(), event)
	if err != nil {
		result.Error = fmt.Sprintf("cannot check policy: %v", err)
		return result
	}

	result.Checked = true
	v := event.Attrs["result"]

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
