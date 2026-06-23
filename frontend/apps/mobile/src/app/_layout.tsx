import { DarkTheme, ThemeProvider } from 'expo-router';
import { Stack } from 'expo-router';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RepositoryProvider } from '@lapor-bot/shared';
import { HttpReportRepository, HttpAuthRepository } from '@lapor-bot/contract';

const queryClient = new QueryClient();

// Use an environment variable or default to localhost for simulator.
const API_URL = process.env.EXPO_PUBLIC_API_URL || 'http://localhost:8080';

const repositories = {
  reports: new HttpReportRepository(API_URL),
  auth: new HttpAuthRepository(API_URL),
};

import '../global.css';

export default function TabLayout() {
  return (
    <ThemeProvider value={DarkTheme}>
      <QueryClientProvider client={queryClient}>
        <RepositoryProvider repositories={repositories}>
          <Stack screenOptions={{ headerShown: false }}>
            <Stack.Screen name="index" />
            <Stack.Screen name="login" />
            <Stack.Screen name="personal" />
            <Stack.Screen name="profile-setup" />
          </Stack>
        </RepositoryProvider>
      </QueryClientProvider>
    </ThemeProvider>
  );
}
