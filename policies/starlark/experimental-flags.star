KNOWN_FLAGS = [
    "layouts",
    "parallel-instances",
    "hotplug",
    "per-user-mount-namespace",
    "refresh-app-awareness",
    "classic-preserves-xdg-runtime-dir",
    "user-daemons",
    "dbus-activation",
    "hidden-snap-folder",
    "move-snap-home-dir",
    "check-disk-space-install",
    "check-disk-space-refresh",
    "check-disk-space-remove",
    "gate-auto-refresh-hook",
    "quota-groups",
    "refresh-app-awareness-ux",
    "confdb",
    "confdb-control",
    "apparmor-prompting"
]

DEPENDENCIES = {
    "confdb-control": ["confdb", "parallel-instances"]
}


def check_policy():
    result = {"violations": [], "plan": []}

    for flag in new_flags:
        if flag not in KNOWN_FLAGS:
            result["violations"].append("unknown flag {}".format(flag))

        # Check if this flag has dependencies
        if flag in DEPENDENCIES:
            for dependency in DEPENDENCIES[flag]:
                # Violation if dependency isn't enabled
                if dependency not in experimental_flags or not experimental_flags[dependency]:
                    result["violations"].append(
                        "flag {} requires {} to be enabled".format(flag, dependency)
                    )

    result["compliant"] = len(result["violations"]) == 0
    if result["compliant"]:
        # Unsafe & very kumbaya, this is a significant attack vector.
        # Plan to enable the flags.
        for flag in new_flags:
            result["plan"].append("snap set system experimental.{}=true".format(flag))

    return result
