package prometheus

// Define the application-specific counters.
var applicationCounters = []struct {
	name        string
	description string
	labels      []string
}{
	{
		name:        "http_requests_total",
		description: "Total amount of HTTP requests",
		labels:      []string{"status_code", "resource"},
	},
}

// Define the application-specific gauges.
var applicationGauges = []struct {
	name        string
	description string
	labels      []string
}{
	{
		name:        "active_sessions_total",
		description: "Total amount of active sessions.",
		labels:      []string{},
	},
	{
		name:        "available_sessions_total",
		description: "Total amount of available sessions.",
		labels:      []string{},
	},
}

// Define the application-specific histograms.
var applicationHistograms = []struct {
	name        string
	description string
	labels      []string
	buckets     []float64
}{
	{
		name:        "http_requests_duration_seconds",
		description: "Distribution of the duration of HTTP requests in seconds",
		labels:      []string{"status_code", "resource"},
		buckets:     []float64{0.000025, 0.00005, 0.000075, 0.0001, 0.0003, 0.0005, 0.00075, 0.001, 0.003, 0.007, 0.1, 1},
	},
}
