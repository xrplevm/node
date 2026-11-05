# shellcheck shell=bash
# Shared remote-state checks for bundle planning and publication.

# asset_state TAG SET sets state to build or skip, and refuses partial sets.
asset_state() {
	local tag=$1 set=$2 assets f have=()
	assets=$(gh release view "$tag" --json assets --jq '.assets[].name') || return
	for f in "exrpd-bundle_${set}_"{Linux,Darwin}_{amd64,arm64}.tar.gz "exrpd-bundle_${set}_checksums.txt"; do
		if grep -qxF "$f" <<<"$assets"; then have+=("$f"); fi
	done
	case ${#have[@]} in
	0) state=build ;;
	5) state=skip ;;
	*)
		echo "::error::$tag has ${#have[@]} of the 5 assets of exrpd-bundle_$set: ${have[*]}. Delete them from the release and rerun" >&2
		return 1
		;;
	esac
	echo "::notice::exrpd-bundle_$set assets: $state"
}

# image_state IMAGE sets state to build or skip. An unknown registry state is an error.
image_state() {
	local image=$1 out
	if out=$(docker manifest inspect "$image" 2>&1); then
		state=skip
	elif grep -qiE 'no such manifest|manifest unknown' <<<"$out"; then
		state=build
	else
		echo "::error::cannot tell whether $image exists: $out" >&2
		return 1
	fi
	echo "::notice::$image image: $state"
}
