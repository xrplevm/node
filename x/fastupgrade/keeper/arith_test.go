package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestMeetsThresholdIsExact(t *testing.T) {
	dec := math.LegacyMustNewDecFromStr
	three := math.NewInt(3)

	// 2/3 as an 18-decimal quotient rounds up to 0.666666666666666667. A
	// ratio comparison would accept it; the exact cross-multiplication does not.
	require.True(t, math.LegacyNewDec(2).Quo(math.LegacyNewDec(3)).GTE(dec("0.666666666666666667")))
	require.False(t, meetsThreshold(math.LegacyNewDec(2), dec("0.666666666666666667"), three))
	require.True(t, meetsThreshold(math.LegacyNewDec(2), dec("0.666666666666666666"), three))

	// Equality passes, one indivisible unit below fails.
	require.True(t, meetsThreshold(math.LegacyNewDec(85), dec("0.85"), math.NewInt(100)))
	require.False(t, meetsThreshold(dec("84.999999999999999999"), dec("0.85"), math.NewInt(100)))

	// A weighted Yes sum one indivisible unit below 1 of 3 tokens.
	require.False(t, meetsThreshold(dec("0.999999999999999999"), dec("0.333333333333333334"), three))
	require.True(t, meetsThreshold(dec("0.999999999999999999"), dec("0.333333333333333333"), three))

	// Small token counts, as on mainnet (1e6 per validator).
	require.True(t, meetsThreshold(dec("1500000"), dec("0.75"), math.NewInt(2_000_000)))
	require.False(t, meetsThreshold(dec("1500000"), dec("0.750000000000000001"), math.NewInt(2_000_000)))

	// An empty bonded pool never meets the threshold, not even with zero Yes.
	require.False(t, meetsThreshold(math.LegacyZeroDec(), dec("0.85"), math.ZeroInt()))
}
