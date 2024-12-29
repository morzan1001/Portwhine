#!/bin/bash

# Install system CA certificate
echo "Installing CA certificate..."
mkdir -p /usr/local/share/ca-certificates
cp /certs/selfsigned-ca.crt /usr/local/share/ca-certificates/
chmod 644 /usr/local/share/ca-certificates/selfsigned-ca.crt
update-ca-certificates

# Check if jq is installed, if not install it
if ! command -v jq &> /dev/null; then
    echo "Installing jq..."
    apk add --no-cache jq
fi

# Check if mc is installed, if not install it
if ! command -v mc &> /dev/null; then
    echo "Installing mc (MinIO Client)..."
    wget https://dl.min.io/client/mc/release/linux-amd64/mc -O /usr/local/bin/mc
    chmod +x /usr/local/bin/mc
fi

# Check if openssl is installed, if not install it
if ! command -v openssl &> /dev/null; then
    echo "Installing openssl..."
    apk add --no-cache openssl
fi

# Create app role first
echo "Creating application role..."
curl -u "$ELASTIC_USERNAME":"$ELASTIC_PASSWORD" -X POST "https://elasticsearch:9200/_security/role/app_role" -H 'Content-Type: application/json' -d '{
  "cluster": ["monitor"],
  "indices": [{
    "names": ["*"],
    "privileges": ["read", "write", "view_index_metadata", "create_index"],
    "allow_restricted_indices": false
  }]
}'

# Create application user with restricted role
echo "Creating application user..."
curl -u "$ELASTIC_USERNAME":"$ELASTIC_PASSWORD" -X POST "https://elasticsearch:9200/_security/user/$APP_DB_USER" -H 'Content-Type: application/json' -d "{
  \"password\": \"$APP_DB_PASSWORD\",
  \"roles\": [\"app_role\"],
  \"full_name\": \"Application User\"
}"

# Create service account for Kibana
SERVICE_ACCOUNT_RESPONSE=$(curl -u "$ELASTIC_USERNAME":"$ELASTIC_PASSWORD" -X POST "https://elasticsearch:9200/_security/service/elastic/kibana/credential/token")
if [ $? -ne 0 ]; then
    echo "Error: Failed to get service account token"
    exit 1
fi

# Extract token
SERVICE_ACCOUNT_TOKEN=$(echo "$SERVICE_ACCOUNT_RESPONSE" | jq -r '.token.value')
if [ -z "$SERVICE_ACCOUNT_TOKEN" ]; then
    echo "Error: Failed to extract token"
    exit 1
fi

# Update or add token to kibana.yml
awk -v token="$SERVICE_ACCOUNT_TOKEN" '
    $0 ~ /elasticsearch.serviceAccountToken:/ {found=1; print "elasticsearch.serviceAccountToken: " token; next}
    {print}
    END {if (!found) print "elasticsearch.serviceAccountToken: " token}
' /kibana.yml > /tmp/kibana.yml.tmp && cat /tmp/kibana.yml.tmp > /kibana.yml

if [ $? -eq 0 ]; then
    echo "Setup completed successfully"
else
    echo "Error: Failed to update config file"
    exit 1
fi

# Create MinIO user
echo "Creating MinIO user..."
mc alias set portwhine http://minio:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"
mc admin user add portwhine "$APP_MINIO_USER" "$APP_MINIO_PASSWORD"
mc admin policy attach portwhine readwrite --user "$APP_MINIO_USER"

if [ $? -eq 0 ]; then
    echo "MinIO user created successfully"
else
    echo "Error: Failed to create MinIO user"
    exit 1
fi

# Create Redis user with no initial permissions
echo "Creating Redis user..."
echo -e "AUTH $REDIS_DEFAULT_PW\r\nACL SETUSER $REDIS_USER on >$REDIS_PASSWORD resetkeys resetchannels\r\nQUIT\r\n" | openssl s_client -connect redis:6379 -quiet -ign_eof
if [ $? -ne 0 ]; then
    echo "Error: Failed to create Redis user"
    exit 1
fi

# Grant permissions to Redis user for specific databases
echo "Granting permissions to Redis user for databases 0 and 1..."
echo -e "AUTH $REDIS_DEFAULT_PW\r\nACL SETUSER $REDIS_USER on >$REDIS_PASSWORD ~* +@read +@write +select +ping &0 &1\r\nQUIT\r\n" | openssl s_client -connect redis:6379 -quiet -ign_eof
if [ $? -ne 0 ]; then
    echo "Error: Failed to grant permissions to Redis user for databases 0 and 1"
    exit 1
fi

echo "Setup completed successfully."