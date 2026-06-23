import React, { useState } from 'react';
import { View, Text, TouchableOpacity, TextInput, ActivityIndicator } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { LogIn, ArrowLeft, Phone, User, AlertCircle } from 'lucide-react-native';
import { useAuth } from '@lapor-bot/shared';
import type { EnrichedReport } from '@lapor-bot/shared';

interface LoginScreenProps {
  onLoginSuccess: (user: EnrichedReport) => void;
  onBack: () => void;
}

export const LoginScreen: React.FC<LoginScreenProps> = ({ onLoginSuccess, onBack }) => {
  const [countryCode, setCountryCode] = useState('62');
  const [phone, setPhone] = useState('');
  const [validationError, setValidationError] = useState<string | null>(null);
  const { login, loading, error: authError } = useAuth();

  const error = validationError || authError;

  const handleSubmit = async () => {
    setValidationError(null);

    if (!countryCode || !/^\d+$/.test(countryCode)) {
      setValidationError('Kode negara wajib diisi (angka saja).');
      return;
    }

    const digits = phone.replace(/\D/g, '');
    if (!digits) {
      setValidationError('Nomor telepon wajib diisi.');
      return;
    }
    if (digits.length < 6 || digits.length > 14) {
      setValidationError('Nomor harus 6-14 digit (tanpa kode negara).');
      return;
    }

    const user = await login(`${countryCode}${digits}`);
    if (user) {
      onLoginSuccess(user);
    }
  };

  return (
    <SafeAreaView className="flex-1 bg-[#07130c]">
      <View className="flex-1 justify-center px-4">
        <View className="bg-[#102018] rounded-3xl p-6 border border-[#22c55e]/20">
          <View className="flex-row items-center gap-3 mb-6">
            <TouchableOpacity
              onPress={onBack}
              className="p-2 rounded-xl bg-[#07130c] border border-gray-800"
            >
              <ArrowLeft size={16} color="#9ca3af" />
            </TouchableOpacity>
            <View>
              <View className="flex-row items-center gap-2">
                <User color="#2dd4bf" size={18} />
                <Text className="text-lg font-bold text-white tracking-wide">Akses Personal</Text>
              </View>
              <Text className="text-[10px] text-gray-500 font-mono mt-0.5 uppercase tracking-wider">
                Masuk dengan nomor HP yang sudah pernah lapor
              </Text>
            </View>
          </View>

          <View className="mb-4">
            <Text className="text-[10px] text-gray-400 font-mono uppercase tracking-wider mb-2">
              Nomor Telepon
            </Text>
            <View className="flex-row gap-2">
              <View className="relative w-20">
                <View className="absolute left-3 top-0 bottom-0 justify-center z-10">
                  <Text className="text-gray-400">+</Text>
                </View>
                <TextInput
                  value={countryCode}
                  onChangeText={(text) => setCountryCode(text.replace(/\D/g, ''))}
                  placeholder="62"
                  placeholderTextColor="#6b7280"
                  maxLength={3}
                  keyboardType="numeric"
                  editable={!loading}
                  className="w-full h-12 pl-7 pr-2 rounded-xl bg-[#07130c] border border-gray-800 text-white font-mono"
                />
              </View>
              <View className="relative flex-1">
                <View className="absolute left-3 top-0 bottom-0 justify-center z-10">
                  <Phone size={14} color="#6b7280" />
                </View>
                <TextInput
                  value={phone}
                  onChangeText={(text) => setPhone(text.replace(/\D/g, ''))}
                  placeholder="8xxxxxx"
                  placeholderTextColor="#6b7280"
                  maxLength={14}
                  keyboardType="numeric"
                  editable={!loading}
                  autoFocus
                  className="w-full h-12 pl-9 pr-3 rounded-xl bg-[#07130c] border border-gray-800 text-white font-mono"
                />
              </View>
            </View>
          </View>

          {error ? (
            <View className="p-3 mb-4 rounded-xl bg-[#f97316]/10 border border-[#f97316]/30 flex-row items-start gap-2">
              <AlertCircle color="#f97316" size={14} />
              <Text className="text-xs text-[#f97316] flex-1">{error}</Text>
            </View>
          ) : null}

          <TouchableOpacity
            onPress={handleSubmit}
            disabled={loading}
            className={`flex-row items-center justify-center gap-2 h-12 rounded-xl bg-[#2dd4bf]/10 border border-[#2dd4bf]/30 ${loading ? 'opacity-50' : ''}`}
          >
            {loading ? (
              <ActivityIndicator color="#2dd4bf" size="small" />
            ) : (
              <LogIn color="#2dd4bf" size={16} />
            )}
            <Text className="text-[#2dd4bf] font-bold text-sm tracking-wider uppercase">
              {loading ? 'Mencari...' : 'Masuk'}
            </Text>
          </TouchableOpacity>

          <Text className="text-[10px] text-gray-600 font-mono mt-5 text-center tracking-wide uppercase">
            Pastikan nomor HP kamu terdaftar di database laporan grup
          </Text>
        </View>
      </View>
    </SafeAreaView>
  );
};
