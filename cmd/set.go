package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/daishe/gitidentity/internal/identity"
)

type setOptions struct {
	name  string
	email string
}

func setCmd(r *rootOptions) *cobra.Command {
	o := &setOptions{}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set local repository identity",
		Long:  "Set local identity for current repository based on gitidentity user configuration file.",
	}
	cmd.Flags().StringVar(&o.name, "name", "", "User name value")
	cmd.Flags().StringVar(&o.email, "email", "", "User email value")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if !setCmdRun(cmd, r, o, args) {
			os.Exit(1)
		}
	}
	return cmd
}

func setCmdRun(cmd *cobra.Command, r *rootOptions, o *setOptions, args []string) bool {
	cfg, err := identity.ReadConfig(r.config)
	if err != nil {
		showErr(cmd, err)
		return false
	}

	stringifiedIdentities := identity.IdentitiesAsStrings(cfg.List)
	addMetadataToStringifiedIdentity(stringifiedIdentities)

	idx, err := selectPrompt("Select identity", stringifiedIdentities)
	if err != nil {
		showErr(cmd, err)
		return false
	}

	if err := identity.ApplyIdentity(cfg.List[idx]); err != nil {
		showErr(cmd, err)
		return false
	}
	return true
}

func addMetadataToStringifiedIdentity(stringifiedIdentities []string) {
	current, err := identity.CurrentIdentity(false)
	currentStr := identity.IdentityAsString(current)
	if err != nil {
		currentStr = "" // setting current to empty will effectively result in skipping 'current' metadata tag
	}
	global, err := identity.GlobalIdentity()
	globalStr := identity.IdentityAsString(global)
	if err != nil {
		globalStr = "" // setting global to empty will effectively result in skipping 'global' metadata tag
	}
	for idx := range stringifiedIdentities {
		if stringifiedIdentities[idx] == currentStr && currentStr == globalStr {
			stringifiedIdentities[idx] += " (current, global)"
		} else if stringifiedIdentities[idx] == currentStr {
			stringifiedIdentities[idx] += " (current)"
		} else if stringifiedIdentities[idx] == globalStr {
			stringifiedIdentities[idx] += " (global)"
		}
	}
}
