// APIMonitor implements functions to periodically monitor the health of
// HTTP services and back-end services.
package APIMonitor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// NewMonitor, constructor for a Monitord daemon.
func NewMonitor(infoLog, errorLog *log.Logger, freq time.Duration) *Monitord {
	mon := new(Monitord)
	// Create a new http.Client, exclusively for the Monitord daemon.
	mon.client = new(http.Client)
	mon.infoLog = infoLog
	mon.errorLog = errorLog
	mon.freq = freq
	// Unbuffered channel to receive unlimited client requests.
	mon.Requests = make(chan ClientReq)
	// Unbuffered channel to receive unlimited health check results.
	mon.hcResults = make(chan []HealthCheckResult)
	mon.HealthChecks = make([]HealthCheck, 0, 10)
	return mon
}

// PingHTTPService, returns an error if an API resource does not respond with
// with the expected response status code.
func (mon *Monitord) PingHTTPService(resource APIResource) error {
	// Configure a timeout for the client's HTTP request. If the request takes
	// more than this time duration, then it should be cancelled.
	ctx, cancel := context.WithTimeout(context.Background(), resource.ReqTimeout)
	// Canceling a context releases resources associated with it, so code should
	// call cancel as soon as the operations running in a context complete.
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, resource.Method, resource.URL, resource.ReqBody)
	if err != nil {
		return fmt.Errorf("unable to create a new %s HTTP request to %s with timeout context: %w", resource.Method, resource.URL, err)
	}

	res, err := mon.client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send HTTP %s request to %s: %w", resource.Method, resource.URL, err)
	}
	defer res.Body.Close()

	if res.StatusCode != resource.ExpectedStatusCode {
		err = fmt.Errorf("HTTP %s request to %s failed with status code %d (%s). Expected status code is %d", resource.Method, resource.URL, res.StatusCode, res.Status, resource.ExpectedStatusCode)
		return err
	}

	return nil

}

// Daemon, runs the Monitord daemon which will periodically perform all the
// health checks defined in its healthChecks[] slice and send the results from
// the health checks through a channel.
func (mon *Monitord) Daemon(ctx context.Context) {
	mon.infoLog.Printf("monitord: health monitor daemon started.")
	tz, tzErr := time.LoadLocation("Europe/Berlin")
	if tzErr != nil {
		mon.infoLog.Printf("monitord: error parsing time location: %v", tzErr)
	}
	for {
		results := make([]HealthCheckResult, 0, 10)
		var result HealthCheckResult

		for _, hc := range mon.HealthChecks {
			var timestamp time.Time
			// No errors while parsing time zone.
			if tzErr == nil {
				timestamp = time.Now().In(tz)
			}
			if tzErr != nil {
				timestamp = time.Now()
			}
			if err := hc.Check(); err != nil {
				mon.infoLog.Printf("monitord: [FAIL] %s health check failed: %v", hc.Name, err)
				// Failed health check!
				result = HealthCheckResult{
					Name:        hc.Name,
					Description: hc.Description,
					Pass:        false,
					Diagnostics: err,
					Timestamp:   timestamp,
				}
			} else {
				mon.infoLog.Printf("monitord: [OK] %s health check passed.", hc.Name)
				// Successful health check.
				result = HealthCheckResult{
					Name:        hc.Name,
					Description: hc.Description,
					Pass:        true,
					Diagnostics: nil,
					Timestamp:   timestamp,
				}
			}
			results = append(results, result)
		}
		// TODO: set a timeout here as well
		// Send results in a blocking way
		mon.hcResults <- results

		// Start a timer which will return with a certain frequency.
		timer := time.NewTimer(mon.freq)
		mon.infoLog.Printf("monitord: next API health check in %v.", mon.freq)

		// After the time of the timer expires, the current time will be sent
		// through its channel, discard the value. The read from the chan will
		// block until the timer expires.
		select {
		case <-ctx.Done():
			mon.infoLog.Printf("monitord: context cancelled. Shutting down.")
			// Close all possible idle connections of the HTTP client.
			mon.client.CloseIdleConnections()
			return
		case <-timer.C:
			// Timer is over, daemon will check again the resources.
			break
		}

	}

}

// CurrentHealth, sends a response with the most current health checks. This
// method is concurrent-safe and should be used by HTTP handlers to retrieve
// the most current health status of the application to render a web page.
func (mon Monitord) CurrentHealth() {
	var response MonitorResp
	response.Errors = fmt.Errorf("No health checks have been received yet.")
	for {
		select {
		// Receive requests from clients for a new health check.
		case req := <-mon.Requests:
			// TODO: make a non-blocking response
			req.RespCh <- response
		case checks := <-mon.hcResults:
			// TODO: check that the length of checks is not equal to 0,
			// otherwise there is an error
			response.HealthChecksResults = checks
			response.Errors = nil
		}
	}
}
