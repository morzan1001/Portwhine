FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/
COPY backend/pkg/ ./pkg/
COPY backend/workers/ffuf/go.mod workers/ffuf/go.sum ./workers/ffuf/
WORKDIR /app/workers/ffuf
RUN go mod download
COPY backend/workers/ffuf/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM alpine:latest
RUN apk add --no-cache ca-certificates curl
# Install ffuf
RUN curl -sL https://github.com/ffuf/ffuf/releases/download/v2.1.0/ffuf_2.1.0_linux_amd64.tar.gz | tar xz -C /usr/local/bin ffuf
# Install common wordlist
RUN mkdir -p /wordlists && \
    curl -sL https://raw.githubusercontent.com/danielmiessler/SecLists/master/Discovery/Web-Content/common.txt -o /wordlists/common.txt
COPY --from=builder /worker /worker
EXPOSE 50051
ENTRYPOINT ["/worker"]
