FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/
COPY backend/workers/subfinder/go.mod backend/workers/subfinder/go.sum ./workers/subfinder/
WORKDIR /app/workers/subfinder
RUN go mod download
COPY backend/workers/subfinder/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM alpine:latest
RUN apk add --no-cache ca-certificates && \
    wget -q https://github.com/projectdiscovery/subfinder/releases/download/v2.6.7/subfinder_2.6.7_linux_amd64.zip -O /tmp/subfinder.zip && \
    unzip /tmp/subfinder.zip -d /usr/local/bin/ && \
    rm /tmp/subfinder.zip && \
    chmod +x /usr/local/bin/subfinder
COPY --from=builder /worker /worker
EXPOSE 50051
ENTRYPOINT ["/worker"]
