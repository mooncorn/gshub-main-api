#!/bin/bash

# Wait until the API container is found
while ! sudo docker ps -a --format '{{.Names}}' | grep -Eq "^api$"; do
	  sleep 5
done

# Execute the startup script
sudo /usr/local/bin/startup.sh