## TODO for SessionManager

### Stability
1. When removing a session with stopSession(), also remove its dir, from the volume in the container (host system). To not have unnecessary persistent data.
2. Let Docker give containers and networks its names automatically, do not give them a name yourself. To avoid collisions.
3. Make random containers and network names longer than 6 characters to avoid collisions.
