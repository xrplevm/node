package cmd

import (
	"bytes"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/xrplevm/node/v12/testutil/constants"
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

func TestRootCmd_RejectsDirectAux(t *testing.T) {
	tt := []struct {
		name     string
		signMode []string
	}{
		{name: "default sign mode"},
		{name: "direct-aux", signMode: []string{"--sign-mode=direct-aux"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// NewRootCmd reads and creates client.toml in the default node home.
			t.Setenv("NODE_HOME", t.TempDir())
			rootCmd, _ := NewRootCmd()
			rootCmd.SilenceUsage = true

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&bytes.Buffer{})
			rootCmd.SetArgs(append([]string{
				"tx", "bank", "send", "from", "to", "1axrp",
				"--aux", "--offline", "--account-number=0", "--sequence=0",
				"--chain-id=" + constants.LocalnetChainID.ChainID,
			}, tc.signMode...))

			require.ErrorContains(t, rootCmd.Execute(), "sign mode direct-aux is not enabled")
			require.Empty(t, out.String())
		})
	}
}
