package cli

import (
	"fmt"
	"os"
	"testing"

	"github.com/erodrigufer/CTForchestrator/internal/ctfsmd"
	"github.com/spf13/viper"
)

// Testing goals:
// 1) A combination of flags and environment variables, in order to see if the
// cobra/viper configuration is working properly
type testExecute struct {
	subTestName     string
	args            []string
	envVariables    map[string]interface{}
	expectedResults map[string]interface{}
}

type mockApplication struct{}

var mockApp mockApplication

func (mockApplication) Run(configValues ctfsmd.UserConfiguration) error {
	return nil
}

// appendTestExecute, appends a subtest to a slice with test for the Execute
// function.
// This function guarantees that only the expectedKeys get changed and the rest
// of the default values map is used to test the validity of a test.
func appendTestExecute(name string, args []string, expectedKeys map[string]interface{}, envVariables map[string]interface{}, tests []testExecute) ([]testExecute, error) {
	// Use the default values as the basis to change only the expectedKeys.
	dv := make(map[string]interface{})
	// Do a deep-copy of the defaultValues map.
	// Important: `dv := defaultValues` is a shallow copy, therefore both maps
	// get modified each time one of them is modified.
	for k, v := range defaultValues {
		dv[k] = v
	}

	for key, value := range expectedKeys {
		// Check if key exists in the default values map. If it does not, return
		// an error right away.
		_, ok := defaultValues[key]
		if !ok {
			return tests, fmt.Errorf("error: key=/%s/ is not a key within the default values map.", key)
		}
		// Change the value from the default values. To the expected value.
		dv[key] = value
	}
	newTest := testExecute{
		subTestName:     name,
		args:            args,
		expectedResults: dv,
		envVariables:    envVariables,
	}
	tests = append(tests, newTest)
	return tests, nil
}

