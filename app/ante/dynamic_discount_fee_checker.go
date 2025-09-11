package ante

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/evm/ante/evm"
	anteinterfaces "github.com/cosmos/evm/ante/interfaces"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

const (
	MinDiscountValue int64 = 0
	MaxDiscountValue int64 = 100
)

type Discount int64

func (d Discount) Int64() int64 {
	return int64(d)
}

func (d Discount) Int() sdkmath.Int {
	return sdkmath.NewInt(int64(d))
}

func (d Discount) IsZero() bool {
	return d.Int64() == 0
}

func (d Discount) IsValid() bool {
	intD := d.Int64()
	return intD >= MinDiscountValue && intD <= MaxDiscountValue
}

func NewDynamicDiscountTxFeeChecker(keeper anteinterfaces.FeeMarketKeeper, discount Discount) ante.TxFeeChecker {
	return func(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
		feeTx, ok := tx.(sdk.FeeTx)

		denom := evmtypes.GetEVMCoinDenom()
		ethCfg := evmtypes.GetEthChainConfig()

		if !ok {
			return nil, 0, fmt.Errorf(" Tx must be a FeeTx")
		}

		fee, priority, err := evm.FeeChecker(ctx, keeper, denom, ethCfg, feeTx)
		if err != nil {
			return nil, 0, err
		}

		if !IsDiscountApplicable(tx) {
			return fee, priority, nil
		}

		return ApplyFeeDiscount(fee, priority, discount, denom)
	}
}

func IsDiscountApplicable(tx sdk.Tx) bool {
	for _, msg := range tx.GetMsgs() {
		if _, ok := msg.(*banktypes.MsgSend); ok {
			return true
		}
	}
	return false
}

func ApplyFeeDiscount(fee sdk.Coins, priority int64, discount Discount, denom string) (sdk.Coins, int64, error) {
	found, feeCoin := fee.Find(denom)
	if !found {
		return fee, priority, fmt.Errorf("fee not found for denom: %s", denom)
	}

	if !discount.IsValid() || discount.IsZero() {
		return fee, priority, fmt.Errorf("invalid discount: %d", discount)
	}

	if fee.IsZero() {
		return fee, priority, nil
	}

	discountAmt := feeCoin.Amount.Mul(discount.Int()).Quo(sdkmath.NewInt(100))
	discountedFeeAmt := feeCoin.Amount.Sub(discountAmt)

	priorityInt := sdkmath.NewInt(priority)
	discountedPriority := priorityInt.Add(priorityInt.Mul(discount.Int()).Quo(sdkmath.NewInt(100)))

	discountedFee := sdk.Coins{
		{
			Denom:  denom,
			Amount: discountedFeeAmt,
		},
	}

	return discountedFee, discountedPriority.Int64(), nil
}
