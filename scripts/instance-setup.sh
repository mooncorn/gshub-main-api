#!/bin/bash

sudo yum update -y
sudo yum install docker -y

# Make sure docker starts up on boot
sudo systemctl enable docker

sudo systemctl start docker

# Create the startup script
STARTUP_SCRIPT="/usr/local/bin/startup.sh"

sudo tee ${STARTUP_SCRIPT} > /dev/null <<'EOF'
#!/bin/bash

# Function to check if Docker is running
is_docker_ready() {
    sudo docker info &>/dev/null
    return $?
}

# Wait for Docker to be ready
until is_docker_ready; do
    echo -n "."
    sleep 1
done

# Get the instance ID
INSTANCE_ID=$(ec2-metadata -i | cut -d ' ' -f 2)

# Pull the latest version of the image
sudo docker pull dasior/server-api

# Check if the container is already running
if sudo docker ps -a --format '{{.Names}}' | grep -Eq "^api\$"; then
    # Stop and remove the existing container
    sudo docker stop api
    sudo docker rm api
fi

# Run the application
sudo docker run --restart always -d -p 3001:3001 --name api -v /var/run/docker.sock:/var/run/docker.sock -e INSTANCE_ID="$INSTANCE_ID" -e APP_ENV="production" dasior/server-api
EOF

# Make the startup script executable
sudo chmod +x ${STARTUP_SCRIPT}

# Create a systemd service to run the startup script at boot
sudo tee /etc/systemd/system/startup.service > /dev/null <<'EOF'
[Unit]
Description=Run startup script

[Service]
ExecStart=/usr/local/bin/startup.sh

[Install]
WantedBy=multi-user.target
EOF

# Enable the service
sudo systemctl enable startup.service

# Initial run of the startup script
sudo /usr/local/bin/startup.sh
