package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

// configurePiperFilesystem, properly configure the filesystem necessary for
// the SSH Piper container in the host system. This filesystem will be mounted
// as a persistent volume into the SSH Piper container afterwards.
func (app *application) configurePiperFilesystem() error {
	// Remove the existant directory in the host machine with the SSH Piper
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
func (app *application) initializeReverseProxy() (err error) {
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

// stopReverseProxy, this method stops the SSH piper reverse proxy container and
// removes its network.
func (app *application) stopReverseProxy() error {
	ctx := context.Background()

	timeout := time.Duration(-1)
	if err := app.client.ContainerStop(ctx, app.sshPiperContainerID, &timeout); err != nil {
		return fmt.Errorf("error stopping SSH Piper container: %w", err)
	}

	if err := app.client.NetworkRemove(ctx, app.networkIDreverseProxy); err != nil {
		return fmt.Errorf("error removing the reverse proxy network: %w", err)
	}

	return nil
}
