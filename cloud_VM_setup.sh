#!/bin/sh

installDocker(){
 curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

 echo \
	   "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
	     $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

apt-get update
apt-get --yes install docker-ce docker-ce-cli containerd.io docker-compose
}

# Reference: https://github.com/golang/go/wiki/Ubuntu
installGo(){
sudo add-apt-repository --yes ppa:longsleep/golang-backports
sudo apt update --yes
sudo apt install --yes golang-go
}

setupFirewall(){
	PORT_HTTP=4000
	PORT_SSH=50000
	ufw allow proto tcp from any to any port ${PORT_HTTP}
	ufw allow proto tcp from any to any port ${PORT_SSH}
}

installDocker
installGo
setupFirewall


