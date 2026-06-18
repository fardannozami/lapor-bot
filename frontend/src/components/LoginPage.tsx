import { useState, type FormEvent } from 'react';
import { LogIn, ArrowLeft, Phone, User, AlertCircle, Loader2 } from 'lucide-react';
import type { EnrichedReport } from '../types';

interface LoginPageProps {
  onLoginSuccess: (user: EnrichedReport) => void;
  onBack: () => void;
}

export function LoginPage({ onLoginSuccess, onBack }: LoginPageProps) {
  const [countryCode, setCountryCode] = useState('62');
  const [phone, setPhone] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!countryCode || !/^\d+$/.test(countryCode)) {
      setError('Kode negara wajib diisi (angka saja).');
      return;
    }

    const digits = phone.replace(/\D/g, '');
    if (!digits) {
      setError('Nomor telepon wajib diisi.');
      return;
    }
    if (digits.length < 6 || digits.length > 14) {
      setError('Nomor harus 6-14 digit (tanpa kode negara).');
      return;
    }

    setLoading(true);
    try {
      const res = await fetch(`/api/user?phone=${countryCode}${digits}`);
      if (res.status === 404) {
        setError('User tidak ditemukan. Pastikan nomor sudah pernah lapor di grup.');
        return;
      }
      if (!res.ok) throw new Error('Gagal menghubungi server.');
      const data = (await res.json()) as EnrichedReport;
      onLoginSuccess(data);
    } catch {
      setError('Gagal menghubungi server. Coba lagi nanti.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[70vh] flex items-center justify-center px-4">
      <div className="w-full max-w-md glass rounded-3xl p-8 border border-system-green/20">
        <div className="flex items-center gap-3 mb-6">
          <button
            onClick={onBack}
            className="p-2 rounded-xl bg-gray-950 hover:bg-gray-900 border border-gray-800 text-gray-400 hover:text-white transition-colors"
            title="Kembali"
          >
            <ArrowLeft size={16} />
          </button>
          <div>
            <h2 className="text-lg font-bold font-orbitron text-white tracking-wide flex items-center gap-2">
              <User className="text-system-blue" size={18} />
              Akses Personal
            </h2>
            <p className="text-xs text-gray-500 font-mono mt-0.5 uppercase tracking-wider">
              Masuk dengan nomor HP yang sudah pernah lapor
            </p>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs text-gray-400 font-mono uppercase tracking-wider mb-1.5">
              Nomor Telepon
            </label>
            <div className="flex gap-2">
              <div className="relative shrink-0">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 text-sm pointer-events-none">
                  +
                </span>
                <input
                  type="text"
                  value={countryCode}
                  onChange={(e) => setCountryCode(e.target.value.replace(/\D/g, ''))}
                  placeholder="62"
                  maxLength={3}
                  className="w-16 pl-6 pr-2 py-2.5 rounded-xl bg-gray-950 border border-gray-800 text-white text-sm font-mono focus:outline-none focus:border-system-blue transition-colors"
                  disabled={loading}
                />
              </div>
              <div className="relative flex-1">
                <Phone className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" size={14} />
                <input
                  type="text"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value.replace(/\D/g, ''))}
                  placeholder="8xxxxxx"
                  maxLength={14}
                  className="w-full pl-9 pr-3 py-2.5 rounded-xl bg-gray-950 border border-gray-800 text-white text-sm font-mono focus:outline-none focus:border-system-blue transition-colors"
                  disabled={loading}
                  autoFocus
                />
              </div>
            </div>
          </div>

          {error && (
            <div className="p-3 rounded-xl bg-system-red/10 border border-system-red/35 flex items-start gap-2">
              <AlertCircle className="text-system-red mt-0.5 shrink-0" size={14} />
              <p className="text-xs text-red-300 font-mono">{error}</p>
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full flex items-center justify-center gap-2 px-5 py-3 rounded-xl bg-system-blue/10 hover:bg-system-blue/20 border border-system-blue/30 text-system-blue font-bold font-orbitron text-sm tracking-wider uppercase transition-colors disabled:opacity-50"
          >
            {loading ? (
              <Loader2 className="animate-spin" size={16} />
            ) : (
              <LogIn size={16} />
            )}
            {loading ? 'Mencari...' : 'Masuk'}
          </button>
        </form>

        <p className="text-[10px] text-gray-600 font-mono mt-5 text-center tracking-wide uppercase">
          Pastikan nomor HP kamu terdaftar di database laporan grup
        </p>
      </div>
    </div>
  );
}
