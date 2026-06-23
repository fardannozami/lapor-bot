import React from 'react';
import { View, Text } from 'react-native';
import { Users, Flame, Activity, Shield } from 'lucide-react-native';
import type { GlobalSummary } from '@lapor-bot/shared';

interface StatsOverviewProps {
  summary: GlobalSummary | null;
  loading: boolean;
}

type StatTone = "blue" | "red" | "green" | "gold";

interface StatItem {
  key: string;
  label: string;
  value: string | number;
  icon: React.ElementType;
  tone: StatTone;
  description: string;
}

const TONE_STYLES: Record<StatTone, { text: string; bg: string }> = {
  blue: { text: "text-[#2dd4bf]", bg: "bg-[#2dd4bf]/20" },
  red: { text: "text-[#f97316]", bg: "bg-[#f97316]/20" },
  green: { text: "text-[#22c55e]", bg: "bg-[#22c55e]/20" },
  gold: { text: "text-[#eab308]", bg: "bg-[#eab308]/20" },
};

const ICON_COLORS: Record<StatTone, string> = {
  blue: "#2dd4bf",
  red: "#f97316",
  green: "#22c55e",
  gold: "#eab308",
};

export const StatsOverview: React.FC<StatsOverviewProps> = ({ summary, loading }) => {
  const stats: StatItem[] = [
    {
      key: "hunters",
      label: "Total Hunters Active",
      value: loading ? "..." : summary?.total_participants ?? 0,
      icon: Users,
      tone: "blue",
      description: "Hunters who reported this season",
    },
    {
      key: "streak",
      label: "Keep Streak 🔥",
      value: loading ? "..." : `${summary?.active_streak_count ?? 0} Players`,
      icon: Flame,
      tone: "red",
      description: "Active streak holders this week",
    },
    {
      key: "workouts",
      label: "Total Workouts Logged",
      value: loading ? "..." : summary?.total_workouts_logged ?? 0,
      icon: Activity,
      tone: "green",
      description: "Total workouts in lifetime history",
    },
    {
      key: "season",
      label: "System Current State",
      value: loading ? "..." : `Season ${summary?.current_season ?? 1} - Day ${summary?.current_day ?? 1}`,
      icon: Shield,
      tone: "gold",
      description: "Active campaign stage progress",
    },
  ];

  return (
    <View className="flex-col gap-4 mb-6">
      {stats.map((stat) => {
        const Icon = stat.icon;
        const tone = TONE_STYLES[stat.tone];
        return (
          <View
            key={stat.key}
            className="p-5 rounded-2xl bg-[#102018] border border-[#23402e]"
          >
            <View className="flex-row items-start justify-between">
              <View className="flex-1">
                <Text className="text-gray-400 text-xs font-medium">{stat.label}</Text>
                <Text className="text-2xl font-bold mt-1 text-white">
                  {stat.value}
                </Text>
              </View>
              <View className={`p-3 rounded-xl ${tone.bg}`}>
                <Icon size={24} color={ICON_COLORS[stat.tone]} />
              </View>
            </View>
            <Text className="text-[10px] text-gray-500 mt-3">{stat.description}</Text>
          </View>
        );
      })}
    </View>
  );
};
