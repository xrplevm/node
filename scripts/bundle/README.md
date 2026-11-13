# Bundle maintainer reference

Run the commands below from the repository root. For node operation, see the [bundle operator guide](../../docs/bundle.md).

## Build

The following local examples use standalone release archives and their `checksums.txt` files in `dist/genesis` and `dist/standalone`. They do not publish anything. Replace the versions and paths with the intended releases.

```sh
(
  set -e
  source scripts/bundle/lib.sh
  verify_archive dist/standalone/checksums.txt dist/standalone/node_12.0.0_Linux_amd64.tar.gz
  mkdir -p dist/target
  tar -xzf dist/standalone/node_12.0.0_Linux_amd64.tar.gz -C dist/target bin/exrpd

  # Two-binary bundle for linux/amd64.
  scripts/bundle/build.sh \
    --genesis-version 11.3.0 \
    --genesis-archive dist/genesis/node_11.3.0_Linux_amd64.tar.gz \
    --genesis-checksums dist/genesis/checksums.txt \
    --upgrade-exrpd dist/target/bin/exrpd --upgrade-version 12.0.0 --plan v12.0.0 \
    --goos linux --goarch amd64 --out dist/two-binary/exrpd

  # Single bundle for linux/amd64.
  scripts/bundle/build.sh \
    --genesis-version 12.0.0 \
    --genesis-archive dist/standalone/node_12.0.0_Linux_amd64.tar.gz \
    --genesis-checksums dist/standalone/checksums.txt \
    --goos linux --goarch amd64 --out dist/single/exrpd
)
```

Use the target release's standalone `bin/exrpd`, not a local node build. `build.sh` requires `--plan` with `--upgrade-exrpd` and `--upgrade-version`, and rejects `--plan` for a single bundle. The genesis plan defaults to `vX.Y.0` for genesis version `X.Y.Z`. Use `--genesis-plan` for a historical exception. For a two-binary bundle, the genesis plan must precede the target plan, and the target version and plan must have the same major and minor numbers.

To build release archives locally, use `release.sh` once per set. Each input directory must contain the standalone archives and their `checksums.txt`. The output directory must be absent or empty.

```sh
scripts/bundle/release.sh --version 12.0.0 --plan v12.0.0 \
  --standalone-dir dist/standalone --out-dir dist/release-single
scripts/bundle/release.sh --version 12.0.0 --plan v12.0.0 \
  --standalone-dir dist/standalone --genesis-version 11.3.0 \
  --genesis-dir dist/genesis --out-dir dist/release-two-binary
```

`release.sh` always requires `--plan`, including for a single set. It accepts `--genesis-plan` with `--genesis-version` for historical exceptions. By default it builds Linux and Darwin archives for amd64 and arm64. Set `PLATFORMS`, for example `PLATFORMS=linux/amd64`, to restrict a local build.

## Test

`cmd/exrpd-bundle` is a separate Go module. The root checks do not cover it. Install the `golangci-lint` version in `.golangci-lint-version`, then run:

```sh
scripts/bundle/check.sh
shellcheck -x scripts/bundle/*.sh
```

The first command runs launcher vet, tests and lint with `GOWORK=off`, then tests the patched Cosmovisor. For a focused launcher test, run `(cd cmd/exrpd-bundle && GOWORK=off go test -count=1 ./...)`.

The [Bundle checks workflow](../../.github/workflows/bundle-e2e.yml) runs unit tests, lint, and E2E on pull requests and pushes to `main` when bundle, upgrade, or lint configuration paths change.

The E2E script builds nothing and downloads nothing. It needs linux/amd64 artifacts: a genesis archive and checksums, the standalone target executable, and at least one bundle executable. With a two-binary bundle, also provide stock Cosmovisor for the migration test. For the build examples above:

```sh
source scripts/bundle/cosmovisor/cosmovisor.env
GOBIN="$PWD/dist/e2e-tools" GOWORK=off go install "$COSMOVISOR_MODULE/cmd/cosmovisor@$COSMOVISOR_VERSION"
scripts/bundle/e2e.sh \
  --genesis-archive dist/genesis/node_11.3.0_Linux_amd64.tar.gz \
  --genesis-checksums dist/genesis/checksums.txt \
  --upgrade-exrpd dist/target/bin/exrpd --upgrade-version 12.0.0 --plan v12.0.0 \
  --bundle dist/two-binary/exrpd --single-bundle dist/single/exrpd \
  --stock-cosmovisor dist/e2e-tools/cosmovisor
```

