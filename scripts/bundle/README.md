# scripts/bundle

Scripts that build, test and publish the `exrpd` bundle. See [docs/bundle-dev.md](../../docs/bundle-dev.md) for the details.

## Who uses which file

| Caller | Files |
| --- | --- |
| PR workflow, job `bundle` | `check.sh` -> `cosmovisor.sh --test` |
| PR workflow, job `bundle-e2e` | `plan.sh checkout`, `plan.sh genesis`, `e2e.sh` |
| Bundle workflow, job `plan` | `plan.sh release`, `plan.sh genesis`, `publication.sh` |
| Bundle workflow, job `assets` | `check.sh`, `release.sh` -> `build.sh` -> `cosmovisor.sh`, `e2e.sh`, `publication.sh` |
| Bundle workflow, job `image` | `lib.sh`, `image-smoke.sh`, `publication.sh` |
| Developer | `build.sh`, `release.sh`, `e2e.sh`, `check.sh` |

## Files

| File | What it does |
| --- | --- |
| `plan.sh` | Prints a plan or a version: the plan of this checkout, the plan of a release tag, or the genesis release for a plan. |
| `build.sh` | Builds one bundle executable for one platform: a genesis `exrpd`, an optional upgrade `exrpd` and Cosmovisor, embedded in the launcher. |
| `release.sh` | Builds one bundle set from published release archives: one archive for each platform and a checksums file. Calls `build.sh` for each platform. |
| `cosmovisor.sh` | Builds Cosmovisor from the pinned module source plus the patch in `cosmovisor/`, or runs its tests. |
| `check.sh` | Runs the launcher vet, tests and lint, then the patched Cosmovisor tests. |
| `e2e.sh` | Runs a real chain upgrade through the bundle on a localnet (linux/amd64). Sets up the run, then runs the scenarios in `e2e/`. |
| `e2e/lib.sh` | Helpers of the E2E: checks, node start and stop, queries, chain setup. |
| `e2e/<scenario>.sh` | One E2E scenario each, run in order on the same chain. The comment at the top of each file says what it tests. |
| `image-smoke.sh` | Smoke test of a bundle image: every embedded binary must run. |
| `publication.sh` | Tells the Bundle workflow which release assets and images are missing. |
| `lib.sh` | Helpers sourced by the scripts: archive checksums, platform and plan checks. Loads the Cosmovisor pins. |
| `cosmovisor/` | The Cosmovisor pin (`cosmovisor.env`) and the patch. |

Lint the E2E through its root file, so that shellcheck follows the scenario files: `shellcheck -x scripts/bundle/*.sh`.
