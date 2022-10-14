// prometheus
package prometheus

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// declareHTTPServer, declares and configures an HTTP server.
func declareHTTPServer(errorLog *log.Logger, addr string) *http.Server {
	// Use the pat.New() function to initialize a new servemux.
	mux := pat.New()

	// Prometheus endpoint with the metrics captured through instrumentation.
	mux.Get("/metrics", promhttp.Handler())

	// chain of middlewares being executed before the mux, e.g.
	// a defer function to recover from a panic from within a client's connec.
	// (the go routine for the client)
	// secureHeaders executes its instructions and then returns the next http
	// Handler in the chain of events, in this case the mux.
	// TODO: consider adding the recoverPanic and secureHeaders middleware
	// return app.recoverPanic(secureHeaders(mux))

	// Initialize a new http.Server struct.
	srv := &http.Server{
		// Address where server listens.
		Addr: addr,
		// Logger for errors.
		ErrorLog: errorLog,
		// Handler that receives the client after accept().
		Handler: mux,

		// Time after which inactive keep-alive connections will be closed.
		// IdleTimeout: time.Minute,
		// // Max. time to read the header and body of a request in the server.
		// ReadTimeout: 5 * time.Second,
		// // Close connection if data is still being written after this time since
		// // accepting the connection.
		// WriteTimeout: 15 * time.Second,
	}

	return srv
}

// ExposeMetrics, expose the metrics with an HTTP server.
func ExposeMetrics(infoLog, errorLog *log.Logger, addr string) (*Instrumentation, error) {
	ins, err := startInstrumentation()
	if err != nil {
		err = fmt.Errorf("error while starting the Prometheus instrumentation: %w", err)
		return ins, err
	}

	srv := declareHTTPServer(errorLog, addr)

	infoLog.Printf("prometheus: Starting Prometheus web HTTP server at %s.", addr)
	infoLog.Printf("prometheus: Metrics are served at %s/metrics.", addr)
	infoLog.Print("prometheus: Always verify that your FIREWALL permits traffic to the metrics HTTP webpage.")

	go func() {
		err := srv.ListenAndServe()
		// Error returned when server is closed, not actually an error, log to
		// infoLog.
		if err == http.ErrServerClosed {
			infoLog.Print(err)
			// An actual error happened, log to errorLog.
		} else {
			errorLog.Print(err)
		}
	}()

	return ins, nil
}
