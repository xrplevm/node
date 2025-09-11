package ante

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	poatypes "github.com/xrplevm/node/v9/x/poa/types"
	"google.golang.org/protobuf/proto"
)

type mockTx struct {
	msgs []sdk.Msg
}

// Implement sdk.Tx interface
func (m mockTx) GetMsgs() []sdk.Msg                  { return m.msgs }
func (m mockTx) GetMsgsV2() ([]proto.Message, error) { return nil, nil }
func (m mockTx) ValidateBasic() error                { return nil }

func TestIsDiscountApplicable(t *testing.T) {
	// Helper type to mock sdk.Tx

	tt := []struct {
		name           string
		tx             sdk.Tx
		expectedResult bool
	}{
		{
			name:           "should return true if a discountable message is found",
			tx:             mockTx{msgs: []sdk.Msg{&banktypes.MsgSend{}}},
			expectedResult: true,
		},
		{
			name:           "should return false if no messages",
			tx:             mockTx{msgs: []sdk.Msg{}},
			expectedResult: false,
		},
		{
			name:           "should return false if no discountable message is found",
			tx:             mockTx{msgs: []sdk.Msg{&poatypes.MsgAddValidator{}}},
			expectedResult: false,
		},
		{
			name:           "should return true if at least one discountable message is found among others",
			tx:             mockTx{msgs: []sdk.Msg{&poatypes.MsgAddValidator{}, &banktypes.MsgSend{}}},
			expectedResult: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := IsDiscountApplicable(tc.tx)
			if result != tc.expectedResult {
				t.Errorf("expected %v, got %v", tc.expectedResult, result)
			}
		})
	}
}

func TestApplyFeeDiscount(t *testing.T) {
	denom := "token"

	discount := Discount(10)
	zeroFeeCoins := sdk.Coins{sdk.NewInt64Coin(denom, 0)}
	feeCoins := sdk.Coins{sdk.NewInt64Coin(denom, 100)}

	tt := []struct {
		name             string
		fee              sdk.Coins
		priority         int64
		discount         Discount
		denom            string
		expectedFee      sdk.Coins
		expectedPriority int64
		expectedErr      bool
		errContains      string
	}{
		{
			name:        "should return an error if no fee denom is found",
			fee:         feeCoins,
			denom:       "unknown",
			expectedErr: true,
			errContains: "fee not found for denom",
		},
		{
			name:        "should return an error if the discount is not valid (lower boundaries)",
			fee:         feeCoins,
			denom:       denom,
			discount:    Discount(-1),
			expectedErr: true,
			errContains: "invalid discount",
		},
		{
			name:        "should return an error if the discount is not valid (upper boundaries)",
			fee:         feeCoins,
			denom:       denom,
			discount:    Discount(101),
			expectedErr: true,
			errContains: "invalid discount",
		},
		{
			name:        "should return an error if the discount is zero",
			fee:         zeroFeeCoins,
			denom:       denom,
			discount:    Discount(0),
			expectedErr: true,
			errContains: "invalid discount",
		},
		{
			name:             "should return the same fee and priority if fee is zero",
			fee:              zeroFeeCoins,
			priority:         10,
			denom:            denom,
			discount:         discount,
			expectedFee:      sdk.Coins{sdk.NewInt64Coin("token", 0)},
			expectedPriority: 10,
		},
		{
			name:             "should return the discounted fee and priority (fee=5, priority=5, discount=10)",
			fee:              sdk.Coins{sdk.NewInt64Coin(denom, 5)},
			priority:         5,
			denom:            denom,
			discount:         discount,
			expectedFee:      sdk.Coins{sdk.NewInt64Coin(denom, 5)},
			expectedPriority: 5,
		},
		{
			name:             "should return the discounted fee and priority (fee=10, priority=10, discount=10)",
			fee:              sdk.Coins{sdk.NewInt64Coin("token", 10)},
			priority:         10,
			denom:            denom,
			discount:         discount,
			expectedFee:      sdk.Coins{sdk.NewInt64Coin("token", 9)},
			expectedPriority: 11,
		},
		{
			name:             "should return the discounted fee and priority (fee=100, priority=100, discount=10)",
			fee:              feeCoins,
			priority:         100,
			denom:            denom,
			discount:         discount,
			expectedFee:      sdk.Coins{sdk.NewInt64Coin("token", 90)},
			expectedPriority: 110,
		},
		{
			name:             "should return the discounted fee and priority (fee=100, priority=0, discount=10)",
			fee:              feeCoins,
			priority:         0,
			denom:            denom,
			discount:         discount,
			expectedFee:      sdk.Coins{sdk.NewInt64Coin("token", 90)},
			expectedPriority: 0,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			fee, priority, err := ApplyFeeDiscount(tc.fee, tc.priority, tc.discount, tc.denom)

			if tc.expectedErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("expected %s, got %s", tc.errContains, err.Error())
				}
			} else {
				if !tc.expectedErr && err != nil {
					t.Errorf("expected nil error, got %v", err)
				}
				if tc.expectedPriority != priority {
					t.Errorf("expected %d, got %d", tc.expectedPriority, priority)
				}
				if !tc.expectedFee.Equal(fee) {
					t.Errorf("expected %v, got %v", tc.expectedFee, fee)
				}
			}
		})
	}
}
