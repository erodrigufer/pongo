package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/bmizerany/pat"

	monitor "github.com/erodrigufer/CTForchestrator/internal/APIMonitor"
)

// declareHTTPServer, declares and configures an HTTP server.
func (app *application) declareHTTPServer() {
	mux := app.routes()
	// Initialize a new http.Server struct.
	// Use app.errorLog for errors instead of default option.
	app.srv = &http.Server{
		// Address where server listens.
		Addr: app.configurations.httpAddr,
		// Logger for errors.
		ErrorLog: app.errorLog,
		// Handler that receives the client after accept().
		Handler: mux,
		// Time after which inactive keep-alive connections will be closed.
		IdleTimeout: time.Minute,
		// Max. time to read the header and body of a request in the server.
		ReadTimeout: 5 * time.Second,
		// Close connection if data is still being written after this time since
		// accepting the connection.
		WriteTimeout: 15 * time.Second,
	}

}

// startHTTPServer, start the HTTP server used by the clients to get sessions
// with a GUI.
func (app *application) startHTTPServer() {
	// Configure an HTTP server for the app.
	app.declareHTTPServer()
	app.infoLog.Printf("main: Starting web HTTP server at %s.", app.configurations.httpAddr)
	app.infoLog.Print("main: Always verify that your FIREWALL permits traffic from the outside to the HTTP webpage.")

	go func() {
		err := app.srv.ListenAndServe()
		// Error returned when server is closed, not actually an error, log to
		// infoLog.
		if err == http.ErrServerClosed {
			app.infoLog.Print(err)
			// An actual error happened, log to errorLog.
		} else {
			app.errorLog.Print(err)
		}
	}()
}

// routes, creates a mux, routing paths and initializes multiple middlewares,
// before returning a servemux, a mux used by an http.Server (an http.Handler).
func (app *application) routes() http.Handler {

	// Use the pat.New() function to initialize a new servemux.
	mux := pat.New()

	mux.Get("/", http.HandlerFunc(app.index))

	// Create routing for healthcheck function to check uptime/status of server.
	mux.Get("/healthcheck", http.HandlerFunc(app.healthcheck))

	// Create routing to request a session.
	mux.Get("/session", http.HandlerFunc(app.sessionFrontend))

	// Create a handler/fileServer for all files in the static directory
	// Type Dir implements the interface required by FileServer and makes the
	// code portable by using the native file system (which could be different
	// for Windows and other Unix systems).
	// ./ui/static/ will be the root (like a chroot jail) of the fileServer
	// it will serve files relative to this path. Nonetheless, a security
	// concern is that symlinks that point outside the 'jail' can also be
	// followed (check documentation of type Dir).
	fileServer := http.FileServer(http.Dir("./ui/static/"))

	// Make the fileServer the handler for all URLs starting with '/static'.
	// http.StripPrefix, will create a new http.Handler that first strips the
	// prefix "/static" from the URL request, and passes the new request to
	// the fileServer.
	mux.Get("/static/", http.StripPrefix("/static", fileServer))

	// chain of middlewares being executed before the mux, e.g.
	// a defer function to recover from a panic from within a client's connec.
	// (the go routine for the client), a logger for all requests and then
	// secureHeaders executes its instructions and then returns the next http
	// Handler in the chain of events, in this case the mux.
	return app.recoverPanic(app.logRequest(app.prometheusMiddleware(secureHeaders(mux))))
}

// index, handler used to render the main landing page.
func (app *application) index(w http.ResponseWriter, r *http.Request) {
	dynamicData := &templateData{}
	// Render page.
	app.render(w, r, "main.page.tmpl", dynamicData)

}

// healthcheck, status check or uptime monitor of server.
func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	// Create a response channel through which this function will get a
	// response from the monitord daemon.
	respCh := make(chan monitor.MonitorResp)
	req := monitor.ClientReq{
		RespCh: respCh,
	}
	// Send request for the most current health checks.
	app.monitor.Requests <- req
	response := <-respCh
	if response.Errors != nil {
		app.serverError(w, fmt.Errorf("Error received from monitord, could not retrieve monitor information."))
		return
	}

	dynamicData := &templateData{}
	// Pass the health check results to the template's dynamic data.
	dynamicData.HealthCheckResults = response.HealthChecksResults
	// Render page.
	app.render(w, r, "healthcheck.page.tmpl", dynamicData)

}

