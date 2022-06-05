package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	configv1 "github.com/daishe/gitidentity/config/v1"
	"github.com/daishe/gitidentity/internal/identity"
)

type addOptions struct {
	name  string
	email string
}

func addCmd(r *rootOptions) *cobra.Command {
	o := &addOptions{}
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add new identity to configuration",
		Long:  "Add new identity to gitidentity user configuration file.",
	}
	cmd.Flags().StringVar(&o.name, "name", "", "User name value")
	cmd.Flags().StringVar(&o.email, "email", "", "User email value")
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if !addCmdRun(cmd, r, o, args) {
			os.Exit(1)
		}
	}
	return cmd
}

func addCmdRun(cmd *cobra.Command, r *rootOptions, o *addOptions, args []string) bool {
	cfg, err := identity.ReadConfig(r.config)
	if errors.Is(err, os.ErrNotExist) {
		cfg, err = identity.EmptyConfig(), nil
	}
	if err != nil {
		showErr(cmd, err)
		return false
	}
	cfg.List = append(cfg.List, &configv1.Identity{Name: o.name, Email: o.email})
	identity.SortIdentities(cfg.List)
	if err := identity.WriteConfig(r.config, cfg); err != nil {
		showErr(cmd, err)
		return false
	}
	return true
}
