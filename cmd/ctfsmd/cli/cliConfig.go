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
	// Set prefix for all env. variables, e.g. 'CTFSMD_NOINSTRUMENTATION', the
	// prefix for all env. variables is thereafter 'CTFSMD'.
	viper.SetEnvPrefix(executableName) // This will automatically be
	// uppercased by viper.

}
