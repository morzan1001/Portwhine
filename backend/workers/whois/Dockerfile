FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/
COPY backend/workers/whois/go.mod backend/workers/whois/go.sum ./workers/whois/
WORKDIR /app/workers/whois
RUN go mod download
COPY backend/workers/whois/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM alpine:latest
RUN apk add --no-cache ca-certificates whois
COPY --from=builder /worker /worker
EXPOSE 50051
ENTRYPOINT ["/worker"]
