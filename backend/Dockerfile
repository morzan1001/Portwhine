# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /operator ./cmd/operator/main.go

# Stage 2: Runtime
FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /operator /operator

EXPOSE 50051

ENTRYPOINT ["/operator"]
