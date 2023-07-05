package app_test

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"math/rand"
	"os"
	"time"
)

func AppStateFn(cdc codec.JSONCodec, simManager *module.SimulationManager) simtypes.AppStateFn {
	return func(r *rand.Rand, accs []simtypes.Account, config simtypes.Config,
	) (appState json.RawMessage, simAccs []simtypes.Account, chainID string, genesisTimestamp time.Time) {
		if simapp.FlagGenesisTimeValue == 0 {
			genesisTimestamp = simtypes.RandTimestamp(r)
		} else {
			genesisTimestamp = time.Unix(simapp.FlagGenesisTimeValue, 0)
		}

		chainID = config.ChainID
		switch {
		case config.ParamsFile != "" && config.GenesisFile != "":
			panic("cannot provide both a genesis file and a params file")

		case config.GenesisFile != "":
			// override the default chain-id from simapp to set it later to the config
			genesisDoc, accounts := simapp.AppStateFromGenesisFileFn(r, cdc, config.GenesisFile)

			if simapp.FlagGenesisTimeValue == 0 {
				// use genesis timestamp if no custom timestamp is provided (i.e no random timestamp)
				genesisTimestamp = genesisDoc.GenesisTime
			}

			appState = genesisDoc.AppState
			chainID = genesisDoc.ChainID
			simAccs = accounts

		case config.ParamsFile != "":
			appParams := make(simtypes.AppParams)
			bz, err := os.ReadFile(config.ParamsFile)
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal(bz, &appParams)
			if err != nil {
				panic(err)
			}
			appState, simAccs = AppStateRandomizedFn(simManager, r, cdc, accs, genesisTimestamp, appParams)

		default:
			appParams := make(simtypes.AppParams)
			appState, simAccs = AppStateRandomizedFn(simManager, r, cdc, accs, genesisTimestamp, appParams)
		}

		appState = OverrideBankGenState(r, cdc, appState, simAccs)

		rawState := make(map[string]json.RawMessage)
		err := json.Unmarshal(appState, &rawState)
		if err != nil {
			panic(err)
		}

		stakingStateBz, ok := rawState[stakingtypes.ModuleName]
		if !ok {
			panic("staking genesis state is missing")
		}

		stakingState := new(stakingtypes.GenesisState)
		err = cdc.UnmarshalJSON(stakingStateBz, stakingState)
		if err != nil {
			panic(err)
		}
		// compute not bonded balance
		notBondedTokens := sdk.ZeroInt()
		for _, val := range stakingState.Validators {
			if val.Status != stakingtypes.Unbonded {
				continue
			}
			notBondedTokens = notBondedTokens.Add(val.GetTokens())
		}
		notBondedCoins := sdk.NewCoin(stakingState.Params.BondDenom, notBondedTokens)
		// edit bank state to make it have the not bonded pool tokens
		bankStateBz, ok := rawState[banktypes.ModuleName]
		// TODO(fdymylja/jonathan): should we panic in this case
		if !ok {
			panic("bank genesis state is missing")
		}
		bankState := new(banktypes.GenesisState)
		err = cdc.UnmarshalJSON(bankStateBz, bankState)
		if err != nil {
			panic(err)
		}

		stakingAddr := authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName).String()
		var found bool
		for _, balance := range bankState.Balances {
			if balance.Address == stakingAddr {
				found = true
				break
			}
		}
		if !found {
			bankState.Balances = append(bankState.Balances, banktypes.Balance{
				Address: stakingAddr,
				Coins:   sdk.NewCoins(notBondedCoins),
			})
		}

		// change appState back
		rawState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingState)
		rawState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankState)

		// replace appstate
		appState, err = json.Marshal(rawState)
		if err != nil {
			panic(err)
		}

		return appState, simAccs, chainID, genesisTimestamp
	}
}

