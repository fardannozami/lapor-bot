import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Timer, Check, X } from 'lucide-react'
import { useWorkouts, useTimer } from '../hooks/useWorkouts'
import ExercisePicker from '../components/ExercisePicker'
import ExerciseCard from '../components/ExerciseCard'

export default function LogWorkout() {
  const navigate = useNavigate()
  const { addWorkout, getLastExerciseHistory } = useWorkouts()
  const timer = useTimer()

  const [title, setTitle] = useState('')
  const [exercises, setExercises] = useState([]) // { exercise, sets: [...] }
  const [showPicker, setShowPicker] = useState(false)
  const [isSaving, setIsSaving] = useState(false)

  // Start timer when first exercise is added
  const handleAddExercise = (exercise) => {
    setExercises(prev => [
      ...prev,
      {
        exercise,
        sets: exercise.type === 'duration'
          ? [{ duration: 60 }]
          : [{ weight: 0, reps: 10 }],
      },
    ])
    setShowPicker(false)
    if (!timer.isRunning && exercises.length === 0) {
      timer.start()
    }
  }

  const handleUpdateSets = (index, newSets) => {
    setExercises(prev => {
      const updated = [...prev]
      updated[index] = { ...updated[index], sets: newSets }
      return updated
    })
  }

  const handleRemoveExercise = (index) => {
    setExercises(prev => prev.filter((_, i) => i !== index))
  }

  const handleSave = () => {
    if (exercises.length === 0) return

    setIsSaving(true)
    timer.pause()

    const workout = {
      title: title || `Workout ${new Date().toLocaleDateString('id-ID')}`,
      exercises: exercises.map(({ exercise, sets }) => ({
        id: exercise.id,
        name: exercise.name,
        category: exercise.category,
        type: exercise.type,
        sets,
      })),
      totalTime: timer.formatted,
      totalSeconds: timer.seconds,
    }

    addWorkout(workout)

    // Navigate back to dashboard
    setTimeout(() => {
      navigate('/', { replace: true })
    }, 300)
  }

  const handleDiscard = () => {
    if (exercises.length > 0) {
      if (!confirm('Buang workout ini?')) return
    }
    timer.reset()
    navigate('/', { replace: true })
  }

  return (
    <div className="flex-1 px-4 pt-4 pb-24 w-full">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <button
          onClick={handleDiscard}
          className="p-2 rounded-lg hover:bg-[var(--bg-hover)] transition-colors text-[var(--text-secondary)]"
        >
          <X size={22} />
        </button>
        <h1 className="font-bold text-lg">Workout Baru</h1>
        <button
          onClick={handleSave}
          disabled={exercises.length === 0 || isSaving}
          className="p-2 rounded-lg hover:bg-[rgba(0,184,148,0.15)] transition-colors text-[var(--success)] disabled:opacity-30 disabled:cursor-not-allowed"
        >
          <Check size={22} />
        </button>
      </div>

      {/* Timer */}
      <div className="flex items-center justify-center gap-3 mb-5">
        <button
          onClick={timer.toggle}
          className={`flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-mono transition-all ${
            timer.isRunning
              ? 'bg-[rgba(108,92,231,0.15)] text-[var(--accent-light)] border border-[var(--accent)]'
              : 'bg-[var(--bg-card)] text-[var(--text-secondary)] border border-[var(--border)]'
          }`}
        >
          <Timer size={16} />
          <span className="text-xl font-bold tracking-wider">{timer.formatted}</span>
        </button>
      </div>

      {/* Title input */}
      <input
        type="text"
        placeholder="Nama workout (opsional)"
        value={title}
        onChange={e => setTitle(e.target.value)}
        className="input mb-4"
      />

      {/* Exercise list */}
      <div className="space-y-3 mb-4">
        {exercises.map((item, index) => (
          <ExerciseCard
            key={`${item.exercise.id}-${index}`}
            exercise={item.exercise}
            sets={item.sets}
            onSetsChange={(newSets) => handleUpdateSets(index, newSets)}
            onRemove={() => handleRemoveExercise(index)}
            lastHistory={getLastExerciseHistory(item.exercise.id)}
          />
        ))}
      </div>

      {/* Add exercise button */}
      <button
        onClick={() => setShowPicker(true)}
        className="btn-secondary w-full flex items-center justify-center gap-2"
      >
        <Plus size={18} />
        Tambah Latihan
      </button>

      {/* Exercise picker modal */}
      {showPicker && (
        <ExercisePicker
          onSelect={handleAddExercise}
          onClose={() => setShowPicker(false)}
        />
      )}
    </div>
  )
}
