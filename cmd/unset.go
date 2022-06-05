package cmd

import (
	"os"

	"github.com/spf13/cobra"

	configv1 "github.com/daishe/gitidentity/config/v1"
	"github.com/daishe/gitidentity/internal/identity"
)

type unsetOptions struct {
	skipName  bool
	skipEmail bool
}

func unsetCmd(r *rootOptions) *cobra.Command {
	o := &unsetOptions{}
	cmd := &cobra.Command{
		Use:   "unset",
		Short: "Unset local repository identity",
		Long:  "Unset local identity in current repository.",
	}
	cmd.Flags().BoolVar(&o.skipName, "skip-name", false, "Skip unsetting user name value")
	cmd.Flags().BoolVar(&o.skipEmail, "skip-email", false, "Skip unsetting user email value")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if !unsetCmdRun(cmd, r, o, args) {
			os.Exit(1)
		}
	}
	return cmd
}

func unsetCmdRun(cmd *cobra.Command, r *rootOptions, o *unsetOptions, args []string) bool {
	if !o.skipName && !o.skipEmail { // fast path for unsetting everything
		if err := identity.ApplyIdentity(&configv1.Identity{}); err != nil {
			showErr(cmd, err)
			return false
		}
		return true
	}

	current, err := identity.CurrentIdentity(false)
	if err != nil {
		showErr(cmd, err)
		return false
	}
	if !o.skipName {
		current.Name = ""
	}
	if !o.skipEmail {
		current.Email = ""
	}
	if err := identity.ApplyIdentity(current); err != nil {
		showErr(cmd, err)
		return false
	}
	return true
}
