# Lapor Bot AI Agent Guidelines

This file is the root reference for AI Agents (Gemini, Claude, Cursor, etc) working on the `lapor-bot` project.

## Monorepo Architecture
- **Backend (Go)**: The WhatsApp bot logic is located in `cmd/bot`, `internal/`, etc.
- **Frontend (TypeScript/Turborepo)**: Located in `frontend/`. It uses a strict "Packages-First" architecture. 
  - `apps/web` (React/Vite)
  - `apps/mobile` (React Native/Expo)
  - `packages/` (Shared UI, logic, contracts)
  
**CRITICAL RULE FOR FRONTEND**: ALL view logic, UI components, and API calls MUST reside in `frontend/packages/`. The `frontend/apps/` directories should ONLY contain build configurations and thin routing/bootstrap wrappers.

## Feature Tracking
Recently added features:
- **Goals Tracking**: Personal and weekly goals. Includes robust refresh mechanisms and WhatsApp group completion notifications. 
- **Mobile App Setup**: Added Expo mobile app structure.

## Further Reading
Before making any changes to the frontend, you **MUST** read:
- `frontend/agents.md`

Before modifying mobile/web specific code, review their respective READMEs.