For a single-only test, omit `--bundle` and `--stock-cosmovisor`. The script deletes its scratch directory at startup. Set `E2E_DIR` to a missing, empty, or previously marked scratch directory and `PORT_BASE` to the start of a free 100-port range when the defaults are unsuitable.

For a focused nesting check on Linux, use an already built bundle and stock Cosmovisor. This check uses temporary homes and starts no node. It also runs as part of the stock migration E2E scenario.

```sh
scripts/bundle/nesting-smoke.sh dist/two-binary/exrpd dist/e2e-tools/cosmovisor
```

## Release

After the standalone release is published, dispatch the [Bundle workflow](../../.github/workflows/bundle.yml). This publishes assets and images. Set `target_tag` to the published target tag, such as `v12.0.0`. Leave `genesis_tag` empty for a single set, or supply the published pre-upgrade tag, such as `v11.3.0`, to request both a single and a two-binary set. The workflow can run from `main` or another branch. It builds the launcher from the selected branch, not the target tag.

For new releases, version `X.Y.Z` uses plan `vX.Y.0`. The workflow reads the target plan from the release tag with `plan.sh release`; check historical exceptions before dispatch. A two-binary bundle requires the target version and plan to have the same major and minor numbers. The workflow does not expose `--genesis-plan`, so its two-binary build uses the default genesis plan.

Each set has four archives, `exrpd-bundle_<set>_<Os>_<arch>.tar.gz`, and `exrpd-bundle_<set>_checksums.txt`. The set is `<version>` for a single bundle or `<genesis>-<version>` for a two-binary bundle. The image tag is `peersyst/exrp:bundle-<set>` and uses the published Linux amd64 archive. A dispatch for another genesis requests a separate two-binary set and reuses the single set.

## Recover a release

The workflow checks assets and images independently. It builds an asset set when none of its five files exist, skips it when all five exist, and refuses a partial set. It uploads archives before the checksums file. It skips an existing image and stops if it cannot determine the registry state. It does not automatically overwrite published assets or images.

- For a transient failure, rerun failed jobs or dispatch again with the same inputs. If the workflow artifact expired, dispatch again to rebuild and test it.
- For a partial asset set, inspect the listed files and check whether the set's image `peersyst/exrp:bundle-<set>` exists. If the image exists, publish a new release and bundle that release, because a rebuild under the same tag can put a different launcher in the archives than in the image. If the image does not exist, delete that set's published files from the release and dispatch again. Do not delete unrelated sets.
- If assets are complete but an image is missing, dispatch again with the same inputs. The workflow can build the missing image from the published archive without rebuilding the set.
- If `genesis_tag` was wrong, remove the wrong `exrpd-bundle_<wrong-genesis>-<target>_*` assets and its `peersyst/exrp:bundle-<wrong-genesis>-<target>` image tag, if published. Then dispatch with the correct genesis. Keep the unrelated single set and image.
- If a defect affects an already published bundle or image, publish a new release and bundle that release. A same-tag rerun leaves existing assets and images unchanged. For a defect found before publication, fix the selected workflow branch and dispatch again.

## Script map

- [`build.sh`](build.sh) builds one bundle executable. [`release.sh`](release.sh) builds a set of release archives and checksums.
- [`plan.sh`](plan.sh) resolves checkout and release plans or selects a genesis release. [`publication.sh`](publication.sh) checks published asset and image state.
- [`check.sh`](check.sh) runs bundle CI checks. [`e2e.sh`](e2e.sh) and [`e2e/`](e2e/) test localnet journeys. [`nesting-smoke.sh`](nesting-smoke.sh) checks refusal under stock Cosmovisor and direct startup. [`image-smoke.sh`](image-smoke.sh) checks a built image. Every installed `exrpd` and Cosmovisor version command must exit successfully.
- [`cosmovisor.sh`](cosmovisor.sh) builds and tests Cosmovisor using the pin in [`cosmovisor.env`](cosmovisor/cosmovisor.env) and the [signal-handling patch](cosmovisor/cosmovisor-v1.7.3-signal.patch). The patch fixes stale signal handlers after an upgrade and stop handling during the upgrade. Remove it when a pinned Cosmovisor release includes the fix.
- [`lib.sh`](lib.sh) provides shared archive, platform, and plan checks.
