package cli

import (
	"github.com/spf13/cobra"
)

// Registering all the flags to the command neuron itself.
func registerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&unpcker.Name, "name", "n", "", "name of the asset that has to be created")
	cmd.PersistentFlags().StringVarP(&unpcker.AssetPath, "asset", "a", "", "path to asset which needs to be packed")
	cmd.PersistentFlags().StringVarP(&unpcker.Environment, "dev", "e", "", "name of environment in which the asset is packed")
	cmd.PersistentFlags().StringVarP(&unpcker.Path, "path", "p", "", "path where the asset has to be created")
	cmd.PersistentFlags().StringVarP(&unpcker.AssetVersion, "version", "v", "1.0", "version the asset that needs to be packed")
	cmd.PersistentFlags().StringVarP(&unpcker.ConfigPath, "config", "c", ".", "path where the config file exists")
}
