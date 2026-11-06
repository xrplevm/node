# shellcheck shell=bash
# Build the tested two-binary bundle (or copy --bundle) and check its manifest. build.sh must
# also refuse the next minor release, which needs its own plan.

build_args=(--genesis-version "$GENESIS_VERSION" --genesis-archive "$genesis_archive" --genesis-checksums "$genesis_checksums"
	--upgrade-exrpd "$UPGRADE_BIN" --upgrade-version "$UPGRADE_VERSION" --plan "$UPGRADE_PLAN" --goos linux --goarch amd64 --out "$UB")
if [[ $mode == unpatched ]]; then
	cv=$(basename "$cosmovisor_archive")
	verify_archive "$cosmovisor_checksums" "$cosmovisor_archive" >/dev/null || die "cosmovisor release checksum"
	mkdir -p "$E2E_DIR/cv" && tar -xzf "$cosmovisor_archive" -C "$E2E_DIR/cv"
	build_args+=(--cosmovisor "$E2E_DIR/cv/cosmovisor" --cosmovisor-source "unpatched upstream release $cv")
fi
IFS=. read -r plan_major plan_minor _ <<<"${UPGRADE_PLAN#v}"
next_minor=$plan_major.$((plan_minor + 1)).0
# Only the major.minor diagnostic proves the check. Other build errors, such as a missing Go, do not.
build_refuses_next_minor() {
	local out
	if out=$("$here/build.sh" "${build_args[@]}" --upgrade-version "$next_minor" 2>&1); then
		echo "build.sh accepted --upgrade-version $next_minor"
		return 1
	fi
	grep -qF "does not have the major.minor of plan" <<<"$out" && return 0
	echo "build.sh failed without the major.minor diagnostic. Output:"
	echo "$out"
	return 1
}
check "build refuses $next_minor, a minor release that needs its own plan" build_refuses_next_minor
if [[ -n $bundle ]]; then
	cp "$bundle" "$UB" || die "copy $bundle"
else
	"$here/build.sh" "${build_args[@]}" || die "build"
fi
check "manifest cosmovisor source ($mode)" cosmovisor_source "$([[ $mode == patched ]] && echo "$COSMOVISOR_PATCH" || echo unpatched)"

tar -xzf "$genesis_archive" -C "$E2E_DIR/genesis-release" bin/exrpd
GENESIS_BIN=$E2E_DIR/genesis-release/bin/exrpd
check "genesis tool is release v$GENESIS_VERSION" test "$("$GENESIS_BIN" version --home "$E2E_DIR/genesis-release")" = "$GENESIS_VERSION"
GENESIS_PLAN=$(manifest genesis.plan)
check "manifest genesis plan $GENESIS_PLAN is the newest plan of v$GENESIS_VERSION" test "$GENESIS_PLAN" = "$("$here/plan.sh" release "$GENESIS_VERSION")"
check "manifest has one upgrade" test "$(manifest 'upgrades | length')" = 1
check "manifest upgrade plan is $UPGRADE_PLAN" test "$(manifest 'upgrades[0].plan')" = "$UPGRADE_PLAN"
check "manifest genesis is the release v$GENESIS_VERSION exrpd" test "$(manifest genesis.sha256)" = "$(sha256 "$GENESIS_BIN")"
check "manifest upgrade is the standalone $UPGRADE_VERSION exrpd" test "$(manifest 'upgrades[0].sha256')" = "$(sha256 "$UPGRADE_BIN")"
check "manifest launcher commit is set" launcher_commit_is_set "$UB"

check "standalone upgrade exrpd version is $UPGRADE_VERSION" test "$("$UPGRADE_BIN" version --home "$E2E_DIR/genesis-release")" = "$UPGRADE_VERSION"
