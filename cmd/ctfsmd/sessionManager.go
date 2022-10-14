package main

import (
	"context"
	"fmt"
	"time"

	prometheus "github.com/erodrigufer/CTForchestrator/internal/prometheus"
)

// initializeSessionManager, this method creates and populates all the channels
// and data structures required for the sm daemons.
func (app *application) initializeSessionManager() {
	sm := new(sessionManager)
	// Create a channel to store the currently available sessions. The channel
	// must be buffered, so that it does not block at the creation of new
	// sessions that are then immediately sent to the channel. An unbuffered
	// channel blocks a sender until a receiver is ready.
	sm.availableSessions = make(chan session, app.configurations.maxAvailableSess)
	// Create a channel to receive requests for a session from clients.
	// The clients will send a clientReq to this channel, so that the
	// session manager can use the received channel as a channel to respond to
	// a particular client. This channel is unbuffered, so that the clients
	// block until the server reads their request, and the number of concurrent
	// requests is unlimited.
	sm.requestSession = make(chan clientReq)
	// Channel sm.activeSessions stores the sessions that have been delivered
	// to clients, so that srd can check periodically if the sessions have
	// exceeded their maximum lifetime, and if so, it terminates the sessions.
	sm.activeSessions = make(chan session, app.configurations.maxActiveSess)
	app.sm = sm

}

// smd, the session manager (sm) daemon (d) is spawned perpetually in a
// goroutine, and is in charge of handling the requests for new sessions from
// all clients.
func (app *application) smd(ctx context.Context) {
	// timeLastRequest maps the IP of a client making a request with the time
	// in which it made its last request. This map is used to block clients from
	// performing too many requests in a very small amount of time.
	timeLastRequest := make(map[string]time.Time)
	for {
		var req clientReq
		select {
		// Read the channel with requests for new session.
		case req = <-app.sm.requestSession:
			// Continue execution for request that just got in.
			break
		case <-ctx.Done():
			// smd daemon received a context termination.
			app.infoLog.Print("smd: shutting down.")
			app.wg.Done()
			// Daemon terminates.
			return
		}

		tlr, ok := timeLastRequest[req.reqInfo.clientAddr]
		if !ok {
			app.infoLog.Printf("smd: client (%s) is establishing a connection for the first time.", req.reqInfo.clientAddr)
		}
		// Check when was the last request from this client.
		if ok {
			if timeDiff := time.Since(tlr); timeDiff < time.Duration(app.configurations.timeBetweenRequests)*time.Minute {
				app.infoLog.Printf("smd: Not enough time has passed since last request by client %s", req.reqInfo.clientAddr)
				response := smResponse{}
				// Send ERR_LAST_REQ = not enough time has passed since last
				// request, so that the client can identify the exact error
				// that took place.
				response.errors = ERR_LAST_REQ
				req.respCh <- response
				continue // Loop back to the beginning, wait for next request.
			}
		}
		// The last request was before the minimum time between request, update
		// time of last request.
		timeLastRequest[req.reqInfo.clientAddr] = time.Now()
		app.infoLog.Printf("smd: Req from %s at time: %v.", req.reqInfo.clientAddr, timeLastRequest[req.reqInfo.clientAddr])

		// Create a struct of type smResponse (session manager response) to
		// send a response back to the client.
		response := smResponse{}

		// Check if there are no more available sessions.
		if len(app.sm.availableSessions) == 0 {
			// Send error to client.
			response.errors = fmt.Errorf("no more sessions are currently available.")
			req.respCh <- response
			continue // Loop back to the beginning, wait for next request.
		}

		// Get a session to deliver to client requesting session.
		response.session = <-app.sm.availableSessions
		prometheus.DecrementGauge(app.instrumentation, "available_sessions_total")
		response.errors = nil
		// Add activation time for new session. Required to kill session after
		// lifetime expires.
		response.session.timeActivated = time.Now()

		// Send requested session back to client wrapped in a smResponse struct.
		req.respCh <- response

		// Send new client's session to the activeSessions channel. If the
		// buffered channel is completely full, then this call will block, since
		// the maximum number of active sessions has been achieved.
		app.sm.activeSessions <- response.session
		prometheus.IncrementGauge(app.instrumentation, "active_sessions_total")
	}
}

// scd, session creator daemon is in charge of guaranteeing that the
// availableSessions ch always has available sessions. scd dynamically creates
// new sessions and adds them to chan availableSessions.
func (app *application) scd(ctx context.Context) {
	for {
		// ss is the next session that will be added to availableSessions chan.
		ss, err := app.createSession()
		if err != nil {
			err = fmt.Errorf("scd: unable to create session: %w", err)
			app.errorLog.Print(err)
			continue
		}
		// scd blocks in the next call if the max. capacity of the channel is
		// achieved, it will unlock and keep creating more sessions as soon as a
		// session is taken from the channel.
		select {
		case app.sm.availableSessions <- ss:
			prometheus.IncrementGauge(app.instrumentation, "available_sessions_total")
			break
		case <-ctx.Done():
			// Stop the session waiting to be sent to the availableSession chan.
			// Otherwise, this session will not be cleaned up when the channels
			// are emptied out.
			if err := app.stopSession(ss); err != nil {
				err = fmt.Errorf("scd: unable to stop session dangling outside of any channel: %w", err)
				app.errorLog.Print(err)
			}
			app.infoLog.Print("scd: shutting down.")
			app.wg.Done()
			return
		}

		app.infoLog.Printf("scd: sent new session %s to availableSessions ch.", ss.name)
		app.infoLog.Printf("scd: current number of sessions in availableSessions ch: %d", len(app.sm.availableSessions))

	}

}

