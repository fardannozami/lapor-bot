# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git for fetching dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary (modernc.org/sqlite is pure Go)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/bot

# Run Stage
FROM alpine:latest

WORKDIR /app

# Install certificates for external connections (required for WhatsApp) and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Create data directory for SQLite and WhatsApp sessions
RUN mkdir -p /app/data

# Copy binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]
