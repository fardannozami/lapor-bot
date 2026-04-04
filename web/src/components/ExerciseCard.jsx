import { useState, useRef } from 'react'
import { Trash2, Plus, Check, RotateCcw } from 'lucide-react'

function SwipeableRow({ children, onDelete }) {
  const startX = useRef(0)
  const currentX = useRef(0)
  const rowRef = useRef(null)
  const [offset, setOffset] = useState(0)
  const [showDelete, setShowDelete] = useState(false)

  const handleTouchStart = (e) => {
    startX.current = e.touches[0].clientX
    currentX.current = startX.current
  }

  const handleTouchMove = (e) => {
    currentX.current = e.touches[0].clientX
    const diff = currentX.current - startX.current
    if (diff < 0) {
      setOffset(Math.max(diff, -80))
    } else if (showDelete) {
      setOffset(Math.min(0, -80 + diff))
    }
  }

  const handleTouchEnd = () => {
    if (offset < -40) {
      setOffset(-72)
      setShowDelete(true)
    } else {
      setOffset(0)
      setShowDelete(false)
    }
  }

  // Mouse support for desktop
  const handleMouseDown = (e) => {
    startX.current = e.clientX
    const handleMouseMove = (ev) => {
      currentX.current = ev.clientX
      const diff = currentX.current - startX.current
      if (diff < 0) setOffset(Math.max(diff, -80))
      else if (showDelete) setOffset(Math.min(0, -80 + diff))
    }
    const handleMouseUp = () => {
      if (offset < -40) { setOffset(-72); setShowDelete(true) }
      else { setOffset(0); setShowDelete(false) }
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }
    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
  }

  return (
    <div className="relative overflow-hidden rounded-lg mb-0.5">
      {/* Delete button behind */}
      <div className="absolute right-0 top-0 bottom-0 w-[72px] flex items-center justify-center bg-[var(--danger)]">
        <button
          onClick={() => { setOffset(0); setShowDelete(false); onDelete() }}
          className="flex flex-col items-center gap-0.5 text-white"
        >
          <Trash2 size={16} />
          <span className="text-[10px] font-medium">Hapus</span>
        </button>
      </div>
      {/* Row content */}
      <div
        ref={rowRef}
        style={{ transform: `translateX(${offset}px)`, transition: offset === 0 || offset === -72 ? 'transform 0.2s ease' : 'none' }}
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
        onMouseDown={handleMouseDown}
        className="relative z-10 bg-[var(--bg-card)]"
      >
        {children}
      </div>
    </div>
  )
}

