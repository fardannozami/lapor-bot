# @lapor-bot/frontend

This is the cross-platform frontend monorepo for Lapor Bot, built with Turborepo. It manages both web and mobile applications from a unified codebase, utilizing shared packages for UI, contracts, and design systems.

## Project Structure

This monorepo uses npm workspaces and strictly enforces a "Packages-First" architecture:

### Apps
The `apps/` directory should **ONLY** contain build configurations, routing/bootstrap files, and thin wrappers. **NO** business logic, view logic, or API calls are allowed here.
- `apps/web`: The web application entry point (Vite/React).
- `apps/mobile`: The mobile application entry point (React Native/Expo).

### Packages
All feature implementation, view logic, and API calls **MUST** reside within the `packages/` directory. Even platform-specific components must be placed here (e.g., inside platform-specific subdirectories or dedicated packages).
- `packages/ui`: UI components and view logic (both shared and platform-specific).
- `packages/design-system`: Design tokens, themes, and foundational styles.
- `packages/contract`: API contracts and schemas shared between the frontend and the backend.
- `packages/shared`: Shared utilities, helpers, types, and API calls.

## Prerequisites

- Node.js (v18+ recommended)
- npm (v11.12.1 is specified in packageManager)

## Scripts and Workflows

Commands are executed using Turbo at the root of the `frontend` directory.

### Install Dependencies

```bash
npm install
```

### Development

To start the development servers for all applications and packages:

```bash
npm run dev
```

### Build

To build all apps and packages:

```bash
npm run build
```

### Lint & Format

To run linters across the project:

```bash
npm run lint
```

To format code with Prettier:

```bash
npm run format
```

## Adding New Packages or Apps

When adding a new package or app to this monorepo, ensure you update the `package.json` inside the new folder to appropriately link any internal dependencies using `"workspace:*"` or the specific package name.

## Current Project Context & Features (June 2026)
Recently added features and architectural focus:
- **Goals Tracking & Profile Setup**: Implemented personal and weekly goals, activity tracking, and a new profile setup flow for users.
- **Personal Page & Leaderboard**: Enhanced UI/UX for activity tracking (daily streak map, heatmap), and refined leaderboard with seasonal and lifetime metrics.
- **Mobile App Setup**: Added Expo mobile app structure in `apps/mobile` and configured environment variables (`EXPO_PUBLIC_API_URL`) to connect with the remote backend API.
- **Clean Architecture Abstraction**: Abstracted web logic into shared modules (`packages/contract`, `packages/design-system`, `packages/shared`) for better reusability between web and mobile.
- **Mobile UI Stabilization**: Resolved NativeWind navigation crash by removing conditional shadows, and added ErrorBoundary for React components in the mobile UI package.
- **Dev Proxy Support**: Added `API_TARGET` support for local frontend development proxy.
