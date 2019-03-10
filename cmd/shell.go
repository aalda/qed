package cmd

import (
	"github.com/bbva/qed/shell"
	"github.com/spf13/cobra"
)

func newShellCommand(ctx *cmdContext) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Interactive shell for QED",
		Long:  `Client interactive shell for interacting with QED `,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &shell.Config{}
			shell, err := shell.NewShell(cfg)
			if err != nil {
				return err
			}
			shell.Run()
			return nil
		},
	}

	return cmd
}
