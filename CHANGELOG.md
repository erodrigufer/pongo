## v1.18.1
* Improve styling of logs. Remove newlines.
* Minor improvements on documentation of methods.
* Change `Fatalln()` calls for `Fatal()`.

## v1.18.0
* Add the `total_available_sessions` gauge in the instrumentation.
	* It is possible to monitor how the number of available sessions dwindles throughout time, which is especially useful in order to know if at any time the system is not able to produce any more sessions.
* Add name of author to footer.

## v1.17.0
* Fix mechanism that starts the monitor.Daemon() in charge of realizing health checks.
	* The daemon is now initialized after booting the system, when the number of available sessions is higher than a given number of sessions.
	* The previous solution of waiting for 40 seconds is now deprecated.

## v1.16.1
* Improve `css` styles. Header style is good. 
* Improve instructions text.
* Change frequency of health checks to 30 minutes.
* Remove health check that always failed (to test UI).
* Add health check of GET request for landing page.

## v1.16.0
* Create the `EmptyInterface` in the prometheus package, to allow the application to run without instrumentation by using the `-noInstrumentation` flag.
* Security fix: Run Prometheus metrics in localhost.
* Add configuration for both Prometheus and Grafana. 
* Minor improvements in the `css` styles.

## v1.15.2
* Implement histograms in the prometheus internal package. 
* Collect `http_requests_duration_seconds` metrics.

## v1.15.1
* Add all the required methods (`Increment` and `Decrement`) for handling Gauges.
* Implement the `active_sessions_total` gauge within `sessionManager.go`

## v1.15.0
* Improve the Prometheus library, in order to be able to decouple the instrumentation from the application if needed.
* Handle the instrumentation through an interface, in order to more easily decide if one wants to use instrumentation or not.
* Define the instrumentation on the file `internal/prometheus/application.go`

## v1.14.0
* Implement the first Prometheus instrumentation -> `http_requests_total` (COUNTER) with the `status_code` label.
* [FIX] If the monitor has not been fully initialized, and no health checks can be delivered to the clients, then a `500 Internal Server Error` is delivered, instead of a 200 OK HTTP response.
* Use `negroni` in order to be able to capture the status code of HTTP responses that have been written.

## v1.13.0
* Expose very rudimentary Prometheus metrics only about the Go environment at `localhost:9999/metrics`.

## v1.12.1
* Start refactoring `main.go` file to get a more clean structure.
* `monitord` is not initialized if an error is returned during the configuration of the system monitor parameters.

## v1.12.0
* Properly integrate the output of the health checks from `monitord` into the HTML templates at `/healthcheck`.

## v1.11.0
* Very big/substantial improvement and feature addition:
	* Automatically retrieve the host's outbound IP address.
	* Use the outbound IP address to setup the internal monitor.
	* Use the outbound IP address in the HTML template for a session.
* Start refactoring/cleaning the code of `main.go`.

## v1.10.0
* Minimum time between requests can now be configured with flags. 
	* This helps to properly calibrate the time between health checks for Monitord.
	* Monitord now also checks if a consecutive request to `/session` gets denied with a 429 response.

## v1.9.0
* Change name of filepath where daemon is installed to `ctfsmd` (this makes the names in the systemd logs more compact).

## v1.8.2
* Improve documentation on README on how to start/stop the `ctfsmd` daemon.

## v1.8.1
* Improve UI:
	* CSS Tweaks
	* Style of button.
* Pass _Lifetime of sessions_ dynamic data to HTML rendering engine.

## sessionManager_v1.8.0
* Monitord internal library to check the health of the service continuously.
	* **IMPORTANT**: The Monitord daemon is working, but it still needs improvements (see open TODOs with edge cases) and must be initialized in a more reliable way than using a time delay (see main.go).
* `--uninstall` option in `mainConfiguration.sh` script now also reverts the configuration of systemd, removes systemd config files for ctfsmd.
* Improve maintainability of css code (using variables) and fix minor issues in HTML/CSS UI.

## sessionManager_v1.7.0
* Control the daemon with systemd as a service unit.
* Automatically configure the daemon in systemd.

## sessionManager_v1.6.1
* Automatic installation without prompts for Docker and Golang.

## sessionManager_v1.6.0
* Add shell scripts to automatically deploy the daemon in a server.

## sessionManager_v1.5.0
* Render HTML pages for all routes.

## sessionManager_v1.4.2
* Add flag to handle frequency of lifetime check by srd.

## sessionManager_v1.4.1
* SSH Piper container gets connected only to SSH piper reverse network. It does not connect to 'bridge' network any more.

## sessionManager_v1.4.0
* The bug with the **bridge** network being in all initialized containers has been solved.
* The flag `lifetimeSess` setups the lifetime in minutes after which a session expires and is removed by srd.

## sessionManager_v1.3.1
* Different log message when client does not receive a session from smd.
	* The previous message was confusing, because it started with the word 'error'.

## sessionManager_v1.3.0
* CLI options to change number of available and active sessions, among others.
* `daemon.json` config file to expand the available IP range for Docker internal networking.

## sessionManager_v1.2.1
* Add 2 new flags to change the default SSH Port and HTTP webpage port.

## sessionManager_v1.2.0
* The server limits the amount of requests that a single IP can do in a given amount of time.
* This block affects all devices sharing the same IP behind a router.

## sessionManager_v1.1.0
* Clients now timeout if the session manager does not respond back in time.
* Clients receive an Internal Server Error HTTP response if a timeout takes place.

## sessionManager_v1.0.1
* Add documentation to configure `ufw` in Ubuntu to host the session manager.

## sessionManager_v1.0.0
* First stable official release. 
* Program closes appropriately at shutdown.

## sessionManagement_v0.12.1
* Close all sessions, except the SSH reverse proxy and its network, properly at shutdown.

## sessionManagement_v0.12.0
* Close all daemons correctly, when the program receives `SIGTERM` or `SIGINT`.

## sessionManagement_v0.11.0
* Create `-debugMode` flag to activate debug mode, where logging is more extensive.
* Otherwise logging is more scarce now in non-debug mode.

## sessionManagement_v0.10.0
* srd: session removal daemon, is periodically checking for active sessions that have expired.

## sessionManagement_v0.9.0
* scd: session creation daemon, constantly creates sessions to keep the amount of available sessions always constant.

## sessionManagement_v0.8.0
* The HTTP server properly sends back session information under the path `/session`
* The session manager now runs as a daemon, which is perpetually managing the requests for new sessions.
* The session manager still does not have the ability to create and destroy sessions.

## sessionManagement_v0.7.0
* Run an HTTP Server, which in the future will serve the session requests from the clients.
	* HTTP Server is still very rudimentary, can only serve /ping route, as a healthcheck
	* It works properly in parallel with the whole Docker reverse proxy infrastructure.

## sessionManagement_v0.6.0
* Start creating methods for an actual session manager, which can create and manage multiple session.
* Integrate session manager into app's data model.
* Test capability of session manager to store many sessions.

## sessionManagement_v0.5.0
* Create a method in helpers.go to eliminate all the session-specific objects (containers and networks).

## sessionManagement_v0.4.0
* Add a container for _priv_server_ to every session created.
* Add a private network for _priv_server_ and _linux_server_ within the session.

## sessionManagement_v0.3.0
* Add a container for nginx _simple_server_ to every session created.
* Add more fields to the _session_ type.
* Handle Dockerfiles for this application in a folder within the repo.

## sessionManagement_v0.2.0
* Add a container for _linux_server_ to every session created.
* Create a session network and add all session containers to this network.

## sessionManagement_v0.1.0
* Handle creation of session with a method.
