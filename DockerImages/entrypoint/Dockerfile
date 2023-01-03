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
ARG pass=lbecgyjsnei3mu7yp1hw
ARG group=admin
ARG home=/home/$user
RUN addgroup $group
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home $home \
    --ingroup $group \
    $user

# Set user password.
RUN echo $user':'$pass | chpasswd

# Prepare entrypoint script.
COPY entrypoint.sh /tmp/entrypoint.sh
RUN chmod 500 "/tmp/entrypoint.sh"

# Set working directory.
WORKDIR $home

# Disable cmd history (the commands executed by the user inside their session
# will not be stored in a file by bash).
RUN echo 'set +o history' >> $home/.bashrc

# Launch SSH service and start with bash.
ENTRYPOINT [ "/bin/bash", "-c", "/tmp/entrypoint.sh" ]
