# Mobile-Focused Development Rules

1. **Mobile Scope ONLY**: The current active development is exclusively focused on the mobile application (`frontend/apps/mobile`). Do NOT modify, touch, or analyze `frontend/apps/web` or any web-only packages unless explicitly instructed.
2. **Monorepo Architecture Compliance**:
   - `frontend/apps/mobile`: This directory should ONLY contain build configurations, routing, and thin bootstrap wrappers.
   - `frontend/packages`: ALL view logic, UI components, state management, and API calls MUST reside here. If a component is specific to Mobile, it still belongs in `packages/` (e.g., in a platform-specific subdirectory or its own dedicated package).
3. **Dependency Management**:
   - Always run `npm install` from the `frontend/` root, not from inside the apps or packages.
   - Use `npm` only. Do not use `yarn` or `pnpm`.
   - To add a dependency to a specific package/app, use `npm install <package> -w <workspace>`.
4. **General Behavior**: 
   - Never break the web app when making changes to shared packages. 
   - Ensure any generated code adheres to formatting rules (run `npm run format`).
