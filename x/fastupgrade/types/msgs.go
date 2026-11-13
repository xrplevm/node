package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var _ sdk.HasValidateBasic = (*MsgUpdateParams)(nil)

// ValidateBasic lets gov reject invalid params at proposal submission.
func (m *MsgUpdateParams) ValidateBasic() error {
	return m.Params.Validate()
}
