# Mobile Architecture

## Structure Overview
- **Backend**: Go (located in `cmd/bot`, `internal/`)
- **Frontend Monorepo**: Turborepo using `npm` workspaces (located in `frontend/`).
  - **`frontend/apps/mobile`**: React Native / Expo application shell. This should strictly contain routing, configuration, and bootstrapping logic.
  - **`frontend/packages/`**: The core of the frontend codebase.
    - **`packages/ui`**: Shared UI components. Mobile-specific UI logic and components should be placed here.
    - **`packages/contract`**: API schemas and endpoints shared across platforms.
    - **`packages/design-system`**: Tokens for colors, spacing, and typography.

## Mobile Development Guidelines
- The mobile app is built with **React Native** and **Expo**.
- **Crucial Pattern**: UI components, screens, state management, and business logic must NEVER be placed in `apps/mobile`. They must always be abstracted into `frontend/packages/` to adhere to the strict Packages-First architecture.
- When creating UI, try to use shared components. If a component must be mobile-specific, place it in `packages/` (e.g., in a dedicated mobile UI package or mobile-specific subdirectory).
