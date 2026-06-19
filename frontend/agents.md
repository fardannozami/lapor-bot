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
- When building UI components in `packages/ui`, ensure they are compatible with the target platforms. If a component is specific to Web or Mobile, name it accordingly or keep it within its respective app folder.
- Use `packages/design-system` for tokens like colors, spacing, and typography to maintain visual consistency across all apps.
- `packages/contract` contains API schemas and endpoints. Ensure both the web and mobile apps utilize these shared contracts when interacting with the backend.

## 5. Working with MCP (Model Context Protocol)
- As this project has goals related to MCP integration (such as UI/API generation), ensure that any generated code places artifacts in their proper semantic location. For example, generated UI components go to `packages/ui`, and generated schemas go to `packages/contract`.

## 6. Testing Changes
- After making significant changes, run `npm run build` to verify that Turborepo successfully builds all packages and applications without errors.
- Do not commit code that breaks the `npm run lint` check or the build pipeline.

Follow these rules closely to maintain a scalable and healthy workspace!
