package main

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

type containerModel struct {
	// name, name of container.
	name string
	// containerConfig, type with container configuration options for creation.
	containerConfig container.Config
	// hostConfig, host-specific configuration option for creation.
	hostConfig container.HostConfig
	// networkConfig, networking configurations for a container.
	networkConfig network.NetworkingConfig
}

// newUpstream, method to configure upstream containers, it adds the --rm,
// --tty and -i flags to any container being initialized. Set the container's
// internal hostname to be the same as its name.
// Parameters: container's name and image used to create container and the
// networkID of a network to which the container will be connected since the
// initialization.
func newUpstream(name, image, networkID string) (newContainer *containerModel) {
	newContainer = new(containerModel)
	newContainer.name = name
	newContainer.containerConfig.Image = image
	// Equivalent to --rm flag in Docker CLI.
	newContainer.hostConfig.AutoRemove = true
	// Initialize a pseudo-terminal in the container.
	newContainer.containerConfig.Tty = true
	// Attach a Stdin to the container to also be able to interact with it.
	newContainer.containerConfig.AttachStdin = true
	// Set the internal hostname of the container to be the same as the
	// container's name.
	newContainer.containerConfig.Hostname = newContainer.name
	// Create a struct with the configuration of an endpoint for the new
	// container.
	endpointsConfig := new(network.EndpointSettings)
	// Add the networkID to the endpoint.
	endpointsConfig.NetworkID = networkID
	// EndpointsConfig is a map, because it can have the configuration
	// parameters of multiple endpoints at the same time in the map.
	endpointsMap := make(map[string]*network.EndpointSettings)
	endpointsMap[networkID] = endpointsConfig
	newContainer.networkConfig.EndpointsConfig = endpointsMap

	return newContainer
}

// containerConnect, connects an existing container to a network.
// Parameters: networkID, network to which container will be connected.
// containerID, ID of container being connected to network.
func (app *application) containerConnect(networkID, containerID string) error {
	ctx := context.Background()
	err := app.client.NetworkConnect(ctx, networkID, containerID, nil)
	if err != nil {
		return err
	}
	app.debugLog.Printf("Successfully connected container %s to network %s.\n", containerID[:10], networkID[:10])
	return nil
}

// listContainers, list all active containers.
func (app *application) listContainers() error {
	ctx := context.Background()

	containers, err := app.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err
	}

	for _, container := range containers {
		fmt.Printf("Name: %s. Image: %s.\n---------\n", container.Names, container.Image)
	}

	return nil
}

// runContainer, runs a given container in detached mode (-d flag). If the
// method succeeds, it returns the ID of the newly created container.
func (app *application) runContainer(newContainer *containerModel) (string, error) {
	ctx := context.Background()

	resp, err := app.client.ContainerCreate(ctx, &(newContainer.containerConfig), &(newContainer.hostConfig), &(newContainer.networkConfig), nil, newContainer.name)
	if err != nil {
		return "", err
	}

	if err := app.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

//createNetwork, creates a local Docker network with the name specified in the
// parameter. It returns the ID of the newly created network (string) and
// an error type.
func (app *application) createNetwork(networkName string) (string, error) {
	ctx := context.Background()

	// The network created so that SSH Piper and the other containers can
	// communicate with one another should use the driver 'bridge'.
	networkOptions := types.NetworkCreate{
		// The bridge driver is used to create virtual networks between
		// containers running in a single host.
		Driver: "bridge",
		// Check documentation for more information, this option should 'try'
		// to catch any issues if there was already a network previously
		// established with the same name.
		CheckDuplicate: true,
	}

	resp, err := app.client.NetworkCreate(ctx, networkName, networkOptions)
	if err != nil {
		// Check if any warnings were returned, if so, print them out.
		if resp.Warning != "" {
			app.errorLog.Printf("Warning while creating a new network: %s", resp.Warning)
		}
		err = fmt.Errorf("an error occured while trying to create a network: %w", err)
		return "", err
	}
	// Check if any warnings were returned, if so, print them out, nonetheless,
	// client.NetworkCreate did not return an error, so keep executing the main
	// program.
	if resp.Warning != "" {
		app.errorLog.Printf("Warning while creating a new network: %s", resp.Warning)
	}

	return resp.ID, nil
}

// createUpstreamContainer, wrapper to create, run and connect to a network
// a new upstream container. Parameters: name of new container, image
// from which to create new upstream container and networkID to which to
// connect the new container.
// Output: containerID of newly created container.
func (app *application) createUpstreamContainer(name, image, networkID string) (string, error) {
	// Create data model for new upstream container.
	// The container gets connected at initialization time to the network
	// defined by 'networkID'. In a previous iteration of this program, this
	// did not happen, which had as a consequence that all containers got
	// connected to the 'bridge' network by default at their initialization.
	// This was a massive problem, since it was actually intended that most of
	// these containers be isolated from one another, not connected to the same
	// network.
	upstreamContainer := newUpstream(name, image, networkID)
	containerID, err := app.runContainer(upstreamContainer)
	if err != nil {
		return "", err
	}
	app.debugLog.Printf("Created upstream container with ID: %s\n", containerID[:10])

	return containerID, nil
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
		nat.Port("2222/tcp"): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: app.configurations.sshPort}},
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
	app.infoLog.Printf("SSH piper reverse proxy container was created (ID: %s) on port %s.", app.sshPiperContainerID[:10], app.configurations.sshPort)

	return nil
}

// runExec, runs a command inside an already running container.
// Parameters: ctx, a context. containerID, the container ID of the container
// in which the command will be executed, cmd ([]string) the command to be
// executed in the container.
func (app *application) runExec(ctx context.Context, containerID string, cmd []string) error {
	// Configuration parameters for the exec process.
	execConfig := types.ExecConfig{
		// Attach output streams to be able to read diagnostic messages of
		// container while executing the command.
		AttachStderr: true,
		AttachStdout: true,
		// AttachStdin: true,
		Cmd: cmd, // Actual command that will be executed.
	}

	// Configure an exec process, the response is a struct with the ID of the
	// exec process.
	response, err := app.client.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return err
	}

	// Do not check if it is detached or if there is a TTY, before running
	// command.
	execStartConfig := types.ExecStartCheck{
		Detach: false,
		Tty:    false,
	}

	err = app.client.ContainerExecStart(ctx, response.ID, execStartConfig)
	if err != nil {
		return err
	}

	return nil
}

// addUpstream, adds an upstream-container to the SSH Piper reverse proxy
// container, and configures the username used by the client to connect to the
// new upstream-container.
// Parameters: containerName, is the name of the container that should be added
// as an upstream-container. usernamePublic, is the username that a client would
// use to connect to the container through the reverse proxy. usernameUpstream,
// is the actual username that the upstream-container uses and that is mapped
// to usernamePublic (they can be different from one another).
func (app *application) addUpstream(containerName, usernamePublic, usernameUpstream string) error {
	// Create the command (as a string slice) that will be executed by the exec
	// process.
	// E.g.: '/sshpiperd pipe add -n userPublic -u container1 \
	// --upstream-username admin'.
	cmd := []string{
		"/sshpiperd",
		"pipe",
		"add",
		"-n",
		usernamePublic,
		"-u",
		containerName,
		"--upstream-username",
		usernameUpstream,
	}

	ctx := context.Background()
	if err := app.runExec(ctx, app.sshPiperContainerID, cmd); err != nil {
		return err
	}

	return nil
}
