import React, { useState, useMemo } from 'react';
import { View, Text, TouchableOpacity, FlatList, TextInput } from 'react-native';
import { Search, Trophy, Award, Flame, CalendarDays } from 'lucide-react-native';
import type { EnrichedReport } from '@lapor-bot/shared';
import { HunterCard } from './HunterCard';
import { ErrorBoundary } from './ErrorBoundary';

interface LeaderboardListProps {
  hunters: EnrichedReport[];
  onSelectHunter: (hunter: EnrichedReport) => void;
  ListHeaderComponent?: React.ReactElement;
  StatsComponent?: React.ReactElement;
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
  StatsComponent,
  refreshing = false,
  onRefresh,
}) => {
  const [search, setSearch] = useState("");
  const [activeTab, setActiveTab] = useState<TabType>("seasonal");

  const filteredAndSorted = useMemo(() => {
    const uniqueMap = new Map();
    hunters.forEach(h => {
      if (!h) return;
      if (!uniqueMap.has(h.user_id)) {
        uniqueMap.set(h.user_id, h);
      }
    });
    let result = Array.from(uniqueMap.values());

    if (search.trim()) {
      const q = search.toLowerCase();
      result = result.filter((h) => String(h.name || "").toLowerCase().includes(q));
    }

    if (activeTab === "seasonal") {
      result = result.filter((h) => (h.seasonal_points || 0) > 0 || (h.seasonal_activity_count || 0) > 0);
    }
    if (activeTab === "week") {
      result = result.filter((h) => (h.week_active_days || 0) > 0);
    }

    result.sort((a, b) => {
      if (activeTab === "seasonal") {
        if (b.seasonal_points === a.seasonal_points) {
          if (b.seasonal_activity_count === a.seasonal_activity_count) {
            return String(a.name || "").localeCompare(String(b.name || ""));
          }
          return (b.seasonal_activity_count || 0) - (a.seasonal_activity_count || 0);
        }
        return (b.seasonal_points || 0) - (a.seasonal_points || 0);
      } else if (activeTab === "lifetime") {
        if (b.total_points === a.total_points) {
          return (b.activity_count || 0) - (a.activity_count || 0);
        }
        return (b.total_points || 0) - (a.total_points || 0);
      } else if (activeTab === "streak") {
        if (b.streak === a.streak) {
          return (b.max_streak || 0) - (a.max_streak || 0);
        }
        return (b.streak || 0) - (a.streak || 0);
      } else {
        if (b.week_active_days === a.week_active_days) {
          return String(a.name || "").localeCompare(String(b.name || ""));
        }
        return (b.week_active_days || 0) - (a.week_active_days || 0);
      }
    });

    return result;
  }, [hunters, search, activeTab]);

  return (
    <View className="flex-1">
      <FlatList
        data={filteredAndSorted}
        keyExtractor={(item, index) => `${item?.user_id || index}-${index}`}
        refreshing={refreshing}
        onRefresh={onRefresh}
        ListHeaderComponent={
          <View>
            {ListHeaderComponent}
            {/* Tabs / Segmented Control */}
            <View className="flex-row justify-between mb-6 bg-[#1a2f24] p-1 rounded-lg">
              {LEADERBOARD_TABS.map((tab) => {
                const active = activeTab === tab.id;
                return (
                  <TouchableOpacity
                    key={tab.id}
                    activeOpacity={0.7}
                    onPress={() => {
                      setActiveTab(tab.id);
                      setSearch("");
                    }}
                    className={`flex-1 items-center justify-center py-2 rounded-md ${
                      active ? "bg-[#2dd4bf]" : "bg-transparent"
                    }`}
                  >
                    <Text className={`text-xs font-semibold ${active ? "text-black" : "text-gray-400"}`}>
                      {tab.label}
                    </Text>
                  </TouchableOpacity>
                );
              })}
            </View>

            {StatsComponent}

            {/* Search */}
            <View className="relative mb-6">
              <View className="absolute left-4 top-4 z-10">
                <Search size={20} color="#6b7280" />
              </View>
              <TextInput
                value={search}
                onChangeText={setSearch}
                placeholder="Search hunter..."
                placeholderTextColor="#6b7280"
                className="w-full bg-[#13281f] border border-gray-800/50 rounded-2xl pl-12 pr-4 py-4 text-white text-sm"
              />
            </View>
          </View>
        }
        renderItem={({ item }) => (
          <ErrorBoundary>
            <HunterCard hunter={item} onClick={() => onSelectHunter(item)} />
          </ErrorBoundary>
        )}
        showsVerticalScrollIndicator={false}
        contentContainerStyle={{ paddingBottom: 100 }}
        ListEmptyComponent={() => (
          <View className="py-10 items-center">
            <Text className="text-gray-500 text-sm text-center">No active hunters match the filters.</Text>
          </View>
        )}
      />
    </View>
  );
};
