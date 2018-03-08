package cmds

import (
	"fmt"

	"github.com/appscode/go/flags"
	api "github.com/appscode/kubed/apis/kubed/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

func NewCmdCheck() *cobra.Command {
	var (
		configPath string
	)

	cmd := &cobra.Command{
		Use:               "check",
		Short:             "Check cluster config",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags.EnsureRequiredFlags(cmd, "clusterconfig")

			cfg, err := api.LoadConfig(configPath)
			if err != nil {
				return err
			}
			err = cfg.Validate()
			if err != nil {
				return err
			}
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return err
			}
			fmt.Println("Cluster config was parsed successfully.")
			fmt.Println()
			fmt.Println(string(data))
			return nil
		},
	}
	cmd.Flags().StringVar(&configPath, "clusterconfig", configPath, "Path to cluster config file")
	return cmd
}
