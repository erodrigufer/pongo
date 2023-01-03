package main

import (
	"context"
	"fmt"
	"net/http"
)

// requestSession, method used by clients to request a session.
// Parameter: r *http.Request, to log the info from the client requesting a new
// session.
// Returns: a session and an error.
func (app *application) requestSession(ctx context.Context, r *http.Request) (session, error) {
	// Channel sent to the session manager in which to receive a responseCh with
	// a new valid session and an error.
	// The channel is buffered to only one smResponse. The channel is buffered,
	// instead of unbuffered because unbuffered channels block if a receiver is
	// not ready, while a buffered channel blocks only if there is not enough
	// capacity left in the channel.
	responseCh := make(chan smResponse, 1)

	// Declare session request that will be sent to smd.
	sessionReq := clientReq{
		respCh: responseCh,
	}

	// Send clientReq to session manager to request a new session.
	select {
	case app.sm.requestSession <- sessionReq:
		break
	case <-ctx.Done():
		// Timeout, smd did not accept request in time.
		// Return empty session and an error.
		var ss session
		err := fmt.Errorf("error: smd did not receive request for new session in time. Canceled request for new session")
		return ss, err
	}
	app.infoLog.Printf("New session requested by: %s", r.RemoteAddr)

	var smResponse smResponse
	select {
	// Blocks until a value is available (a value sent by the session manager).
	case smResponse = <-sessionReq.respCh:
		break
	// Timeout.
	case <-ctx.Done():
		// Timeout, smd did not respond back in time.
		// Return empty session and an error.
		var ss session
		err := fmt.Errorf("error: smd did not respond to request for new session back in time. ")
		return ss, err
	}
	// Check if the session manager is notifying of an error, if so return the
	// error.
	err := smResponse.errors
	if err != nil {
		err := fmt.Errorf("client did not receive a valid session from smd: %w", err)
		return smResponse.session, err
	}

	return smResponse.session, nil
}
