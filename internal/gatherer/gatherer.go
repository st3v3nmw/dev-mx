package gatherer

import (
	"fmt"

	"github.com/canonical/starlark/starlark"
)

func Get(policy string) (starlark.StringDict, error) {
	switch policy {
	case "snaps":
		return getSnapContext()
	case "experimental-flags":
		return getExperimentalFlagsContext()
	}

	return nil, fmt.Errorf("unknown policy %s", policy)
}

func getSnapContext() (starlark.StringDict, error) {
	actual, err := getSnapData()
	if err != nil {
		return nil, err
	}

	desired, err := getInstalledSnapsFromView(":manage-snaps")
	if err != nil {
		return nil, err
	}

	newSnaps := &starlark.List{}
	for snap := range desired.Installed {
		_, found := actual[snap]
		if !found {
			newSnaps.Append(starlark.String(snap))
		}
	}

	installed := &starlark.Dict{}
	for snap, info := range actual {
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

	return starlark.StringDict{
		"new_snaps":       newSnaps,
		"installed_snaps": installed,
	}, nil
}

func getExperimentalFlagsContext() (starlark.StringDict, error) {
	actual, err := getExperimentalFlagData()
	if err != nil {
		return nil, err
	}

	desired, err := getExperimentalFlagsFromView(":manage-experimental-flags")
	if err != nil {
		return nil, err
	}

	newFlags := &starlark.List{}
	for flag, newValue := range desired.Flags {
		oldValue, found := actual[flag]
		if (found && newValue != oldValue) || (!found && newValue == true) {
			newFlags.Append(starlark.String(flag))
		}
	}

	flags := &starlark.Dict{}
	for flag, on := range actual {
		flags.SetKey(starlark.String(flag), starlark.Bool(on))
	}

	return starlark.StringDict{
		"new_flags":          newFlags,
		"experimental_flags": flags,
	}, nil
}
