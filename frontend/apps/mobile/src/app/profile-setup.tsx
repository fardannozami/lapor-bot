import React from 'react';
import { useRouter } from 'expo-router';
import { ProfileSetupScreen } from '../../../../packages/ui/src/mobile/screens/ProfileSetupScreen';
import { useAuth } from '@lapor-bot/shared';
import { ActivityIndicator, View } from 'react-native';

export default function ProfileSetupRoute() {
  const router = useRouter();
  const { user } = useAuth();

  if (!user) {
    return (
      <View style={{ flex: 1, backgroundColor: '#07130c', justifyContent: 'center', alignItems: 'center' }}>
        <ActivityIndicator size="large" color="#2dd4bf" />
      </View>
    );
  }

  return (
    <ProfileSetupScreen 
      user={user}
      onBack={() => {
        router.replace('/login');
      }}
      onComplete={(updatedUser) => {
        router.replace('/personal');
      }}
    />
  );
}
