// pongo, defines objects and methods shared throughout this project. For
// instance, a general data model (e.g. configuration parameters) which should
// remain consistent in multiple different packages.
package pongo

// UserConfiguration, user configurations handled through flags or env.
// variables by the cli package using cobra and viper.
type UserConfiguration struct {
	// debugMode, run the daemon in debug mode. More extensive logging.
	DebugMode bool
	// sshPort, port in which the SSH Piper will work as an SSH proxy.
	SSHPort string
	// httpAddr, IP and port in which the HTTP service will be hosted, e.g.
	// ':4000'.
	HTTPAddr string
	// maxAvailableSess, the size of the channel that handles the available
	// sessions. scd (session creation daemon) will try to always keep this
	// amount of available sessions ready to be deployed.
	MaxAvailableSess int
	// maxActiveSess, the size of the channel that handles the currently active
	// sessions. srd (session removal daemon) will check periodically to remove
	// sessions from the activeSessions chan which have exceeded their max.
	// lifetime.
	// IMPORTANT: No more sessions can be active than the size of this channel,
	// otherwise the other daemons will block.
	MaxActiveSess int
	// lifetimeSess, is the lifetime of a session in minutes. After this time
	// has elapsed since the activation of the session by a client, the session
	// will expire and it will be removed by srd (session removal daemon).
	LifetimeSess int
	// srdFreq, is the frequency (in min) with which the session removal
	// daemon (srd) checks if some active sessions have expired.
	SRDFreq int
	// timeBetweenRequests, is the minimum time in minutes that has to pass
	// between requests coming from the same user-agent with a particular IP
	// address, so that the user-agent does not get its request for a new
	// session denied (429 Too Many Requests).
	TimeBetweenRequests int
	// noInstrumentation, if true, no instrumentation will be performed in the
	// application.
	NoInstrumentation bool
}
