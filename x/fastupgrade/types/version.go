package types

import (
	"strconv"
	"strings"
)

// Version is a strict vMAJOR.MINOR.PATCH upgrade name.
type Version struct {
	Major, Minor, Patch uint64
}

// ParseVersion parses a strict vMAJOR.MINOR.PATCH name: three decimal
// components without signs or leading zeros.
func ParseVersion(name string) (Version, bool) {
	rest, ok := strings.CutPrefix(name, "v")
	if !ok {
		return Version{}, false
	}
	majorStr, rest, ok := strings.Cut(rest, ".")
	if !ok {
		return Version{}, false
	}
	minorStr, patchStr, ok := strings.Cut(rest, ".")
	if !ok {
		return Version{}, false
	}
	major, ok := parseComponent(majorStr)
	if !ok {
		return Version{}, false
	}
	minor, ok := parseComponent(minorStr)
	if !ok {
		return Version{}, false
	}
	patch, ok := parseComponent(patchStr)
	if !ok {
		return Version{}, false
	}
	return Version{Major: major, Minor: minor, Patch: patch}, true
}

func parseComponent(s string) (uint64, bool) {
	if s == "" || (len(s) > 1 && s[0] == '0') {
		return 0, false
	}
	n, err := strconv.ParseUint(s, 10, 64)
	return n, err == nil
}

// IsMinorBumpOf reports whether v has the same major and a higher minor than base.
func (v Version) IsMinorBumpOf(base Version) bool {
	return v.Major == base.Major && v.Minor > base.Minor
}
