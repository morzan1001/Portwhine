#!/bin/bash
set -euo pipefail

CERT="--cert /usr/share/elasticsearch/config/certs/elasticsearch.crt"
KEY="--key /usr/share/elasticsearch/config/certs/elasticsearch.key"
CA="--cacert /usr/share/elasticsearch/config/certs/ca.crt"
AUTH="-u ${ELASTIC_USERNAME}:${ELASTIC_PASSWORD}"
ES_URL="https://localhost:9200"

# Escape special characters for JSON strings
json_escape() {
  printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g' -e 's/\t/\\t/g'
}

# Start Elasticsearch in background
/usr/local/bin/docker-entrypoint.sh elasticsearch -d -p /tmp/pid

# Wait for Elasticsearch
echo "Waiting for Elasticsearch..."
until curl -fsS $CERT $KEY $CA $AUTH "$ES_URL" 2>/dev/null | grep -q "You Know, for Search"; do
  sleep 3
done
echo "Elasticsearch is ready."

# Helper: Create/update user with role
create_user() {
  local user="$1" pass role
  pass=$(json_escape "$2")
  role="$3"
  curl -fsS $CERT $KEY $CA $AUTH -X PUT "$ES_URL/_security/user/$user" \
    -H 'Content-Type: application/json' \
    -d "{\"password\":\"$pass\",\"roles\":[\"$role\"],\"enabled\":true}" >/dev/null
  echo "Created user: $user"
}

# Helper: Set password for built-in user
set_password() {
  local user="$1" pass
  pass=$(json_escape "$2")
  curl -fsS $CERT $KEY $CA $AUTH -X POST "$ES_URL/_security/user/$user/_password" \
    -H 'Content-Type: application/json' \
    -d "{\"password\":\"$pass\"}" >/dev/null
  echo "Set password for: $user"
}

# Create users
create_user "$APP_DB_USER" "$APP_DB_PASSWORD" "portwhine_api"
set_password "kibana_system" "$KIBANA_SYSTEM_PASSWORD"

[ -n "${METRICBEAT_WRITER_USER:-}" ] && [ -n "${METRICBEAT_WRITER_PASSWORD:-}" ] && \
  create_user "$METRICBEAT_WRITER_USER" "$METRICBEAT_WRITER_PASSWORD" "metricbeat_writer"

[ -n "${METRICBEAT_SETUP_USER:-}" ] && [ -n "${METRICBEAT_SETUP_PASSWORD:-}" ] && \
  create_user "$METRICBEAT_SETUP_USER" "$METRICBEAT_SETUP_PASSWORD" "metricbeat_setup"

[ -n "${KIBANA_USER:-}" ] && [ -n "${KIBANA_USER_PASSWORD:-}" ] && \
  create_user "$KIBANA_USER" "$KIBANA_USER_PASSWORD" "kibana_user"

echo "User setup complete."

# Keep container running (wait for ES process)
tail -f /dev/null --pid=$(cat /tmp/pid)