// srd, session removal daemon is in charge of periodically checking the
// activeSessions chan and stopping the sessions that are older than a given
// time scale. It guarantees that all session will not live more than a given
// time (after their activation).
func (app *application) srd(ctx context.Context) {
	// freq, the frequency with which srd will be called to clean the
	// activeSessions chan, convert the frequency parsed from flags to a
	// time.Duration value.
	freq := time.Minute * time.Duration(app.configurations.srdFreq)
	// maxLifetime, is the maxLifetime of a session, convert lifetime parsed
	// from flags to time.Duration value.
	maxLifetime := time.Minute * time.Duration(app.configurations.lifetimeSess)

	oldestSession := session{}
	getNewSession := true
	// for-loop for the timer.
	for {
		// Start a timer which will return after freq.
		timer := time.NewTimer(freq)
		app.infoLog.Printf("srd: next check in %v.", freq)
		// After the time of the timer expires, the current time will be sent
		// through its channel, discard the value. The read from the chan will
		// block until the timer expires.
		select {
		case <-timer.C:
			break
		case <-ctx.Done():
			app.infoLog.Print("srd: shutting down.")
			app.wg.Done()
			return
		}
		app.infoLog.Print("srd: checking max. lifetime of active sessions.")
		// for-loop to check multiple sessions, until the oldest session has not
		// expired.
		keepCheckingSessions := true
		for keepCheckingSessions {
			// This first check is necessary in case that in the last iteration
			// the oldestSession did not expire. So that in the next iteration,
			// no newer sessions would be read first from the channel. And the
			// program checks first if the last oldestSession expired yet.
			if getNewSession {
				select {
				// activeSessions is a FIFO, read the oldest session.
				case oldestSession = <-app.sm.activeSessions:
				// If there are no activeSessions, re-start the loop and start
				// a new timer (default). The read-call on activeSessions never
				// blocks. Break out of the inner-loop.
				default:
					// It breaks out of inner-loop in the next if-statement.
					keepCheckingSessions = false
				}
			}
			// If no oldestSession was read from the channel break out of inner
			// loop. Re-start timer in outer-loop. This statement is required,
			// because writing the break inside the 'default' case, would only
			// break out of the 'select'.
			if !keepCheckingSessions {
				break
			}
			// Calculate the time elapsed since the activation of the oldest
			// session.
			timeElapsed := time.Since(oldestSession.timeActivated)
			// Check if oldest session expired.
			if timeElapsed > maxLifetime {
				if err := app.stopSession(oldestSession); err != nil {
					err = fmt.Errorf("srd: unable to stop expired session (%s): %w", oldestSession.name, err)
					app.errorLog.Print(err)
					// error stopping oldest expired session, break out of
					// the inner-loop and re-start the timer, try again
					// later.
					keepCheckingSessions = false
				}
				app.infoLog.Printf("srd: expired session (%s) successfully stopped.", oldestSession.name)
				prometheus.DecrementGauge(app.instrumentation, "active_sessions_total")
				// Check the next oldest session. 'continue', restarts the inner
				// loop. Guarantee that the value to get next activateSessions
				// is true.
				getNewSession = true
				continue
			} else {
				// If the oldestSession has not expired yet, continue to the
				// next outer for-loop iteration and restart the timer. Break
				// out of the inner for-loop.
				keepCheckingSessions = false
				// Do not get new sessions values from the activeSessions
				// channel, since the current oldestSession has not expired yet.
				getNewSession = false
			}
		}
	}
}

// stopAllSessions, stops all active and available sessions.
// This method is used at shutdown to stop all remaining sessions.
func (app *application) stopAllSessions() {
	availableSessions := true
	for availableSessions {
		select {
		case ss := <-app.sm.availableSessions:
			// Close session
			if err := app.stopSession(ss); err != nil {
				err = fmt.Errorf("error stopping session at shutdown: %w", err)
				app.errorLog.Print(err)
				continue // Try next session in the channel.
			}
		default:
			// No more sessions in channel, break out of for-loop.
			availableSessions = false
			break
		}
	}
	app.infoLog.Print("Finish stopping sessions from channel 'availableSessions'.")

	activeSessions := true
	for activeSessions {
		select {
		case ss := <-app.sm.activeSessions:
			// Close session
			if err := app.stopSession(ss); err != nil {
				err = fmt.Errorf("error stopping session at shutdown: %w", err)
				app.errorLog.Print(err)
				continue // Try next session in the channel.
			}
		default:
			// No more sessions in channel, break out of for-loop.
			activeSessions = false
			break
		}
	}
	app.infoLog.Print("Finish stopping sessions from channel 'activeSessions'.")

	// Stop SSH Piper container and remove its network.
	if err := app.stopReverseProxy(); err != nil {
		app.errorLog.Print(err)
	}

}
