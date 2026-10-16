package integration

import (
	"math/big"
	"sort"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/evm/precompiles/staking"
	testutiltypes "github.com/cosmos/evm/testutil/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v12/testutil/integration/exrp/utils"
)

// The staking precompile is not registered in v12, because it calls the staking
// msg server directly and skips the POA ante handler. Even if governance
// activates its address, every staking transaction through it must fail
// without changing the validator set or delegations.
func (s *TestSuite) TestStakingPrecompile_NotRegistered() {
	authority := sdktypes.AccAddress(address.Module("gov")).String()
	sender := s.keyring.GetKey(1)
	validators := s.Network().GetValidators()
	require.GreaterOrEqual(s.T(), len(validators), 2)
	srcVal, dstVal := validators[0].OperatorAddress, validators[1].OperatorAddress

	// The mistake this protects against: governance activates 0x800 in the EVM
	// params. The params must stay sorted to pass validation.
	params := s.Network().EvmKeeper().GetParams(s.Network().GetContext())
	params.ActiveStaticPrecompiles = append(params.ActiveStaticPrecompiles, evmtypes.StakingPrecompileAddress)
	sort.Strings(params.ActiveStaticPrecompiles)
	proposal, err := utils.SubmitAndAwaitProposalResolution(s.factory, s.Network(), s.keyring.GetKeys(), "activate staking precompile",
		&evmtypes.MsgUpdateParams{Authority: authority, Params: params})
	require.NoError(s.T(), err)
	require.Equal(s.T(), govv1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposal.Status)

	// stakingSnapshot reads the state that the staking methods below could change:
	// every validator, and the sender's delegations.
	stakingSnapshot := func() ([]stakingtypes.Validator, stakingtypes.DelegationResponses) {
		validatorsRes, err := s.Network().GetStakingClient().Validators(
			s.Network().GetContext(),
			&stakingtypes.QueryValidatorsRequest{},
		)
		require.NoError(s.T(), err)
		delegationsRes, err := s.Network().GetStakingClient().DelegatorDelegations(
			s.Network().GetContext(),
			&stakingtypes.QueryDelegatorDelegationsRequest{DelegatorAddr: sender.AccAddr.String()},
		)
		require.NoError(s.T(), err)
		return validatorsRes.Validators, delegationsRes.DelegationResponses
	}
	validatorsBefore, delegationsBefore := stakingSnapshot()

	description := staking.Description{Moniker: "test"}
	amount := big.NewInt(1)

	// Every staking transaction method. The arguments only need to ABI-encode,
	// because the call never reaches staking code.
	tt := []struct {
		name   string
		method string
		args   []interface{}
	}{
		{
			name:   "staking precompile - createValidator",
			method: staking.CreateValidatorMethod,
			args: []interface{}{
				description,
				staking.Commission{Rate: big.NewInt(0), MaxRate: big.NewInt(0), MaxChangeRate: big.NewInt(0)},
				amount,
				sender.Addr,
				"nyfUvTSZk4o5QHsTClk0gBsTvn3Cnt2Z7raJZORs7p4=",
				amount,
			},
		},
		{
			name:   "staking precompile - editValidator",
			method: staking.EditValidatorMethod,
			args:   []interface{}{description, sender.Addr, big.NewInt(-1), big.NewInt(-1)},
		},
		{
			name:   "staking precompile - delegate",
			method: staking.DelegateMethod,
			args:   []interface{}{sender.Addr, srcVal, amount},
		},
		{
			name:   "staking precompile - undelegate",
			method: staking.UndelegateMethod,
			args:   []interface{}{sender.Addr, srcVal, amount},
		},
		{
			name:   "staking precompile - redelegate",
			method: staking.RedelegateMethod,
			args:   []interface{}{sender.Addr, srcVal, dstVal, amount},
		},
		{
			name:   "staking precompile - cancelUnbondingDelegation",
			method: staking.CancelUnbondingDelegationMethod,
			args:   []interface{}{sender.Addr, srcVal, amount, big.NewInt(1)},
		},
	}

	stakingPrecompile := common.HexToAddress(evmtypes.StakingPrecompileAddress)
	for _, tc := range tt {
		s.Run(tc.name, func() {
			// Call the staking method through the EVM, as a contract call to 0x800.
			tx := evmtypes.EvmTxArgs{To: &stakingPrecompile, GasLimit: 500_000}
			call := testutiltypes.CallArgs{ContractABI: staking.ABI, MethodName: tc.method, Args: tc.args}
			res, err := s.factory.ExecuteContractCall(sender.Priv, tx, call)

			// 0x800 is active but has no code, so the EVM panics before any staking
			// code runs. The node recovers the panic and marks only this tx as failed.
			require.ErrorContains(s.T(), err, "precompiled contract not stored in memory")
			require.Equal(s.T(), sdkerrors.ErrPanic.ABCICode(), res.Code)

			// The chain keeps producing blocks.
			require.NoError(s.T(), s.Network().NextBlock())

			// No validator or delegation changed.
			validatorsAfter, delegationsAfter := stakingSnapshot()
			require.Equal(s.T(), validatorsBefore, validatorsAfter)
			require.Equal(s.T(), delegationsBefore, delegationsAfter)
		})
	}
}
