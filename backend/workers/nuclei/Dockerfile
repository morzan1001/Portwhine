FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/
COPY backend/pkg/ ./pkg/
COPY backend/workers/nuclei/go.mod backend/workers/nuclei/go.sum ./workers/nuclei/
WORKDIR /app/workers/nuclei
RUN go mod download
COPY backend/workers/nuclei/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM alpine:latest
RUN apk add --no-cache ca-certificates && \
    wget -q https://github.com/projectdiscovery/nuclei/releases/download/v3.3.7/nuclei_3.3.7_linux_amd64.zip -O /tmp/nuclei.zip && \
    unzip /tmp/nuclei.zip -d /usr/local/bin/ && \
    rm /tmp/nuclei.zip && \
    chmod +x /usr/local/bin/nuclei
COPY --from=builder /worker /worker
EXPOSE 50051
ENTRYPOINT ["/worker"]
