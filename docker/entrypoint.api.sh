#!/bin/sh

# Get the group ID of the docker socket
if [ -S /var/run/docker.sock ]; then
    SOCKET_GID=$(stat -c '%g' /var/run/docker.sock)
    
    # Check if the group with this GID exists
    if ! getent group $SOCKET_GID > /dev/null; then
        # Create the group if it doesn't exist
        addgroup -g $SOCKET_GID docker_sock
    fi
    
    # Get the group name
    GROUP_NAME=$(getent group $SOCKET_GID | cut -d: -f1)
    
    # Add appuser to the group
    addgroup appuser $GROUP_NAME
fi

# Execute the command as appuser
exec su-exec appuser "$@"
