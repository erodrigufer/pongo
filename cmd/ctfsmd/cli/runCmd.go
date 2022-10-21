package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: fmt.Sprintf("Runs %s in the local machine.", executableName),
	Long:  fmt.Sprintf("Runs %s in the local machine.", executableName),
	// Command does not accept any positional arguments )no arguments other
	// than flags. If any arguments are submitted, the command will return an
	// error.
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run ctfsmd.")
		noInstrumentation := viper.GetBool("NoInstrumentation")
		if noInstrumentation {
			fmt.Println("Running without instrumentation.")
		} else {
			fmt.Println("Running with instrumentation.")
		}
	},
}

func configureRunCmd(parentCmd *cobra.Command) {
	viper.SetDefault("NoInstrumentation", false)
	// Add local flag to check if application should run with or without
	// instrumentation. Default behaviour is to run with instrumentation.
	runCmd.Flags().Bool("no-instrumentation", false, "No instrumentation (Prometheus monitoring) will be performed by the application while running.")

	// TODO: handle the viper methods error cases properly.
	if err := viper.BindPFlag("NoInstrumentation", runCmd.Flags().Lookup("no-instrumentation")); err != nil {
		fmt.Println("error binding flag")
	}
	if err := viper.BindEnv("NoInstrumentation"); err != nil {
		fmt.Println("error binding env. variable")
	}

	// Add runCmd as a child from its parent command.
	parentCmd.AddCommand(runCmd)

}
