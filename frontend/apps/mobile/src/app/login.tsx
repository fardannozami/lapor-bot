import React from 'react';
import { useRouter } from 'expo-router';
import { LoginScreen } from '../../../../packages/ui/src/mobile/screens/LoginScreen';
import { useAuth } from '@lapor-bot/shared';

export default function LoginRoute() {
  const router = useRouter();
  const { setCurrentUser } = useAuth(); // If available, or we just pass user to router params.
  // Actually Expo router we can pass user object or fetch it again in Personal.
  // A better way is state management or local storage, but `useAuth` probably caches the user.

  return (
    <LoginScreen 
      onBack={() => router.back()}
      onLoginSuccess={(user) => {
        const isPhoneLike = !user.name || /^[\d+]/.test(user.name);
        if (isPhoneLike) {
          router.replace('/profile-setup');
        } else {
          router.replace('/personal');
        }
      }}
    />
  );
}
