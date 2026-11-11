package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"cosmossdk.io/x/tx/signing/aminojson"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/xrplevm/node/v12/x/fastupgrade/types"
)

func TestDefaultParams(t *testing.T) {
	p := types.DefaultParams()
	require.False(t, p.Enabled)
	require.Equal(t, uint64(150), p.TimelockBlocks)
	require.Equal(t, "0.850000000000000000", p.Threshold)
	require.Equal(t, uint64(10), p.PostponeBlocks)
	require.NoError(t, p.Validate())
	require.NoError(t, types.DefaultGenesis().Validate())
}

func TestParamsValidate(t *testing.T) {
	valid := types.DefaultParams()
	for _, tc := range []struct {
		name   string
		mutate func(*types.Params)
		ok     bool
	}{
		{"timelock min", func(p *types.Params) { p.TimelockBlocks = 2 }, true},
		{"timelock below min", func(p *types.Params) { p.TimelockBlocks = 1 }, false},
		{"timelock max", func(p *types.Params) { p.TimelockBlocks = 100000 }, true},
		{"timelock above max", func(p *types.Params) { p.TimelockBlocks = 100001 }, false},
		{"postpone min", func(p *types.Params) { p.PostponeBlocks = 1 }, true},
		{"postpone below min", func(p *types.Params) { p.PostponeBlocks = 0 }, false},
		{"postpone max", func(p *types.Params) { p.PostponeBlocks = 100000 }, true},
		{"postpone above max", func(p *types.Params) { p.PostponeBlocks = 100001 }, false},
		{"threshold one", func(p *types.Params) { p.Threshold = "1" }, true},
		{"threshold smallest", func(p *types.Params) { p.Threshold = "0.000000000000000001" }, true},
		{"threshold zero", func(p *types.Params) { p.Threshold = "0" }, false},
		{"threshold negative", func(p *types.Params) { p.Threshold = "-0.5" }, false},
		{"threshold above one", func(p *types.Params) { p.Threshold = "1.000000000000000001" }, false},
		{"threshold 19 decimals", func(p *types.Params) { p.Threshold = "0.8500000000000000001" }, false},
		{"threshold empty", func(p *types.Params) { p.Threshold = "" }, false},
		{"threshold garbage", func(p *types.Params) { p.Threshold = "abc" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := valid
			tc.mutate(&p)
			err := p.Validate()
			if tc.ok {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, types.ErrInvalidParams)
			require.ErrorIs(t, (&types.MsgUpdateParams{Params: p}).ValidateBasic(), types.ErrInvalidParams)
		})
	}
}

func liveGovParams() govv1.Params {
	p := govv1.DefaultParams()
	p.Quorum = "0.667000000000000000"
	p.Threshold = "0.667000000000000000"
	p.VetoThreshold = "0.334000000000000000"
	p.ExpeditedThreshold = "0.800000000000000000"
	return p
}

