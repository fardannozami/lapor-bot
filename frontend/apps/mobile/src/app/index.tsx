import React from 'react';
import { useRouter } from 'expo-router';
import { HomeScreen } from '../../../../packages/ui/src/mobile/screens/HomeScreen';

export default function IndexRoute() {
  const router = useRouter();

  return (
    <HomeScreen 
      onLoginPress={() => router.push('/login')}
      onHunterPress={(hunter) => {
        // Here we could navigate to a detailed profile view if wanted.
        console.log("Pressed hunter:", hunter.name);
      }}
    />
  );
}
