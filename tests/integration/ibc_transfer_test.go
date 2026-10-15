package integration

import (
	"errors"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/evm/contracts"
	"github.com/cosmos/evm/testutil/keyring"
	testutiltypes "github.com/cosmos/evm/testutil/types"
	erc20types "github.com/cosmos/evm/x/erc20/types"
	"github.com/cosmos/evm/x/vm/statedb"
	evmtypes "github.com/cosmos/evm/x/vm/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v10/modules/core/05-port/types"
	"github.com/ethereum/go-ethereum/common"
)

// These tests cover the ICS-20 transfer stack after v12 removed the erc20 IBC
// middleware. They call the receive, acknowledgement, and timeout callbacks of
// the stack that the app routes to, so every middleware wired in app.go runs
// them. sendOut stands in for the send side. No light client or channel is
// needed: the rate-limit and transfer callbacks used here read no channel state.

const (
	ourChannel   = "channel-0" // this chain's end of the channel
	theirChannel = "channel-7" // the counterparty's end of the channel
	theirAccount = "counterparty-account"
)

var transferAmount = sdkmath.NewInt(500)

// returnPath is one way a token that this chain sent over IBC comes back.
type returnPath struct {
	name string
	run  func(s *TestSuite, ctx sdk.Context, owner sdk.AccAddress, coin sdk.Coin)
}

var returnPaths = []returnPath{
	{"counterparty sends it back", (*TestSuite).receiveReturningCoin},
	{"outgoing packet times out", (*TestSuite).timeOutOutgoing},
	{"outgoing packet gets an error ack", (*TestSuite).failOutgoingWithErrorAck},
}

func (s *TestSuite) TestIBCTransfer_NewVoucherIsNotRegisteredAsERC20() {
	ctx := s.network.GetContext()
	alice := s.keyring.GetKey(0)

	s.receiveFromCounterparty(ctx, "uatom", transferAmount, alice.AccAddr)

	voucher := voucherDenom("uatom")
	s.requireAmount("Alice's voucher balance", transferAmount, s.bankBalance(ctx, alice.AccAddr, voucher))
	s.Require().False(s.network.ERC20Keeper().IsDenomRegistered(ctx, voucher), "the voucher got an ERC20 token pair")
}

// Voucher pairs registered before v12 keep working. Their ERC20 precompile has
// no storage of its own and reads the bank balance.
func (s *TestSuite) TestIBCTransfer_RegisteredVoucherKeepsItsERC20() {
	ctx := s.network.GetContext()
	alice := s.keyring.GetKey(0)
	voucher := voucherDenom("uatom")
	pair, err := s.network.ERC20Keeper().RegisterERC20Extension(ctx, voucher)
	s.Require().NoError(err)

	s.receiveFromCounterparty(ctx, "uatom", transferAmount, alice.AccAddr)

	s.requireAmount("Alice's voucher balance", transferAmount, s.bankBalance(ctx, alice.AccAddr, voucher))
	s.requireAmount("Alice's balance on the voucher ERC20", transferAmount, s.erc20Balance(ctx, pair.GetERC20Contract(), alice.Addr))
}

// XRP is a bank coin, so it needs no conversion when it comes back.
func (s *TestSuite) TestIBCTransfer_XRPComesBack() {
	for _, path := range returnPaths {
		s.Run(path.name, func() {
			s.SetupTest()
			ctx := s.network.GetContext()
			alice := s.keyring.GetKey(0)
			xrp := sdk.NewCoin(s.network.GetBaseDenom(), transferAmount)
			balanceBefore := s.bankBalance(ctx, alice.AccAddr, xrp.Denom)

			s.sendOut(ctx, alice.AccAddr, xrp)
			path.run(s, ctx, alice.AccAddr, xrp)

			s.requireAmount("Alice's XRP balance", balanceBefore, s.bankBalance(ctx, alice.AccAddr, xrp.Denom))
			s.requireAmount("XRP in escrow", sdkmath.ZeroInt(), s.escrowed(ctx, xrp.Denom))
		})
	}
}

