import { useState } from 'react'
import { useWorkouts } from '../hooks/useWorkouts'
import { ChevronLeft, ChevronRight, Trash2, Clock, Dumbbell } from 'lucide-react'

export default function History() {
  const { workouts, deleteWorkout } = useWorkouts()
  const [expandedId, setExpandedId] = useState(null)

  // Calendar state
  const [currentMonth, setCurrentMonth] = useState(new Date())

  const year = currentMonth.getFullYear()
  const month = currentMonth.getMonth()

  // Get days with workouts
  const workoutDates = new Set(
    workouts.map(w => w.createdAt.split('T')[0])
  )

  // Generate calendar grid
  const firstDay = new Date(year, month, 1).getDay() // 0=Sun
  const daysInMonth = new Date(year, month + 1, 0).getDate()
  const today = new Date().toISOString().split('T')[0]

  const calendarDays = []
  for (let i = 0; i < firstDay; i++) calendarDays.push(null)
  for (let d = 1; d <= daysInMonth; d++) calendarDays.push(d)

  const prevMonth = () => setCurrentMonth(new Date(year, month - 1))
  const nextMonth = () => setCurrentMonth(new Date(year, month + 1))

  const monthName = currentMonth.toLocaleDateString('id-ID', { month: 'long', year: 'numeric' })

  const formatDate = (iso) => {
    return new Date(iso).toLocaleDateString('id-ID', {
      weekday: 'short',
      day: 'numeric',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const formatDuration = (secs) => {
    if (!secs) return '—'
    const m = Math.floor(secs / 60)
    const s = secs % 60
    if (m >= 60) {
      const h = Math.floor(m / 60)
      return `${h}j ${m % 60}m`
    }
    return `${m}m ${s}s`
  }

  const handleDelete = (id, e) => {
    e.stopPropagation()
    if (confirm('Hapus workout ini?')) {
      deleteWorkout(id)
    }
  }

  return (
    <div className="flex-1 px-4 pt-4 pb-24 w-full">
      <h1 className="text-2xl font-bold mb-4">Riwayat 📋</h1>

      {/* Calendar */}
      <div className="card mb-5">
        <div className="flex items-center justify-between mb-3">
          <button onClick={prevMonth} className="p-1.5 rounded-lg hover:bg-[var(--bg-hover)] transition-colors">
            <ChevronLeft size={18} className="text-[var(--text-secondary)]" />
          </button>
          <span className="font-semibold text-sm">{monthName}</span>
          <button onClick={nextMonth} className="p-1.5 rounded-lg hover:bg-[var(--bg-hover)] transition-colors">
            <ChevronRight size={18} className="text-[var(--text-secondary)]" />
          </button>
        </div>

        {/* Day headers */}
        <div className="grid grid-cols-7 gap-1 mb-1">
          {['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab'].map(d => (
            <div key={d} className="text-center text-[10px] text-[var(--text-muted)] font-medium py-1">
              {d}
            </div>
          ))}
        </div>

        {/* Calendar grid */}
        <div className="grid grid-cols-7 gap-1">
          {calendarDays.map((day, i) => {
            if (!day) return <div key={`empty-${i}`} />
            const dateStr = `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`
            const hasWorkout = workoutDates.has(dateStr)
            const isToday = dateStr === today

            return (
              <div
                key={i}
                className={`relative text-center py-1.5 rounded-lg text-xs transition-all ${
                  isToday
                    ? 'bg-[var(--accent)] text-white font-bold'
                    : hasWorkout
                    ? 'text-[var(--text-primary)] font-medium'
                    : 'text-[var(--text-muted)]'
                }`}
              >
                {day}
                {hasWorkout && !isToday && (
                  <div className="absolute bottom-0.5 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full bg-[var(--success)]" />
                )}
                {hasWorkout && isToday && (
                  <div className="absolute bottom-0.5 left-1/2 -translate-x-1/2 w-1 h-1 rounded-full bg-white" />
                )}
              </div>
            )
          })}
        </div>
      </div>

      {/* Workout list */}
      {workouts.length === 0 ? (
        <div className="text-center py-12 text-[var(--text-muted)]">
          <Dumbbell size={40} className="mx-auto mb-3 opacity-30" />
          <p className="text-sm">Belum ada riwayat workout</p>
        </div>
      ) : (
        <div className="space-y-2">
          {workouts.map((w) => (
            <div key={w.id} className="card animate-fade-in">
              <div
                className="flex items-center gap-3 cursor-pointer"
                onClick={() => setExpandedId(expandedId === w.id ? null : w.id)}
              >
                <div className="w-10 h-10 rounded-xl flex items-center justify-center text-lg shrink-0"
                  style={{ background: 'linear-gradient(135deg, var(--accent), #8b5cf6)' }}>
                  🏋️
                </div>
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-[15px] truncate">{w.title}</div>
                  <div className="text-xs text-[var(--text-muted)] flex items-center gap-2 mt-0.5">
                    <span>{w.exercises?.length || 0} latihan</span>
                    <span>•</span>
                    <span className="flex items-center gap-0.5">
                      <Clock size={11} />
                      {w.totalTime || '—'}
                    </span>
                  </div>
                </div>
                <div className="text-right shrink-0">
                  <div className="text-xs text-[var(--text-muted)]">{formatDate(w.createdAt)}</div>
                </div>
              </div>

              {/* Expanded details */}
              {expandedId === w.id && (
                <div className="mt-3 pt-3 border-t border-[var(--border)] animate-fade-in">
                  {w.exercises?.map((ex, i) => (
                    <div key={i} className="mb-2">
                      <div className="text-sm font-medium text-[var(--accent-light)]">{ex.name}</div>
                      <div className="text-xs text-[var(--text-muted)] ml-2 mt-0.5">
                        {ex.sets?.map((set, si) => (
                          <span key={si} className="inline-block mr-2">
                            {ex.type === 'duration'
                              ? `Set ${si + 1}: ${set.duration}s`
                              : `Set ${si + 1}: ${set.weight}kg × ${set.reps}`
                            }
                          </span>
                        ))}
                      </div>
                    </div>
                  ))}
                  <button
                    onClick={(e) => handleDelete(w.id, e)}
                    className="btn-danger w-full mt-3 flex items-center justify-center gap-1.5 text-xs"
                  >
                    <Trash2 size={14} />
                    Hapus Workout
                  </button>
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
