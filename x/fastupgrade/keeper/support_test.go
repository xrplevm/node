package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestMeetsThreshold(t *testing.T) {
	dec := math.LegacyMustNewDecFromStr

	// 2/3 as an 18-decimal quotient rounds up to 0.666666666666666667. A
	// ratio comparison would accept that threshold. The first case shows that
	// the exact cross-multiplication does not.
	require.True(t, math.LegacyNewDec(2).Quo(math.LegacyNewDec(3)).GTE(dec("0.666666666666666667")))

	tt := []struct {
		name      string
		yes       string
		threshold string
		bonded    int64
		expected  bool
	}{
		{
			name:      "should fail - 2 of 3 is below the rounded-up quotient",
			yes:       "2",
			threshold: "0.666666666666666667",
			bonded:    3,
		},
		{
			name:      "should pass - 2 of 3 is above the rounded-down quotient",
			yes:       "2",
			threshold: "0.666666666666666666",
			bonded:    3,
			expected:  true,
		},
		{
			name:      "should pass - yes equals the threshold",
			yes:       "85",
			threshold: "0.85",
			bonded:    100,
			expected:  true,
		},
		{
			name:      "should fail - yes is one indivisible unit below the threshold",
			yes:       "84.999999999999999999",
			threshold: "0.85",
			bonded:    100,
		},
		{
			name:      "should fail - weighted yes is one indivisible unit below 1 of 3 tokens",
			yes:       "0.999999999999999999",
			threshold: "0.333333333333333334",
			bonded:    3,
		},
		{
			name:      "should pass - weighted yes meets the rounded-down third",
			yes:       "0.999999999999999999",
			threshold: "0.333333333333333333",
			bonded:    3,
			expected:  true,
		},
		{
			// Small token counts, as on mainnet (1e6 per validator).
			name:      "should pass - small token counts at the threshold",
			yes:       "1500000",
			threshold: "0.75",
			bonded:    2_000_000,
			expected:  true,
		},
		{
			name:      "should fail - small token counts one unit below the threshold",
			yes:       "1500000",
			threshold: "0.750000000000000001",
			bonded:    2_000_000,
		},
		{
			name:      "should fail - empty bonded pool, also with zero yes",
			yes:       "0",
			threshold: "0.85",
			bonded:    0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, meetsThreshold(dec(tc.yes), dec(tc.threshold), math.NewInt(tc.bonded)))
		})
	}
}
