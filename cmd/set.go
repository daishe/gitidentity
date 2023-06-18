package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	configv2 "github.com/daishe/gitidentity/config/v2"
	"github.com/daishe/gitidentity/internal/identity"
	"github.com/daishe/gitidentity/internal/runcmd"
)

type setOptions struct {
	noAuto   bool
	onlyAuto bool
}

func setCmd(r *rootOptions) *cobra.Command {
	o := &setOptions{}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set local repository identity",
		Long:  "Set local identity for current repository based on gitidentity user configuration file.",

		Run: func(cmd *cobra.Command, args []string) {
			if !setCmdRun(cmd, r, o, args) {
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVar(&o.noAuto, "no-auto", false, "do not to apply auto applicable identities")
	cmd.Flags().BoolVar(&o.onlyAuto, "only-auto", false, "attempt to only apply auto applicable identities")
	return cmd
}

func setCmdRun(cmd *cobra.Command, r *rootOptions, o *setOptions, args []string) bool {
	if o.noAuto && o.onlyAuto {
		showErr(cmd, fmt.Errorf("conflicting options %q and %q", "no-auto", "only-auto"))
		return false
	}

	cfg, err := identity.ReadConfig(r.config)
	if err != nil {
		showErr(cmd, err)
		return false
	}

	if !o.noAuto {
		i, err := setCmdRun_auto(cmd.Context(), cfg.GetList())
		if err != nil {
			showErr(cmd, err)
			return false
		}
		if i != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "Automatically selected identity:", identity.IdentityAsString(i))
			return true
		}
	}
	if o.onlyAuto {
		showErr(cmd, fmt.Errorf("no matching identity"))
		return false
	}

	_, err = setCmdRun_manual(cmd.Context(), cfg.GetList())
	if err != nil {
		showErr(cmd, err)
		return false
	}
	return true
}

func setCmdRun_auto(ctx context.Context, list []*configv2.Identity) (*configv2.Identity, error) {
	gi, err := runcmd.GitInfoFromDir(ctx)
	if err != nil {
		return nil, err
	}

	i, err := identity.FirstAutoMatchingIdentity(ctx, list, gi)
	if err != nil {
		return nil, err
	}
	if i == nil {
		return nil, nil
	}

	if err := identity.ApplyIdentity(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}

func setCmdRun_manual(ctx context.Context, list []*configv2.Identity) (*configv2.Identity, error) {
	i, err := selectIdentityPrompt(ctx, list)
	if err != nil {
		return nil, err
	}
	if i == nil {
		return nil, fmt.Errorf("no identity selected")
	}

	if err := identity.ApplyIdentity(ctx, i); err != nil {
		return nil, err
	}
	return i, err
}
