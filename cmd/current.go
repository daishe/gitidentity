package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/daishe/gitidentity/internal/identity"
)

type currentOptions struct {
	all  bool
	json bool
}

func currentCmd(r *rootOptions) *cobra.Command {
	o := &currentOptions{}
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Show current identity",
		Long:  "Show set identity in current repository.",

		Run: func(cmd *cobra.Command, args []string) {
			if !currentCmdRun(cmd, r, o, args) {
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVar(&o.all, "all", false, "include all Git configs, not only the local one")
	cmd.Flags().BoolVar(&o.json, "json", false, "use JSON output that includes all information")
	return cmd
}

func currentCmdRun(cmd *cobra.Command, r *rootOptions, o *currentOptions, args []string) bool {
	i, err := identity.CurrentIdentity(cmd.Context(), o.all)
	if errors.Is(err, identity.ErrNoCurrentIdentity) {
		return true
	}
	if err != nil {
		showErr(cmd, err)
		return false
	}

	if !o.json {
		fmt.Fprintln(cmd.OutOrStdout(), identity.IdentityAsString(i))
		return true
	}

	out, err := identity.MarshalIdentity(i)
	if err != nil {
		showErr(cmd, err)
		return false
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(out))
	return true
}
