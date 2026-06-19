# @lapor-bot/frontend

This is the cross-platform frontend monorepo for Lapor Bot, built with Turborepo. It manages both web and mobile applications from a unified codebase, utilizing shared packages for UI, contracts, and design systems.

## Project Structure

This monorepo uses npm workspaces and is structured as follows:

### Apps
- `apps/web`: The web application (Next.js/React).
- `apps/mobile`: The mobile application (React Native/Expo).

### Packages
- `packages/ui`: Shared UI components used across applications.
- `packages/design-system`: Design tokens, themes, and foundational styles.
- `packages/contract`: API contracts and schemas shared between the frontend and the backend.
- `packages/shared`: Shared utilities, helpers, and types.

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
