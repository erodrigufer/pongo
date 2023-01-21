package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	monitor "github.com/erodrigufer/pongo/internal/APIMonitor"
)

// startMonitord, starts all the daemons required to monitor the health of the
// application.
func (app *application) startMonitord(ctx context.Context) {
	// Always check first if there was an error while configuring the
	// system monitor. If an error happened while configuring the monitor, then
	// the daemons that comprise the monitord system are not started.
	go func() {
		if app.appState.monitorConfigErr == nil {
			// Check in a for-loop, until the amount of available sessions is
			// higher than x, this check would only take place once, when the
			// system is booting. Otherwise return, if the monitor.Daemon()
			// returns, so that the for-loop does not keep going indefinitely.
			for {
				if len(app.sm.availableSessions) > 2 {
					app.monitor.Daemon(ctx)
					return
				}
			}
		}
	}()
	go func() {
		if app.appState.monitorConfigErr == nil {
			// TODO: maybe add a ctx for CurrentHealth() as well.
			app.monitor.CurrentHealth()
		}
	}()

}

// getOutboundIP, returns the preferred outbound IP of the host machine.
// This function is untested in a machine with multiple outbound interfaces.
func getOutboundIP() (string, error) {
	// Use UDP, so that no handshake is necessary, as a matter of fact, the
	// remote machine being contacted does not even have to exist for this
	// function to work.
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// LocalAddr returns the local network address, if known.
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	// TODO: add checks for correct/valid IP

	return localAddr.IP.String(), nil
}

// setupMonitor, initialize a monitor. Add its loggers and period for sampling
// health checks.
func (app *application) setupMonitor() error {
	// The frequency with which the monitor will perform the health checks is
	// a multiple bigger than 1 of the minimum time between requests from the
	// same client, so that the monitor does not get a 429 Too Many Requests
	// response unnecessarily.
	app.monitor = monitor.NewMonitor(app.infoLog, app.errorLog, time.Duration(app.configurations.TimeBetweenRequests)*6*time.Minute)

	if app.outboundIP == "" {
		return fmt.Errorf("monitor could not be configured due to invalid outboundIP")
	}
	serverSessionURL := fmt.Sprintf("http://%s%s/session", app.outboundIP, app.configurations.HTTPAddr)
	landingPage := fmt.Sprintf("http://%s%s/", app.outboundIP, app.configurations.HTTPAddr)

	// Create the health checks that will be performed.
	// GET a session successfully.
	sessionResource := monitor.APIResource{
		Method: "GET",
		URL:    serverSessionURL,
		// Empty body.
		ReqBody:            strings.NewReader(""),
		ExpectedStatusCode: 200,
		ReqTimeout:         60 * time.Second,
	}
	// GET session request denied due to 'Too many requests' (429).
	sessionResourceBlocked := monitor.APIResource{
		Method: "GET",
		URL:    serverSessionURL,
		// Empty body.
		ReqBody:            strings.NewReader(""),
		ExpectedStatusCode: 429, // Too many requests.
		ReqTimeout:         60 * time.Second,
	}

	// GET the landing page successfully.
	landingPageResource := monitor.APIResource{
		Method: "GET",
		URL:    landingPage,
		// Empty body.
		ReqBody:            strings.NewReader(""),
		ExpectedStatusCode: 200,
		ReqTimeout:         60 * time.Second,
	}

	h1 := monitor.HealthCheck{
		Name:        "GET session",
		Description: "Check if it is possible to GET a new session. Afterwards, check if a subsequent GET request for a session gets denied with a 429 response.",
		Check: func() error {
			if err := app.monitor.PingHTTPService(sessionResource); err != nil {
				return err
			}
			// Sleep and try the same resource. It should deliver a 429
			// (Too Many Requests) response.
			time.Sleep(3 * time.Second)
			if err := app.monitor.PingHTTPService(sessionResourceBlocked); err != nil {
				return err
			}

			return nil
		},
	}
	// Add health check to slice with all health checks.
	app.monitor.HealthChecks = append(app.monitor.HealthChecks, h1)

	h2 := monitor.HealthCheck{
		Name:        "GET landing page '/'",
		Description: "Check if it is possible to GET the landing page of the server.",
		Check: func() error {
			if err := app.monitor.PingHTTPService(landingPageResource); err != nil {
				return err
			}

			return nil
		},
	}
	// Add health check to slice with all health checks.
	app.monitor.HealthChecks = append(app.monitor.HealthChecks, h2)

	return nil
}
