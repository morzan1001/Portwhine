FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy root module files (needed for replace directive)
COPY backend/go.mod backend/go.sum ./
COPY backend/gen/ ./gen/
COPY backend/pkg/ ./pkg/

# Copy worker module
COPY backend/workers/screenshot/go.mod backend/workers/screenshot/go.sum ./workers/screenshot/

WORKDIR /app/workers/screenshot
RUN go mod download

COPY backend/workers/screenshot/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker .

FROM alpine:latest

RUN apk add --no-cache ca-certificates chromium nss freetype harfbuzz font-noto-all

ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/lib/chromium/

COPY --from=builder /worker /worker

EXPOSE 50051

ENTRYPOINT ["/worker"]
