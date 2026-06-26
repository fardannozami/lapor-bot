# Agent Guidelines for `@lapor-bot/frontend`

Hello AI Agents! If you are working in this directory, please adhere to the following guidelines to ensure consistency, prevent breakages, and maintain the cross-platform architecture of this project.

## 1. Monorepo Architecture Overview
- This is a **Turborepo** project using **npm workspaces**.
- Apps are in `apps/` (web, mobile).
- Shared packages are in `packages/` (ui, contract, design-system, shared).
- Always consider the impact of modifying shared packages. A change in `packages/ui` will affect both `apps/web` and `apps/mobile`.

## 2. Dependency Management
- To install dependencies, ALWAYS run `npm install` from the `frontend/` root, not from inside the `apps/*` or `packages/*` directories.
- To add a dependency to a specific package/app, either use `npm install <package> -w <workspace>` or manually edit the package's `package.json` and then run `npm install` at the root.
- Cross-workspace dependencies should use the package's name.

## 3. Tooling Constraints
- **Do not use `yarn` or `pnpm`**. This project strictly uses `npm` as the package manager (`packageManager` field in root `package.json`).
- Development servers should be run using `npm run dev` from the `frontend/` root. This command uses `turbo run dev` and will start all necessary applications concurrently.
- Linting and Formatting are handled globally. Run `npm run lint` and `npm run format`. Ensure any generated code adheres to this by running the formatter afterwards.

## 4. Cross-Platform Considerations (Web & Mobile)
- **MANDATORY ARCHITECTURAL RULE**: Every new feature implementation across the backend, Android, iOS, and web platforms MUST consistently utilize the `packages/` directory for view logic and API calls.
- You MUST NOT place any business logic, API calls, or view components in the `apps/` directory. The `apps/` folder should only contain build configuration, routing/bootstrap files, and thin wrappers that render components from `packages/`.
- If a component is specific to Web or Mobile, it still belongs in `packages/` (e.g., in a platform-specific subdirectory like `packages/ui/src/web` or its own dedicated UI package), NOT in the app folder.
- Use `packages/design-system` for tokens like colors, spacing, and typography to maintain visual consistency across all apps.
- `packages/contract` contains API schemas and endpoints. Ensure both the web and mobile apps utilize these shared contracts when interacting with the backend.

## 5. Working with MCP (Model Context Protocol)
- As this project has goals related to MCP integration (such as UI/API generation), ensure that any generated code places artifacts in their proper semantic location. For example, generated UI components go to `packages/ui`, and generated schemas go to `packages/contract`.

## 6. Testing Changes
- After making significant changes, run `npm run build` to verify that Turborepo successfully builds all packages and applications without errors.
- Do not commit code that breaks the `npm run lint` check or the build pipeline.

Follow these rules closely to maintain a scalable and healthy workspace!

## 7. Current Project Context & Features (June 2026)
Recently added features and architectural focus:
- **Goals Tracking & Profile Setup**: Implemented personal and weekly goals, activity tracking, and a new profile setup flow for users.
- **Personal Page & Leaderboard**: Enhanced UI/UX for activity tracking (daily streak map, heatmap), and refined leaderboard with seasonal and lifetime metrics.
- **Mobile App Setup**: Added Expo mobile app structure in `apps/mobile` and configured environment variables (`EXPO_PUBLIC_API_URL`) to connect with the remote backend API.
- **Clean Architecture Abstraction**: Abstracted web logic into shared modules (`packages/contract`, `packages/design-system`, `packages/shared`) for better reusability between web and mobile.
- **Mobile UI Stabilization**: Resolved NativeWind navigation crash by removing conditional shadows, and added ErrorBoundary for React components in the mobile UI package.
- **Dev Proxy Support**: Added `API_TARGET` support for local frontend development proxy.
