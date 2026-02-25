FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/
COPY backend/pkg/ ./pkg/
COPY backend/workers/ssh-audit/go.mod backend/workers/ssh-audit/go.sum ./workers/ssh-audit/
WORKDIR /app/workers/ssh-audit
RUN go mod download
COPY backend/workers/ssh-audit/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM python:3.12-alpine
RUN apk add --no-cache ca-certificates && \
    pip install --no-cache-dir ssh-audit
COPY --from=builder /worker /worker
EXPOSE 50051
ENTRYPOINT ["/worker"]