// A native ERC20 is a contract deployed on this chain and registered for
// conversion. Without the erc20 middleware, tokens that come back over IBC
// stay as the erc20: bank coin until the owner converts them.
func (s *TestSuite) TestIBCTransfer_NativeERC20ComesBack() {
	minted := sdkmath.NewInt(1000)

	for _, path := range returnPaths {
		s.Run(path.name, func() {
			s.SetupTest()
			alice := s.keyring.GetKey(0)
			token := s.deployERC20(alice)

			ctx := s.network.GetContext()
			s.mintERC20(ctx, token, alice, minted)
			s.registerERC20(ctx, token)
			coin := sdk.NewCoin(erc20types.CreateDenom(token.Hex()), transferAmount)

			// MsgTransfer escrows bank coins, so Alice converts before she sends.
			s.convertERC20ToCoin(ctx, token, alice, transferAmount)
			s.requireCoinsBackedByLockedERC20(ctx, token)

			s.sendOut(ctx, alice.AccAddr, coin)

			// The tokens come back as the erc20: coin. Her ERC20 balance does
			// not change until she converts them.
			path.run(s, ctx, alice.AccAddr, coin)
			s.requireAmount("Alice's erc20: coin balance", transferAmount, s.bankBalance(ctx, alice.AccAddr, coin.Denom))
			s.requireAmount("Alice's ERC20 balance", minted.Sub(transferAmount), s.erc20Balance(ctx, token, alice.Addr))
			s.requireAmount("erc20: coins in escrow", sdkmath.ZeroInt(), s.escrowed(ctx, coin.Denom))
			s.requireCoinsBackedByLockedERC20(ctx, token)

			// Converting the coins gives back all the ERC20 she minted.
			s.convertCoinToERC20(ctx, coin, alice)
			s.requireAmount("Alice's ERC20 balance", minted, s.erc20Balance(ctx, token, alice.Addr))
			s.requireAmount("erc20: coin supply", sdkmath.ZeroInt(), s.network.BankKeeper().GetSupply(ctx, coin.Denom).Amount)
			s.requireCoinsBackedByLockedERC20(ctx, token)
		})
	}
}

// voucherDenom is the denom that this chain gives a token that the
// counterparty sends over ourChannel.
func voucherDenom(baseDenom string) string {
	return transfertypes.NewDenom(baseDenom, transfertypes.NewHop(transfertypes.PortID, ourChannel)).IBCDenom()
}

func (s *TestSuite) transferStack() porttypes.IBCModule {
	stack, ok := s.network.IBCKeeper().PortKeeper.Route(transfertypes.ModuleName)
	s.Require().True(ok)
	return stack
}

func (s *TestSuite) relayer() sdk.AccAddress {
	return s.keyring.GetAccAddr(2)
}

// receiveFromCounterparty delivers a packet that the counterparty sent over
// ourChannel. denomPath is the denom as the counterparty names it.
func (s *TestSuite) receiveFromCounterparty(ctx sdk.Context, denomPath string, amount sdkmath.Int, receiver sdk.AccAddress) {
	data := transfertypes.NewFungibleTokenPacketData(denomPath, amount.String(), theirAccount, receiver.String(), "")
	packet := channeltypes.NewPacket(data.GetBytes(), 1, transfertypes.PortID, theirChannel, transfertypes.PortID, ourChannel, clienttypes.ZeroHeight(), 1)

	ack := s.transferStack().OnRecvPacket(ctx, transfertypes.V1, packet, s.relayer())
	s.Require().True(ack.Success(), string(ack.Acknowledgement()))
}

// sendOut puts coin in the escrow of ourChannel, as MsgTransfer does before
// the packet is relayed. MsgTransfer itself needs an open channel, which this
// suite does not have.
func (s *TestSuite) sendOut(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) {
	escrow := transfertypes.GetEscrowAddress(transfertypes.PortID, ourChannel)
	s.Require().NoError(s.network.TransferKeeper().EscrowCoin(ctx, sender, escrow, coin))
}

// outgoingPacket is the packet that sendOut stands for.
func outgoingPacket(sender sdk.AccAddress, coin sdk.Coin) channeltypes.Packet {
	data := transfertypes.NewFungibleTokenPacketData(coin.Denom, coin.Amount.String(), sender.String(), theirAccount, "")
	return channeltypes.NewPacket(data.GetBytes(), 1, transfertypes.PortID, ourChannel, transfertypes.PortID, theirChannel, clienttypes.ZeroHeight(), 1)
}

// receiveReturningCoin delivers the coin back from the counterparty. The
// counterparty names it with its own port and channel in front.
func (s *TestSuite) receiveReturningCoin(ctx sdk.Context, owner sdk.AccAddress, coin sdk.Coin) {
	denomPath := transfertypes.NewDenom(coin.Denom, transfertypes.NewHop(transfertypes.PortID, theirChannel)).Path()
	s.receiveFromCounterparty(ctx, denomPath, coin.Amount, owner)
}

func (s *TestSuite) timeOutOutgoing(ctx sdk.Context, owner sdk.AccAddress, coin sdk.Coin) {
	s.Require().NoError(s.transferStack().OnTimeoutPacket(ctx, transfertypes.V1, outgoingPacket(owner, coin), s.relayer()))
}

