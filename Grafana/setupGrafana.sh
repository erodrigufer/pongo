#!/bin/sh

GRAFANA_PORT=3000

installGrafana(){
	sudo apt-get install -y apt-transport-https
	sudo apt-get install -y software-properties-common wget
	wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -

	echo "deb https://packages.grafana.com/enterprise/deb stable main" | sudo tee -a /etc/apt/sources.list.d/grafana.list

	sudo apt-get update
	sudo apt-get install grafana-enterprise

}

startService(){
	sudo systemctl daemon-reload
	sudo systemctl start grafana-server
	# Grafana will start after boot automatically.
	sudo systemctl enable grafana-server.service
	sudo systemctl status grafana-server

}

installGrafana
startService
sudo ufw allow ${GRAFANA_PORT}/tcp
