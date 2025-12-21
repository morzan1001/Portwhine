#!/bin/sh

set -e
set -x

# Create certs directory if it doesn't exist
mkdir -p /certs
CONFIG_FILE="/certs/openssl.cnf"
CA_CONFIG_FILE="/certs/ca.cnf"

# Install envsubst and openssl
apk add --no-cache gettext openssl

# Cleanup existing certificates to ensure fresh generation on every start
echo "Removing existing certificates..."
rm -f /certs/*.crt /certs/*.key /certs/*.srl /certs/*.csr 2>/dev/null || true

# CA
echo "Generating CA..."
openssl genrsa -out /certs/ca.key 4096
openssl req -x509 -new -nodes -key /certs/ca.key -sha256 -days 3650 -out /certs/ca.crt -config $CA_CONFIG_FILE -extensions v3_ca

# Function to generate certs
generate_cert() {
    NAME=$1
    export COMMON_NAME=$2
    export SAN=$3

    echo "Generating cert for $NAME..."
    openssl genrsa -out /certs/$NAME.key 2048
    
    openssl req -new -key /certs/$NAME.key -out /certs/$NAME.csr -config $CONFIG_FILE -extensions v3_req
    openssl x509 -req -in /certs/$NAME.csr -CA /certs/ca.crt -CAkey /certs/ca.key -CAcreateserial -out /certs/$NAME.crt -days 3650 -sha256 -extfile $CONFIG_FILE -extensions v3_req
    
    # Cleanup
    rm /certs/$NAME.csr
    
    # Set permissions
    chmod 644 /certs/$NAME.crt
    chmod 644 /certs/$NAME.key
}

# Read services from config file and generate certs
SERVICES_CONF_TEMPLATE="/certs/services.conf.template"
SERVICES_CONF="/certs/services.conf"

if [ -f "$SERVICES_CONF_TEMPLATE" ]; then
    # Substitute environment variables in template
    envsubst < "$SERVICES_CONF_TEMPLATE" > "$SERVICES_CONF"

    while IFS=";" read -r NAME CN SAN || [ -n "$NAME" ]; do
        # Skip comments and empty lines
        case "$NAME" in
            \#*|"") continue ;;
        esac
        
        generate_cert "$NAME" "$CN" "$SAN"
    done < "$SERVICES_CONF"
else
    echo "Error: Services configuration template not found at $SERVICES_CONF_TEMPLATE"
    exit 1
fi

echo "Certificates generated successfully."
chmod -R 755 /certs
