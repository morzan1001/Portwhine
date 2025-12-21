#!/bin/sh

# Install mc if not present
if ! command -v mc >/dev/null 2>&1; then
    echo "Installing mc..."
    curl -f https://dl.min.io/client/mc/release/linux-amd64/mc -o /usr/local/bin/mc
    chmod +x /usr/local/bin/mc
fi

# Start MinIO in the background
minio server /data --console-address ":9001" &
MINIO_PID=$!

# Wait for MinIO to start
echo "Waiting for MinIO..."
until curl -sS --cacert /root/.minio/certs/CAs/public.crt https://localhost:9000/minio/health/live >/dev/null; do
    sleep 2
done

# Make sure mc trusts our CA (strict TLS, no --insecure)
mkdir -p /root/.mc/certs/CAs
cp -f /root/.minio/certs/CAs/public.crt /root/.mc/certs/CAs/public.crt

# Configure alias
echo "Configuring MinIO..."
mc alias set portwhine https://localhost:9000 "${MINIO_ROOT_USER}" "${MINIO_ROOT_PASSWORD}"

# Create user
echo "Creating user ${APP_MINIO_USER}..."
mc admin user add portwhine "${APP_MINIO_USER}" "${APP_MINIO_PASSWORD}"
mc admin policy attach portwhine readwrite --user "${APP_MINIO_USER}"

# Create bucket
mc mb portwhine/portwhine-data --ignore-existing

echo "MinIO configuration complete."

# Wait for the MinIO process
wait $MINIO_PID
