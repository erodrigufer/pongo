#!/bin/bash

# Reference:
# 1) The filesystem:
# https://www.pathname.com/fhs/pub/fhs-2.3.html
# 2) /bin, /sbin, /usr/bin, /usr/local/bin, ...
# https://askubuntu.com/questions/308045/differences-between-bin-sbin-usr-bin-usr-sbin-usr-local-bin-usr-local

SYSTEM_USER=ctfsmd

DAEMON_EXECUTABLE_NAME=ctfsmd

# Check:
# https://askubuntu.com/questions/308045/differences-between-bin-sbin-usr-bin-usr-sbin-usr-local-bin-usr-local
# Path to copy executable (binary) of daemon.
INSTALLATION_PATH=/usr/local/bin/${DAEMON_EXECUTABLE_NAME}/

INSTALLATION_FILE=${INSTALLATION_PATH}${DAEMON_EXECUTABLE_NAME}

# Path for static resources (HTML/CSS).
STATIC_PATH=/var/local/${DAEMON_EXECUTABLE_NAME}

DOCKER_IMAGE_PATH=${STATIC_PATH}/image

# ---------------------------------------------------------------------

# Print an INFO message to the log.
print_info(){
	# Print 'INFO' with green color.
	printf "[\033[1;32mINFO\033[0m] %s\n" "$1"
}

# Print an 'ERROR' message but DO NOT exit.
error_log(){
	# Print 'ERROR' with red color.
	printf "[\033[1;31mERROR\033[0m] %s\n" "$1"

}

# Print an 'FATAL' message and exit.
fatal() {
	# Print 'FATAL' with red color.
	printf "[\033[1;31mFATAL\033[0m] %s\n" "$1"
	exit 1
}

# Some commands are debian-based (even probably BSD compliant)
# so if the distro is not debian-based exit
check_distribution(){
	# print Linux_Standard_Base release ID
	# use grep in egrep mode (-e), being case-insensitive (-i) and check for 
	# either ubuntu or debian
	echo "* Checking system distribution..."
	lsb_release -i | grep -i -E "(ubuntu|debian|kali)" || fatal "System must be debian-based to install daemon."

	# About egrep: with egrep it is easier to write the OR
	# logic ( | ) otherwise all these characters must be escaped
	# with the normal regex grammar of grep.

	# Docker is a dependency for this system.
	which docker > /dev/null || fatal "Docker is not present in the host system. Install Docker!"
	# Go is a dependency for this system.
	which go > /dev/null || fatal "golang is not present in the host system. Install go!"
}

create_system_user(){
	
	echo "* Checking existence of correct system user for daemon..."
	# If the system user already exists, the function returns 0.
	id ${SYSTEM_USER} > /dev/null && return 0

	echo "* Creating system user (UID=${SYSTEM_USER}) ..."
	# add a new system user, without login, without a home directory
	# and with its own group
	sudo adduser --system --no-create-home --group ${SYSTEM_USER} || fatal "System user creation failed."

	# Some remarks:
	# - After creating the user, no home folder for this user should be found on
	# /home
	# - running `sudo -u ${SYSTEM_USER} whoami` should display `${SYSTEM_USER}`
	# - `grep ${SYSTEM_USER} /etc/passwd /etc/shadow` should show info about 
	# the user

}

# Remove any running containers and networks from the Docker daemon, before
# continuing with the installation. It is better to have the Docker daemon in 
# the same initial state.
resetDockerState(){
	echo "* Resetting the Docker daemon..."
	docker stop $(docker ps -q)
	docker network prune -f
}

# Remove daemon's executable, all other files and systemd configuration.
uninstall_daemon(){
	echo "* Un-installing ${DAEMON_EXECUTABLE_NAME}..."
	sudo rm -rf ${INSTALLATION_PATH} || fatal "Program's files could not be removed during un-installation."
	print_info "De-installation of the daemon finished properly."

	echo "* Removing daemon static files..."
	sudo rm -rf ${STATIC_PATH} || fatal "Static files could not be removed during un-installation."
	print_info "Removal of static files finished properly."

	uninstall_systemd
	return 0
}

# Build the daemon.
build_daemon(){
	go build -o ./${DAEMON_EXECUTABLE_NAME} ./cmd/ctfsmd || fatal "Daemon's build failed."
}

# Execute this function if installation fails, functions removes all 
# installation files.
defer_installation(){
	sudo rm -rf ${INSTALLATION_PATH}

	exit 1
}

