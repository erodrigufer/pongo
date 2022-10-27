package cli

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
)

// Testing goals:
// 1) A combination of flags and environment variables, in order to see if the
// cobra/viper configuration is working properly
type testExecute struct {
	subTestName     string
	args            []string
	expectedResults map[string]interface{}
}

// appendTestExecute, appends a subtest to a slice with test for the Execute
// function.
// This function guarantees that only the expectedKeys get changed and the rest
// of the default values map is used to test the validity of a test.
func appendTestExecute(name string, args []string, expectedKeys map[string]interface{}, tests []testExecute) ([]testExecute, error) {
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
	}
	tests = append(tests, newTest)
	return tests, nil
}

func TestExecute(t *testing.T) {
	var tests = []testExecute{
		{
			subTestName:     "Default values (no flags)",
			args:            []string{},    // Do not provide any flags.
			expectedResults: defaultValues, // The viper values should be the
			// same as the previously defined default values map.
		},
	}

	var err error
	tests, err = appendTestExecute(
		"No instrumentation flag",
		[]string{"run", "--no-instrumentation"},
		map[string]interface{}{"NoInstrumentation": true},
		tests)
	if err != nil {
		t.Fatalf("error appending test: %v", err)
	}

	tests, err = appendTestExecute(
		"Debug mode",
		[]string{"run", "--debug"},
		map[string]interface{}{"Debug": true},
		tests)
	if err != nil {
		// If there was an error appending a test run a Fatal() method, so that
		// no further subtests get executed, until all subtests are properly
		// defined.
		t.Fatalf("error appending test: %v", err)
	}
	tests, err = appendTestExecute(
		"No instrumentation and debug flag",
		[]string{"run", "--debug", "--no-instrumentation"},
		map[string]interface{}{
			"Debug":             true,
			"NoInstrumentation": true},
		tests)
	if err != nil {
		t.Fatalf("error appending test: %v", err)
	}
	tests, err = appendTestExecute(
		"SSH Port and no instrumentation",
		[]string{"run", "--ssh", "800", "--no-instrumentation"},
		map[string]interface{}{
			"SSH":               "800",
			"NoInstrumentation": true},
		tests)
	if err != nil {
		t.Fatalf("error appending test: %v", err)
	}
	tests, err = appendTestExecute(
		"Sessions flags",
		[]string{"run", "--maxAvailableSess", "44", "--maxActiveSess", "213"},
		map[string]interface{}{
			"MaxAvailableSess": 44,
			"MaxActiveSess":    213},
		tests)
	if err != nil {
		t.Fatalf("error appending test: %v", err)
	}
	tests, err = appendTestExecute(
		"Time requests, expiration of sessions, HTTP and Debug flags.",
		[]string{"run", "--lifetimeSess", "2", "--srdFreq", "16", "--timeReq", "40", "--http", ":88", "--debug"},
		map[string]interface{}{
			"LifetimeSess": 2,
			"SRDFreq":      16,
			"TimeReq":      40,
			"HTTP":         ":88",
			"Debug":        true},
		tests)
	if err != nil {
		t.Fatalf("error appending test: %v", err)
	}

	// Store the original os.Args, where the first element is the name of the
	// executable (os.Args[0]).
	originalArgs := os.Args
	// Loop over the test cases.
	for _, tt := range tests {
		t.Run(tt.subTestName, func(t *testing.T) {

			// Reset the internal viper key handling data structure so that in
			// each iteration of the for-loop the keys are not set yet.
			viper.Reset()
			// Reset the flags as well, otherwise the test panics.
			runCmd.ResetFlags()
			os.Args = []string{originalArgs[0]}
			// Append the arguments with flags specific to a subtest, the first
			// argument is the name of the executable.
			os.Args = append(os.Args, tt.args...)

			err := Execute()
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
