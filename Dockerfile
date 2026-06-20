# Stage 1: Build React Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
# Stub every workspace package.json so npm ci can resolve local links
COPY frontend/packages/contract/package.json ./packages/contract/
COPY frontend/packages/design-system/package.json ./packages/design-system/
COPY frontend/packages/shared/package.json ./packages/shared/
COPY frontend/packages/ui/package.json ./packages/ui/
COPY frontend/apps/web/package.json ./apps/web/
RUN npm ci
# Now copy sources
COPY frontend/packages/ ./packages/
COPY frontend/apps/ ./apps/
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
COPY --from=frontend-builder /app/frontend/apps/web/dist ./frontend/dist

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]
