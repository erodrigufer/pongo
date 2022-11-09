package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

// configurePiperFilesystem, properly configure the filesystem necessary for
// the SSH Piper container in the host system. This filesystem will be mounted
// as a persistent volume into the SSH Piper container afterwards.
func (app *application) configurePiperFilesystem() error {
	// Remove the existent directory in the host machine with the SSH Piper
	// persistent data, and all its subdirectories. If the path does not exist,
	// os.RemoveAll() returns nil.
	// WARNING: This step will permanently remove any existing data from
	// previous SSH Piper sessions.
	err := os.RemoveAll(app.sshPiperFileSystem)
	if err != nil {
		app.errorLog.Print(err)
		// Removing folder failed, but still try to create a new one, do not
		// return back to callee.
	}
	// Create the directory in the host, which is mounted in the SSH Piper
	// container as a volume, in order to store persistent information required
	// for SSH Piper to work. If, the directory already exists, then MkdirAll()
	// does nothing and returns nil.
	err = os.MkdirAll(app.sshPiperFileSystem, 0750)
	// The os.IsExist function returns a boolean indicating if the error
	// returned by a function or method of the os package was due to the
	// a-priori existance of a file or directory. In this case we are checking
	// that the error is not that the directory already existed.
	if err != nil && !os.IsExist(err) {
		app.errorLog.Fatal(err) // Fatal error, exit application.
	}

	return nil
}

// initializeReverseProxy, starts the reverse proxy SSH Piper and does all the
// required a-priori configuration, like creating a network, configuring the
// filesystem, etc.
func (app *application) initializeReverseProxy() error {
	var err error
	// Create the network for the reverse proxy (SSH Piper) and the upstream
	// containers.
	reverseProxyName := "reverseProxy" // Name of the network.
	app.networkIDreverseProxy, err = app.createNetwork(reverseProxyName)
	if err != nil {
		return err
	}
	app.debugLog.Printf("Created reverse proxy network with ID: %s\n", app.networkIDreverseProxy[:10])

	// Configure the filesystem required for the SSH Piper container.
	if err = app.configurePiperFilesystem(); err != nil {
		return err
	}

	// Create the SSH Piper reverse proxy container.
	namePiperContainer := "piperSSH"
	if err = app.createPiperContainer(namePiperContainer); err != nil {
		return err
	}

	return nil
}

// createPiperContainer, creates and runs the container that will act as the SSH
// reverse proxy. It connects the container to the network with which it
// interacts with all upstream-containers (app.networkIDreverseProxy).
// Parameter: name, the name that will be given to the container running SSH
// Piper.
func (app *application) createPiperContainer(name string) (err error) {
	// TODO: first make sure that the image necessary for SSH Piper is already
	// pulled and built on the system -> farmer1992/sshpiperd
	sshPiperProxy := new(containerModel)
	sshPiperProxy.name = name
	// Container's image name.
	sshPiperProxy.containerConfig.Image = app.images.sshPiperImage
	// Set the internal hostname of the container.
	sshPiperProxy.containerConfig.Hostname = sshPiperProxy.name
	// Equivalent to --rm flag in Docker CLI, automatically remove container
	// after stopping it.
	sshPiperProxy.hostConfig.AutoRemove = true
	// Expose port 2222/tcp of the SSH Piper container.
	sshPiperProxy.containerConfig.ExposedPorts = nat.PortSet{nat.Port("2222/tcp"): {}}
	// Bind host's port (defined through flags) with SSH Piper 2222 port,
	// this is the port that clients will use to access the upstream-containers
	// through the SSH reverse proxy.
	sshPiperProxy.hostConfig.PortBindings = nat.PortMap{
		nat.Port("2222/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: app.configurations.SSHPort}},
	}
	// Bind volume to SSH Piper container, for persistent data management.
	// Bind RSA SSH key of host, so that SSH Piper always works with the same
	// key, otherwise SSH Piper generates a new key each time, which makes the
	// client think that the keys have changed for the same IP, which is pretty
	// bad because the SSH client blocks the connection attempt, in order to
	// prevent a man-in-the-middle attack.
	sshPiperProxy.hostConfig.Binds = []string{"/tmp/sshpiper:/var/sshpiper", "/etc/ssh/ssh_host_rsa_key:/etc/ssh/ssh_host_rsa_key"}

	// Networking configurations for container, so that the SSH piper container
	// is automatically connected to the SSH reverse proxy network after
	// initialization, and does not get connected to the 'bridge' network.
	// Create a struct with the configuration of an endpoint for the new
	// container.
	endpointsConfig := new(network.EndpointSettings)
	// Add the networkID of the reverse proxy network to the endpoint.
	endpointsConfig.NetworkID = app.networkIDreverseProxy
	// EndpointsConfig is a map, because it can have the configuration
	// parameters of multiple endpoints at the same time in the map.
	endpointsMap := make(map[string]*network.EndpointSettings)
	endpointsMap[app.networkIDreverseProxy] = endpointsConfig
	sshPiperProxy.networkConfig.EndpointsConfig = endpointsMap

	// Initialize a pseudo-terminal in the container.
	// sshPiperProxy.containerConfig.Tty = true
	// Attach a Stdin to the container to also be able to interact with it.
	// sshPiperProxy.containerConfig.AttachStdin = true

	// Run the previously configured SSH Piper reverse proxy container.
	app.sshPiperContainerID, err = app.runContainer(sshPiperProxy)
	if err != nil {
		return err
	}
	app.infoLog.Printf("SSH piper reverse proxy container was created (ID: %s) on port %s.", app.sshPiperContainerID[:10], app.configurations.SSHPort)

	return nil
}

// stopReverseProxy, this method stops the SSH piper reverse proxy container and
// removes its network.
func (app *application) stopReverseProxy() error {
	ctx := context.Background()

	// Check documentation for client.ContainerStop:
	// "In case the container fails to stop gracefully within a time frame
	// specified by the timeout argument, it is forcefully terminated (killed).
	// timeout, err := time.ParseDuration("30s")
	// if err != nil {
	// 	return fmt.Errorf("error: parsing time for timeout for stopping container: %w", err)
	// }

	// TODO: research if this is correct.
	// https://vsupalov.com/docker-compose-stop-slow/
	timeout := time.Duration(-1)

	if err := app.client.ContainerStop(ctx, app.sshPiperContainerID, &timeout); err != nil {
		return fmt.Errorf("error stopping SSH Piper container: %w", err)
	}

	if err := app.client.NetworkRemove(ctx, app.networkIDreverseProxy); err != nil {
		return fmt.Errorf("error removing the reverse proxy network: %w", err)
	}

	return nil
}
