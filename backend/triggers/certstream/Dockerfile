FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy root module files (needed for replace directive)
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/

# Copy trigger module
COPY backend/triggers/certstream/go.mod backend/triggers/certstream/go.sum ./triggers/certstream/

WORKDIR /app/triggers/certstream
RUN go mod download

COPY backend/triggers/certstream/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /trigger .

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /trigger /trigger

EXPOSE 50051

ENTRYPOINT ["/trigger"]
