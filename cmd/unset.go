package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/daishe/gitidentity/internal/identity"
)

type unsetOptions struct {
}

func unsetCmd(r *rootOptions) *cobra.Command {
	o := &unsetOptions{}
	cmd := &cobra.Command{
		Use:   "unset",
		Short: "Unset local repository identity",
		Long:  "Unset local identity in current repository.",

		Run: func(cmd *cobra.Command, args []string) {
			if !unsetCmdRun(cmd, r, o, args) {
				os.Exit(1)
			}
		},
	}

	return cmd
}

func unsetCmdRun(cmd *cobra.Command, r *rootOptions, o *unsetOptions, args []string) bool {
	if err := identity.UnsetCurrentIdentity(cmd.Context()); err != nil {
		showErr(cmd, err)
		return false
	}
	return true
}
