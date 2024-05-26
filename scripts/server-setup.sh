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

# Run server-api
sudo docker run -d -p 3001:3001 --name server-api -v /var/run/docker.sock:/var/run/docker.sock dasior/server-api
