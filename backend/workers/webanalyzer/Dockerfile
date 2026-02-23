FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy root module files (needed for replace directive)
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/

# Copy worker module
COPY backend/workers/webanalyzer/go.mod workers/webanalyzer/go.sum ./workers/webanalyzer/

WORKDIR /app/workers/webanalyzer
RUN go mod download

COPY backend/workers/webanalyzer/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /worker /worker

EXPOSE 50051

ENTRYPOINT ["/worker"]
