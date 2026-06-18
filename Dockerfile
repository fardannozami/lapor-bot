# Stage 1: Build React Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go Backend
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/bot

# Stage 3: Run Stage
FROM alpine:latest
WORKDIR /app

# Install certificates for external connections (required for WhatsApp), timezone data, and curl for healthchecks
RUN apk --no-cache add ca-certificates tzdata curl

# Create data directory for SQLite and WhatsApp sessions
RUN mkdir -p /app/data

# Copy binary from builder
COPY --from=backend-builder /app/main .

# Copy built React assets
COPY --from=frontend-builder /frontend/dist ./frontend/dist

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]
