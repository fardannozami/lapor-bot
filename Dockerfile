# Stage 1: Build React Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY package*.json ./
COPY packages/shared/package.json ./packages/shared/
COPY frontend/web/package.json ./frontend/web/
RUN npm ci
COPY packages/shared/ ./packages/shared/
COPY frontend/web/ ./frontend/web/
RUN npm run build --workspace=@lapor-bot/web

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
COPY --from=frontend-builder /app/frontend/web/dist ./frontend/dist

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]
