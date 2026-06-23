import React, { useState, useEffect } from 'react';
import { View, Text, TouchableOpacity, TextInput, ActivityIndicator, ScrollView } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { User, Compass, Target, ChevronRight, Check, ArrowLeft } from 'lucide-react-native';
import { useRepositories } from '@lapor-bot/shared';
import type { EnrichedReport } from '@lapor-bot/shared';

interface JobInfo {
  id: string;
  name: string;
  icon: string;
  description: string;
  trait: string;
}

interface ProfileSetupScreenProps {
  user: EnrichedReport;
  onComplete: (user: EnrichedReport) => void;
  onBack: () => void;
}

const STEPS = [
  { step: 1, label: "Nama", icon: User },
  { step: 2, label: "Job", icon: Compass },
  { step: 3, label: "Goal", icon: Target },
] as const;

export const ProfileSetupScreen: React.FC<ProfileSetupScreenProps> = ({ user, onComplete, onBack }) => {
  const { reports: repo } = useRepositories();
  const phone = user.user_id;

  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [name, setName] = useState(user.name ?? "");
  const [selectedJobId, setSelectedJobId] = useState<string | null>(user.job_class || null);
  const [targetDays, setTargetDays] = useState("3");
  const [activity, setActivity] = useState("Olahraga");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [jobs, setJobs] = useState<JobInfo[]>([]);
  const [jobsLoading, setJobsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    repo
      .listJobs()
      .then((data) => {
        if (!cancelled) setJobs(data);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : "Gagal memuat daftar job");
      })
      .finally(() => {
        if (!cancelled) setJobsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [repo]);

  const handleStep1 = async () => {
    const trimmed = name.trim();
    if (!trimmed) return;
    setError(null);
    setLoading(true);
    try {
      await repo.updateName(phone, trimmed);
      setStep(2);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal memperbarui nama");
    } finally {
      setLoading(false);
    }
  };

  const handleStep2 = async () => {
    setError(null);
    setLoading(true);
    try {
      if (selectedJobId) {
        await repo.selectJob(phone, selectedJobId);
      }
      setStep(3);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal memilih job");
    } finally {
      setLoading(false);
    }
  };

  const handleStep3 = async () => {
    setError(null);
    setLoading(true);
    try {
      await repo.setGoal(phone, Number(targetDays), activity.trim() || "Olahraga");
      const refreshedUser = await repo.fetchUserByPhone(phone);
      onComplete(refreshedUser);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Gagal menyimpan goal");
    } finally {
      setLoading(false);
    }
  };

  const handleBack = () => {
    if (step === 1) {
      onBack();
    } else {
      setStep((prev) => (prev - 1) as 1 | 2 | 3);
      setError(null);
    }
  };

  return (
    <SafeAreaView className="flex-1 bg-[#07130c]">
      <ScrollView className="flex-1 px-4 py-6" contentContainerStyle={{ paddingBottom: 40 }}>
        {/* Back button */}
        <TouchableOpacity
          onPress={handleBack}
          disabled={loading}
          className={`flex-row items-center gap-2 self-start px-3 py-2 rounded-xl bg-[#07130c] border border-gray-800 mb-6 ${loading ? 'opacity-50' : ''}`}
        >
          <ArrowLeft size={14} color="#9ca3af" />
          <Text className="text-gray-400 font-mono text-xs">
            {step === 1 ? "Kembali ke Login" : "Sebelumnya"}
          </Text>
        </TouchableOpacity>

        {/* Progress indicator */}
        <View className="flex-row items-center justify-center mb-8">
          {STEPS.map((s, idx) => (
            <React.Fragment key={s.step}>
              <View className="items-center gap-2">
                <View
                  className={`w-10 h-10 rounded-full items-center justify-center border-2 ${
                    step >= s.step
                      ? "border-[#22c55e] bg-[#22c55e]/20"
                      : "border-gray-700 bg-[#07130c]"
                  }`}
                >
                  {step > s.step ? (
                    <Check size={18} color="#22c55e" />
                  ) : (
                    <s.icon size={18} color={step >= s.step ? "#22c55e" : "#4b5563"} />
                  )}
                </View>
                <Text
                  className={`text-[10px] font-mono uppercase tracking-wider ${
                    step >= s.step ? "text-[#22c55e]" : "text-gray-600"
                  }`}
                >
                  {s.label}
                </Text>
              </View>
              {idx < STEPS.length - 1 && (
                <View
                  className={`w-10 h-0.5 mt-[-18px] ${
                    step > s.step ? "bg-[#22c55e]" : "bg-gray-800"
                  }`}
                />
              )}
            </React.Fragment>
          ))}
        </View>

        {/* Error display */}
        {error && (
          <View className="mb-6 p-4 rounded-2xl bg-[#f97316]/10 border border-[#f97316]/30">
            <Text className="font-bold font-mono uppercase text-xs text-[#f97316] mb-1">Error</Text>
            <Text className="text-xs text-[#f97316]">{error}</Text>
          </View>
        )}

        {/* Step 1: Name */}
        {step === 1 && (
          <View className="bg-[#102018] rounded-3xl p-6 border border-gray-800 overflow-hidden">
            <View className="flex-row items-center gap-3 mb-6">
              <View className="w-10 h-10 rounded-xl bg-[#22c55e]/15 border border-[#22c55e]/30 items-center justify-center">
                <User size={20} color="#22c55e" />
              </View>
              <Text className="text-xl font-bold text-white">Nama Hunter</Text>
            </View>
            <Text className="text-xs text-gray-500 font-mono mb-6">
              Nama ini akan ditampilkan di leaderboard dan profil personal kamu.
            </Text>
            <View className="mb-6">
              <Text className="text-[10px] text-gray-500 font-mono uppercase tracking-wider mb-2">
                Nama Kamu
              </Text>
              <TextInput
                value={name}
                onChangeText={setName}
                placeholder="Masukkan nama hunter..."
                placeholderTextColor="#6b7280"
                className="w-full px-4 py-3 h-12 rounded-2xl bg-[#07130c] border border-gray-800 text-white font-mono text-sm"
              />
            </View>
            <TouchableOpacity
              onPress={handleStep1}
              disabled={loading || !name.trim()}
              className={`flex-row items-center justify-center gap-2 h-12 rounded-2xl bg-[#22c55e] ${
                loading || !name.trim() ? "opacity-50" : ""
              }`}
            >
              {loading ? (
                <ActivityIndicator color="#07130c" size="small" />
              ) : (
                <>
                  <Text className="text-[#07130c] font-bold text-sm">Lanjut</Text>
                  <ChevronRight size={18} color="#07130c" />
                </>
              )}
            </TouchableOpacity>
          </View>
        )}

        {/* Step 2: Job Selection */}
        {step === 2 && (
          <View className="bg-[#102018] rounded-3xl p-6 border border-gray-800">
            <View className="flex-row items-center gap-3 mb-6">
              <View className="w-10 h-10 rounded-xl bg-[#eab308]/15 border border-[#eab308]/30 items-center justify-center">
                <Compass size={20} color="#eab308" />
              </View>
              <Text className="text-xl font-bold text-white">Pilih Job</Text>
            </View>
            <Text className="text-xs text-gray-500 font-mono mb-6">
              Job menentukan peran hunter kamu. Pilih satu, atau skip untuk melanjutkan.
            </Text>

            {jobsLoading ? (
              <View className="items-center py-12">
                <ActivityIndicator size="large" color="#22c55e" />
              </View>
            ) : (
              <View className="flex-row flex-wrap justify-between mb-6">
                {jobs.map((job) => {
                  const isSelected = selectedJobId === job.id;
                  return (
                    <TouchableOpacity
                      key={job.id}
                      onPress={() => setSelectedJobId(job.id)}
                      className={`w-[48%] p-4 rounded-2xl border mb-3 ${
                        isSelected ? "border-[#eab308] bg-[#eab308]/10" : "border-gray-800 bg-[#07130c]"
                      }`}
                    >
                      <Text className="text-2xl mb-2">{job.icon}</Text>
                      <Text className="text-sm font-bold text-white mb-1">{job.name}</Text>
                      <Text className="text-[10px] text-gray-500 font-mono mb-2" numberOfLines={3}>
                        {job.description}
                      </Text>
                      <Text className="text-[10px] text-[#eab308] font-mono uppercase tracking-wider">
                        {job.trait}
                      </Text>
                    </TouchableOpacity>
                  );
                })}
              </View>
            )}

            <View className="flex-row gap-3">
              <TouchableOpacity
                onPress={() => setStep(3)}
                disabled={loading}
                className="flex-1 h-12 items-center justify-center rounded-2xl bg-[#07130c] border border-gray-800"
              >
                <Text className="text-gray-400 font-mono text-xs">Skip</Text>
              </TouchableOpacity>
              <TouchableOpacity
                onPress={handleStep2}
                disabled={loading || jobsLoading}
                className={`flex-1 flex-row items-center justify-center gap-2 h-12 rounded-2xl bg-[#eab308] ${
                  loading || jobsLoading ? "opacity-50" : ""
                }`}
              >
                {loading ? (
                  <ActivityIndicator color="#07130c" size="small" />
                ) : (
                  <>
                    <Text className="text-[#07130c] font-bold text-sm">Lanjut</Text>
                    <ChevronRight size={18} color="#07130c" />
                  </>
                )}
              </TouchableOpacity>
            </View>
          </View>
        )}

        {/* Step 3: Goal Setup */}
        {step === 3 && (
          <View className="bg-[#102018] rounded-3xl p-6 border border-gray-800">
            <View className="flex-row items-center gap-3 mb-6">
              <View className="w-10 h-10 rounded-xl bg-[#a855f7]/15 border border-[#a855f7]/30 items-center justify-center">
                <Target size={20} color="#a855f7" />
              </View>
              <Text className="text-xl font-bold text-white">Setup Goal</Text>
            </View>
            <Text className="text-xs text-gray-500 font-mono mb-6">
              Tetapkan target mingguan untuk konsistensi latihan kamu.
            </Text>

            <View className="mb-6">
              <Text className="text-[10px] text-gray-500 font-mono uppercase tracking-wider mb-2">
                Target Hari / Minggu (1-7)
              </Text>
              <TextInput
                value={targetDays}
                onChangeText={setTargetDays}
                placeholder="3"
                placeholderTextColor="#6b7280"
                keyboardType="numeric"
                className="w-full px-4 h-12 rounded-2xl bg-[#07130c] border border-gray-800 text-white font-mono text-sm mb-4"
              />

              <Text className="text-[10px] text-gray-500 font-mono uppercase tracking-wider mb-2">
                Aktivitas
              </Text>
              <TextInput
                value={activity}
                onChangeText={setActivity}
                placeholder="Contoh: Olahraga, Lari, Gym..."
                placeholderTextColor="#6b7280"
                className="w-full px-4 h-12 rounded-2xl bg-[#07130c] border border-gray-800 text-white font-mono text-sm"
              />
            </View>

            <TouchableOpacity
              onPress={handleStep3}
              disabled={loading}
              className={`flex-row items-center justify-center gap-2 h-12 rounded-2xl bg-[#a855f7] ${
                loading ? "opacity-50" : ""
              }`}
            >
              {loading ? (
                <ActivityIndicator color="#fff" size="small" />
              ) : (
                <>
                  <Check size={18} color="#fff" />
                  <Text className="text-white font-bold text-sm">Selesai</Text>
                </>
              )}
            </TouchableOpacity>
          </View>
        )}
      </ScrollView>
    </SafeAreaView>
  );
};
