import React, { useState } from 'react';
import { View, Text, TouchableOpacity, ActivityIndicator } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { HeartPulse, LogIn, RefreshCw, AlertCircle, Settings } from 'lucide-react-native';
import { useReports } from '@lapor-bot/shared';
import { StatsOverview } from '../components/StatsOverview';
import { LeaderboardList } from '../components/LeaderboardList';
import { ErrorBoundary } from '../components/ErrorBoundary';
import type { EnrichedReport } from '@lapor-bot/shared';

interface HomeScreenProps {
  onLoginPress: () => void;
  onHunterPress: (hunter: EnrichedReport) => void;
}

export const HomeScreen: React.FC<HomeScreenProps> = ({ onLoginPress, onHunterPress }) => {
  const { summary, hunters, loading, refreshing, error, refresh } = useReports();

  const seasonTitle = summary ? `Season ${summary.current_season}` : "SWEG Healthy Club";

  const renderHeader = () => (
    <View className="mb-2">
      <View className="flex-row items-center justify-between mb-4 mt-2">
        <TouchableOpacity>
          <Settings color="#2dd4bf" size={24} />
        </TouchableOpacity>
        <TouchableOpacity onPress={onLoginPress}>
          <Text className="text-[#2dd4bf] text-lg font-semibold">Masuk</Text>
        </TouchableOpacity>
      </View>
      <Text className="text-3xl font-bold text-white mb-1">{seasonTitle}</Text>
      <Text className="text-sm text-gray-400 font-medium mb-6">
        Healthy with sports consistency leaderboard
      </Text>

      {error ? (
        <View className="p-4 rounded-2xl bg-[#f97316]/10 border border-[#f97316]/30 flex-row gap-3 items-start mb-6">
          <AlertCircle color="#f97316" size={18} />
          <View className="flex-1">
            <Text className="font-bold text-xs font-mono uppercase text-[#f97316]">System Error</Text>
            <Text className="text-xs text-[#f97316] mt-1">{error.message}</Text>
          </View>
        </View>
      ) : null}
    </View>
  );

  return (
    <SafeAreaView className="flex-1 bg-[#07130c]">
      <View className="flex-1 px-4 pt-4">
        {loading && hunters.length === 0 ? (
          <View className="flex-1 items-center justify-center">
            <ActivityIndicator size="large" color="#2dd4bf" />
            <Text className="text-sm text-gray-400 mt-4 font-bold uppercase tracking-widest">
              Retrieving Status...
            </Text>
          </View>
        ) : (
          <ErrorBoundary>
            <LeaderboardList
              hunters={hunters}
              onSelectHunter={onHunterPress}
              ListHeaderComponent={renderHeader()}
              StatsComponent={<StatsOverview summary={summary ?? null} loading={loading} />}
              refreshing={refreshing}
              onRefresh={refresh}
            />
          </ErrorBoundary>
        )}
      </View>
    </SafeAreaView>
  );
};
