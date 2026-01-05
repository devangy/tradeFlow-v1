#!/usr/bin/env bash



# ec2 startup script to setup every thing

set -euxo pipefail

echo "EC2 startup script started"

# update system packages
echo "startup script started!"
sudo apt update -y
sudo apt upgrade -y
echo "packages update complete"

# install tools required
# Add Docker's official GPG key:
apt update
apt install ca-certificates curl
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
tee /etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "$${UBUNTU_CODENAME:-$${VERSION_CODENAME}}")
Components: stable
Signed-By: /etc/apt/keyrings/docker.asc
EOF

apt update

apt install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# start docker and enable it on every boot
systemctl start docker
systemctl enable docker
systemctl status docker

# adding ubuntu user to docker group to run without sudo in ssh
usermod -aG docker ubuntu

# Create a folder
mkdir actions-runner && cd actions-runner# Download the latest runner package
curl -o actions-runner-linux-x64-2.330.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.330.0/actions-runner-linux-x64-2.330.0.tar.gz# Optional: Validate the hash
echo "af5c33fa94f3cc33b8e97937939136a6b04197e6dadfcfb3b6e33ae1bf41e79a  actions-runner-linux-x64-2.330.0.tar.gz" | shasum -a 256 -c# Extract the installer
tar xzf ./actions-runner-linux-x64-2.330.0.tar.gz

# Create the runner and start the configuration experience
./config.sh --url https://github.com/devangy/market --token ${runner_token} # Last step, run it!
./run.sh

echo "EC2 startup script completed"