func TestCheckGovCompatibility(t *testing.T) {
	dec := math.LegacyMustNewDecFromStr
	require.NoError(t, types.CheckGovCompatibility(types.DefaultThreshold, liveGovParams()))
	require.NoError(t, types.CheckGovCompatibility(types.DefaultThreshold, govv1.DefaultParams()))

	for _, tc := range []struct {
		name      string
		threshold string
		mutate    func(*govv1.Params)
		ok        bool
	}{
		{"equal to threshold rejected", "0.7", func(p *govv1.Params) { p.Threshold = "0.7"; p.ExpeditedThreshold = "0.69" }, false},
		{"just above threshold", "0.700000000000000001", func(p *govv1.Params) { p.Threshold = "0.7"; p.ExpeditedThreshold = "0.69" }, true},
		{"equal to expedited rejected", "0.8", nil, false},
		{"just above expedited", "0.800000000000000001", nil, true},
		{"equal to quorum accepted", "0.9", func(p *govv1.Params) { p.Quorum = "0.9" }, true},
		{"below quorum rejected", "0.899999999999999999", func(p *govv1.Params) { p.Quorum = "0.9" }, false},
		{"equal to one minus veto accepted", "0.9", func(p *govv1.Params) { p.VetoThreshold = "0.1" }, true},
		{"below one minus veto rejected", "0.899999999999999999", func(p *govv1.Params) { p.VetoThreshold = "0.1" }, false},
		{"unparsable gov param rejected", "0.85", func(p *govv1.Params) { p.Quorum = "x" }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gp := liveGovParams()
			if tc.mutate != nil {
				tc.mutate(&gp)
			}
			err := types.CheckGovCompatibility(dec(tc.threshold), gp)
			if tc.ok {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, types.ErrIncompatibleWithGov)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	for _, tc := range []struct {
		name string
		want types.Version
		ok   bool
	}{
		{"v12.0.0", types.Version{12, 0, 0}, true},
		{"v12.10.3", types.Version{12, 10, 3}, true},
		{"v0.0.0", types.Version{}, true},
		{"v18446744073709551615.0.0", types.Version{Major: 18446744073709551615}, true},
		{"v18446744073709551616.0.0", types.Version{}, false},
		{"12.1.0", types.Version{}, false},
		{"V12.1.0", types.Version{}, false},
		{"v12.1", types.Version{}, false},
		{"v12.1.0.0", types.Version{}, false},
		{"v12.01.0", types.Version{}, false},
		{"v012.1.0", types.Version{}, false},
		{"v12.1.00", types.Version{}, false},
		{"v12.1.0-rc1", types.Version{}, false},
		{"v12.+1.0", types.Version{}, false},
		{"v12..0", types.Version{}, false},
		{"v12.1.0 ", types.Version{}, false},
		{"", types.Version{}, false},
	} {
		got, ok := types.ParseVersion(tc.name)
		require.Equal(t, tc.ok, ok, tc.name)
		require.Equal(t, tc.want, got, tc.name)
	}
}

func TestIsMinorBumpOf(t *testing.T) {
	base := types.Version{12, 3, 4}
	require.True(t, types.Version{12, 4, 0}.IsMinorBumpOf(base))
	require.True(t, types.Version{12, 9, 9}.IsMinorBumpOf(base))
	require.False(t, types.Version{12, 3, 5}.IsMinorBumpOf(base))
	require.False(t, types.Version{12, 2, 9}.IsMinorBumpOf(base))
	require.False(t, types.Version{13, 4, 0}.IsMinorBumpOf(base))
	require.False(t, types.Version{11, 4, 0}.IsMinorBumpOf(base))
}

// Clients such as autocli decode responses with descriptors from
// gogoproto.HybridResolver. Every message field must resolve to a full
// descriptor: a placeholder decodes as an empty message.
func TestDescriptorsResolveForClients(t *testing.T) {
	for _, name := range []protoreflect.FullName{
		"fastupgrade.GenesisState",
		"fastupgrade.QueryParamsResponse",
		"fastupgrade.QueryTimerResponse",
		"fastupgrade.MsgUpdateParams",
	} {
		d, err := gogoproto.HybridResolver.FindDescriptorByName(name)
		require.NoError(t, err, name)
		md, ok := d.(protoreflect.MessageDescriptor)
		require.True(t, ok, name)
		fields := md.Fields()
		for i := range fields.Len() {
			if m := fields.Get(i).Message(); m != nil {
				require.False(t, m.IsPlaceholder(), "%s.%s", name, fields.Get(i).Name())
			}
		}
	}
}

func TestTimerResponseDecodesDynamically(t *testing.T) {
	bz, err := (&types.QueryTimerResponse{Timer: &types.Timer{ProposalId: 3, UpgradeHeight: 252}}).Marshal()
	require.NoError(t, err)

	d, err := gogoproto.HybridResolver.FindDescriptorByName("fastupgrade.QueryTimerResponse")
	require.NoError(t, err)
	md, ok := d.(protoreflect.MessageDescriptor)
	require.True(t, ok)
	msg := dynamicpb.NewMessage(md)
	require.NoError(t, protov2.Unmarshal(bz, msg))

	out, err := aminojson.NewEncoder(aminojson.EncoderOptions{}).Marshal(msg)
	require.NoError(t, err)
	require.JSONEq(t, `{"timer":{"proposal_id":"3","upgrade_height":"252"}}`, string(out))
}
