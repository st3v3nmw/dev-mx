package gatherer

import "github.com/canonical/starlark/starlark"

func Get(policy string) starlark.StringDict {
	switch policy {
	case "snaps":
		return getSnapContext()
	}

	return starlark.StringDict{}
}
