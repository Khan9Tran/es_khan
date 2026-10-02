# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build statically linked standalone binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o eskhan ./cmd/eskhan

# Runtime Stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/eskhan .

# Expose Web IDE port
EXPOSE 8989

# Run as non-root or standard app
ENTRYPOINT ["./eskhan", "-host", "0.0.0.0", "-port", "8989", "-no-browser"]
