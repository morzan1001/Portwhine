# Build all docker containers
build-container:
	# Build the base image
	docker build -f docker/Dockerfile.base -t base:1.0 .
	# Build the API image without using cache
	docker build --no-cache -f docker/Dockerfile.api -t api:1.0 .
	# Build the certstream_trigger image without using cache
	docker build --no-cache -f docker/Dockerfile.certstream -t certstream:1.0 .
	# Build the ip_address_tritter image without using cache
	docker build --no-cache -f docker/Dockerfile.ipaddress -t ipaddress:1.0 .
	# Build the nmap image without using cache
	docker build --no-cache -f docker/Dockerfile.nmap -t nmap:1.0 .
	# Build the ffuf image without using cache
	docker build --no-cache -f docker/Dockerfile.ffuf -t ffuf:1.0 .
	# Build the humble image without using cache
	docker build --no-cache -f docker/Dockerfile.humble -t humble:1.0 .
	# Build the testssl image without using cache
	docker build --no-cache -f docker/Dockerfile.testssl -t testssl:1.0 .
	# Build the webappanalyzer image without using cache
	docker build --no-cache -f docker/Dockerfile.webappanalyzer -t webappanalyzer:1.0 .
	# Build the screenshot image without using cache
	docker build --no-cache -f docker/Dockerfile.screenshot -t screenshot:1.0 .
	# Build the resolver image without using cache
	docker build --no-cache -f docker/Dockerfile.resolver -t resolver:1.0 .
	# Build the frontend image without using cache
	docker build --no-cache -f docker/Dockerfile.frontend -t frontend:1.0 .

# Create docker network if it doesn't exist
create-docker-network:
	docker network inspect portwhine > /dev/null 2>&1 || docker network create portwhine

# Build containers and start the services
start: build-container create-docker-network
	docker compose up -d

# Stop and remove all running containers
stop:
	docker compose down

# Rebuild and restart the services
restart: stop start

# Show logs of the running services
logs:
	docker compose logs -f

# Clean up all docker images and containers
clean:
	docker compose down --rmi all --volumes --remove-orphans
	docker system prune -f

# Run pylint on the codebase
lint:
	pylint **/*.py

# Generate self-signed certificates
generate-certs:
	# Create CA key and certificate
	openssl genpkey -algorithm RSA -out certs/selfsigned-ca.key -pkeyopt rsa_keygen_bits:2048
	openssl req -x509 -new -nodes -key certs/selfsigned-ca.key -sha256 -days 365 -out certs/selfsigned-ca.crt -subj "/C=DE/ST=Some-State/L=Locality/O=Organization/OU=OrgUnit/CN=CA"

	# Create server key and certificate signing request (CSR)
	openssl genpkey -algorithm RSA -out certs/selfsigned-server.key -pkeyopt rsa_keygen_bits:2048
	openssl req -new -key certs/selfsigned-server.key -out certs/selfsigned-server.csr -config certs/openssl.cnf

	# Sign server certificate with CA
	openssl x509 -req -in certs/selfsigned-server.csr -CA certs/selfsigned-ca.crt -CAkey certs/selfsigned-ca.key -CAcreateserial -out certs/selfsigned-server.crt -days 365 -sha256 -extfile certs/openssl.cnf -extensions v3_ca

	# Clean up CSR and serial files
	rm -f certs/selfsigned-server.csr certs/selfsigned-ca.srl

run-build-runner-frontend:
	cd frontend/portwhine && dart run build_runner build --delete-conflicting-outputs

sort-frontend:
	cd frontend/portwhine && dart run import_sorter:main

genSplashAndIcon:
	cd frontend/portwhine && dart run flutter_launcher_icons:main -f flutter_launcher_icons.yaml
	cd frontend/portwhine && dart run flutter_native_splash:create --path=flutter_native_splash.yaml