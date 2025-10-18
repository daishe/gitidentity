package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/daishe/gitidentity/internal/identity"
)

type currentOptions struct {
	all    bool
	format string
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
	cmd.Flags().StringVar(&o.format, "format", "short", "output format, possible values are: short, JSON or YAML")
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

	var format identity.Format
	switch strings.ToLower(o.format) {
	case "short":
		fmt.Fprintln(cmd.OutOrStdout(), identity.IdentityAsString(i))
		return true
	case "json":
		format = identity.FormatJSON
	case "yaml":
		format = identity.FormatYAML
	default:
		showErr(cmd, fmt.Errorf("unknown output format %q", o.format))
		return false
	}

	out, err := identity.MarshalIdentity(i, format)
	if err != nil {
		showErr(cmd, err)
		return false
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(out))
	return true
}
