# shellcheck shell=bash
# Helpers shared by the bundle scripts, sourced by them. release_plan and build_one run
# the scripts next to this file. It also sets the Cosmovisor pins of
# cosmovisor/cosmovisor.env.
# The scripts set their own shell options.

lib_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
# shellcheck source=SCRIPTDIR/cosmovisor/cosmovisor.env
source "$lib_dir/cosmovisor/cosmovisor.env"

# The launcher accepts only plans named vX.Y.Z without leading zeros.
PLAN_RE='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'

# is_plan NAME: NAME is a plan a bundle can carry.
is_plan() { [[ $1 =~ $PLAN_RE ]]; }

# upgrade_names [FILE...]: print the UpgradeName values of constants.go files, or of
# constants.go text on stdin.
upgrade_names() { sed -n 's/^[[:space:]]*UpgradeName[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' "$@"; }

# require_plan_version WHAT VERSION PLAN: fail unless release VERSION has the major.minor
# of PLAN. The staged upgrade binary lives in upgrades/<plan>. Major and minor releases
# need their own on-chain plan, so only releases with the plan's major.minor fit there:
# the plan's release, its patch releases and their pre-releases. WHAT names VERSION in
# the message.
require_plan_version() {
	local plan_version=${3#v} base_version=${2%%-*}
	if [[ ! $2 =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ || ${base_version%.*} != "${plan_version%.*}" ]]; then
		echo "$1 $2 does not have the major.minor of plan $3 in app/upgrades: a new major or minor release needs its upgrade handler (see docs/bundle-dev.md, Releases and bundles)" >&2
		return 1
	fi
}

sha256() { if command -v sha256sum >/dev/null; then sha256sum "$1"; else shasum -a 256 "$1"; fi | cut -d' ' -f1; }

# os_title GOOS: the Os of release archive names.
os_title() { if [[ $1 == darwin ]]; then echo Darwin; else echo Linux; fi; }

# check_platform GOOS GOARCH: fail unless a bundle can be built for GOOS/GOARCH.
check_platform() {
	case $1/$2 in
	linux/amd64 | linux/arm64 | darwin/amd64 | darwin/arm64) ;;
	*) echo "unsupported platform $1/$2" >&2; return 2 ;;
	esac
}

# verify_archive CHECKSUMS FILE: check FILE against its entry in the release CHECKSUMS
# file and print its sha256.
verify_archive() {
	local name want got
	name=$(basename "$2")
	want=$(awk -v f="$name" '$2 == f { print $1 }' "$1") || return
	got=$(sha256 "$2") || return
	[[ -n $want && $got == "$want" ]] || { echo "$name sha256 $got does not match $(basename "$1") ($want)" >&2; return 1; }
	echo "$got"
}

# release_plan VERSION [PLAN]: print PLAN, by default the plan of release tag vVERSION
# (release-plan.sh), and fail unless it is vX.Y.Z. Call it as plan=$(release_plan ...),
# so that set -e stops the caller when it fails.
release_plan() {
	local plan=${2:-}
	if [[ -z $plan ]]; then
		plan=$("$lib_dir/release-plan.sh" "$1") || return
	fi
	# Any plan name other than vX.Y.Z would make every command of the bundle fail.
	if ! is_plan "$plan"; then
		echo "plan $plan is not named vX.Y.Z, so this bundle cannot carry it" >&2
		return 1
	fi
	echo "$plan"
}

# resolve_plans WHAT VERSION PLAN GENESIS_VERSION [GENESIS_PLAN]: print "PLAN GENESIS_PLAN"
# for a two-binary bundle whose upgrade is release VERSION, staged for PLAN. GENESIS_PLAN
# is by default the plan of the genesis release tag (release_plan). Fail unless release
# VERSION fits PLAN (require_plan_version, WHAT names VERSION), both plans are vX.Y.Z and
# the genesis plan comes before PLAN. Call it as plans=$(resolve_plans ...), so that
# set -e stops the caller when it fails.
resolve_plans() {
	local plan genesis_plan
	plan=$(release_plan "$2" "$3") || return
	require_plan_version "$1" "$2" "$plan" || return
	genesis_plan=$(release_plan "$4" "${5:-}") || return
	# The launcher requires plans in order: the genesis runs the data before the upgrade plan.
	if [[ $genesis_plan == "$plan" || $(printf '%s\n%s\n' "$genesis_plan" "$plan" | sort -V | head -n 1) != "$genesis_plan" ]]; then
		echo "genesis $4 runs plan $genesis_plan, which does not come before plan $plan of this bundle: choose a genesis release from before plan $plan" >&2
		return 1
	fi
	echo "$plan $genesis_plan"
}

# launcher_commit: the commit of this checkout, which builds the launcher, with -dirty
# when cmd/exrpd-bundle has uncommitted changes.
launcher_commit() {
	local repo=$lib_dir/../.. commit changes
	commit=$(git -C "$repo" rev-parse HEAD) || return
	changes=$(git -C "$repo" status --porcelain -- cmd/exrpd-bundle) || return
	echo "$commit${changes:+-dirty}"
}

