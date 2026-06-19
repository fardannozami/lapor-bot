import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RepositoryProvider } from '@lapor-bot/shared'
import { HttpReportRepository, HttpAuthRepository } from '@lapor-bot/contract'
import './index.css'
import App from './App.tsx'

const queryClient = new QueryClient();

// Instantiate the repositories with empty baseURL since we are using vite proxy
const repositories = {
  reports: new HttpReportRepository(''),
  auth: new HttpAuthRepository(''),
};

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RepositoryProvider repositories={repositories}>
        <App />
      </RepositoryProvider>
    </QueryClientProvider>
  </StrictMode>,
)
