package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var executableName string = "ctfsmd"

var rootCmdLongDescr string = fmt.Sprintf(`%s is a dynamic container manager for capture the flag (CTF) events.
   
It can be used to provide each participant of a CTF event with an individual 
and isolated container.`, executableName)

var rootCmd = &cobra.Command{
	Use:   executableName,
	Short: fmt.Sprintf("%s - CTF Session Manager Daemon", executableName),
	Long:  rootCmdLongDescr,
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
