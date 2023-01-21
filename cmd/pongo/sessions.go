package main

import (
	"context"
	"fmt"
	"time"

	"github.com/erodrigufer/pongo/internal/sysutils"
)

// charsetUsername, valid character-set for generating random usernames.
// IMPORTANT: Only lower-case alphabetic characters can be used as valid
// usernames in the SSH system we are using.
const charsetUsername = "abcdefghijklmnopqrstuvwxyz"

// charsetPassword, valid character-set for generating random passwords.
// When special characters are used in the passwords (like $, ! ...) it is
// impossible to connect to the container with SSH. Avoid special characters!
const charsetPassword = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// stopSession, stops all the containers that form a session and removes all
// the session-specific networks created for that session.
func (app *application) stopSession(ss session) error {
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
	// Remove all the session-specific containers, which are stored in a slice
	// of strings with the containers' IDs.
	for _, containerID := range ss.containersIDs {
		if err := app.client.ContainerStop(ctx, containerID, &timeout); err != nil {
			return fmt.Errorf("error: unable to stop container (with container ID %s): %w", containerID, err)
		}

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
	var err error
	newSession.username, err = sysutils.NewRandomUsername(usernameLength, charsetUsername)
	if err != nil {
		return newSession, fmt.Errorf("error: could not create a new session: %w", err)
	}
	newSession.password, err = sysutils.NewRandomPassword(passwordLength, charsetPassword)
	if err != nil {
		return newSession, fmt.Errorf("error: could not create a new session: %w", err)
	}
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

// LEGACY: code was used in a previous iteration of this codebase, where
// sessions had their own internal networks as well with multiple containers
// inside a session.
//
// createSessionNetwork, creates the network that will be used by all
// containers inside the same session. Parameter: networkName, name for the new
// network. Output: networkID (string) and error.
// func (app *application) createSessionNetwork(networkName string) (string, error) {
// 	networkID, err := app.createNetwork(networkName)
// 	if err != nil {
// 		return "", err
// 	}
// 	app.debugLog.Printf("Created session network with ID: %s.\n", networkID[:10])
// 	return networkID, nil
// }

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
