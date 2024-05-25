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
