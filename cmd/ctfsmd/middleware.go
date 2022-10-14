package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	prometheus "github.com/erodrigufer/CTForchestrator/internal/prometheus"
	"github.com/urfave/negroni"
)

// secureHeaders, sets two header values for all incoming server requests
// These two header values should instruct the client's web browser to implement
// some additional security measures to protect against XSS and clickjacking.
func secureHeaders(next http.Handler) http.Handler {
	// Explanation:
	// http.HandlerFunc is a type that works as an adapter to allow a function f
	// to be returned as an http.Handler (which is an interface, that requires
	// the ServeHTTP() method), since the method ServeHTTP of the type
	// HandlerFunc simply calls the HandlerFunc which has a signature:
	// func(ResponseWriter, *Request)
	// In the above code we are type casting an anonymous function into a
	// HandlerFunc, so that when this function is called with ServeHTTP, it will
	// execute, and in the last step call the ServeHTTP method of the next http
	// handler in the chain of handlers
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

// logRequest, logs every client's request.
// Log IP address of client, protocol used, HTTP method and requested URL.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// r.RemoteAddr is the IP address of the client doing the request.
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL)

		next.ServeHTTP(w, r)
	})
}

// recoverPanic, sends an Internal Server Error message code to a client, when
// the server has to close an HTTP connection with a client due to a panic
// inside the goroutine handling the client.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event
		// of a panic as Go unwinds the stack).
		defer func() {
			// Use the built-in recover function to check if there has been a
			// panic or not.
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500
				// Internal Server Error response.
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// prometheusMiddleware, captures all HTTP responses status codes and resource
// paths and increments a Prometheus counter.
func (app *application) prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startDuration := time.Now()
		// negroni must we used as a wrapper around the http.ResponseWriter,
		// because otherwise there is no way of having access to the status code
		// of a response that has been sent to the client.
		// Check:
		// * https://stackoverflow.com/questions/53272536/how-do-i-get-response-statuscode-in-golang-middleware
		// * https://github.com/urfave/negroni/blob/master/response_writer.go
		extW := negroni.NewResponseWriter(w)
		next.ServeHTTP(extW, r)

		// Increment counter for HTTP requests that have received a response.
		// This method is after the ServeHTTP() method, so that it can gather
		// the status code and further information, after the requests have been
		// processed by other middlewares or endpoints.
		// Transform status codes into strings, and use them as labels.
		statusCodeLabel := strconv.Itoa(extW.Status())
		// URL of resource being requested. Remove any trailing '?' Otherwise,
		// some URL have a trailing '?' when the user does a POST request.
		resourceLabel := strings.Trim(r.URL.String(), "?")
		if err := prometheus.IncrementCounter(app.instrumentation, "http_requests_total", statusCodeLabel, resourceLabel); err != nil {
			app.errorLog.Printf("prometheus: unable to increment counter http_requests_total: %v", err)
		}
		endDuration := time.Since(startDuration).Seconds()
		if err := prometheus.ObserveHistogram(app.instrumentation, endDuration, "http_requests_duration_seconds", statusCodeLabel, resourceLabel); err != nil {
			app.errorLog.Printf("prometheus: unable to observe value for histogram http_requests_duration_seconds: %v", err)
		}

	})
}
