package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ctfsmd",
	Short: "ctfsmd - CTF Session Manager Daemon",
	Long: `ctfsmd is a dynamic container manager for capture the flag events.
   
It can be used to provide each participant of a CTF event with an individual 
and isolated personalized container.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	// Configures the CLI, e.g. define all the children commands to the
	// root command.
	configCLI()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "There was an error while executing the CLI of ctfsmd: %v", err)
		os.Exit(1)
	}
}
