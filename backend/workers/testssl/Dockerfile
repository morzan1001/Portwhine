FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/
COPY backend/pkg/ ./pkg/
COPY backend/workers/testssl/go.mod workers/testssl/go.sum ./workers/testssl/
WORKDIR /app/workers/testssl
RUN go mod download
COPY backend/workers/testssl/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM alpine:latest
RUN apk add --no-cache ca-certificates bash coreutils openssl bind-tools curl procps && \
    curl -sL https://github.com/testssl/testssl.sh/archive/refs/tags/v3.2.2.tar.gz | tar xz -C /opt && \
    ln -s /opt/testssl.sh-3.2.2/testssl.sh /usr/local/bin/testssl.sh && \
    chmod +x /opt/testssl.sh-3.2.2/testssl.sh
COPY --from=builder /worker /worker
EXPOSE 50051
ENTRYPOINT ["/worker"]
