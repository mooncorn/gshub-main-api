#!/bin/bash

sudo yum update -y

sudo yum install docker -y

# Make sure docker starts up on boot
sudo systemctl enable docker

sudo systemctl start docker

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

# Run api
sudo docker run -d -p 3001:3001 --name api -v /var/run/docker.sock:/var/run/docker.sock -e INSTANCE_ID="$INSTANCE_ID" dasior/server-api
