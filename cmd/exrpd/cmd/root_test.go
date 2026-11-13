package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
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
		name    string
		args    []string
		appToml string
	}{
		{name: "aux flag with default sign mode", args: []string{"--aux"}},
		{name: "aux flag with direct-aux", args: []string{"--aux", "--sign-mode=direct-aux"}},
		{name: "aux in app.toml", appToml: "aux = true\n"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			if tc.appToml != "" {
				require.NoError(t, os.MkdirAll(filepath.Join(home, "config"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(home, "config", "app.toml"), []byte(tc.appToml), 0o600))
			}
			// NewRootCmd reads and creates client.toml in the default node home.
			t.Setenv("NODE_HOME", home)
			rootCmd, _ := NewRootCmd()
			rootCmd.SilenceUsage = true
			// svrcmd.Execute adds --home in production.
			rootCmd.PersistentFlags().String(flags.FlagHome, home, "")

			var out bytes.Buffer
			rootCmd.SetOut(&out)
			rootCmd.SetErr(&bytes.Buffer{})
			rootCmd.SetArgs(append([]string{
				"tx", "bank", "send", "from", "to", "1axrp",
				"--offline", "--account-number=0", "--sequence=0",
				"--chain-id=" + constants.LocalnetChainID.ChainID,
			}, tc.args...))

			err := rootCmd.ExecuteContext(svrcmd.CreateExecuteContext(context.Background()))
			require.ErrorContains(t, err, "sign mode direct-aux is not enabled")
			require.Empty(t, out.String())
		})
	}
}