func (s *TestSuite) failOutgoingWithErrorAck(ctx sdk.Context, owner sdk.AccAddress, coin sdk.Coin) {
	ack := channeltypes.NewErrorAcknowledgement(errors.New("counterparty rejected the packet")).Acknowledgement()
	s.Require().NoError(s.transferStack().OnAcknowledgementPacket(ctx, transfertypes.V1, outgoingPacket(owner, coin), ack, s.relayer()))
}

// deployERC20 deploys an ERC20 with owner as its minter and commits the block.
func (s *TestSuite) deployERC20(owner keyring.Key) common.Address {
	token, err := s.factory.DeployContract(owner.Priv, evmtypes.EvmTxArgs{}, testutiltypes.ContractDeploymentData{
		Contract:        contracts.ERC20MinterBurnerDecimalsContract,
		ConstructorArgs: []interface{}{"Token", "TKN", uint8(18)},
	})
	s.Require().NoError(err)
	s.Require().NoError(s.network.NextBlock())
	return token
}

func (s *TestSuite) mintERC20(ctx sdk.Context, token common.Address, owner keyring.Key, amount sdkmath.Int) {
	evmKeeper := s.network.EvmKeeper()
	stateDB := statedb.New(ctx, &evmKeeper, statedb.NewEmptyTxConfig())
	_, err := evmKeeper.CallEVM(ctx, stateDB, contracts.ERC20MinterBurnerDecimalsContract.ABI, owner.Addr, token, true, false, nil, "mint", owner.Addr, amount.BigInt())
	s.Require().NoError(err)
}

// registerERC20 registers token for conversion, as a passed governance
// proposal does.
func (s *TestSuite) registerERC20(ctx sdk.Context, token common.Address) {
	erc20Keeper := s.network.ERC20Keeper()
	_, err := erc20Keeper.RegisterERC20(ctx, &erc20types.MsgRegisterERC20{
		Signer:         sdk.AccAddress(address.Module("gov")).String(),
		Erc20Addresses: []string{token.Hex()},
	})
	s.Require().NoError(err)
}

func (s *TestSuite) convertERC20ToCoin(ctx sdk.Context, token common.Address, owner keyring.Key, amount sdkmath.Int) {
	_, err := s.network.ERC20Keeper().ConvertERC20(ctx, &erc20types.MsgConvertERC20{
		ContractAddress: token.Hex(),
		Amount:          amount,
		Receiver:        owner.AccAddr.String(),
		Sender:          owner.Addr.Hex(),
	})
	s.Require().NoError(err)
}

func (s *TestSuite) convertCoinToERC20(ctx sdk.Context, coin sdk.Coin, owner keyring.Key) {
	_, err := s.network.ERC20Keeper().ConvertCoin(ctx, &erc20types.MsgConvertCoin{
		Coin:     coin,
		Receiver: owner.Addr.Hex(),
		Sender:   owner.AccAddr.String(),
	})
	s.Require().NoError(err)
}

func (s *TestSuite) requireAmount(what string, want, got sdkmath.Int) {
	s.T().Helper()
	s.Require().Truef(want.Equal(got), "%s: want %s, got %s", what, want, got)
}

func (s *TestSuite) bankBalance(ctx sdk.Context, holder sdk.AccAddress, denom string) sdkmath.Int {
	return s.network.BankKeeper().GetBalance(ctx, holder, denom).Amount
}

// escrowed is the amount of denom that IBC escrow holds for transfers still in flight.
func (s *TestSuite) escrowed(ctx sdk.Context, denom string) sdkmath.Int {
	return s.network.TransferKeeper().GetTotalEscrowForDenom(ctx, denom).Amount
}

func (s *TestSuite) erc20Balance(ctx sdk.Context, token, holder common.Address) sdkmath.Int {
	balance := s.network.ERC20Keeper().BalanceOf(ctx, contracts.ERC20MinterBurnerDecimalsContract.ABI, token, holder)
	s.Require().NotNil(balance)
	return sdkmath.NewIntFromBigInt(balance)
}

// requireCoinsBackedByLockedERC20 checks that every erc20: coin in existence,
// escrowed ones included, has one ERC20 token locked in the erc20 module
// behind it. A mint or burn bug breaks this equality.
func (s *TestSuite) requireCoinsBackedByLockedERC20(ctx sdk.Context, token common.Address) {
	supply := s.network.BankKeeper().GetSupply(ctx, erc20types.CreateDenom(token.Hex())).Amount
	locked := s.erc20Balance(ctx, token, erc20types.ModuleAddress)
	s.requireAmount("ERC20 locked in the erc20 module, which must equal the erc20: coin supply", supply, locked)
}
