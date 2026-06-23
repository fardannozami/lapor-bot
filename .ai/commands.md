# Common Commands for Mobile Development

Always run these commands from the `frontend/` directory root, as it is a Turborepo.

- **Install Dependencies**:
  ```bash
  npm install
  ```
  *(Always run from `frontend/` root, even for mobile)*

- **Start Mobile Dev Server**:
  ```bash
  npm run dev --filter mobile
  ```
  *(Starts the Expo development server for the mobile app)*

- **Linting**:
  ```bash
  npm run lint
  ```

- **Formatting**:
  ```bash
  npm run format
  ```

- **Build**:
  ```bash
  npm run build --filter mobile
  ```
