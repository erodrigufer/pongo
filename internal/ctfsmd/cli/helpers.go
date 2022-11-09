package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// bindFlag, binds a Cobra flag to a key managed by Viper.
// Parameters:
// viperKey, key used by Viper to manage configuration parameter.
// flagKey, key used as the flag.
func bindFlag(cmd *cobra.Command, viperKey, flagKey string) error {
	// Bind flag to viper key.
	if err := viper.BindPFlag(viperKey, cmd.Flags().Lookup(flagKey)); err != nil {
		return fmt.Errorf("error binding flag: %w", err)
	}

	return nil
}