export default function ExerciseCard({ exercise, sets, onSetsChange, onRemove, lastHistory }) {
  const isDuration = exercise.type === 'duration'
  const [completedSets, setCompletedSets] = useState(new Set())

  const addSet = () => {
    const lastSet = sets.length > 0 ? sets[sets.length - 1] : null
    const newSet = isDuration
      ? { duration: lastSet?.duration || 60 }
      : { weight: lastSet?.weight || 0, reps: lastSet?.reps || 10 }
    onSetsChange([...sets, newSet])
  }

  const updateSet = (index, field, value) => {
    const updated = [...sets]
    updated[index] = { ...updated[index], [field]: value }
    onSetsChange(updated)
  }

  const removeSet = (index) => {
    onSetsChange(sets.filter((_, i) => i !== index))
    setCompletedSets(prev => {
      const next = new Set(prev)
      next.delete(index)
      return next
    })
  }

  const toggleComplete = (index) => {
    setCompletedSets(prev => {
      const next = new Set(prev)
      if (next.has(index)) next.delete(index)
      else next.add(index)
      return next
    })
  }

  const prefillFromHistory = () => {
    if (!lastHistory?.sets) return
    const newSets = lastHistory.sets.map(s => ({ ...s }))
    onSetsChange(newSets)
  }

  const formatPrevious = (set) => {
    if (!set) return '—'
    if (isDuration) return `${set.duration}s`
    return `${set.weight} × ${set.reps}`
  }

  const formatHistoryDate = (dateStr) => {
    if (!dateStr) return ''
    const d = new Date(dateStr)
    const now = new Date()
    const diffDays = Math.floor((now - d) / (1000 * 60 * 60 * 24))
    if (diffDays === 0) return 'Hari ini'
    if (diffDays === 1) return 'Kemarin'
    if (diffDays < 7) return `${diffDays} hari lalu`
    return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })
  }

  const gridCols = isDuration ? '36px 1fr 1fr 36px' : '36px 1fr 1fr 1fr 36px'

  return (
    <div className="card animate-fade-in !p-0 overflow-hidden">
      {/* Header */}
      <div className="flex items-center justify-between px-4 pt-3 pb-2">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-lg flex items-center justify-center text-sm"
            style={{ background: 'linear-gradient(135deg, var(--accent), #8b5cf6)' }}>
            {isDuration ? '⏱' : '🏋️'}
          </div>
          <div>
            <h3 className="font-semibold text-[14px] leading-tight">{exercise.name}</h3>
            <p className="text-[11px] text-[var(--text-muted)]">{exercise.category}</p>
          </div>
        </div>
        <div className="flex items-center gap-1">
          {lastHistory && (
            <button
              onClick={prefillFromHistory}
              title="Isi dari workout terakhir"
              className="p-1.5 rounded-lg hover:bg-[rgba(108,92,231,0.15)] transition-colors text-[var(--text-muted)] hover:text-[var(--accent-light)]"
            >
              <RotateCcw size={14} />
            </button>
          )}
          <button
            onClick={onRemove}
            className="p-1.5 rounded-lg hover:bg-[rgba(225,112,85,0.15)] transition-colors text-[var(--text-muted)] hover:text-[var(--danger)]"
          >
            <Trash2 size={14} />
          </button>
        </div>
      </div>

      {/* Last history info */}
      {lastHistory && (
        <div className="px-4 pb-2">
          <span className="text-[10px] text-[var(--text-muted)] bg-[var(--bg-secondary)] px-2 py-0.5 rounded-md">
            Terakhir: {formatHistoryDate(lastHistory.date)}
          </span>
        </div>
      )}

      {/* Sets table - Hevy style */}
      <div className="px-2 pb-2">
        {/* Header row */}
        <div
          className="grid items-center text-[10px] text-[var(--text-muted)] font-semibold uppercase tracking-wider px-2 py-1.5"
          style={{ gridTemplateColumns: gridCols }}
        >
          <span className="text-center">SET</span>
          <span className="text-center">SEBELUMNYA</span>
          {isDuration ? (
            <span className="text-center">DURASI</span>
          ) : (
            <>
              <span className="text-center">KG</span>
              <span className="text-center">REPS</span>
            </>
          )}
          <span />
        </div>

        {/* Set rows - swipeable */}
        {sets.map((set, i) => {
          const prevSet = lastHistory?.sets?.[i]
          const isComplete = completedSets.has(i)

          return (
            <SwipeableRow key={i} onDelete={() => removeSet(i)}>
              <div
                className={`grid items-center px-2 py-1 transition-colors ${
                  isComplete ? 'bg-[rgba(0,184,148,0.08)]' : ''
                }`}
                style={{ gridTemplateColumns: gridCols }}
              >
                {/* Set number */}
                <span className={`text-xs text-center font-bold rounded-md py-0.5 ${
                  isComplete ? 'text-[var(--success)]' : 'text-[var(--text-muted)]'
                }`}>
                  {i + 1}
                </span>

                {/* Previous */}
                <span className="text-xs text-center text-[var(--text-muted)] font-mono">
                  {formatPrevious(prevSet)}
                </span>

                {/* Current inputs */}
                {isDuration ? (
                  <input
                    type="number"
                    value={set.duration}
                    onChange={e => updateSet(i, 'duration', Number(e.target.value))}
                    className="bg-[var(--bg-secondary)] border border-[var(--border)] rounded-md text-center text-sm py-1.5 text-[var(--text-primary)] outline-none focus:border-[var(--accent)] transition-colors w-full mx-1"
                    min="0"
                  />
                ) : (
                  <>
                    <input
                      type="number"
                      value={set.weight}
                      onChange={e => updateSet(i, 'weight', Number(e.target.value))}
                      className="bg-[var(--bg-secondary)] border border-[var(--border)] rounded-md text-center text-sm py-1.5 text-[var(--text-primary)] outline-none focus:border-[var(--accent)] transition-colors w-full mx-1"
                      min="0"
                      step="0.5"
                    />
                    <input
                      type="number"
                      value={set.reps}
                      onChange={e => updateSet(i, 'reps', Number(e.target.value))}
                      className="bg-[var(--bg-secondary)] border border-[var(--border)] rounded-md text-center text-sm py-1.5 text-[var(--text-primary)] outline-none focus:border-[var(--accent)] transition-colors w-full mx-1"
                      min="0"
                    />
                  </>
                )}

                {/* Complete check */}
                <button
                  onClick={() => toggleComplete(i)}
                  className={`w-7 h-7 rounded-md flex items-center justify-center mx-auto transition-all ${
                    isComplete
                      ? 'bg-[var(--success)] text-white'
                      : 'border border-[var(--border)] text-[var(--text-muted)] hover:border-[var(--success)] hover:text-[var(--success)]'
                  }`}
                >
                  <Check size={14} strokeWidth={isComplete ? 3 : 2} />
                </button>
              </div>
            </SwipeableRow>
          )
        })}
      </div>

      {/* Add set button */}
      <button
        onClick={addSet}
        className="w-full py-2.5 border-t border-[var(--border)] text-sm text-[var(--text-muted)] hover:text-[var(--accent-light)] hover:bg-[var(--bg-hover)] transition-all flex items-center justify-center gap-1.5"
      >
        <Plus size={14} />
        Tambah Set
      </button>
    </div>
  )
}
