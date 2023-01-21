## v0.6.1
* Fix dependabot security alert regarding `containerd` dependency.
* Bump docker version in go.mod

## v0.6.0
* Rename the project/program to `pongo`

## v0.5.1
* Improve styling of README file.

## v0.5.0
* It is not possible any more to access the `admin` account from within the entrypoint Docker container.

## v0.4.3
* Update docker and go-semver versions in go.mod and go.sum files. 

## v0.4.2
* Move the type definitions and the function to create a new templates cache to an internal package.

## v0.4.1
* Very minor style improvements in the documentation.

## v0.4.0
* The program now builds the container image for the upstream container using the Docker SDK when initially configuring the program (see `/internal/docker/build`).
* The internal package used to build docker images also provides a test for the `ImageBuild()` function.

## v0.3.0
* Fix issue when stopping sessions. The method stopping the sessions was trying to stop some networks that did not exist within the session model of the current application (they were used in a previous iteration of this application when the sessions were made up of multiple containers and networks).

## v0.2.1
* Move `cli` to `/internal` (best practices in repository structure).

## v0.2.0
* Move all static files (HTML and CSS) needed for the web page to `/var/local/ctfsmd`.

## v0.1.3
* Use `crypto/rand` to generate passwords and usernames.
* Move the functions used to generate random strings to an internal package.

## v0.1.2
* Upgrade docker Go client package to next minor version.

## v0.1.1
* Migrate CLI to `cobra` and `viper`.
* Application can now be controlled by both flags and env. variables.

## v0.0.3
* Remove old containers' references from the code's data model.
* Refactor SSH Piper reverse proxy code to `reverseProxy.go`.

## v0.0.2
* Display revision also in the webpage.

## v0.0.1
* Add revision to first daemon's logs when booting the daemon for the first time.
