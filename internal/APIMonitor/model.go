package APIMonitor

import (
	"io"
	"log"
	"net/http"
	"time"
)

// Monitord, a monitor daemon that monitors an API.
type Monitord struct {
	// client, HTTP client to perform HTTP requests to monitor.
	client *http.Client
	// healthChecks, this slice contains all the health checks that the daemon
	// should periodically perform.
	// healthchecks is NOT concurrent-safe, only populate the slice when
	// invoking the constructor for Monitord.
	HealthChecks []HealthCheck
	// freq, frequency with which the monitor daemon will monitor the API
	// resources.
	freq time.Duration
	// infoLog, a logger to print info messages.
	infoLog *log.Logger
	// errorLog, a logger to print error messages.
	errorLog *log.Logger
	// Requests, a channel with Requests for the current health status coming
	// from the clients.
	Requests chan ClientReq

	// hcResults, is a channel for a slice with all performed health checks
	hcResults chan []HealthCheckResult
}

// HealthCheck, single health check performed by Monitord.
type HealthCheck struct {
	// Name, name of health check.
	Name string
	// Description, human-readable description of a health check.
	Description string
	// Check, is the function executed by Monitord when performing a specific
	// health check.
	Check func() error
}

// HealthCheckResult, struct defines all the information necessarry to
// interpret the results from the health checks periodically being performed by
// the Monitord daemon.
type HealthCheckResult struct {
	// Name, of a particular health check.
	Name string
	// Description, human-readable description of a health check.
	Description string
	// Pass, did the health check passed successfully?
	Pass bool
	// Diagnostics, why did the health check failed?
	Diagnostics error
	// Timestamp, when was the health check performed?
	Timestamp time.Time
}

// APIResource, describes a single resource in an HTTP API.
type APIResource struct {
	// Method, to request API resource.
	Method string
	// URL, where resource resides.
	URL string
	// ReqBody, request body to send to resource.
	ReqBody io.Reader
	// ExpectedStatusCode, HTTP status code expected as response from resource.
	ExpectedStatusCode int
	// ReqTimeout, time after which an HTTP requests is cancelled due to a
	// timeout.
	ReqTimeout time.Duration
}

// ClientReq, request sent by a client to the Monitord daemon to get the current
// health status.
type ClientReq struct {
	// RespCh, a channel provided by the client to Monitord to receive a
	// response back from Monitord with all the health checks.
	RespCh chan MonitorResp
}

// MonitorResp, the response sent back to the client with the requested health
// checks.
type MonitorResp struct {
	// HealthChecksResults, slice with all the HealthCheckResults for a particular
	// timestamp.
	HealthChecksResults []HealthCheckResult
	// Errors, if any Errors happened while getting the health checks, then
	// Errors != nil, the client should consider the data received as erroneous.
	Errors error
}
