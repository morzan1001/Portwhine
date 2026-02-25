FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy root module files (needed for replace directive)
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/
COPY backend/pkg/ ./pkg/

# Copy worker module
COPY backend/workers/webhook/go.mod backend/workers/webhook/go.sum ./workers/webhook/

WORKDIR /app/workers/webhook
RUN go mod download

COPY backend/workers/webhook/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /worker /worker

EXPOSE 50051

ENTRYPOINT ["/worker"]
