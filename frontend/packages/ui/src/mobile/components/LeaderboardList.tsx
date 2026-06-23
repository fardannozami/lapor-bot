import React, { useState, useMemo } from 'react';
import { View, Text, TouchableOpacity, FlatList, TextInput } from 'react-native';
import { Search, Trophy, Award, Flame, CalendarDays } from 'lucide-react-native';
import type { EnrichedReport } from '@lapor-bot/shared';
import { HunterCard } from './HunterCard';

interface LeaderboardListProps {
  hunters: EnrichedReport[];
  onSelectHunter: (hunter: EnrichedReport) => void;
  ListHeaderComponent?: React.ReactElement;
  refreshing?: boolean;
  onRefresh?: () => void;
}

type TabType = "seasonal" | "lifetime" | "streak" | "week";

const LEADERBOARD_TABS: { id: TabType; label: string; icon: any }[] = [
  { id: "seasonal", label: "Season", icon: Trophy },
  { id: "lifetime", label: "Lifetime", icon: Award },
  { id: "streak", label: "Streak", icon: Flame },
  { id: "week", label: "Week", icon: CalendarDays },
];

export const LeaderboardList: React.FC<LeaderboardListProps> = ({ 
  hunters, 
  onSelectHunter, 
  ListHeaderComponent,
  refreshing = false,
  onRefresh,
}) => {
  const [search, setSearch] = useState("");
  const [activeTab, setActiveTab] = useState<TabType>("seasonal");

  const filteredAndSorted = useMemo(() => {
    // Deduplicate hunters by user_id to prevent duplicate rows in the UI
    const uniqueMap = new Map();
    hunters.forEach(h => {
      if (!uniqueMap.has(h.user_id)) {
        uniqueMap.set(h.user_id, h);
      }
    });
    let result = Array.from(uniqueMap.values());

    if (search.trim()) {
      const q = search.toLowerCase();
      result = result.filter((h) => h.name.toLowerCase().includes(q));
    }

    if (activeTab === "seasonal") {
      result = result.filter((h) => h.seasonal_points > 0 || h.seasonal_activity_count > 0);
    }
    if (activeTab === "week") {
      result = result.filter((h) => h.week_active_days > 0);
    }

    result.sort((a, b) => {
      if (activeTab === "seasonal") {
        if (b.seasonal_points === a.seasonal_points) {
          if (b.seasonal_activity_count === a.seasonal_activity_count) {
            return a.name.localeCompare(b.name);
          }
          return b.seasonal_activity_count - a.seasonal_activity_count;
        }
        return b.seasonal_points - a.seasonal_points;
      } else if (activeTab === "lifetime") {
        if (b.total_points === a.total_points) {
          return b.activity_count - a.activity_count;
        }
        return b.total_points - a.total_points;
      } else if (activeTab === "streak") {
        if (b.streak === a.streak) {
          return b.max_streak - a.max_streak;
        }
        return b.streak - a.streak;
      } else {
        if (b.week_active_days === a.week_active_days) {
          return a.name.localeCompare(b.name);
        }
        return b.week_active_days - a.week_active_days;
      }
    });

    return result;
  }, [hunters, search, activeTab]);

  return (
    <View className="flex-1">
      <FlatList
        data={filteredAndSorted}
        keyExtractor={(item, index) => `${item.user_id}-${index}`}
        refreshing={refreshing}
        onRefresh={onRefresh}
        ListHeaderComponent={
          <View>
            {ListHeaderComponent}
            {/* Tabs */}
            <View className="flex-row justify-between mb-4 bg-[#102018] p-1 rounded-xl">
              {LEADERBOARD_TABS.map((tab) => {
                const Icon = tab.icon;
                const active = activeTab === tab.id;
                return (
                  <TouchableOpacity
                    key={tab.id}
                    onPress={() => {
                      setActiveTab(tab.id);
                      setSearch("");
                    }}
                    className={`flex-1 flex-row items-center justify-center gap-1 py-2 rounded-lg ${
                      active ? "bg-[#1f3a2c]" : ""
                    }`}
                  >
                    <Icon size={12} color={active ? "#2dd4bf" : "#6b7280"} />
                    <Text className={`text-[10px] font-bold uppercase ${active ? "text-[#2dd4bf]" : "text-gray-500"}`}>
                      {tab.label}
                    </Text>
                  </TouchableOpacity>
                );
              })}
            </View>

            {/* Search */}
            <View className="relative mb-6">
              <View className="absolute left-3 top-3 z-10">
                <Search size={16} color="#6b7280" />
              </View>
              <TextInput
                value={search}
                onChangeText={setSearch}
                placeholder="Search hunter..."
                placeholderTextColor="#6b7280"
                className="w-full bg-[#102018] border border-gray-800 rounded-xl pl-10 pr-4 py-3 text-white text-xs"
              />
            </View>
          </View>
        }
        renderItem={({ item }) => (
          <HunterCard hunter={item} onClick={() => onSelectHunter(item)} />
        )}
        showsVerticalScrollIndicator={false}
        contentContainerStyle={{ paddingBottom: 100 }}
        ListEmptyComponent={() => (
          <View className="py-10 items-center">
            <Text className="text-gray-500 text-xs text-center">No active hunters match the filters.</Text>
          </View>
        )}
      />
    </View>
  );
};
