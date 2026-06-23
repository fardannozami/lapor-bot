import React from 'react';
import { useRouter } from 'expo-router';
import { PersonalScreen } from '../../../../packages/ui/src/mobile/screens/PersonalScreen';
import { useAuth } from '@lapor-bot/shared';
import { ActivityIndicator, View } from 'react-native';

export default function PersonalRoute() {
  const router = useRouter();
  const { user, logout } = useAuth();

  if (!user) {
    return (
      <View style={{ flex: 1, backgroundColor: '#07130c', justifyContent: 'center', alignItems: 'center' }}>
        <ActivityIndicator size="large" color="#2dd4bf" />
      </View>
    );
  }

  return (
    <PersonalScreen 
      user={user}
      onLogout={() => {
        logout?.();
        router.replace('/');
      }}
    />
  );
}
