package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"

	configv2 "github.com/daishe/gitidentity/config/v2"
	"github.com/daishe/gitidentity/internal/identity"
	"github.com/daishe/gitidentity/internal/runcmd"
)

type addOptions struct {
	id     string
	name   string
	email  string
	values []string
}

func addCmd(r *rootOptions) *cobra.Command {
	o := &addOptions{}
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add new identity to configuration",
		Long:  "Add new identity to gitidentity user configuration file.",

		Run: func(cmd *cobra.Command, args []string) {
			if !addCmdRun(cmd, r, o, args) {
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&o.id, "id", "", "identifier of identity for easier identification (automatically generated from name and email when empty)")
	cmd.Flags().StringVar(&o.name, "name", "", "user name value")
	cmd.Flags().StringVar(&o.email, "email", "", "user email value")
	cmd.Flags().StringArrayVar(&o.values, "value", nil, "extra git config value")
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

	i := &configv2.Identity{
		Identifier: o.id,
		Values:     make(map[string]string, len(o.values)+2),
	}
	for _, v := range o.values {
		k, d, _ := strings.Cut(v, "=")
		i.Values[k] = d
	}
	i.Values[runcmd.GitNameKey] = o.name
	i.Values[runcmd.GitEmailKey] = o.email

	cfg.List = append(cfg.GetList(), i)
	identity.SortIdentities(cfg.GetList())
	if err := identity.WriteConfig(r.config, cfg); err != nil {
		showErr(cmd, err)
		return false
	}
	return true
}
