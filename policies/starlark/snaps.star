BLOCK = ["shady-wallet"]

CONFLICTING = {"firefox": ["hello-world"]}


def init():
    engine.observe("snaps", check_policy)


def check_policy(event):
    result = {"violations": [], "plan": []}

    for snap in event.new_snaps:
        if snap in BLOCK:
            result["violations"].append(
                "installation of {} blocked by policy".format(snap)
            )

        # Check if the snap being installed conflicts with any installed snaps
        for installed_snap in event.installed_snaps:
            if installed_snap in CONFLICTING:
                conflicting_snaps = CONFLICTING[installed_snap]
                if snap in conflicting_snaps:
                    result["violations"].append(
                        "cannot install {} because {} is already installed".format(
                            snap, installed_snap
                        )
                    )

        # Check if any installed snaps would conflict with the snap being installed
        if snap in CONFLICTING:
            conflicting_snaps = CONFLICTING[snap]
            for conflicting_snap in conflicting_snaps:
                if conflicting_snap in event.installed_snaps:
                    result["violations"].append(
                        "cannot install {} because it conflicts with installed snap {}".format(
                            snap, conflicting_snap
                        )
                    )

    result["compliant"] = len(result["violations"]) == 0
    if result["compliant"]:
        # Unsafe & very kumbaya, this is a significant attack vector.
        # Plan to install the snaps.
        for snap in event.new_snaps:
            result["plan"].append("snap install {}".format(snap))

    engine.set_result(result)
