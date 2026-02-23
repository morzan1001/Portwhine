FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/
COPY backend/workers/nikto/go.mod workers/nikto/go.sum ./workers/nikto/
WORKDIR /app/workers/nikto
RUN go mod download
COPY backend/workers/nikto/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM alpine:latest
RUN apk add --no-cache ca-certificates perl perl-net-ssleay && \
    wget -q https://github.com/sullo/nikto/archive/refs/tags/2.5.0.tar.gz -O /tmp/nikto.tar.gz && \
    tar xzf /tmp/nikto.tar.gz -C /opt && \
    rm /tmp/nikto.tar.gz && \
    ln -s /opt/nikto-2.5.0/program/nikto.pl /usr/local/bin/nikto && \
    chmod +x /opt/nikto-2.5.0/program/nikto.pl
COPY --from=builder /worker /worker
EXPOSE 50051
ENTRYPOINT ["/worker"]
