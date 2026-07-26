import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  RepositoryProvider,
  AuthProvider,
  loadToken,
  handleUnauthorized,
} from "@lapor-bot/shared";
import { HttpReportRepository, HttpAuthRepository } from "@lapor-bot/contract";
import "./index.css";
import App from "./App.tsx";

const queryClient = new QueryClient();

const getToken = () => loadToken();

const authRepo = new HttpAuthRepository("", getToken, handleUnauthorized);
const reportRepo = new HttpReportRepository("", getToken, handleUnauthorized);

const repositories = {
  reports: reportRepo,
  auth: authRepo,
};

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <AuthProvider authRepo={authRepo}>
        <RepositoryProvider repositories={repositories}>
          <App />
        </RepositoryProvider>
      </AuthProvider>
    </QueryClientProvider>
  </StrictMode>,
);