// AppStateRandomizedFn creates calls each module's GenesisState generator function
// and creates the simulation params
func AppStateRandomizedFn(
	simManager *module.SimulationManager, r *rand.Rand, cdc codec.JSONCodec,
	accs []simtypes.Account, genesisTimestamp time.Time, appParams simtypes.AppParams,
) (json.RawMessage, []simtypes.Account) {
	numAccs := int64(len(accs))
	genesisState := simapp.NewDefaultGenesisState(cdc)

	// generate a random amount of initial stake coins and a random initial
	// number of bonded accounts
	var (
		numInitiallyBonded int64
		initialStake       sdk.Int
	)
	appParams.GetOrGenerate(
		cdc, simappparams.StakePerAccount, &initialStake, r,
		func(r *rand.Rand) { initialStake = sdk.DefaultPowerReduction /*sdk.NewInt(1 r.Int63n(1e12))*/ },
	)
	appParams.GetOrGenerate(
		cdc, simappparams.InitiallyBondedValidators, &numInitiallyBonded, r,
		func(r *rand.Rand) { numInitiallyBonded = numAccs /*int64(r.Intn(300))*/ },
	)

	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}

	fmt.Printf(
		`Selected randomly generated parameters for simulated genesis:
{
  stake_per_account: "%d",
  initially_bonded_validators: "%d"
}
`, initialStake, numInitiallyBonded,
	)

	simState := &module.SimulationState{
		AppParams:    appParams,
		Cdc:          cdc,
		Rand:         r,
		GenState:     genesisState,
		Accounts:     accs,
		InitialStake: initialStake,
		NumBonded:    numInitiallyBonded,
		GenTimestamp: genesisTimestamp,
	}

	simManager.GenerateGenesisStates(simState)

	appState, err := json.Marshal(genesisState)
	if err != nil {
		panic(err)
	}

	return appState, accs
}

func OverrideBankGenState(r *rand.Rand, cdc codec.JSONCodec, appState json.RawMessage, accs []simtypes.Account) json.RawMessage {
	rawState := make(map[string]json.RawMessage)
	err := json.Unmarshal(appState, &rawState)
	if err != nil {
		panic(err)
	}

	bankStateBz, ok := rawState[banktypes.ModuleName]
	if !ok {
		panic("bank genesis state is missing")
	}
	bankState := new(banktypes.GenesisState)
	cdc.MustUnmarshalJSON(bankStateBz, bankState)

	stakingStateBz, ok := rawState[stakingtypes.ModuleName]
	if !ok {
		panic("staking genesis state is missing")
	}
	stakingState := new(stakingtypes.GenesisState)
	cdc.MustUnmarshalJSON(stakingStateBz, stakingState)

	bankState.Params.SetSendEnabledParam(evmtypes.DefaultEVMDenom, false)
	bankState.Params.SetSendEnabledParam(sdk.DefaultBondDenom, false)

	genesisBalances := []banktypes.Balance{}
	balancePerAccount := sdk.NewInt(r.Int63n(1e12))
	totalEvmSupply := balancePerAccount.Mul(sdk.NewInt(int64(len(accs))))
	totalBondSupply := sdk.DefaultPowerReduction.Mul(sdk.NewInt(int64(len(stakingState.Validators))))
	for _, acc := range accs {
		isValidator := false
		for _, val := range stakingState.Validators {
			accValAddr := sdk.ValAddress(acc.Address)
			if accValAddr.String() == val.OperatorAddress {
				isValidator = true
				break
			}
		}
		evmCoin := sdk.NewCoin(evmtypes.DefaultEVMDenom, balancePerAccount)
		if !isValidator {
			genesisBalances = append(genesisBalances, banktypes.Balance{
				Address: acc.Address.String(),
				Coins:   sdk.NewCoins(evmCoin),
			})
		} else {
			genesisBalances = append(genesisBalances, banktypes.Balance{
				Address: acc.Address.String(),
				Coins: sdk.NewCoins(
					evmCoin,
					// sdk.NewCoin(sdk.DefaultBondDenom, sdk.DefaultPowerReduction),
				),
			})
		}
	}

	bankState.Params = banktypes.Params{
		SendEnabled: banktypes.SendEnabledParams{
			{
				Denom:   sdk.DefaultBondDenom,
				Enabled: false,
			},
			{
				Denom:   evmtypes.DefaultEVMDenom,
				Enabled: false,
			},
		},
		DefaultSendEnabled: false,
	}
	bankState.Balances = genesisBalances
	bankState.Supply = sdk.NewCoins(
		sdk.NewCoin(evmtypes.DefaultEVMDenom, totalEvmSupply),
		sdk.NewCoin(sdk.DefaultBondDenom, totalBondSupply),
	)
	bankState.Supply = bankState.Supply.Add()

	rawState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankState)

	// replace appstate
	appState, err = json.Marshal(rawState)
	if err != nil {
		panic(err)
	}

	return appState
}
