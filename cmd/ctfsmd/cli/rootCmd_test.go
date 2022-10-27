package cli

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

// Testing goals:
// 1) A combination of flags and environment variables, in order to see if the
// cobra/viper configuration is working properly

func TestExecute(t *testing.T) {
	tests := []struct {
		subTestName     string
		args            []string
		expectedResults map[string]interface{}
	}{
		{
			subTestName:     "Default values (no flags)",
			args:            []string{},
			expectedResults: defaultValues,
		},
		{
			subTestName: "No instrumentation flag",
			args:        []string{"run", "--no-instrumentation"},
			expectedResults: map[string]interface{}{
				"Debug":             false,
				"NoInstrumentation": true,
			},
		},
		{
			subTestName: "Debug mode",
			args:        []string{"run", "--debug"},
			expectedResults: map[string]interface{}{
				"Debug":             true,
				"NoInstrumentation": false,
			},
		},
		{
			subTestName: "No instrumentation and debug flag",
			args:        []string{"run", "--no-instrumentation", "--debug"},
			expectedResults: map[string]interface{}{
				"Debug":             true,
				"NoInstrumentation": true,
			},
		},

		{
			subTestName: "SSH Port",
			args:        []string{"run", "--ssh", "70000"},
			expectedResults: map[string]interface{}{
				"SSH":               "70000",
				"NoInstrumentation": false,
			},
		},
		{
			subTestName: "Max Available Session",
			args:        []string{"run", "--maxAvailableSess", "10"},
			expectedResults: map[string]interface{}{
				"MaxAvailableSess":  10,
				"NoInstrumentation": false,
			},
		},
	}

	originalArgs := os.Args
	// Loop over the test cases.
	for _, tt := range tests {
		t.Run(tt.subTestName, func(t *testing.T) {

			// Reset the internal viper key handling data structure so that in
			// each iteration of the for-loop the keys are not set yet.
			viper.Reset()
			runCmd.ResetFlags()
			// os.Args = []string{originalArgs[0], "run", "--no-instrumentation"}
			os.Args = []string{originalArgs[0]}
			os.Args = append(os.Args, tt.args...)
			// os.Args[1:] = tt.args

			// Inject the arguments directly into the command definition,
			// elegant way of mocking the input of flags.
			// rootCmd.SetArgs([]string{})
			// rootCmd.SetArgs(tt.args)

			err := Execute()

			if err != nil {
				t.Errorf("error: %v", err)
			}
			for key, expectedResult := range tt.expectedResults {
				if viperData := viper.Get(key); viperData != expectedResult {
					t.Errorf("/%s/(=/%v/) is not equal to expected value /%v/.", key, viperData, expectedResult)
				}
			}
		})

	}
}
