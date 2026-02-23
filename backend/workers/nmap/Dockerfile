FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/
COPY backend/pkg/ ./pkg/
COPY backend/workers/nmap/go.mod backend/workers/nmap/go.sum ./workers/nmap/
WORKDIR /app/workers/nmap
RUN go mod download
COPY backend/workers/nmap/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM alpine:latest
RUN apk add --no-cache ca-certificates nmap
COPY --from=builder /worker /worker
EXPOSE 50051
ENTRYPOINT ["/worker"]
