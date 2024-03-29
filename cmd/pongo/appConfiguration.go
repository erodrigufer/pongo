package main

import (
	"fmt"
	"io"
	"log"
	"os"

	dock "github.com/erodrigufer/pongo/internal/docker/build"

	"github.com/docker/docker/client"
	semver "github.com/erodrigufer/go-semver"
	"github.com/erodrigufer/pongo/internal/pongo"
	dyntemplate "github.com/erodrigufer/pongo/internal/pongo/templates"
	prometheus "github.com/erodrigufer/pongo/internal/prometheus"
)

// pathEntryImage, path where the Dockerfile and further files needed to build
// the entrypoint container can be found.
const pathEntryImage = "/var/local/pongo/image"

// setupApplication, configures the info and error loggers of the application
// type and it initializes the client that communicates with the Docker daemon.
// It configure all needed general parameters for the application, e.g. the
// public port to which clients will connect.
// Parameters: port, the port to which the clients will connect through SSH.
func (app *application) setupApplication(configValues pongo.UserConfiguration) error {
	// Fetch configValues
	app.configurations = configValues

	// Create a logger for INFO messages, the prefix "INFO" and a tab will be
	// displayed before each log message. The flags Ldate and Ltime provide the
	// local date and time.
	app.infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	// Create an ERROR messages logger, additionally use the Lshortfile flag to
	// display the file's name and line number for the error.
	app.errorLog = log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Create a DEBUG messages logger if the -debugMode flag was set, otherwise
	// discard all logs. Additionally use the Lshortfile flag to
	// display the file's name and line number for the debug message.
	if app.configurations.DebugMode {
		app.debugLog = log.New(os.Stdout, "DEBUG\t", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		app.debugLog = log.New(io.Discard, "DEBUG\t", log.Ldate|log.Ltime|log.Lshortfile)
	}

	// Print daemon initialization log, including build revision (if possible).
	app.infoLog.Print("----------------------------------------------------")
	app.infoLog.Print("pongo -CTF session manager daemon- is initializing.")
	buildRev, err := semver.GetRevision()
	if err == nil {
		app.buildRev = buildRev
		app.infoLog.Printf("pongo revision: %s", buildRev)
	}

	// Initialize Docker daemon client.
	app.client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("error while initializing a Docker client: %v", err)
	}

	// The directory in which all the SSH Piper persistent data will be stored.
	// This directory will be mounted as a volume into the SSH Piper container.
	app.sshPiperFileSystem = "/tmp/sshpiper"
	// Docker image used to create SSH Piper container.
	app.images.sshPiperImage = "sshpiperd"
	// Docker image used to create the entrypoint container.
	app.images.entrypointImage = "entrypoint"

	// Initialize the HTML templates cache.
	app.templateCache, err = dyntemplate.NewTemplateCache("/var/local/pongo/html/")
	if err != nil {
		return fmt.Errorf("error while creating HTML templates cache: %v", err)
	}

	// Start data structures required for session manager daemons (smd).
	app.initializeSessionManager()

	app.outboundIP, err = getOutboundIP()
	if err != nil {
		app.errorLog.Printf("error while retrieving outbound IP of host machine: %v", err)
	}

	// Create a new system health monitor using its constructor.
	if app.appState.monitorConfigErr = app.setupMonitor(); app.appState.monitorConfigErr != nil {
		app.errorLog.Printf("error while configuring the health monitor: %v", err)
	}

	if !app.configurations.NoInstrumentation {
		// Expose the Prometheus metrics.
		app.instrumentation, err = prometheus.ExposeMetrics(app.infoLog, app.errorLog, "localhost:9999")
		if err != nil {
			app.errorLog.Printf("error while starting the Prometheus instrumentation: %v", err)
		}
	} else {
		// Do not expose the Prometheus metrics.
		app.infoLog.Print("prometheus: Running without exposing instrumentation metrics.")
		// NoOpsInstrumentation() will fulfill the instrumentation interface,
		// but it will not require any system resources when it is used in the
		// different instrumentation methods throughout the codebase. Moreover,
		// the codebase does not have to implement any kind of logic on the
		// different places where it wants to perform instrumentation, the
		// interface decides if it needs to collect data or not.
		app.instrumentation = prometheus.NoOpsInstrumentation()
	}

	app.infoLog.Print("Building Docker image for upstream container, this step can take up to several minutes.")
	// Build the Docker image used for the upstream container.
	if err := dock.ImageBuild(app.client, pathEntryImage, app.images.entrypointImage); err != nil {
		return fmt.Errorf("error building Docker image for upstream container: %v", err)
	}

	return nil

}