# build_one OUT GOOS GOARCH GENESIS_VERSION GENESIS_PLAN GENESIS_ARCHIVE GENESIS_CHECKSUMS
#           PLAN UPGRADE_EXRPD UPGRADE_VERSION [COSMOVISOR COSMOVISOR_SOURCE]:
# build the bundle executable OUT for GOOS/GOARCH. The genesis exrpd of release
# GENESIS_VERSION, which runs GENESIS_PLAN, comes from GENESIS_ARCHIVE, checked against
# GENESIS_CHECKSUMS. UPGRADE_EXRPD is release UPGRADE_VERSION, staged for PLAN. With PLAN,
# UPGRADE_EXRPD and UPGRADE_VERSION empty, it builds a single bundle: no upgrade in the
# payload or the manifest. COSMOVISOR is a prebuilt Cosmovisor that COSMOVISOR_SOURCE
# describes; without it, Cosmovisor is built with cosmovisor.sh. The arguments are the
# values of the build.sh flags and the plans of release_plan or resolve_plans.
# It runs in a subshell, so its scratch directory, EXIT trap and variables stay its own.
# Call it as a plain command: it relies on set -e of the calling script.
build_one() (
	out=$1 goos=$2 goarch=$3 genesis_version=$4 genesis_plan=$5 genesis_archive=$6
	genesis_checksums=$7 plan=$8 upgrade_exrpd=$9 upgrade_version=${10}
	cosmovisor=${11:-} cosmovisor_source=${12:-}
	size() { wc -c <"$1" | tr -d ' '; }
	files=(exrpd-genesis cosmovisor)
	if [[ -n $upgrade_exrpd ]]; then files+=(exrpd-upgrade); fi
	commit=$(launcher_commit)

	archive_name="node_${genesis_version}_$(os_title "$goos")_${goarch}.tar.gz"
	[[ $(basename "$genesis_archive") == "$archive_name" ]] || { echo "expected genesis archive $archive_name" >&2; exit 1; }
	got=$(verify_archive "$genesis_checksums" "$genesis_archive")

	work=$(mktemp -d)
	trap 'rm -rf "$work"' EXIT
	cp -R "$lib_dir/../../cmd/exrpd-bundle" "$work/mod"
	payload=$work/mod/payload raw=$work/raw
	rm -rf "$payload" && mkdir -p "$payload" "$raw"

	tar -xzf "$genesis_archive" -C "$work" bin/exrpd
	mv "$work/bin/exrpd" "$raw/exrpd-genesis"
	if [[ -n $upgrade_exrpd ]]; then cp "$upgrade_exrpd" "$raw/exrpd-upgrade"; fi
	if [[ -n $cosmovisor ]]; then
		cp "$cosmovisor" "$raw/cosmovisor"
	else
		"$lib_dir/cosmovisor.sh" "$goos" "$goarch" "$raw/cosmovisor"
		cosmovisor_source="$COSMOVISOR_MODULE@$COSMOVISOR_VERSION $COSMOVISOR_SUM + $COSMOVISOR_PATCH sha256:$(sha256 "$lib_dir/cosmovisor/$COSMOVISOR_PATCH")"
	fi

	# Payload files are gzip-compressed. sha256 and size are those of the executable itself.
	for f in "${files[@]}"; do
		gzip -n -9 -c "$raw/$f" >"$payload/$f.gz"
	done
	entry() { # plan version name source
		jq -cn --arg plan "$1" --arg version "$2" --arg file "$3.gz" --arg sha256 "$(sha256 "$raw/$3")" \
			--argjson size "$(size "$raw/$3")" --arg source "$4" \
			'{version: $version, file: $file, sha256: $sha256, size: $size, source: $source} + if $plan == "" then {} else {plan: $plan} end'
	}
	{
		printf '{"platform":"%s",' "$goos/$goarch"
		printf '"launcher":%s,' "$(jq -cn --arg commit "$commit" '{commit: $commit}')"
		printf '"genesis":%s,' "$(entry "$genesis_plan" "$genesis_version" exrpd-genesis "github.com/xrplevm/node release v$genesis_version $archive_name sha256:$got")"
		if [[ -n $upgrade_exrpd ]]; then
			printf '"upgrades":[%s],' "$(entry "$plan" "$upgrade_version" exrpd-upgrade "exrpd $upgrade_version standalone release binary")"
		fi
		printf '"cosmovisor":%s}\n' "$(entry "" "$COSMOVISOR_VERSION" cosmovisor "$cosmovisor_source")"
	} >"$payload/manifest.json"
	jq -e . "$payload/manifest.json" >/dev/null

	mkdir -p "$(dirname "$out")"
	out=$(cd "$(dirname "$out")" && pwd)/$(basename "$out")
	(cd "$work/mod" && GOWORK=off GOFLAGS=-mod=readonly CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch \
		go build -trimpath -buildvcs=false -ldflags "-s -w" -o "$out" .)
	echo "built $out ($(size "$out") bytes) for $goos/$goarch"
)
