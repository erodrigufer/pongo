package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: fmt.Sprintf("Runs %s in the local host.", executableName),
	Long:  fmt.Sprintf("Runs %s in the local host.", executableName),
	// Command does not accept any positional arguments )no arguments other
	// than flags. If any arguments are submitted, the command will return an
	// error.
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run ctfsmd.")
		if noInstrumentationRunFlag {
			fmt.Println("Running without instrumentation.")
		} else {
			fmt.Println("Running with instrumentation.")
		}
	},
}

var noInstrumentationRunFlag bool

func configureRunCmd(parentCmd *cobra.Command) {
	// Add local flag to check if application should run with or without
	// instrumentation. Default behaviour is to run with instrumentation.
	runCmd.Flags().BoolVar(&noInstrumentationRunFlag, "no-instrumentation", false, "No instrumentation (Prometheus monitoring) will be performed by the application while running.")
	// Add runCmd as a child from its parent command.
	parentCmd.AddCommand(runCmd)

}