# Move static files to path where static files will be deployed.
deploy_static_files(){
	echo "* Moving static files to ${STATIC_PATH}..."
	sudo mkdir -p ${STATIC_PATH}

	# This will copy the whole directory with all static files.
	# (By writing a '/' at the end of the source directory, it will not copy
	# the directory itself as well).
	sudo cp -R ./ui/* ${STATIC_PATH} || fatal "Copying static files failed."
}

# Check if daemon is already installed in the system, otherwise install binary 
# on installation path.
install_daemon(){
	
	# check if daemon is already istalled
	[ -f ${INSTALLATION_FILE} ] && { print_info "daemon is already installed in the system."; exit 0; }

	echo "* Installing daemon at ${INSTALLATION_FILE}..."
	sudo mkdir -p ${INSTALLATION_PATH} || fatal "Creating directory path for daemon's executable failed."
	
	# Copy daemon's executable to installation path.
	sudo cp ${DAEMON_EXECUTABLE_NAME} ${INSTALLATION_PATH} || fatal "Copying daemon's executable failed."

	deploy_static_files

	# Change file ownership to root, only root can modify executable, root and 
	# ${SYSTEM_USER} can execute the file.
	# If any of this commands fails, remove binary (security risk!).
	sudo chown root:${SYSTEM_USER} ${INSTALLATION_FILE} || { error_log "chown of daemon's executable failed."; defer_installation ; } 

	# Change file permission.
	sudo chmod 750 ${INSTALLATION_FILE} || { error_log "chmod of daemon's executable failed." ; defer_installation ; } 

	sudo chown -R ${SYSTEM_USER}:${SYSTEM_USER} ${STATIC_PATH} || { error_log "chown of static files failed."; defer_installation ; } 
	sudo chmod 660 ${STATIC_PATH} || { error_log "chmod of static files failed."; defer_installation ; }

	print_info "Daemon installed properly."
		
}

# Copy the default Docker image used as entrypoint to the folder where the
# program looks for the image that it should build to use as entrypoint.
copyDockerImage(){
	mkdir -p ${DOCKER_IMAGE_PATH}
	echo "* Copying default Docker entrypoint image..."
	cp ./DockerImages/entrypoint/* ${DOCKER_IMAGE_PATH} || fatal "copying default Docker image for entrypoint"
}

# Pull or create all necessary Docker images for the session manager daemon.
create_docker_images(){
	print_info "Pulling and building required Docker images."

	# docker pull farmer1992/sshpiperd || fatal "Pulling sshpiperd image failed."

	docker build --tag sshpiperd:latest ./DockerImages/sshpiper || fatal "sshpiperd Docker image failed."

}

# Copy the configuration file for dockerd in order to extend the range of 
# internal network IP addresses.
configure_dockerd(){
	echo "* Configuring Docker daemon to extend IP range."
	mkdir -p /etc/docker || fail "Unable to create '/etc/docker' folder."

	# If a configuration file already exists, store a backup of that file, 
	# before over-writting it with the new configuration file.
	[ -f /etc/docker/daemon.json ] && sudo cp /etc/docker/daemon.json /etc/docker/daemon.json.bak

	sudo cp ./daemon.json /etc/docker || fail "Unable to copy daemon config file"

	# Restart docker daemon to use new 'daemon.json' config file.
	systemctl restart docker
	print_info "Docker Daemon configured correctly."
}

preconfiguration_daemon(){
	resetDockerState

	check_distribution

	create_system_user

	build_daemon

	create_docker_images

	configure_dockerd

	copyDockerImage

}

# Configure systemd to handle ctfSessionManagerd as a daemon.
configure_systemd(){
	echo "* Configuring service as a daemon..."
	# Copy service file to systemd folder.
	sudo cp ./ctfsmd.service /etc/systemd/system || fail "Unable to copy systemd service unit file."
	sudo chmod 644 /etc/systemd/system/ctfsmd.service || fail "Unable to chmod of service unit file."
	# Notify systemd of the new daemon and enable it.
	sudo systemctl daemon-reload || fail "Unable to reload systemd daemons."
	sudo systemctl enable ctfsmd || fail "Error while enabling ctfsmd."
	print_info "ctfsmd was correctly configured and enabled. Start the daemon with 'systemctl start ctfsmd'"
}

# Uninstall any changes or configurations for systemd.
uninstall_systemd(){
	echo "* Removing ctfsmd systemd configuration... "
	# If the daemon is running, stop it first.
	sudo systemctl stop ctfsmd
	# This command properly removes ctfsmd from the systemd configuration.
	sudo systemctl disable ctfsmd || error_log "Error while removing ctfsmd systemd configuration."

	sudo systemctl daemon-reload || error_log "Error while reloading the daemon configuration (systemd)"
	sudo systemctl reset-failed || error_log "Error (reset-failed command systemd)"
}

install(){
	
	# Remove any previous version, before starting installation process.
	uninstall_daemon

	preconfiguration_daemon
	
	install_daemon

	configure_systemd

	exit 0
	
}

print_usage(){
	
	local PROGNAME=$(basename $0)
	echo "${PROGNAME} usage:  ${PROGNAME} [-i|--install] [-h|--help] [-u|--uninstall] 

	-h --help 		Display usage.
	-i --install	Install the daemon, if the daemon is already installed, 
					it un-installs it and installs a new version.
	-u --uninstall 	Un-install the daemon from the system.
	"

	exit 0

}

[ "$1" = '-i' ] && install
[ "$1" = '--install' ] && install
[ "$1" = '-h' ] && print_usage
[ "$1" = '--help' ] && print_usage
[ "$1" = '-u' ] && { uninstall_daemon; exit 0; }
[ "$1" = '--uninstall' ] && {  uninstall_daemon; exit 0; }
print_usage
