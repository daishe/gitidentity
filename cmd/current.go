package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/daishe/gitidentity/internal/identity"
)

type currentOptions struct {
	all bool
}

func currentCmd(r *rootOptions) *cobra.Command {
	o := &currentOptions{}
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show current identity",
		Long:  "Show set identity in current repository.",
	}
	cmd.Flags().BoolVar(&o.all, "all", false, "Include all Git configs, not only the local one")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if !currentCmdRun(cmd, r, o, args) {
			os.Exit(1)
		}
	}
	return cmd
}

func currentCmdRun(cmd *cobra.Command, r *rootOptions, o *currentOptions, args []string) bool {
	i, err := identity.CurrentIdentity(o.all)
	if err != nil {
		showErr(cmd, err)
		return false
	}
	fmt.Fprintln(cmd.OutOrStdout(), identity.IdentityAsString(i))
	return true
}
