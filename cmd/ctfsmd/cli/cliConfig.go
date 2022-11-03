package cli

import (
	"fmt"

	"github.com/spf13/viper"
)

func configCLI(app Application) error {
	configureViper()
	if err := configureRunCmd(rootCmd, app); err != nil {
		return fmt.Errorf("error configuring Run command: %w", err)
	}
	configureRevisionCmd(rootCmd)

	return nil
}

func configureViper() {
	// Set prefix for all env. variables, e.g. 'CTFSMD_NOINSTRUMENTATION', the
	// prefix for all env. variables is thereafter 'CTFSMD'.
	viper.SetEnvPrefix(executableName) // This will automatically be
	// uppercased by viper.

}
