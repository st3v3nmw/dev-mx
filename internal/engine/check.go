package engine

import (
	"encoding/json"
	"log"
	"os/exec"

	"github.com/canonical/starlark/starlark"
)

type SnapInfo struct {
	Channel   string `json:"channel"`
	Developer string `json:"developer"`
	Summary   string `json:"summary"`
	Version   string `json:"version"`
}

type InstalledSnaps struct {
	Installed map[string]SnapInfo `json:"installed"`
}

func CheckSnapPolicy(snap string) ValidationResult {
	// Declare context to pass to Starlark
	cmd := exec.Command("snapctl", "get", "--view", ":observe-snaps", "installed", "-d")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	var data InstalledSnaps
	if err := json.Unmarshal(output, &data); err != nil {
		log.Fatal(err)
	}

	installed := &starlark.Dict{}
	for snap, info := range data.Installed {
		meta := &starlark.Dict{}
		meta.SetKey(starlark.String("channel"), starlark.String(info.Channel))
		meta.SetKey(starlark.String("developer"), starlark.String(info.Developer))
		meta.SetKey(starlark.String("summary"), starlark.String(info.Summary))
		meta.SetKey(starlark.String("version"), starlark.String(info.Version))

		installed.SetKey(starlark.String(snap), meta)
	}

	context := starlark.StringDict{
		"snap":      starlark.String(snap),
		"installed": installed,
	}

	// Execute Starlark program
	thread := &starlark.Thread{Name: "Check Password Policy"}
	globals, err := starlark.ExecFile(thread, "policies/starlark/snaps.star", nil, context)
	if err != nil {
		log.Fatalf("Error executing program: %v\n", err)
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
		log.Fatal("Error extracting .compliant value")
	}
	compliant := bool(compliantVal.(starlark.Bool))

	violationsVal, found, err := dict.Get(starlark.String("violations"))
	if !found || err != nil {
		log.Fatal("Error extracting .violations value")
	}

	violationsList := violationsVal.(*starlark.List)
	violations := make([]string, violationsList.Len())
	for i := 0; i < violationsList.Len(); i++ {
		violations[i] = string(violationsList.Index(i).(starlark.String))
	}

	return ValidationResult{
		Compliant:  compliant,
		Violations: violations,
	}
}
