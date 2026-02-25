FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy root module files (needed for replace directive)
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/

# Copy worker module
COPY backend/workers/humble/go.mod backend/workers/humble/go.sum ./workers/humble/

WORKDIR /app/workers/humble
RUN go mod download

COPY backend/workers/humble/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM python:3.12-alpine

RUN apk add --no-cache ca-certificates git && \
    git clone https://github.com/rfc-st/humble.git /opt/humble && \
    pip install --no-cache-dir -r /opt/humble/requirements.txt && \
    apk del git
ENV PYTHONPATH=/opt

COPY --from=builder /worker /worker

EXPOSE 50051

ENTRYPOINT ["/worker"]
