package cli

import (
	"github.com/spf13/viper"
)

func configCLI() {
	configureViper()
	configureRunCmd(rootCmd)
	rootCmd.AddCommand(revisionCmd)
}

func configureViper() {
	// TODO: use the variable that defines the executable's name to change
	// this value.
	// Set prefix for all env. variables, e.g. 'CTFSMD_NOINSTRUMENTATION'
	viper.SetEnvPrefix("ctfsmd") // This will automatically be
	// uppercased by viper.

}
