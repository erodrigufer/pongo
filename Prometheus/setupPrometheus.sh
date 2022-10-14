#!/bin/sh

# Reference:
# * https://computingforgeeks.com/install-prometheus-server-on-debian-ubuntu-linux/

PROMETHEUS_PORT=9090

installBinaries(){
	# Create prometheus group and system user.
	sudo groupadd --system prometheus
	sudo useradd -s /sbin/nologin --system -g prometheus prometheus

	# Create folders needed for config files and to store dynamic data.
	sudo mkdir -p /var/lib/prometheus
	sudo mkdir -p /etc/prometheus

	# Download prometheus into tmp directory.
	mkdir -p /tmp/prometheus && cd /tmp/prometheus
	curl -s https://api.github.com/repos/prometheus/prometheus/releases/latest | grep browser_download_url | grep linux-amd64 | cut -d '"' -f 4 | wget -qi -

	# Extract the downloaded package.
	tar xvf prometheus*.tar.gz
	cd prometheus*/

	# Move the binaries to /usr/local/bin
	sudo mv ./prometheus ./promtool /usr/local/bin/

	# Move the consoles.
	sudo mv consoles/ console_libraries/ /etc/prometheus/

	# Change files' permission and owners.
	sudo chown -R prometheus:prometheus /etc/prometheus/
	sudo chmod -R 774 /etc/prometheus/
	sudo chown -R prometheus:prometheus /var/lib/prometheus/

}

# Move the configuration template to /etc/prometheus
sudo cp ./prometheus.yml /etc/prometheus/prometheus.yml
sudo cp ./web.yml /etc/prometheus/web.yml

# Copy systemd service file.
sudo cp ./prometheus.service /etc/systemd/system/prometheus.service

# Allow Prometheus port with firewall.
sudo ufw allow ${PROMETHEUS_PORT}/tcp

# Configure systemctl and start prometheus.
sudo systemctl daemon-reload
sudo systemctl enable prometheus
sudo systemctl start prometheus
