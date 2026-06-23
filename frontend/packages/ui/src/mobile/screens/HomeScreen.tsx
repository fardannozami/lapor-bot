import React, { useState } from 'react';
import { View, Text, TouchableOpacity, ActivityIndicator } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { HeartPulse, LogIn, RefreshCw, AlertCircle } from 'lucide-react-native';
import { useReports } from '@lapor-bot/shared';
import { StatsOverview } from '../components/StatsOverview';
import { LeaderboardList } from '../components/LeaderboardList';
import type { EnrichedReport } from '@lapor-bot/shared';

interface HomeScreenProps {
  onLoginPress: () => void;
  onHunterPress: (hunter: EnrichedReport) => void;
}

export const HomeScreen: React.FC<HomeScreenProps> = ({ onLoginPress, onHunterPress }) => {
  const { summary, hunters, loading, refreshing, error, refresh } = useReports();

  const seasonTitle = summary ? `Season ${summary.current_season}` : "SWEG Healthy Club";

  const renderHeader = () => (
    <View className="mb-6">
      <View className="flex-row items-center justify-between mb-4">
        <View className="flex-row items-center gap-2">
          <HeartPulse color="#22c55e" size={20} />
          <Text className="text-xl font-bold text-white uppercase tracking-widest">{seasonTitle}</Text>
        </View>
        <TouchableOpacity
          onPress={onLoginPress}
          className="flex-row items-center gap-2 px-3 py-2 rounded-xl bg-[#2dd4bf]/10 border border-[#2dd4bf]/30"
        >
          <LogIn color="#2dd4bf" size={14} />
          <Text className="text-[#2dd4bf] text-xs font-mono">Masuk</Text>
        </TouchableOpacity>
      </View>
      <Text className="text-[10px] text-gray-500 font-mono uppercase tracking-wider mb-6">
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

      <StatsOverview summary={summary ?? null} loading={loading} />
    </View>
  );

  return (
    <SafeAreaView className="flex-1 bg-[#07130c]">
      <View className="flex-1 px-4 pt-4">
        {loading && hunters.length === 0 ? (
          <View className="flex-1 items-center justify-center">
            <ActivityIndicator size="large" color="#2dd4bf" />
            <Text className="text-xs text-gray-400 mt-4 font-mono uppercase tracking-widest">
              Retrieving Status...
            </Text>
          </View>
        ) : (
          <LeaderboardList
            hunters={hunters}
            onSelectHunter={onHunterPress}
            ListHeaderComponent={renderHeader()}
            refreshing={refreshing}
            onRefresh={refresh}
          />
        )}
      </View>
    </SafeAreaView>
  );
};
