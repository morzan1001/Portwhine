.PHONY: proto proto-lint build-backend build-operator build-pwctl run test clean deps build-workers build-triggers build-all

# Proto generation
proto:
	buf generate

# Lint proto files
proto-lint:
	buf lint

# Build backend (operator + CLI)
build-backend:
	cd backend && go build -o ../bin/operator ./cmd/operator
	cd backend && go build -o ../bin/pwctl ./cmd/pwctl

# Build operator only
build-operator:
	cd backend && go build -o ../bin/operator ./cmd/operator

# Build CLI only
build-pwctl:
	cd backend && go build -o ../bin/pwctl ./cmd/pwctl

# Run operator
run:
	cd backend && go run ./cmd/operator

# Run tests
test:
	cd backend && go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	cd backend && go mod tidy

# Build all worker Docker images
build-workers:
	docker build -f backend/workers/resolver/Dockerfile -t portwhine/resolver-worker:latest .
	docker build -f backend/workers/nmap/Dockerfile -t portwhine/nmap-worker:latest .
	docker build -f backend/workers/ffuf/Dockerfile -t portwhine/ffuf-worker:latest .
	docker build -f backend/workers/humble/Dockerfile -t portwhine/humble-worker:latest .
	docker build -f backend/workers/testssl/Dockerfile -t portwhine/testssl-worker:latest .
	docker build -f backend/workers/screenshot/Dockerfile -t portwhine/screenshot-worker:latest .
	docker build -f backend/workers/webanalyzer/Dockerfile -t portwhine/webanalyzer-worker:latest .

# Build all trigger Docker images
build-triggers:
	docker build -f backend/triggers/ipaddress/Dockerfile -t portwhine/ipaddress-trigger:latest .
	docker build -f backend/triggers/certstream/Dockerfile -t portwhine/certstream-trigger:latest .

# Build everything
build-all: proto build-backend build-workers build-triggers
