package v4

import (
	"context"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"errors"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensusparamskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	erc20keeper "github.com/evmos/evmos/v20/x/erc20/keeper"
	evmkeeper "github.com/evmos/evmos/v20/x/evm/keeper"
	"github.com/evmos/evmos/v20/x/evm/types"
)

const (
	XrpAddress      = "0xD4949664cD82660AaE99bEdc034a0deA8A0bd517"
	XrpOwnerAddress = "ethm1zrxl239wa6ad5xge3gs68rt98227xgnjq0xyw2"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v13
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	upgradeKey *storetypes.KVStoreKey,
	consensusParamsKeeper consensusparamskeeper.Keeper,
	authAddr string,
	ek *evmkeeper.Keeper,
	erc20k erc20keeper.Keeper,
	gk govkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Fix previous consensusparams upgrade
		logger.Info("Fixing previous consensusparams upgrade...")
		storesvc := runtime.NewKVStoreService(upgradeKey)
		consensuskeeper := consensusparamskeeper.NewKeeper(
			cdc,
			storesvc,
			authAddr,
			runtime.EventService{},
		)
		consensusParams, err := consensuskeeper.ParamsStore.Get(ctx)
		if err != nil {
			return nil, err
		}
		err = consensusParamsKeeper.ParamsStore.Set(ctx, consensusParams)
		if err != nil {
			return nil, err
		}

		logger.Debug("Enabling gov precompile...")
		if err := EnableGovPrecompile(ctx, ek); err != nil {
			logger.Error("error while enabling gov precompile", "error", err.Error())
		}

		// run the sdk v0.50 migrations
		logger.Debug("Running module migrations...")
		vm, err = mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		logger.Debug("Updating expedited prop params...")
		if err := UpdateExpeditedPropsParams(ctx, gk); err != nil {
			logger.Error("error while updating gov params", "error", err.Error())
		}

		logger.Debug("Assigning XRP owner address...")
		err = AssignXrpOwnerAddress(ctx, erc20k, sdk.MustAccAddressFromBech32(XrpOwnerAddress))
		if err != nil {
			logger.Error("error while assigning XRP owner address", "error", err.Error())
			return vm, err
		}

		logger.Debug("Re-registering ERC-20 precompile code hashes...")
		params := erc20k.GetParams(ctx)
		if err := erc20k.SetParams(ctx, params); err != nil {
			logger.Error("error while re-registering ERC-20 precompile code hashes", "error", err.Error())
			return vm, err
		}
		return vm, nil
	}
}

func EnableGovPrecompile(ctx sdk.Context, ek *evmkeeper.Keeper) error {
	// Enable gov precompile
	params := ek.GetParams(ctx)
	params.ActiveStaticPrecompiles = append(params.ActiveStaticPrecompiles, types.GovPrecompileAddress)
	if err := params.Validate(); err != nil {
		return err
	}
	return ek.SetParams(ctx, params)
}

func UpdateExpeditedPropsParams(ctx sdk.Context, gk govkeeper.Keeper) error {
	params, err := gk.Params.Get(ctx)
	if err != nil {
		return err
	}

	// use the same denom as the min deposit denom
	// also amount must be greater than MinDeposit amount
	denom := params.MinDeposit[0].Denom
	expDepAmt := params.ExpeditedMinDeposit[0].Amount
	if expDepAmt.LTE(params.MinDeposit[0].Amount) {
		expDepAmt = params.MinDeposit[0].Amount.MulRaw(govv1.DefaultMinExpeditedDepositTokensRatio)
	}
	params.ExpeditedMinDeposit = sdk.NewCoins(sdk.NewCoin(denom, expDepAmt))

	// if expedited voting period > voting period
	// set expedited voting period to be half the voting period
	if params.ExpeditedVotingPeriod != nil && params.VotingPeriod != nil && *params.ExpeditedVotingPeriod > *params.VotingPeriod {
		expPeriod := *params.VotingPeriod / 2
		params.ExpeditedVotingPeriod = &expPeriod
	}

	if err := params.ValidateBasic(); err != nil {
		return err
	}
	return gk.Params.Set(ctx, params)
}

func AssignXrpOwnerAddress(ctx sdk.Context, ek erc20keeper.Keeper, address sdk.AccAddress) error {
	tokenPairId := ek.GetTokenPairID(ctx, XrpAddress)
	tokenPair, found := ek.GetTokenPair(ctx, tokenPairId)
	if !found {
		return errors.New("token pair not found")
	}
	ek.SetTokenPairOwnerAddress(ctx, tokenPair, address.String())
	return nil
}
