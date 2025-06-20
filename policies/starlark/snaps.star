CONFLICTING = {"firefox": ["hello-world"]}


def check_policy():
    violations = []

    # Check if the snap being installed conflicts with any installed snaps
    for installed_snap in installed:
        if installed_snap in CONFLICTING:
            conflicting_snaps = CONFLICTING[installed_snap]
            if snap in conflicting_snaps:
                violations.append(
                    "Cannot install '{}' because '{}' is already installed".format(
                        snap, installed_snap
                    )
                )

    # Check if any installed snaps would conflict with the snap being installed
    if snap in CONFLICTING:
        conflicting_snaps = CONFLICTING[snap]
        for conflicting_snap in conflicting_snaps:
            if conflicting_snap in installed:
                violations.append(
                    "Cannot install '{}' because it conflicts with installed snap '{}'".format(
                        snap, conflicting_snap
                    )
                )

    return {
        "compliant": len(violations) == 0,
        "violations": violations,
    }
