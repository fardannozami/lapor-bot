# Codebase Overview: WhatsApp Gateway

## 1. Project Summary
A Go-based backend service acting as a gateway for WhatsApp interactions, leveraging the `whatsmeow` library. It provides a RESTful API (via Gin) to manage WhatsApp sessions, pair devices, and send messages.

## 2. Tech Stack
- **Language**: Go 1.25.1
- **Web Framework**: Gin (`github.com/gin-gonic/gin`)
- **WhatsApp Library**: WhatsMeow (`go.mau.fi/whatsmeow`)
- **Database**: SQLite (via `modernc.org/sqlite`)
- **Serialization**: Protocol Buffers (`google.golang.org/protobuf`)

## 3. Architecture Structure
The project follows a modular, layered architecture (Clean Architecture inspired):

```
.
├── cmd/
│   └── server/          # Entry point
│       └── main.go      # Dependency injection and server startup
├── internal/
│   ├── app/             # Application Layer
│   │   ├── http/        # HTTP Handlers & Router (Gin)
│   │   └── usecase/     # Business Logic (Pairing, Messaging, Sessions)
│   ├── config/          # Configuration Management
│   ├── domain/          # Domain Entities (e.g., phone)
│   └── infra/           # Infrastructure Layer
│       └── wa/          # WhatsApp specific implementation (Manager)
├── go.mod               # Dependencies
└── data/                # Likely storage for SQLite/Logs
```

## 4. Key Components

### Entry Point (`cmd/server/main.go`)
- Loads configuration.
- Initializes the **WhatsApp Manager** (`wa.Manager`) with SQLite.
- Starts a background process to auto-connect existing sessions.
- Initializes all **Use Cases** (Pairing, Session Management, Messaging).
- Sets up **HTTP Handlers** and starts the Gin server.

### Business Logic (`internal/app/usecase`)
The application logic is broken down into specific use cases:
- **Pairing**: `PairCodeUsecase`, `PairStreamUsecase` (handling QR/pairing codes).
- **Session Management**: `ListSessionsUsecase`, `DeleteSessionUsecase`, `StopSessionUsecase`, `DeleteSessionForceUsecase`, `MeUsecase`.
- **Messaging**: `SendTextUsecase` (sending text messages).
- **Clients**: `ListClientsUsecase`.

### Infrastructure (`internal/infra/wa`)
- **Manager**: Wraps `whatsmeow` client management, handling multi-device sessions and database persistence.

### API Layer (`internal/app/http`)
- Exposes endpoints mapping to the use cases.
- `NewHandler` aggregates all use cases.
- `NewRouter` defines the HTTP routes.
