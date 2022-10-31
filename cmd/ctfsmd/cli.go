package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type tui struct{}

func (t tui) Run() error {
	// Initialize object application and perform general configurations.
	app := new(application)
	if err := app.setupApplication(); err != nil {
		app.errorLog.Fatal(err)
	}
	// Close Docker client, when application returns.
	defer func() {
		err := app.client.Close()
		if err != nil {
			err = fmt.Errorf("main: error closing Docker client: %w", err)
			app.errorLog.Print(err)
		}
		if err == nil {
			app.infoLog.Print("main: Properly closed Docker client.")
		}
	}()

	// Initialize reverse proxy container (SSH Piper), this method does all the
	// required a-priori configuration.
	if err := app.initializeReverseProxy(); err != nil {
		app.errorLog.Fatal(err)
	}

	// Spawn session manager daemon (smd).
	// Add 1 to the wait-group, so that in the shutdown phase we can be sure
	// when all daemons have correctly returned.
	ctx, cancelDaemons := context.WithCancel(context.Background())
	defer cancelDaemons()
	app.wg.Add(1)
	go app.smd(ctx)
	// Spawn session creation daemon (scd).
	app.wg.Add(1)
	go app.scd(ctx)
	// Spawn session removal daemon (srd).
	app.wg.Add(1)
	go app.srd(ctx)

	app.startHTTPServer()

	app.startMonitord(ctx)

	signalChan := make(chan os.Signal, 1)
	// When main() returns, no more signals will be accepted nor handled.
	defer func() {
		signal.Stop(signalChan)
	}()
	// SIGINT and SIGTERM will be passed to signalChan chan.
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Block waiting for either SIGINT or SIGTERM to arrive.
	<-signalChan
	// Cancel the server first to avoid getting new requests, when starting to
	// cancel the daemons, otherwise new session requests could arrive when
	// the process of stopping the daemons is on its way.
	ctxSrv, cancelSrv := context.WithTimeout(context.Background(), time.Minute)
	defer cancelSrv()
	if err := app.srv.Shutdown(ctxSrv); err != nil {
		err = fmt.Errorf("main: error in server shutdown: %w", err)
		app.errorLog.Print(err)
	}
	// Cancel all daemons by cancelling a context shared with all of them.
	cancelDaemons()
	// Wait until all daemons have returned before starting to close all
	// remaining sessions.
	app.wg.Wait()
	app.infoLog.Print("main: All daemons have shutdown correctly.")
	app.stopAllSessions()

	return nil
}
