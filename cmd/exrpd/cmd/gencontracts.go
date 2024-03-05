package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethermint "github.com/evmos/evmos/v15/types"
	evmtypes "github.com/evmos/evmos/v15/x/evm/types"
	"github.com/spf13/cobra"
)

func AddGenesisContractsCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-contracts [witness][,[witness]] [threshold]",
		Short: "Adds genesis contracts to genesis.json",
		Long: `Adds genesis contracts to genesis.json. First it adds the safe contracts and afterwards
			the bridge contract paired with its proxy.`,

		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			witnesses, witnessesEvm, err := ParseWitnesses(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse witnesses: %w", err)
			}

			threshold, err := ParseThreshold(args[1], len(witnesses))
			if err != nil {
				return fmt.Errorf("failed to parse threshold: %w", err)
			}

			balances := []banktypes.Balance{}
			addresses := []sdk.AccAddress{}
			genAccounts := []authtypes.GenesisAccount{}
			evmGenAccounts := []evmtypes.GenesisAccount{}

			contracts, err := getGenesisContracts(witnessesEvm, threshold)
			if err != nil {
				return fmt.Errorf("failed to get genesis contracts: %w", err)
			}

			for _, contract := range contracts {
				addr, err := sdk.AccAddressFromHexUnsafe(contract.address)
				if err != nil {
					return fmt.Errorf("failed to parse address: %w", err)
				}

				data, err := hex.DecodeString(contract.bytecode)
				if err != nil {
					return fmt.Errorf("failed to decode contract bytecode: %w", err)
				}

				baseAccount := authtypes.NewBaseAccount(addr, nil, 0, 0)
				genAccount := &ethermint.EthAccount{
					BaseAccount: baseAccount,
					CodeHash:    common.BytesToHash(crypto.Keccak256(data)).Hex(),
				}

				if err := genAccount.Validate(); err != nil {
					return fmt.Errorf("failed to validate new genesis account: %w", err)
				}

				if contract.name == safeContractName {
					coins, err := sdk.ParseCoinsNormalized(safeInitCoins)
					if err != nil {
						return fmt.Errorf("failed to parse coins: %w", err)
					}
					balances = append(balances, banktypes.Balance{Address: addr.String(), Coins: coins.Sort()})
				}

				addresses = append(addresses, addr)
				genAccounts = append(genAccounts, genAccount)
				evmGenAccounts = append(evmGenAccounts, evmtypes.GenesisAccount{
					Address: "0x" + contract.address,
					Code:    contract.bytecode,
					Storage: contract.memory,
				})
			}

			for _, witness := range witnesses {
				baseAccount := authtypes.NewBaseAccount(witness, nil, 0, 0)
				genAccount := &ethermint.EthAccount{
					BaseAccount: baseAccount,
					CodeHash:    common.BytesToHash(evmtypes.EmptyCodeHash).Hex(),
				}
				genAccounts = append(genAccounts, genAccount)

				coins, err := sdk.ParseCoinsNormalized(witnessInitCoins)
				if err != nil {
					return fmt.Errorf("failed to parse coins: %w", err)
				}
				balances = append(balances, banktypes.Balance{Address: witness.String(), Coins: coins.Sort()})
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// ----------------------------------------------------------------
			// Auth genesis state update
			// ----------------------------------------------------------------
			authGenState := authtypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)

			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
			if err != nil {
				return fmt.Errorf("failed to get accounts from any: %w", err)
			}

			for _, addr := range addresses {
				if accs.Contains(addr) {
					return fmt.Errorf("cannot add account at existing address %s", addr)
				}
			}

			// Add the new accounts to the set of genesis accounts and sanitize the
			// accounts afterwards.
			accs = append(accs, genAccounts...)
			accs = authtypes.SanitizeGenesisAccounts(accs)

			genAccs, err := authtypes.PackAccounts(accs)
			if err != nil {
				return fmt.Errorf("failed to convert accounts into any's: %w", err)
			}
			authGenState.Accounts = genAccs

			authGenStateBz, err := clientCtx.Codec.MarshalJSON(&authGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}

			appState[authtypes.ModuleName] = authGenStateBz

			// ----------------------------------------------------------------
			// EVM genesis state update
			// ----------------------------------------------------------------
			evmGenState := &evmtypes.GenesisState{}
			if appState[evmtypes.ModuleName] != nil {
				err = clientCtx.Codec.UnmarshalJSON(appState[evmtypes.ModuleName], evmGenState)
				if err != nil {
					return fmt.Errorf("failed to unmarshal evm genesis state: %w", err)
				}
			} else {
				evmGenState = evmtypes.NewGenesisState(evmtypes.DefaultParams(), []evmtypes.GenesisAccount{})
			}

			// Add evm accounts to genesis state
			evmGenState.Accounts = append(evmGenState.Accounts, evmGenAccounts...)

			evmGenStateBz, err := clientCtx.Codec.MarshalJSON(evmGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal evm genesis state: %w", err)
			}
			appState[evmtypes.ModuleName] = evmGenStateBz

			// ----------------------------------------------------------------
			// Bank genesis state update
			// ----------------------------------------------------------------
			bankGenState := banktypes.GetGenesisStateFromAppState(clientCtx.Codec, appState)
			bankGenState.Balances = append(bankGenState.Balances, balances...)
			bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)
			for _, balance := range balances {
				bankGenState.Supply = bankGenState.Supply.Add(balance.Coins...)
			}

			bankGenStateBz, err := clientCtx.Codec.MarshalJSON(bankGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal bank genesis state: %w", err)
			}
			appState[banktypes.ModuleName] = bankGenStateBz

			// ----------------------------------------------------------------
			// Update and export app genesis state
			// ----------------------------------------------------------------
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func ParseWitnesses(witnessesStr string) ([]sdk.AccAddress, []string, error) {
	witnessesStr = strings.TrimSpace(witnessesStr)
	if len(witnessesStr) == 0 {
		return nil, nil, nil
	}

	witnessesStrs := strings.Split(witnessesStr, ",")
	addresses := make([]sdk.AccAddress, len(witnessesStrs))
	evmAddresses := make([]string, len(witnessesStrs))
	for i, witnessStr := range witnessesStrs {
		// Is it in evm format?
		witnessStr = strings.TrimPrefix(witnessStr, "0x")
		addr, err := sdk.AccAddressFromHexUnsafe(witnessStr)
		if err != nil {
			// Is it in bech32 format?
			witness, err := sdk.AccAddressFromBech32(witnessStr)
			if err != nil {
				return nil, nil, err
			}

			evmAddresses[i] = hex.EncodeToString(witness.Bytes())
			addresses[i] = witness
		} else {
			evmAddresses[i] = witnessStr
			addresses[i] = addr
		}
	}

	return addresses, evmAddresses, nil
}

func ParseThreshold(thresholdStr string, witnessesLength int) (int64, error) {
	threshold, err := strconv.ParseInt(thresholdStr, 10, 64)
	if err != nil {
		return 0, err
	}

	if threshold > int64(witnessesLength) || threshold <= 0 {
		return 0, fmt.Errorf("threshold must be greater than 0 and less than or equal of witnesses length %d", witnessesLength)
	}

	return threshold, nil
}
