package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func (s *TestSuite) TestRemoveValidator(t *testing.T) {
	tt := []struct {
		name string
		valAddress string
		expectedError error
	}{
		{
			name: "remove unexisting validator - validator without balance",
		},
		{
			name: "remove unexisting validator - validator with balance",
		},
	}

	for _, tc := range tt {
		s.Run(tc.name, func() {
			// Submit MsgRemoveValidator proposal
			err := s.Network().PoaKeeper().ExecuteRemoveValidator(
				s.Network().GetContext(),
				tc.valAddress,
			)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

