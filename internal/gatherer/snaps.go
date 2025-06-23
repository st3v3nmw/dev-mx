package gatherer

import (
	"encoding/json"
	"log"
	"os/exec"

	"github.com/canonical/starlark/starlark"
)

type snapInfo struct {
	Channel     string `json:"channel"`
	Developer   string `json:"developer"`
	Id          string `json:"id"`
	InstallDate string `json:"install-date"`
	Revision    string `json:"revision"`
	Status      string `json:"status"`
	Summary     string `json:"summary"`
	Version     string `json:"version"`
}

type snaps struct {
	Actual  map[string]snapInfo `json:"installed"`
	Desired map[string]snapInfo `json:"install"`
}

func getInstalledSnaps() *starlark.Dict {
	cmd := exec.Command("snapctl", "get", "--view", ":observe-snaps", "installed", "-d")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Error getting installed snaps: %v\n", err)
	}

	var data snaps
	if err := json.Unmarshal(output, &data); err != nil {
		log.Fatalf("Error unmarshaling installed snaps: %v\n", err)
	}

	installed := &starlark.Dict{}
	for snap, info := range data.Actual {
		meta := &starlark.Dict{}
		meta.SetKey(starlark.String("channel"), starlark.String(info.Channel))
		meta.SetKey(starlark.String("developer"), starlark.String(info.Developer))
		meta.SetKey(starlark.String("id"), starlark.String(info.Id))
		meta.SetKey(starlark.String("install-date"), starlark.String(info.InstallDate))
		meta.SetKey(starlark.String("revision"), starlark.String(info.Revision))
		meta.SetKey(starlark.String("status"), starlark.String(info.Status))
		meta.SetKey(starlark.String("summary"), starlark.String(info.Summary))
		meta.SetKey(starlark.String("version"), starlark.String(info.Version))

		installed.SetKey(starlark.String(snap), meta)
	}

	return installed
}

func getSnapContext() starlark.StringDict {
	return starlark.StringDict{
		"installed": getInstalledSnaps(),
	}
}
