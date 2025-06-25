package gatherer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
)

type snaps struct {
	Installed map[string]snapInfo `json:"installed"`
}

type experimentalFlags struct {
	Flags map[string]bool `json:"flags"`
}

func getView(view, path string, previous bool) ([]byte, error) {
	args := []string{"get", "--view", view, path, "-d"}

	if previous {
		args = append(args, "--previous")
	}

	cmd := exec.Command("snapctl", args...)
	output, err := cmd.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return nil, fmt.Errorf("error calling snapctl: %s", string(exitError.Stderr))
		}

		return nil, fmt.Errorf("error calling snapctl: %w", err)
	}

	return output, nil
}

func getInstalledSnapsFromView(view string) (*snaps, error) {
	output, err := getView(view, "installed", false)
	if err != nil {
		return nil, fmt.Errorf("cannot get installed snaps (%s): %w", view, err)
	}

	var data snaps
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("cannot unmarshal installed snaps (%s): %w", view, err)
	}

	return &data, nil
}

func getExperimentalFlagsFromView(view string) (*experimentalFlags, error) {
	output, err := getView(view, "flags", false)
	if err != nil {
		return nil, fmt.Errorf("cannot get experimental flags (%s): %w", view, err)
	}

	var data experimentalFlags
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("cannot unmarshal experimental flags (%s): %w", view, err)
	}

	return &data, nil
}
