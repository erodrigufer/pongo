FROM ubuntu:latest

# Install required programs.
# Any programs that the user should be capable of using should be installed in
# this section.
RUN apt-get update
RUN apt-get install bash
RUN apt-get install -y openssh-server
# RUN apt-get install -y nmap
RUN apt-get install -y net-tools
RUN apt-get install -y iputils-ping         
RUN apt-get install -y curl

# Create a group and user.
ARG user=admin
ARG group=admin
RUN addgroup $group
RUN adduser \
    --disabled-password \
    --gecos "" \
    --ingroup $group \
	--no-create-home \
    $user

# Prepare entrypoint script.
COPY entrypoint.sh /tmp/entrypoint.sh
# Make entrypoint script executable.
RUN chmod 500 "/tmp/entrypoint.sh"

# Launch SSH service and start with bash.
ENTRYPOINT [ "/bin/bash", "-c", "/tmp/entrypoint.sh" ]
