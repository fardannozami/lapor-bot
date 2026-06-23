import React from 'react';
import { View, Text, TouchableOpacity } from 'react-native';
import { Flame, Trophy, Activity, ArrowUpRight } from 'lucide-react-native';
import type { EnrichedReport } from '@lapor-bot/shared';
import { getJobColor } from '@lapor-bot/shared';

interface HunterCardProps {
  hunter: EnrichedReport;
  onClick: () => void;
}

export const HunterCard: React.FC<HunterCardProps> = ({ hunter, onClick }) => {
  const getRankStyle = (rankName: string) => {
    if (rankName.includes('S-Rank') || rankName.includes('Monarch')) return 'border-[#eab308]';
    if (rankName.includes('A-Rank')) return 'border-[#a855f7]';
    if (rankName.includes('B-Rank')) return 'border-[#2dd4bf]';
    return 'border-gray-800';
  };

  const jobColorText = getJobColor(hunter.job_class);

  return (
    <TouchableOpacity
      activeOpacity={0.8}
      onPress={onClick}
      className={`relative p-5 rounded-2xl bg-[#102018] border mb-4 ${getRankStyle(hunter.rank_name)}`}
    >
      {/* Active today pulse indicator */}
      {hunter.is_active_today && (
        <View className="absolute top-3 right-3 h-2.5 w-2.5 rounded-full bg-[#22c55e]" />
      )}

      {/* Card Content */}
      <View>
        {/* Level and Title */}
        <View className="flex-row justify-between items-center mb-3">
          <Text className="text-[10px] text-gray-500 font-bold uppercase">
            {hunter.rank_name}
          </Text>
          <Text className="text-xs font-bold text-[#2dd4bf]">
            Lv.{hunter.level}
          </Text>
        </View>

        {/* Profile Info */}
        <Text className="text-lg font-bold text-white mb-1 pr-6" numberOfLines={1}>
          {hunter.name}
        </Text>
        <Text className="text-[10px] text-gray-500 mb-4">{hunter.user_id}</Text>

        {/* Job Class Badge */}
        <View className="mb-4 self-start">
          <View className={`flex-row items-center px-2 py-1 rounded-full border ${jobColorText}`}>
            <Text className={`text-[10px] ${jobColorText}`}>{hunter.job_icon} {hunter.job_name}</Text>
          </View>
        </View>
      </View>

      {/* Stats footer */}
      <View className="border-t border-gray-900 pt-3 flex-row justify-between items-center">
        {/* Streak */}
        <View className="flex-row items-center gap-1">
          <Flame size={12} color="#f97316" />
          <Text className="text-xs font-bold text-[#f97316]">{hunter.streak}w</Text>
        </View>

        {/* Active Days */}
        <View className="flex-row items-center gap-1">
          <Activity size={12} color="#2dd4bf" />
          <Text className="text-xs font-bold text-[#2dd4bf]">{hunter.seasonal_activity_count}d</Text>
        </View>

        {/* Points */}
        <View className="flex-row items-center gap-1">
          <Trophy size={12} color="#eab308" />
          <Text className="text-xs font-bold text-[#eab308]">{hunter.seasonal_points}p</Text>
        </View>
      </View>
      
      <View className="absolute bottom-3 right-3">
        <ArrowUpRight size={14} color="#2dd4bf" />
      </View>
    </TouchableOpacity>
  );
};
