package cmd

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestCheckAuxSignMode(t *testing.T) {
	tt := []struct {
		name    string
		txFlags bool
		args    []string
		wantErr bool
	}{
		{name: "aux with default sign mode", txFlags: true, args: []string{"--aux"}, wantErr: true},
		{name: "aux with empty sign mode", txFlags: true, args: []string{"--aux", "--sign-mode="}, wantErr: true},
		{name: "aux with direct-aux", txFlags: true, args: []string{"--aux", "--sign-mode=direct-aux"}, wantErr: true},
		{name: "aux with amino-json", txFlags: true, args: []string{"--aux", "--sign-mode=amino-json"}},
		{name: "aux disabled", txFlags: true, args: []string{"--aux=false"}},
		{name: "command without tx flags"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			if tc.txFlags {
				flags.AddTxFlagsToCmd(cmd)
			}
			require.NoError(t, cmd.ParseFlags(tc.args))

			err := checkAuxSignMode(cmd)
			if tc.wantErr {
				require.ErrorContains(t, err, "sign mode direct-aux is not enabled")
				return
			}
			require.NoError(t, err)
		})
	}
}