func TestExecute(t *testing.T) {
	var tests = []testExecute{
		{
			subTestName:     "Default values (no flags, no env.)",
			args:            []string{"run"}, // Do not provide any flags.
			expectedResults: defaultValues,   // The viper values should be the
			// same as the previously defined default values map.
			envVariables: map[string]interface{}{},
		},
	}

	var err error
	// TODO: currently all flag tests are failing.
	// tests, err = appendTestExecute(
	// 	"No instrumentation flag",
	// 	[]string{"run", "--no-instrumentation"},
	// 	map[string]interface{}{"NoInstrumentation": true},
	// 	map[string]interface{}{},
	// 	tests)
	// if err != nil {
	// 	t.Fatalf("error appending test: %v", err)
	// }

	// tests, err = appendTestExecute(
	// 	"Debug mode",
	// 	[]string{"run", "--debug"},
	// 	map[string]interface{}{"Debug": true},
	// 	map[string]interface{}{},
	// 	tests)
	// if err != nil {
	// 	// If there was an error appending a test run a Fatal() method, so that
	// 	// no further subtests get executed, until all subtests are properly
	// 	// defined.
	// 	t.Fatalf("error appending test: %v", err)
	// }
	// tests, err = appendTestExecute(
	// 	"No instrumentation and debug flag",
	// 	[]string{"run", "--debug", "--no-instrumentation"},
	// 	map[string]interface{}{
	// 		"Debug":             true,
	// 		"NoInstrumentation": true},
	// 	map[string]interface{}{},
	// 	tests)
	// if err != nil {
	// 	t.Fatalf("error appending test: %v", err)
	// }
	// tests, err = appendTestExecute(
	// 	"SSH Port and no instrumentation flags.",
	// 	[]string{"run", "--ssh", "800", "--no-instrumentation"},
	// 	map[string]interface{}{
	// 		"SSH":               "800",
	// 		"NoInstrumentation": true},
	// 	map[string]interface{}{},
	// 	tests)
	// if err != nil {
	// 	t.Fatalf("error appending test: %v", err)
	// }
	// tests, err = appendTestExecute(
	// 	"Sessions flags",
	// 	[]string{"run", "--maxAvailableSess", "44", "--maxActiveSess", "213"},
	// 	map[string]interface{}{
	// 		"MaxAvailableSess": 44,
	// 		"MaxActiveSess":    213},
	// 	map[string]interface{}{},
	// 	tests)
	// if err != nil {
	// 	t.Fatalf("error appending test: %v", err)
	// }
	// tests, err = appendTestExecute(
	// 	"Time requests, expiration of sessions, HTTP and Debug flags.",
	// 	[]string{"run", "--lifetimeSess", "2", "--srdFreq", "16", "--timeReq", "40", "--http", ":88", "--debug"},
	// 	map[string]interface{}{
	// 		"LifetimeSess": 2,
	// 		"SRDFreq":      16,
	// 		"TimeReq":      40,
	// 		"HTTP":         ":88",
	// 		"Debug":        true},
	// 	map[string]interface{}{},
	// 	tests)
	// if err != nil {
	// 	t.Fatalf("error appending test: %v", err)
	// }

	// TODO: these two net tests fail, because the env. variables are being set
	// as strings, so when the internal values are Get() from viper they are
	// strings and the equality test fails because they have different types,
	// e.g. bool and string. The test works for the SSH value, since the actual
	// value is also a string.
	// tests, err = appendTestExecute(
	// 	"No Instrumentation with env. and debug with flag.",
	// 	[]string{"run", "--debug"},
	// 	map[string]interface{}{
	// 		"NoInstrumentation": true,
	// 		"Debug":             true},
	// 	map[string]interface{}{"CTFSMD_NOINSTRUMENTATION": "true"},
	// 	tests)
	// if err != nil {
	// 	t.Fatalf("error appending test: %v", err)
	// }
	// tests, err = appendTestExecute(
	// 	"TimeReq. with env.",
	// 	[]string{"run"},
	// 	map[string]interface{}{
	// 		"TimeReq": 77},
	// 	map[string]interface{}{"CTFSMD_TIMEREQ": "77"},
	// 	tests)
	// if err != nil {
	// 	t.Fatalf("error appending test: %v", err)
	// }
	tests, err = appendTestExecute(
		"SSH with env.",
		[]string{"run"},
		map[string]interface{}{
			"SSH": "5"},
		map[string]interface{}{"CTFSMD_SSH": "5"},
		tests)
	if err != nil {
		t.Fatalf("error appending test: %v", err)
	}
	tests, err = appendTestExecute(
		"HTTP with env.",
		[]string{"run"},
		map[string]interface{}{
			"HTTP": ":78"},
		map[string]interface{}{"CTFSMD_HTTP": ":78"},
		tests)
	if err != nil {
		t.Fatalf("error appending test: %v", err)
	}
	// TODO: flag tests are failing.
	// In this subtest the HTTP value should be the one of the flag, and not the
	// one of the env. variable.
	// tests, err = appendTestExecute(
	// 	"HTTP with env. and flag, SSH with env.",
	// 	[]string{"run", "--http", ":88"},
	// 	map[string]interface{}{
	// 		"HTTP": ":88",
	// 		"SSH":  "1"},
	// 	map[string]interface{}{
	// 		"CTFSMD_HTTP": ":78",
	// 		"CTFSMD_SSH":  "1"},
	// 	tests)
	// if err != nil {
	// 	t.Fatalf("error appending test: %v", err)
	// }

	// Store the original os.Args, where the first element is the name of the
	// executable (os.Args[0]).
	originalArgs := os.Args
	// Loop over the test cases.
	for _, tt := range tests {
		t.Run(tt.subTestName, func(t *testing.T) {
			// Reset/clear the environment before every subtest.
			os.Clearenv()
			for k, v := range tt.envVariables {
				if err := os.Setenv(k, v.(string)); err != nil {
					t.Errorf("error setting env. variable: %v", err)
				}
			}
			// Reset the internal viper key handling data structure so that in
			// each iteration of the for-loop the keys are not set yet.
			viper.Reset()
			// Reset the flags as well, otherwise the test panics.
			// TODO: fix issue, before we were resetting runCmd, now tests fail.
			rootCmd.ResetFlags()
			os.Args = []string{originalArgs[0]}
			// Append the arguments with flags specific to a subtest, the first
			// argument is the name of the executable.
			os.Args = append(os.Args, tt.args...)

			err := Execute(mockApp)
			if err != nil {
				t.Errorf("error: %v", err)
			}

			// Check that every internal value handled by viper with a key has
			// the expected result after providing flags and/or env. variables.
			for key, expectedResult := range tt.expectedResults {
				if viperData := viper.Get(key); viperData != expectedResult {
					t.Errorf("/%s/(=/%v/) is not equal to expected value /%v/.", key, viperData, expectedResult)
				}
			}
		})

	}
}
