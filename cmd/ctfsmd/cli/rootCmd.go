package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/subosito/gotenv"
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
	if err := gotenv.Load(); err != nil {
		// TODO: handle the error case for gotenv (no file and so on) properly.
		// gotenv returns an error, when there is no .env file, or when the
		// env. variables are poorly defined, e.g.
		// 'CTFSMD_NOINTRUMENTATION true', here the '=' is missing.
		fmt.Println("error", err)
	}
	// error: open .env-is-not-exist: no such file or directory
	// Configures the CLI, e.g. define all the children commands to the
	// root command.
	configCLI()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "There was an error while executing the CLI of ctfsmd: %v\n", err)
		os.Exit(1)
	}
}
