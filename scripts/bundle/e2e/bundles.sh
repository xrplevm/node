# shellcheck shell=bash disable=SC2015

verify_archive "$genesis_checksums" "$genesis_archive" >/dev/null || die "genesis release checksum"
tar -xzf "$genesis_archive" -C "$E2E_DIR/genesis-release" bin/exrpd || die "extract genesis release"
GENESIS_BIN=$E2E_DIR/genesis-release/bin/exrpd

[[ $(timeout 30 "$GENESIS_BIN" version --home "$E2E_DIR/genesis-release") == "$GENESIS_VERSION" ]] || die "genesis version differs from $GENESIS_VERSION"
[[ $(timeout 30 "$UPGRADE_BIN" version --home "$E2E_DIR/genesis-release") == "$UPGRADE_VERSION" ]] || die "upgrade version differs from $UPGRADE_VERSION"

genesis_sha=$(sha256 "$GENESIS_BIN") || die "hash genesis exrpd"
upgrade_sha=$(sha256 "$UPGRADE_BIN") || die "hash upgrade exrpd"

if [[ -n $bundle ]]; then
	cp "$bundle" "$tested_bundle" || die "copy $bundle"
	two_manifest=$(timeout 30 "$tested_bundle" version --bundle --home "$E2E_DIR/no-home" | jq -e '.manifest') || die "read two-binary manifest"
	GENESIS_PLAN=$(jq -r '.genesis.plan' <<<"$two_manifest")
	is_plan "$GENESIS_PLAN" && [[ $GENESIS_PLAN != "$UPGRADE_PLAN" && $(printf '%s\n%s\n' "$GENESIS_PLAN" "$UPGRADE_PLAN" | sort -V | head -n 1) == "$GENESIS_PLAN" ]] ||
		die "two-binary genesis plan must precede $UPGRADE_PLAN"

	jq -e --arg genesis_plan "$GENESIS_PLAN" --arg genesis_version "$GENESIS_VERSION" --arg genesis_sha "$genesis_sha" \
		--arg plan "$UPGRADE_PLAN" --arg version "$UPGRADE_VERSION" --arg sha "$upgrade_sha" --arg patch "$COSMOVISOR_PATCH" '
		.genesis.plan == $genesis_plan and .genesis.version == $genesis_version and .genesis.sha256 == $genesis_sha
		and (.upgrades | length) == 1 and .upgrades[0].plan == $plan and .upgrades[0].version == $version and .upgrades[0].sha256 == $sha
		and (.launcher.commit | test("^[0-9a-f]{40}(-dirty)?$")) and (.cosmovisor.source | contains($patch))
	' <<<"$two_manifest" >/dev/null || die "two-binary metadata does not match supplied artifacts"

	preflight_home=$E2E_DIR/preflight-two
	timeout 120 "$tested_bundle" bundle setup --home "$preflight_home" >>"$node_log" 2>&1 || die "extract two-binary payload"
	cmp -s "$GENESIS_BIN" "$preflight_home/cosmovisor/upgrades/$GENESIS_PLAN/bin/exrpd" || die "two-binary genesis payload differs"
	cmp -s "$UPGRADE_BIN" "$preflight_home/cosmovisor/$UPGRADE_DIR/bin/exrpd" || die "two-binary upgrade payload differs"
	cv_sha=$(jq -r '.cosmovisor.sha256' <<<"$two_manifest")
	[[ $(sha256 "$preflight_home/cosmovisor/supervisor/${cv_sha:0:16}/cosmovisor") == "$cv_sha" ]] || die "two-binary supervisor payload differs"

	pass "two-binary artifact metadata and payload match"
fi

if [[ -n $single_bundle ]]; then
	single_bundle_bin=$E2E_DIR/tested/single-exrpd
	cp "$single_bundle" "$single_bundle_bin" || die "copy $single_bundle"
	single_manifest=$(timeout 30 "$single_bundle_bin" version --bundle --home "$E2E_DIR/no-home" | jq -e '.manifest') || die "read single manifest"
	jq -e --arg plan "$UPGRADE_PLAN" --arg version "$UPGRADE_VERSION" --arg sha "$upgrade_sha" --arg patch "$COSMOVISOR_PATCH" '
		(has("upgrades") | not) and .genesis.plan == $plan and .genesis.version == $version and .genesis.sha256 == $sha
		and (.launcher.commit | test("^[0-9a-f]{40}(-dirty)?$")) and (.cosmovisor.source | contains($patch))
	' <<<"$single_manifest" >/dev/null || die "single metadata does not match supplied artifacts"

	preflight_home=$E2E_DIR/preflight-single
	timeout 120 "$single_bundle_bin" bundle setup --home "$preflight_home" >>"$node_log" 2>&1 || die "extract single payload"
	cmp -s "$UPGRADE_BIN" "$preflight_home/cosmovisor/$UPGRADE_DIR/bin/exrpd" || die "single genesis payload differs"
	cv_sha=$(jq -r '.cosmovisor.sha256' <<<"$single_manifest")
	[[ $(sha256 "$preflight_home/cosmovisor/supervisor/${cv_sha:0:16}/cosmovisor") == "$cv_sha" ]] || die "single supervisor payload differs"

	pass "single artifact metadata and payload match"
fi
