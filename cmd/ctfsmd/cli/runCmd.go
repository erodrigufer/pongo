package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// defaultValues, default values used by viper for the different keys.
// If a user does not set a key with either a flag or an env. variable, the
// value for that key will default to the value defined in this map.
var defaultValues = map[string]interface{}{
	"NoInstrumentation": false,
	"SSH":               "50000",
	"HTTP":              ":4000",
	"MaxAvailableSess":  15,
	"MaxActiveSess":     140,
	"LifetimeSess":      150,
	"SRDFreq":           10,
	"TimeReq":           5,
	"Debug":             false,
}

type Application interface {
	Run() error
}

func newRunCmd(app Application) *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: fmt.Sprintf("Runs %s in the local machine.", executableName),
		Long:  fmt.Sprintf("Runs %s in the local machine.", executableName),
		// Command does not accept any positional arguments, no arguments other
		// than flags. If any arguments are submitted, the command will return an
		// error.
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := app.Run(); err != nil {
				return err
			}
			return nil
		},
	}
	return runCmd
}

// configureRunCmd, configures the flags and environment variables used by the
// run command, sets its defaults and adds the run command as a child command
// of root command.
func configureRunCmd(parentCmd *cobra.Command, app Application) error {
	setDefaultsRunCmd()

	runCmd := newRunCmd(app)

	if err := configureFlagsRunCmd(runCmd); err != nil {
		return fmt.Errorf("error configuring flags for Run command: %w", err)
	}

	if err := bindEnvRunCmd(); err != nil {
		return fmt.Errorf("error binding env. variables for Run command: %w", err)
	}

	// Add runCmd as a child from its parent command.
	parentCmd.AddCommand(runCmd)

	return nil
}

// setDefaultsRunCmd, set the default values for all configuration parameters
// handled by viper.
func setDefaultsRunCmd() {
	for key, value := range defaultValues {
		viper.SetDefault(key, value)
	}
}

// configureFlagsRunCmd, configure the flags used by the Run command. Bind all
// the flags to their respective keys handled by viper.
func configureFlagsRunCmd(runCmd *cobra.Command) error {
	// Add local flag to check if application should run with or without
	// instrumentation. Default behaviour is to run with instrumentation.
	runCmd.Flags().Bool("no-instrumentation", false, "No instrumentation (Prometheus monitoring) will be performed by the application while running.")
	// Bind flag to viper key.
	if err := bindFlag(runCmd, "NoInstrumentation", "no-instrumentation"); err != nil {
		return err
	}
	// Debug mode.
	runCmd.Flags().Bool("debug", false, "Run daemon in 'debug' mode. Logging will be more extensive and frequent.")
	if err := bindFlag(runCmd, "Debug", "debug"); err != nil {
		return err
	}
	// Ports options.
	runCmd.Flags().String("ssh", "50000", "Port in which SSH Piper will work as an SSH proxy. Clients connect to this port.")
	if err := bindFlag(runCmd, "SSH", "ssh"); err != nil {
		return err
	}
	runCmd.Flags().String("http", ":4000", "Port in which SSH Piper will work as an SSH proxy. Clients connect to this port.")
	if err := bindFlag(runCmd, "HTTP", "http"); err != nil {
		return err
	}
	// Sessions.
	runCmd.Flags().Int("maxAvailableSess", 15, "Number of sessions always running in the background, which are directly available to be delivered to clients.")
	if err := bindFlag(runCmd, "MaxAvailableSess", "maxAvailableSess"); err != nil {
		return err
	}
	runCmd.Flags().Int("maxActiveSess", 140, "Total max. number of sessions that can be simultaneously actively being used by clients.")
	if err := bindFlag(runCmd, "MaxActiveSess", "maxActiveSess"); err != nil {
		return err
	}
	// Lifetime of sessions and frequency to check for session expiration.
	runCmd.Flags().Int("lifetimeSess", 150, "Lifetime of session (in min) after which the session will expire.")
	if err := bindFlag(runCmd, "LifetimeSess", "lifetimeSess"); err != nil {
		return err
	}
	runCmd.Flags().Int("srdFreq", 10, "Frequency (in min) with which srd checks for expired sessions.")
	if err := bindFlag(runCmd, "SRDFreq", "srdFreq"); err != nil {
		return err
	}
	// Minimum time between requests coming from the same client.
	runCmd.Flags().Int("timeReq", 5, "Minimum time (in min) allowed between requests coming from the same client IP.")
	if err := bindFlag(runCmd, "TimeReq", "timeReq"); err != nil {
		return err
	}
	return nil
}

// bindEnvRunCmd, bind environment variables of run command.
func bindEnvRunCmd() error {
	if err := viper.BindEnv("NoInstrumentation"); err != nil {
		return fmt.Errorf("error binding env. variable: %w", err)
	}
	if err := viper.BindEnv("SSH"); err != nil {
		return fmt.Errorf("error binding env. variable: %w", err)
	}
	if err := viper.BindEnv("HTTP"); err != nil {
		return fmt.Errorf("error binding env. variable: %w", err)
	}
	if err := viper.BindEnv("MaxAvailableSess"); err != nil {
		return fmt.Errorf("error binding env. variable: %w", err)
	}
	if err := viper.BindEnv("MaxActiveSess"); err != nil {
		return fmt.Errorf("error binding env. variable: %w", err)
	}
	if err := viper.BindEnv("LifetimeSess"); err != nil {
		return fmt.Errorf("error binding env. variable: %w", err)
	}
	if err := viper.BindEnv("SRDFreq"); err != nil {
		return fmt.Errorf("error binding env. variable: %w", err)
	}
	if err := viper.BindEnv("TimeReq"); err != nil {
		return fmt.Errorf("error binding env. variable: %w", err)
	}
	if err := viper.BindEnv("Debug"); err != nil {
		return fmt.Errorf("error binding env. variable: %w", err)
	}

	return nil
}
