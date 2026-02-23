FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy root module files (needed for replace directive)
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/

# Copy trigger module
COPY backend/triggers/ipaddress/go.mod triggers/ipaddress/go.sum ./triggers/ipaddress/

WORKDIR /app/triggers/ipaddress
RUN go mod download

COPY backend/triggers/ipaddress/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /trigger .

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /trigger /trigger

EXPOSE 50051

ENTRYPOINT ["/trigger"]
