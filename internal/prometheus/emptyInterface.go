package prometheus

type EmptyInterface struct{}

func (ei EmptyInterface) noOp() bool {
	return true
}

func (ei EmptyInterface) incrementCounter(a string, b []string) error {
	return nil
}

func (ei EmptyInterface) incrementGauge(a string, b []string) error {
	return nil
}

func (ei EmptyInterface) decreaseGauge(a string, b []string) error {
	return nil
}

func (ei EmptyInterface) observeHistogram(a float64, b string, c []string) error {
	return nil
}

// NoOpsInstrumentation, returns an *EmptyInterface that can be used to not
// perform any instrumentation operations.
func NoOpsInstrumentation() *EmptyInterface {
	return new(EmptyInterface)
}
