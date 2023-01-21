package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/docker/docker/client"
	monitor "github.com/erodrigufer/pongo/internal/APIMonitor"
	"github.com/erodrigufer/pongo/internal/pongo"
	prometheus "github.com/erodrigufer/pongo/internal/prometheus"
)

// application, type used for dependency injection and to avoid using globals.
type application struct {
	// errorLog, error log handler.
	errorLog *log.Logger
	// infoLog, info log handler.
	infoLog *log.Logger
	// debugLog, debug log handler. Activated only for debugging. Logging
	// activity will therefore be more frequent.
	debugLog *log.Logger
	// buildRev, hash of last commit from which executable was built.
	buildRev string
	// srv, is the HTTP server that handles clients' requests for sessions.
	srv *http.Server
	// client, is the client that communicates with the Docker daemon.
	client *client.Client
	// sm, handles the session management.
	sm *sessionManager
	// wg, is a WaitGroup that waits for all daemons to return.
	wg sync.WaitGroup
	// configurations, are the user configurations handled through flags.
	configurations pongo.UserConfiguration
	// networkIDreverseProxy, the networkID of the network that the SSH reverse
	// proxy uses to communicate with the upstream containers.
	networkIDreverseProxy string
	// SSHPiperFileSystem, is the directory in which the persistent data
	// required for the SSH Piper container will be stored in the host.
	sshPiperFileSystem string
	// sshPiperContainerID, is the ID of the container running as the SSH
	// reverse proxy.
	sshPiperContainerID string
	// images, Docker images used to create all the containers used in this
	// application.
	images images
	// templateCache, cache map with html templates.
	templateCache map[string]*template.Template
	// monitor, monitors the health of the application with periodic checks.
	monitor *monitor.Monitord
	// outboundIP, the outbound IP used by the host machine.
	outboundIP string
	// appState, defines the states of different subsystems of the app.
	appState appSubsystState
	// instrumentation, defines the interface used to interact with the
	// Prometheus instrumentation.
	instrumentation prometheus.InstrumentationAPI
}

// appSubsystState, stores the state of different subsystems that make up the
// application, e.g. if there was an error while initializing a subsystem.
type appSubsystState struct {
	// monitorConfigErr, defines if there was an error while configuring the
	// system monitor (monitord).
	monitorConfigErr error
}

// images, contains the names of all the Docker images used in this
// application.
type images struct {
	// sshPiperImage, the Docker image used to create the SSH Piper reverse
	// proxy container (it might be an image hosted in a remote Docker repo).
	sshPiperImage string
	// entrypointImage, the Docker image used as the upstream-container from
	// SSH Piper in which the user initiates his session.
	entrypointImage string
}

// session, stores all the relevant information for a unique session, its
// containersIDs, networks, SSH username, SSH password, etc.
type session struct {
	// name, unique identifier/name for a session.
	name string
	// username, username used by client to log into session with SSH.
	username string
	// password, password used by client to log into session with SSH.
	password string
	// containersIDs, is a slice with the containers' IDs of all the
	// containers that are part of the session.
	containersIDs []string
	// timeCreated, time at which session was created.
	timeCreated time.Time
	// timeActivated, time at which session was activated, i.e. the session was
	// given to a client for use after a client's request.
	timeActivated time.Time
}

// sessionManager, manages the creation and allocation of sessions for the
// clients.
type sessionManager struct {
	// availableSessions, is a channel in which all available sessions are
	// stored, so that the sessionManager can read a unique session out of the
	// channel for every client requesting a session.
	availableSessions chan session
	// requestSession, is the channel to which clients can send a clientReq with
	// a channel from which they will eventually get a reply with their username
	// and password for a newly created session.
	requestSession chan clientReq
	// activeSessions, is a channel in which all currently active sessions
	// (sessions sent to a client) are stored, so that the srm (session removal
	// daemon) can later check its creation time to eventually remove them.
	activeSessions chan session
}

// clientReq, data structure sent to smd by each client that requests a new
// sessions.
type clientReq struct {
	// respCh, is the channel where the client expects to receive an answer back
	// from the session manager daemon (smd) with a new session.
	respCh chan smResponse
	// reqInfo, extra information from the client for smd with each request.
	reqInfo reqInfo
}

// reqInfo, contains extra information that the client shares with smd with each
// request, for instance, the IP address of the client performing the request.
type reqInfo struct {
	// clientAddr, IP address of client sending request.
	clientAddr string
}

// ERR_LAST_REQ, error code used to identify an error received when a user tries
// to request a new session too soon after receiving a session.
var ERR_LAST_REQ error = fmt.Errorf("Not enough time has passed since last request.")

// smResponse, is a wrapper for the response that a client receives from the
// session manager (sm), in order to send both a session and an error back.
// If err == nil, then the new session was sent in the field session.
type smResponse struct {
	// session, session delivered to a client by the sm.
	session session
	// errors, if any errors happened during the delivery of a session to a
	// client, like no more available sessions, errors != nil.
	errors error
	// errCode, a code to identify the error.
	// errCode int
}
