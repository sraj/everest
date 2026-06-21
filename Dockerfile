# Backend Dockerfile for Everest
# Uses chromedp for thumbnail generation, requires Chrome/Chromium

FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Resolve any new dependencies
RUN go mod tidy

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# Production image with Chrome for chromedp
FROM alpine:3.19

WORKDIR /app

# Install Chrome/Chromium and dependencies for chromedp
RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    freetype-dev \
    harfbuzz \
    ca-certificates \
    ttf-freefont \
    ttf-dejavu \
    font-noto \
    font-noto-cjk \
    wget

# Set Chrome path for chromedp
ENV CHROME_PATH=/usr/bin/chromium-browser

# Copy binary from builder
COPY --from=builder /app/server .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./server"]
