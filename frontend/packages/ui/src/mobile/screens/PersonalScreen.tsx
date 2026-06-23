import React, { useState, useEffect } from 'react';
import { View, Text, TouchableOpacity, ScrollView, ActivityIndicator, TextInput } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { ArrowLeft, LogOut, Flame, Target, CalendarDays, Edit2, CheckCircle2, Circle } from 'lucide-react-native';
import { useRepositories } from '@lapor-bot/shared';
import type { EnrichedReport } from '@lapor-bot/shared';

interface PersonalScreenProps {
  user: EnrichedReport;
  onLogout: () => void;
  onUserRefresh?: (user: EnrichedReport) => void;
}

export const PersonalScreen: React.FC<PersonalScreenProps> = ({ user, onLogout, onUserRefresh }) => {
  const { reports: repo } = useRepositories();
  const phone = user.user_id;

  const [localUser, setLocalUser] = useState<EnrichedReport>(user);
  const [editingName, setEditingName] = useState(false);
  const [nameValue, setNameValue] = useState(user.name || "");
  const [loading, setLoading] = useState(false);

  const [showGoalForm, setShowGoalForm] = useState(false);
  const [targetDays, setTargetDays] = useState(user.active_goal?.target_days?.toString() || "3");
  const [activity, setActivity] = useState(user.active_goal?.activity || "Olahraga");
  const [goalLoading, setGoalLoading] = useState(false);

  useEffect(() => {
    setLocalUser(user);
    setNameValue(user.name || "");
  }, [user]);

  const refreshUser = async () => {
    try {
      const refreshed = await repo.fetchUserByPhone(phone);
      setLocalUser(refreshed);
      onUserRefresh?.(refreshed);
    } catch (err) {
      console.error(err);
    }
  };

  const handleUpdateName = async () => {
    const v = nameValue.trim();
    if (!v) return;
    setLoading(true);
    try {
      await repo.updateName(phone, v);
      await refreshUser();
      setEditingName(false);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const handleSetGoal = async () => {
    setGoalLoading(true);
    try {
      await repo.setGoal(phone, Number(targetDays), activity || "Olahraga");
      await refreshUser();
      setShowGoalForm(false);
    } catch (e) {
      console.error(e);
    } finally {
      setGoalLoading(false);
    }
  };

  const xpPercent = Math.min(100, Math.round((localUser.xp_progress.CurrentXP / localUser.xp_progress.RequiredXP) * 100));
  const activeGoal = localUser.active_goal;

  return (
    <SafeAreaView className="flex-1 bg-[#07130c]">
      <ScrollView className="flex-1 px-4 py-6" contentContainerStyle={{ paddingBottom: 40 }}>
        {/* Header */}
        <View className="flex-row justify-between items-center mb-6">
          <TouchableOpacity onPress={onLogout} className="flex-row items-center gap-2 px-3 py-2 rounded-xl bg-[#07130c] border border-gray-800">
            <ArrowLeft size={14} color="#9ca3af" />
            <Text className="text-gray-400 font-mono text-xs">Kembali</Text>
          </TouchableOpacity>

          <TouchableOpacity onPress={onLogout} className="flex-row items-center gap-2 px-3 py-2 rounded-xl bg-[#f97316]/10 border border-[#f97316]/20">
            <LogOut size={14} color="#f97316" />
            <Text className="text-[#f97316] font-mono text-xs">Keluar</Text>
          </TouchableOpacity>
        </View>

        {/* Profile Card */}
        <View className="bg-[#102018] rounded-3xl p-6 mb-6 border border-[#2dd4bf]/30">
          <View className="flex-row items-start gap-4 mb-6">
            <View className="w-16 h-16 rounded-3xl bg-[#07130c] border border-gray-800 items-center justify-center">
              <Text className="text-3xl">{localUser.job_icon}</Text>
            </View>
            <View className="flex-1">
              <Text className="text-[10px] text-[#22c55e] font-mono uppercase tracking-widest">Personal Hunter Profile</Text>
              
              <View className="flex-row items-center gap-2 mt-1">
                {editingName ? (
                  <View className="flex-1 flex-row items-center gap-2">
                    <TextInput
                      value={nameValue}
                      onChangeText={setNameValue}
                      className="flex-1 h-8 px-2 rounded-lg bg-[#07130c] border border-gray-800 text-white font-mono text-sm"
                    />
                    <TouchableOpacity onPress={handleUpdateName} disabled={loading} className="bg-[#22c55e] px-3 py-1.5 rounded-lg">
                      {loading ? <ActivityIndicator size="small" color="#07130c" /> : <Text className="text-[#07130c] font-bold text-xs">Simpan</Text>}
                    </TouchableOpacity>
                  </View>
                ) : (
                  <>
                    <Text className="text-2xl font-bold text-white tracking-wide" numberOfLines={1}>{localUser.name}</Text>
                    <TouchableOpacity onPress={() => setEditingName(true)} className="p-1 rounded-md border border-gray-800">
                      <Edit2 size={12} color="#9ca3af" />
                    </TouchableOpacity>
                  </>
                )}
              </View>

              <Text className="text-sm text-gray-400 font-mono mt-1">
                {localUser.job_name} {localUser.level_icon} Lv.{localUser.level} • {localUser.rank_icon} {localUser.rank_name}
              </Text>
              <Text className="text-[10px] text-gray-600 font-mono mt-1">{localUser.user_id}</Text>
            </View>
          </View>

          {/* Level Progress */}
          <View className="bg-[#07130c] rounded-2xl p-4 border border-gray-800">
            <View className="flex-row justify-between mb-3">
              <View>
                <Text className="text-[10px] text-gray-500 font-mono uppercase">Level Progress</Text>
                <Text className="text-sm font-bold text-white mt-1">{localUser.level_icon} {localUser.level_name}</Text>
              </View>
              <View className="items-end">
                <Text className="text-lg font-bold text-[#eab308]">{localUser.total_points}</Text>
                <Text className="text-[10px] text-gray-500 font-mono uppercase">Lifetime XP</Text>
              </View>
            </View>
            <View className="h-2 rounded-full bg-gray-900 border border-gray-800 overflow-hidden">
              <View className="h-full bg-[#2dd4bf]" style={{ width: `${xpPercent}%` }} />
            </View>
            <View className="flex-row justify-between mt-2">
              <Text className="text-[10px] text-gray-500 font-mono">{localUser.xp_progress.CurrentXP} XP</Text>
              <Text className="text-[10px] text-[#2dd4bf] font-bold font-mono">{xpPercent}%</Text>
              <Text className="text-[10px] text-gray-500 font-mono">{localUser.xp_progress.RequiredXP} XP</Text>
            </View>
          </View>
        </View>

        {/* Stats Grid */}
        <View className="flex-row flex-wrap justify-between mb-6">
          <View className="w-[48%] bg-[#07130c] rounded-2xl p-4 border border-gray-800 mb-4">
            <Text className="text-xl font-bold text-[#eab308]">{localUser.seasonal_points}</Text>
            <Text className="text-[10px] text-gray-500 font-mono uppercase mt-1">Season Points</Text>
          </View>
          <View className="w-[48%] bg-[#07130c] rounded-2xl p-4 border border-gray-800 mb-4">
            <Text className="text-xl font-bold text-[#f97316]">{localUser.current_daily_streak ?? 0} hari</Text>
            <Text className="text-[10px] text-gray-500 font-mono uppercase mt-1">Daily Streak</Text>
          </View>
          <View className="w-[48%] bg-[#07130c] rounded-2xl p-4 border border-gray-800">
            <Text className="text-xl font-bold text-[#22c55e]">{localUser.streak} minggu</Text>
            <Text className="text-[10px] text-gray-500 font-mono uppercase mt-1">Weekly Streak</Text>
          </View>
          <View className="w-[48%] bg-[#07130c] rounded-2xl p-4 border border-gray-800">
            <Text className="text-xl font-bold text-[#2dd4bf]">{localUser.active_days_in_window ?? 0}</Text>
            <Text className="text-[10px] text-gray-500 font-mono uppercase mt-1">Active Window</Text>
          </View>
        </View>

        {/* Weekly Goal */}
        <View className="bg-[#102018] rounded-3xl p-6 border border-gray-800">
          <View className="flex-row justify-between mb-5">
            <View>
              <View className="flex-row items-center gap-2">
                <Target size={18} color="#eab308" />
                <Text className="text-lg font-bold text-white">Weekly Goal</Text>
              </View>
              <Text className="text-xs text-gray-500 font-mono mt-1">Goal mingguan pribadi.</Text>
            </View>
            {activeGoal && (
              <View className="items-end">
                <Text className={`text-sm font-bold ${activeGoal.is_completed ? 'text-[#eab308]' : 'text-white'}`}>
                  {activeGoal.is_completed ? '✅ Selesai' : `${activeGoal.completed_days}/${activeGoal.target_days}`}
                </Text>
                <Text className="text-[10px] text-gray-500 font-mono uppercase">Progress</Text>
              </View>
            )}
          </View>

          {!activeGoal ? (
            <View className="bg-[#07130c] rounded-2xl p-4 border border-gray-800 items-center">
              <CalendarDays size={22} color="#4b5563" className="mb-2" />
              <Text className="text-sm text-gray-400 font-mono">Belum ada goal aktif.</Text>
            </View>
          ) : activeGoal.is_completed ? (
            <View className="bg-[#eab308]/10 rounded-2xl p-4 border border-[#eab308]/40 items-center mb-4">
              <Text className="text-lg font-bold text-[#eab308]">🎉 GOAL TERCAPAI! 🎉</Text>
              <Text className="text-xs text-gray-300 font-mono mt-1 text-center">
                {activeGoal.target_days}x {activeGoal.activity} berhasil diselesaikan!
              </Text>
            </View>
          ) : (
            <View className="bg-[#07130c] rounded-2xl p-4 border border-gray-800 mb-4">
              <View className="flex-row justify-between mb-2">
                <Text className="text-sm text-white font-bold font-mono">{activeGoal.target_days}x {activeGoal.activity}</Text>
                <Text className="text-xs text-[#eab308] font-mono">{activeGoal.percent}%</Text>
              </View>
              <View className="h-2 rounded-full bg-gray-900 overflow-hidden">
                <View className="h-full bg-[#eab308]" style={{ width: `${activeGoal.percent}%` }} />
              </View>
              <Text className="mt-2 text-[10px] text-gray-500 font-mono uppercase">
                Sisa {activeGoal.remaining_days} hari lagi.
              </Text>
            </View>
          )}

          <TouchableOpacity
            onPress={() => setShowGoalForm(!showGoalForm)}
            className="mt-4 py-3 rounded-xl border border-gray-700 items-center"
          >
            <Text className="text-gray-300 text-xs">{showGoalForm ? "Batal" : "Atur / Ubah Goal"}</Text>
          </TouchableOpacity>

          {showGoalForm && (
            <View className="mt-4 bg-[#07130c] rounded-2xl p-4 border border-gray-800">
              <Text className="text-[10px] text-gray-500 font-mono mb-2">Target Hari (1-7)</Text>
              <TextInput
                value={targetDays}
                onChangeText={setTargetDays}
                keyboardType="numeric"
                className="w-full h-10 px-3 rounded-xl bg-[#102018] border border-gray-800 text-white font-mono text-xs mb-3"
              />
              <Text className="text-[10px] text-gray-500 font-mono mb-2">Aktivitas</Text>
              <TextInput
                value={activity}
                onChangeText={setActivity}
                className="w-full h-10 px-3 rounded-xl bg-[#102018] border border-gray-800 text-white font-mono text-xs mb-4"
              />
              <TouchableOpacity
                onPress={handleSetGoal}
                disabled={goalLoading}
                className="w-full py-3 rounded-xl bg-[#eab308] items-center"
              >
                {goalLoading ? <ActivityIndicator size="small" color="#07130c" /> : <Text className="text-[#07130c] font-bold">Simpan Goal</Text>}
              </TouchableOpacity>
            </View>
          )}
        </View>

      </ScrollView>
    </SafeAreaView>
  );
};