// sessionFrontend, requests a new session from the session manager, and sends
// the username and password as a response back to the client.
func (app *application) sessionFrontend(w http.ResponseWriter, r *http.Request) {
	// IMPORTANT: The timeout time for sending and receiving a new session
	// should not be larger than the 'IdleTimeout' and 'WriteTimeout' declare
	// for the HTTP server (app.srv). Otherwise, the server closes the
	// current connection with the client, before we are able to send an
	// 'Internal Server Error' with app.serverError(). The client therefore does
	// not receive an error and retries to get a session.
	// TLDR: the timeout here should be smaller than the IdleTimeout and
	// WriteTimeout of the HTTP server.
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	ss, err := app.requestSession(ctx, r)
	if err != nil {
		// Check if the client tried to get a new session in an amount of time
		// shorter than the minimum permitted between requests.
		if errors.Is(err, ERR_LAST_REQ) {
			app.infoLog.Print(err)
			// Send a 429 Too many requests HTTP error.
			w.WriteHeader(429)
			app.render(w, r, "timeRequestError.page.tmpl", &templateData{})
			return
		}
		app.serverError(w, err)
		// An error occured, return from the handler to not send more data to
		// the client.
		return
	}
	app.infoLog.Printf("Session (%s) delivered to %s.", ss.name, r.RemoteAddr)

	dynamicData := &templateData{
		Username: ss.username,
		Password: ss.password,
	}
	app.render(w, r, "session.page.tmpl", dynamicData)
}

// serverError, sends an error message and stack trace to the error logger and
// then sends a generic 500 Internal Server Error response to the client.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	// The first parameter of Output equals the calldepth, which is the count
	// of the number of frames to skip when computing the file name
	// and line number. So basically, just go back on the stack trace to display
	// the name of function (file) which called the error logging helper
	// function.
	app.errorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError, sends a specific error status to the client.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// notFound, is a convenience wrapper for a 404 Not Found resource error.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// getIP, parses the IP from the string with the IP and port contained in
// *http.Request. The parameter unparsed is this same string from the request.
// Return the parsed IP as a string and an error.
func getIP(unparsed string) (string, error) {
	// Parse the IP, discard the port.
	parsedIP, _, err := net.SplitHostPort(unparsed)
	if err != nil {
		return "", fmt.Errorf("parsing IP with net.SplitHostPort() failed: %w", err)
	}

	return parsedIP, nil
}

// addDefaultData, default data is automatically added to the dynamic data every
// time a template is rendered. This dynamic data is then passed to app.render.
func (app *application) addDefaultData(td *templateData) *templateData {
	// If pointer is nil, create a new instance of templateData.
	if td == nil {
		td = &templateData{}
	}
	td.CurrentYear = time.Now().Year()

	// Lifetime of a session in minutes.
	td.LifetimeSess = app.configurations.lifetimeSess

	// SSH reverse proxy port used by the clients.
	td.Port = app.configurations.sshPort

	// Outbound IP used by host machine.
	if app.outboundIP != "" {
		td.OutboundIP = app.outboundIP
	} else {
		// If the string is empty, add a generic identifier.
		td.OutboundIP = "<IP>"
	}

	return td
}

// render, retrieves the appropriate template set from the cache based on the
// page name (like 'index.page.tmpl') used as input. If no entry exists in the
// cache with the provided name, call the serverError helper method. Also
// provide the dynamicData to the templates through a parameter.
func (app *application) render(w http.ResponseWriter, r *http.Request, name string, dynamicData *templateData) {
	// Check if the template exists in the template cache map.
	ts, ok := app.templateCache[name]
	// The object did not exist in the cache map.
	if !ok {
		app.serverError(w, fmt.Errorf("The HTML template %s does not exist", name))
		return
	}
	// Initialize a buffer to first execute template into buffer, if there is an
	// error, then the data will not be half-written to the client, but instead
	// will remain in the buffer, and an Internal Server Error will be sent to
	// the client. Something like this can happen, when there is an error in a
	// template, then the Execute() method will return an error.
	buf := new(bytes.Buffer)
	// Execute the template set, passing in any dynamic data.
	err := ts.Execute(buf, app.addDefaultData(dynamicData))
	if err != nil {
		app.serverError(w, err)
		return // Do not send the template back to the client.
	}
	// There was no error while executing/rendering the template, so send the
	// whole template back to the client.
	buf.WriteTo(w)

}
