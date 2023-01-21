# ctfsmd
CTF Session Manager Daemon (ctfsmd).

## Table of contents
<!-- vim-markdown-toc GitLab -->

* [Installation](#installation)
	- [Dependencies](#dependencies)
	- [Installation steps](#installation-steps)
* [Firewall configuration](#firewall-configuration)
* [Running/stopping ctfsmd](#runningstopping-ctfsmd)
* [Logs with journalctl](#logs-with-journalctl)
* [IP ranges expansion in Docker](#ip-ranges-expansion-in-docker)
	- [Important considerations](#important-considerations)

<!-- vim-markdown-toc -->

## Installation
### Dependencies
1. Go +1.19
2. Docker:
	* Server: Docker Engine 20.10.18 (API 1.41)
	* Client: Docker Engine 20.10.18 (API 1.41)
3. (_optional_) Prometheus +2.38
4. (_optional_) Grafana +9.1.6

### Installation steps
1. Check that the host system has all the required dependencies.
2. Run `./main_configuration.sh --install` with sudo rights.
	* Take into consideration the IP ranges of Docker containers already running in the system (check [this section](#ip-ranges-expansion-in-Docker) for more details.).

## Firewall configuration
If `ufw` is running in Ubuntu as a firewall, add the following rule to allow clients to access the HTTP website to acquire sessions:

```bash
$ ufw allow proto tcp from any to any port <PORT>

<PORT>: the port at which the service can be accessed.
```

## Running/stopping ctfsmd
* Start/stop daemon with `systemctl`
```
$ systemctl start ctfsmd

$ systemctl stop ctfsmd
```

**Important notice**: sometimes some of the containers of a session are not properly stopped when `ctfsmd` is shut down. In that case, run `docker ps -a` to see which containers are still active, and stop the containers with `docker stop`. Finally, after all containers have been properly stopped, execute `docker network prune -f` to remove all unused Docker networks. 

You can close all currently running Docker containers with the command: `docker stop $(docker ps -q)`.

## Logs with journalctl
In order to see the logs of the daemon use `journalctl`.

* See a periodically updated log of the most current events:
```
$ journalctl -f -t ctfsmd

-f  : Show most current logs and update periodically.
-t  : Show only the logs of this particular service.
```

## IP ranges expansion in Docker
* Copy the file `daemon.json` at `/etc/docker/` on the Docker host to expand the range of available private IPs for all the containers running services, otherwise the session manager runs out of available IPs for the containers.
* Restart the Docker daemon afterwards, either with: `systemctl restart docker`, or `systemctl reload docker` or `service docker restart`.

### Important considerations
* If a Docker daemon is already using part of the IP range declared on the new `/etc/docker/daemon.json` file, there will be an unsolvable conflict which will prevent the Docker daemon from correctly running.
* In order to fix this:
	
	a. Run `route -n` and check the current routing table in the system. If some current Docker containers are assigned to the IPs that you want to use, there will be a problem.
	
	b. Change the IP ranges declared on `/etc/docker/daemon.json`, so that they do not collide with the IP ranges of other already running Docker containers, as discovered in the previous step.
