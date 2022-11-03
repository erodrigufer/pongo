package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/docker/docker/client"
	"github.com/erodrigufer/CTForchestrator/internal/ctfsmd"
	prometheus "github.com/erodrigufer/CTForchestrator/internal/prometheus"
	semver "github.com/erodrigufer/go-semver"
)

// charsetUsername, valid character-set for generating random usernames.
// IMPORTANT: Only lower-case alphabetic characters can be used as valid
// usernames in the SSH system we are using.
const charsetUsername = "abcdefghijklmnopqrstuvwxyz"

// charsetPassword, valid character-set for generating random passwords.
// When special characters are used in the passwords (like $, ! ...) it is
// impossible to connect to the container with SSH. Avoid special characters!
const charsetPassword = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// setupApplication, configures the info and error loggers of the application
// type and it initializes the client that communicates with the Docker daemon.
// It configure all needed general parameters for the application, e.g. the
// public port to which clients will connect.
// Parameters: port, the port to which the clients will connect through SSH.
func (app *application) setupApplication(configValues ctfsmd.UserConfiguration) error {
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
	app.infoLog.Print("ctfsmd -CTF session manager daemon- is initializing.")
	buildRev, err := semver.GetRevision()
	if err == nil {
		app.buildRev = buildRev
		app.infoLog.Printf("ctfsmd revision: %s", buildRev)
	}

	// Initialize Docker daemon client.
	app.client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	// The directory in which all the SSH Piper persistent data will be stored.
	// This directory will be mounted as a volume into the SSH Piper container.
	app.sshPiperFileSystem = "/tmp/sshpiper"
	// Docker image used to create SSH Piper container.
	app.images.sshPiperImage = "sshpiperd"
	// Docker image used to create the entrypoint container.
	app.images.entrypointImage = "entrypoint"

	// Create a new and unique random seed, which can be used throughout the
	// application each time that a random string has to be generated.
	// For more information, check:
	// https://www.calhoun.io/creating-random-strings-in-go/
	app.seededRand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	// Initialize the HTML templates cache.
	app.templateCache, err = newTemplateCache("./ui/html/")
	if err != nil {
		return err
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

	return nil

}

// stopSession, stops all the containers that form a session and removes all
// the session-specific networks created for that session.
func (app *application) stopSession(ss session) error {
	ctx := context.Background()

	// TODO: use a positive value, to have a hard timeout, check documentation.
	timeout := time.Duration(-1)
	// Remove all the session-specific containers, which are stored in a slice
	// of strings with the containers' IDs.
	for _, containerID := range ss.containersIDs {
		if err := app.client.ContainerStop(ctx, containerID, &timeout); err != nil {
			return err
		}

	}

	// Remove the session-specific networks from the host as well.
	// Remove the internal session's network.
	if err := app.client.NetworkRemove(ctx, ss.networkID); err != nil {
		return err
	}
	// Remove the session's private network.
	if err := app.client.NetworkRemove(ctx, ss.privateNetworkID); err != nil {
		return err
	}

	return nil
}

// createSession, creates a new session, therefore initializing all required
// containers, and connecting them to the required networks.
// If no error is returned, the session was correctly created and a struct of
// type session is returned.
func (app *application) createSession() (session, error) {
	// Create a session and populate its fields.
	usernameLength := 15
	passwordLength := 15
	var newSession session
	// Create random username and password.
	newSession.username = app.createRandomUsername(usernameLength)
	newSession.password = app.createRandomPassword(passwordLength)
	// Time of creation might be used to delete very old sessions in the future.
	newSession.timeCreated = time.Now()

	// Give the session a name, in this case the first 6 characters of the
	// randomly generated username.
	newSession.name = newSession.username[:6]

	// Create an upstream container for the entrypoint and connect it to the
	// reverse proxy network.
	entrypointID, err := app.createUpstreamContainer(newSession.name, app.images.entrypointImage, app.networkIDreverseProxy)
	if err != nil {
		return newSession, err
	}
	// Append the container ID of the entrypoint container to the slice with
	// all the container IDs for this session.
	newSession.containersIDs = append(newSession.containersIDs, entrypointID)

	// Create a new user account with the randomly generated username and
	// password in the new upstream-container.
	if err := app.createUser(entrypointID, newSession.username, newSession.password); err != nil {
		return newSession, err
	}
	app.debugLog.Printf("Created new user (%s) in container (%s).\n", newSession.username, newSession.name)

	// Add a container as an upstream-container to the reverse proxy.
	if err := app.addUpstream(newSession.name, newSession.username, newSession.username); err != nil {
		return newSession, err
	}
	app.debugLog.Printf("Added container %s as an upstream-container.\n", newSession.name)

	app.infoLog.Printf("New session created with username: %s. Password: %s. ID: %s.", newSession.username, newSession.password, newSession.name)

	return newSession, nil

}

// createSessionNetwork, creates the network that will be used by all
// containers inside the same session. Parameter: networkName, name for the new
// network. Output: networkID (string) and error.
func (app *application) createSessionNetwork(networkName string) (string, error) {
	networkID, err := app.createNetwork(networkName)
	if err != nil {
		return "", err
	}
	app.debugLog.Printf("Created session network with ID: %s.\n", networkID[:10])
	return networkID, nil
}

// createRandomString, returns a random string.
// Parameters: length is the number of characters of the string that should be
// returned. charset, is the valid character set from which to generate a
// random string.
func (app *application) createRandomString(length int, charset string) string {
	// Make a slice of length length, in which to store random characters.
	b := make([]byte, length)
	for i := range b {
		// Pick a single character from the character-set through indexing the
		// string of the character-set. The index is a random number between 0
		// and the length of the character-set minue 1.
		b[i] = charset[app.seededRand.Intn(len(charset))]
	}

	return string(b)
}

// createRandomPassword, returns a random password of given character length.
func (app *application) createRandomPassword(length int) string {
	return app.createRandomString(length, charsetPassword)
}

//createRandomUsername, returns a random password of the given character length.
func (app *application) createRandomUsername(length int) string {
	return app.createRandomString(length, charsetUsername)
}

// createUser, creates a new user with a given password in an upstream
// container by running a command inside the upstream container which creates a
// new user (docker exec) . Parameters: containerID of container being
// configured, username to create with a given password.
func (app *application) createUser(containerID, username, password string) error {
	// Command to create user.
	cmdCreateUser := []string{
		"useradd",
		"--create-home", // Create a home directory for the new user.
		"--user-group",  // Add user to group with its own name.
		"--shell",       // Use bash as the default user shell.
		"/bin/bash",
		username, // Specify the new user's username.
	}

	ctx := context.Background()
	if err := app.runExec(ctx, containerID, cmdCreateUser); err != nil {
		return err
	}

	// Format input that will be piped into chpasswd command.
	accountDetails := fmt.Sprintf("%s:%s", username, password)
	wholeCmd := fmt.Sprintf("echo %s | chpasswd", accountDetails)

	// Command to change password from a given user.
	// 'echo $user:$password | chpasswd'.
	// Without the 'bash -c' it was not working, probably because of the
	// piping.
	cmdChangePwd := []string{
		"bash",
		"-c",
		wholeCmd,
	}

	ctx = context.Background()
	if err := app.runExec(ctx, containerID, cmdChangePwd); err != nil {
		return err
	}

	return nil
}
