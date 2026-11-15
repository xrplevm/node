# shellcheck shell=bash
# Shared bundle helpers. Callers set their own shell options.

# The launcher accepts only plans named vX.Y.Z without leading zeros.
PLAN_RE='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'
is_plan() { [[ $1 =~ $PLAN_RE ]]; }

sha256() { if command -v sha256sum >/dev/null; then sha256sum "$1"; else shasum -a 256 "$1"; fi | cut -d' ' -f1; }

# os_title GOOS: the Os part of release archive names.
os_title() { if [[ $1 == darwin ]]; then echo Darwin; else echo Linux; fi; }

check_platform() {
	case $1/$2 in
	linux/amd64 | linux/arm64 | darwin/amd64 | darwin/arm64) ;;
	*) echo "unsupported platform $1/$2" >&2; return 2 ;;
	esac
}

# verify_archive CHECKSUMS FILE: check FILE against its line in CHECKSUMS and print its sha256.
verify_archive() {
	local name want got
	name=$(basename "$2")
	want=$(awk -v f="$name" '$2 == f { print $1 }' "$1") || return
	got=$(sha256 "$2") || return
	[[ -n $want && $got == "$want" ]] || { echo "$name sha256 $got does not match $(basename "$1") ($want)" >&2; return 1; }
	echo "$got"
}

# release_plan PLAN validates and prints a plan.
release_plan() {
	local plan=$1

	if ! is_plan "$plan"; then
		echo "plan $plan is not named vX.Y.Z, so this bundle cannot carry it" >&2
		return 1
	fi

	echo "$plan"
}

genesis_plan_for() {
	local version=${1%%-*}
	release_plan "${2:-v${version%.*}.0}"
}

# resolve_plans WHAT VERSION PLAN GENESIS_VERSION GENESIS_PLAN: print "PLAN GENESIS_PLAN" of a
# two-binary bundle, release VERSION staged for PLAN. Fails unless VERSION has the major.minor
# of PLAN (a new major or minor release needs its own plan) and the genesis plan comes before
# PLAN. WHAT names VERSION in the message. Call it as plans=$(resolve_plans ...).
resolve_plans() {
	local plan genesis_plan plan_version base_version=${2%%-*}
	plan=$(release_plan "$3") || return
	plan_version=${plan#v}

	if [[ ! $2 =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ || ${base_version%.*} != "${plan_version%.*}" ]]; then
		echo "$1 $2 does not have the major.minor of plan $plan in app/upgrades: a new major or minor release needs its upgrade handler (see scripts/bundle/README.md, Release)" >&2
		return 1
	fi

	genesis_plan=$(genesis_plan_for "$4" "$5") || return

	if [[ $genesis_plan == "$plan" || $(printf '%s\n%s\n' "$genesis_plan" "$plan" | sort -V | head -n 1) != "$genesis_plan" ]]; then
		echo "genesis $4 runs plan $genesis_plan, which does not come before plan $plan of this bundle: choose a genesis release from before plan $plan" >&2
		return 1
	fi

	echo "$plan $genesis_plan"
}
