import React from 'react';
import { View, Text, TouchableOpacity } from 'react-native';
import { Flame, Trophy, Activity, ArrowUpRight, Award, CalendarDays } from 'lucide-react-native';
import type { EnrichedReport } from '@lapor-bot/shared';
import { getJobColor } from '@lapor-bot/shared';

interface HunterCardProps {
  hunter: EnrichedReport;
  onClick: () => void;
  activeTab?: string;
  rankIndex?: number;
}

const RankBadge = ({ rank }: { rank?: number }) => {
  if (rank === undefined) return null;
  
  if (rank === 0) {
    return (
      <View className="w-10 h-10 rounded-full bg-[#eab308]/20 border border-[#eab308]/50 items-center justify-center">
        <Text className="text-[#eab308] text-sm font-bold">1st</Text>
      </View>
    );
  }
  if (rank === 1) {
    return (
      <View className="w-10 h-10 rounded-full bg-slate-300/20 border border-slate-300/40 items-center justify-center">
        <Text className="text-slate-300 text-sm font-bold">2nd</Text>
      </View>
    );
  }
  if (rank === 2) {
    return (
      <View className="w-10 h-10 rounded-full bg-amber-700/20 border border-amber-600/40 items-center justify-center">
        <Text className="text-[#d97706] text-sm font-bold">3rd</Text>
      </View>
    );
  }
  return (
    <View className="w-10 h-10 items-center justify-center">
      <Text className="text-gray-400 text-base font-bold">{rank + 1}</Text>
    </View>
  );
};

export const HunterCard: React.FC<HunterCardProps> = ({ hunter, onClick, activeTab = "seasonal", rankIndex }) => {
  const getRankStyle = (rankName: string | undefined | null) => {
    if (!rankName) return 'border-gray-800';
    const rankStr = String(rankName);
    if (rankStr.includes('S-Rank') || rankStr.includes('Monarch')) return 'border-[#eab308]';
    if (rankStr.includes('A-Rank')) return 'border-[#a855f7]';
    if (rankStr.includes('B-Rank')) return 'border-[#2dd4bf]';
    return 'border-gray-800';
  };

  const jobColorText = getJobColor(hunter?.job_class || "");

  return (
    <TouchableOpacity
      onPress={onClick}
      activeOpacity={0.7}
      className={`relative p-5 rounded-3xl bg-[#13281f] border mb-4 ${getRankStyle(hunter?.rank_name)}`}
    >
      {/* Active today pulse indicator */}
      {hunter.is_active_today && (
        <View className="absolute top-4 right-4 h-2.5 w-2.5 rounded-full bg-[#22c55e]" />
      )}

      {/* Top Row: Rank Badge + Profile Info */}
      <View className="flex-row items-center mb-3">
        {rankIndex !== undefined && (
          <View className="mr-4">
            <RankBadge rank={rankIndex} />
          </View>
        )}
        <View className="flex-1">
          <View className="flex-row justify-between items-center mb-1 pr-4">
            <Text className="text-xs text-gray-400 font-bold uppercase">
              {hunter.rank_name}
            </Text>
            <Text className="text-sm font-bold text-[#2dd4bf]">
              Lv.{hunter.level}
            </Text>
          </View>

          <Text className="text-lg font-bold text-white mb-0.5 pr-2" numberOfLines={1}>
            {hunter.name}
          </Text>
          <Text className="text-xs text-gray-500">{hunter.user_id}</Text>
        </View>
      </View>

      {/* Job Class Badge */}
      <View className={`mb-4 self-start ${rankIndex !== undefined ? 'ml-[56px]' : ''}`}>
        <View className={`flex-row items-center px-3 py-1.5 rounded-full border ${jobColorText}`}>
          <Text className={`text-xs font-medium ${jobColorText}`}>{hunter.job_icon} {hunter.job_name}</Text>
        </View>
      </View>

      {/* Stats footer */}
      <View className="border-t border-gray-800/50 pt-4 flex-row justify-between items-center">
        {activeTab === "seasonal" && (
          <>
            <View className="flex-row items-center gap-1.5">
              <Flame size={16} color="#f97316" />
              <Text className="text-sm font-bold text-[#f97316]">{hunter.streak}w</Text>
            </View>
            <View className="flex-row items-center gap-1.5">
              <Activity size={16} color="#2dd4bf" />
              <Text className="text-sm font-bold text-[#2dd4bf]">{hunter.seasonal_activity_count}d</Text>
            </View>
            <View className="flex-row items-center gap-1.5">
              <Trophy size={16} color="#eab308" />
              <Text className="text-sm font-bold text-[#eab308]">{hunter.seasonal_points}p</Text>
            </View>
          </>
        )}

        {activeTab === "lifetime" && (
          <>
            <View className="flex-col items-center justify-center">
              <Text className="text-sm font-bold text-white">
                {hunter.level_icon} {hunter.level_name}
              </Text>
              <Text className="text-[10px] text-gray-500 mt-0.5">level</Text>
            </View>

            <View className="flex-col items-center justify-center">
              <View className="flex-row items-center gap-1">
                <Flame size={12} color="#f97316" />
                <Text className="text-sm font-bold text-gray-300">{hunter.max_streak}</Text>
              </View>
              <Text className="text-[10px] text-gray-500 mt-0.5">best streak</Text>
            </View>

            <View className="flex-col items-center justify-center">
              <Text className="text-sm font-bold text-gray-300">{hunter.total_active_days}</Text>
              <Text className="text-[10px] text-gray-500 mt-0.5">hari lifetime</Text>
            </View>

            <View className="flex-col items-center justify-center">
              <View className="flex-row items-center gap-1">
                <Award size={12} color="#eab308" />
                <Text className="text-sm font-bold text-[#eab308]">{hunter.total_points}</Text>
              </View>
              <Text className="text-[10px] text-gray-500 mt-0.5">total XP</Text>
            </View>
          </>
        )}

        {activeTab === "streak" && (
          <>
            <View className="flex-row items-center gap-1.5">
              <Activity size={16} color="#2dd4bf" />
              <Text className="text-sm font-bold text-[#2dd4bf]">{hunter.total_active_days}d Total</Text>
            </View>
            <View className="flex-row items-center gap-1.5">
              <Flame size={16} color="#f97316" />
              <Text className="text-sm font-bold text-[#f97316]">{hunter.max_streak}w Best</Text>
            </View>
            <View className="flex-row items-center gap-1.5">
              <Flame size={16} color="#eab308" />
              <Text className="text-sm font-bold text-[#eab308]">{hunter.streak}w Cur</Text>
            </View>
          </>
        )}

        {activeTab === "week" && (
          <>
            <View className="flex-row items-center gap-1.5">
              <Flame size={16} color="#f97316" />
              <Text className="text-sm font-bold text-[#f97316]">{hunter.streak}w</Text>
            </View>
            <View className="flex-row items-center gap-1.5">
              <Activity size={16} color="#2dd4bf" />
              <Text className="text-sm font-bold text-[#2dd4bf]">{hunter.week_active_days}d/7d</Text>
            </View>
            <View className="flex-row items-center gap-1.5">
              <CalendarDays size={16} color="#eab308" />
              <Text className="text-sm font-bold text-[#eab308]">{hunter.estimated_weekly_points}p est</Text>
            </View>
          </>
        )}
      </View>
    </TouchableOpacity>
  );
};

