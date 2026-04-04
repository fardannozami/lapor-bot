import { useWorkouts } from '../hooks/useWorkouts'
import { useNavigate } from 'react-router-dom'
import { Dumbbell, Flame, TrendingUp, ChevronRight, Plus, Calendar } from 'lucide-react'

export default function Dashboard() {
  const { getRecentWorkouts, getStats } = useWorkouts()
  const navigate = useNavigate()
  const recent = getRecentWorkouts(5)
  const stats = getStats()

  const formatDate = (iso) => {
    const d = new Date(iso)
    const today = new Date()
    const yesterday = new Date(today)
    yesterday.setDate(yesterday.getDate() - 1)

    if (d.toDateString() === today.toDateString()) return 'Hari ini'
    if (d.toDateString() === yesterday.toDateString()) return 'Kemarin'
    return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
  }

  return (
    <div className="flex-1 px-4 pt-4 pb-24 w-full">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Workout Logger 💪</h1>
        <p className="text-sm text-[var(--text-secondary)] mt-1">Catat latihanmu setiap hari</p>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-3 gap-3 mb-6">
        <div className="card text-center">
          <Dumbbell size={20} className="mx-auto mb-1.5 text-[var(--accent-light)]" />
          <div className="text-xl font-bold">{stats.total}</div>
          <div className="text-[11px] text-[var(--text-muted)]">Total</div>
        </div>
        <div className="card text-center">
          <Flame size={20} className="mx-auto mb-1.5 text-[var(--warning)]" />
          <div className="text-xl font-bold">{stats.streak}</div>
          <div className="text-[11px] text-[var(--text-muted)]">Streak</div>
        </div>
        <div className="card text-center">
          <Calendar size={20} className="mx-auto mb-1.5 text-[var(--success)]" />
          <div className="text-xl font-bold">{stats.todayCount}</div>
          <div className="text-[11px] text-[var(--text-muted)]">Hari ini</div>
        </div>
      </div>

      {/* Start workout CTA */}
      <button
        onClick={() => navigate('/workout/new')}
        className="btn-primary w-full py-4 text-base mb-6"
        style={{ animation: 'pulse-glow 2s ease-in-out infinite' }}
      >
        <Plus size={22} />
        Mulai Workout
      </button>

      {/* Recent workouts */}
      <div className="mb-4">
        <div className="flex items-center justify-between mb-3">
          <h2 className="font-semibold text-[var(--text-secondary)]">Workout Terakhir</h2>
          {recent.length > 0 && (
            <button
              onClick={() => navigate('/history')}
              className="text-xs text-[var(--accent-light)] hover:underline flex items-center gap-0.5"
            >
              Lihat semua <ChevronRight size={14} />
            </button>
          )}
        </div>

        {recent.length === 0 ? (
          <div className="card text-center py-8">
            <Dumbbell size={32} className="mx-auto mb-2 text-[var(--text-muted)]" />
            <p className="text-[var(--text-muted)] text-sm">Belum ada workout</p>
            <p className="text-[var(--text-muted)] text-xs mt-1">Tap "Mulai Workout" untuk memulai!</p>
          </div>
        ) : (
          <div className="space-y-2">
            {recent.map((w) => (
              <div key={w.id} className="card flex items-center gap-3 animate-fade-in cursor-pointer" onClick={() => navigate(`/history`)}>
                <div className="w-10 h-10 rounded-xl flex items-center justify-center text-lg"
                  style={{ background: 'linear-gradient(135deg, var(--accent), #8b5cf6)' }}>
                  🏋️
                </div>
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-[15px] truncate">{w.title}</div>
                  <div className="text-xs text-[var(--text-muted)] flex items-center gap-2">
                    <span>{w.exercises?.length || 0} latihan</span>
                    <span>•</span>
                    <span>{w.totalTime || '—'}</span>
                  </div>
                </div>
                <div className="text-xs text-[var(--text-muted)]">{formatDate(w.createdAt)}</div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
