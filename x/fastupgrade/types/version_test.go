package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

func TestParseVersion(t *testing.T) {
	tt := []struct {
		name            string
		input           string
		expectedVersion types.Version
		expectedOk      bool
	}{
		{
			name:            "should pass - major version",
			input:           "v12.0.0",
			expectedVersion: types.Version{Major: 12},
			expectedOk:      true,
		},
		{
			name:            "should pass - minor with two digits",
			input:           "v12.10.3",
			expectedVersion: types.Version{Major: 12, Minor: 10, Patch: 3},
			expectedOk:      true,
		},
		{
			name:       "should pass - all components zero",
			input:      "v0.0.0",
			expectedOk: true,
		},
		{
			name:            "should pass - largest uint64 component",
			input:           "v18446744073709551615.0.0",
			expectedVersion: types.Version{Major: 18446744073709551615},
			expectedOk:      true,
		},
		{
			name:  "should fail - component overflows uint64",
			input: "v18446744073709551616.0.0",
		},
		{
			name:  "should fail - no v prefix",
			input: "12.1.0",
		},
		{
			name:  "should fail - uppercase prefix",
			input: "V12.1.0",
		},
		{
			name:  "should fail - two components",
			input: "v12.1",
		},
		{
			name:  "should fail - four components",
			input: "v12.1.0.0",
		},
		{
			name:  "should fail - leading zero in minor",
			input: "v12.01.0",
		},
		{
			name:  "should fail - leading zero in major",
			input: "v012.1.0",
		},
		{
			name:  "should fail - leading zero in patch",
			input: "v12.1.00",
		},
		{
			name:  "should fail - suffix",
			input: "v12.1.0-rc1",
		},
		{
			name:  "should fail - signed component",
			input: "v12.+1.0",
		},
		{
			name:  "should fail - empty component",
			input: "v12..0",
		},
		{
			name:  "should fail - trailing space",
			input: "v12.1.0 ",
		},
		{
			name:  "should fail - empty name",
			input: "",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			version, ok := types.ParseVersion(tc.input)

			require.Equal(t, tc.expectedOk, ok)
			require.Equal(t, tc.expectedVersion, version)
		})
	}
}

func TestVersion_IsMinorBumpOf(t *testing.T) {
	base := types.Version{Major: 12, Minor: 3, Patch: 4}

	tt := []struct {
		name     string
		target   types.Version
		expected bool
	}{
		{
			name:     "should pass - next minor",
			target:   types.Version{Major: 12, Minor: 4, Patch: 0},
			expected: true,
		},
		{
			name:     "should pass - later minor",
			target:   types.Version{Major: 12, Minor: 9, Patch: 9},
			expected: true,
		},
		{
			name:   "should fail - patch only",
			target: types.Version{Major: 12, Minor: 3, Patch: 5},
		},
		{
			name:   "should fail - lower minor",
			target: types.Version{Major: 12, Minor: 2, Patch: 9},
		},
		{
			name:   "should fail - higher major",
			target: types.Version{Major: 13, Minor: 4, Patch: 0},
		},
		{
			name:   "should fail - lower major",
			target: types.Version{Major: 11, Minor: 4, Patch: 0},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.target.IsMinorBumpOf(base))
		})
	}
}
