import { useState, useEffect, useCallback, useRef } from 'react'

const STORAGE_KEY = 'workout-logger'

function loadData() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { workouts: [], templates: [] }
    return JSON.parse(raw)
  } catch {
    return { workouts: [], templates: [] }
  }
}

function saveData(data) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
}

export function useWorkouts() {
  const [data, setData] = useState(loadData)

  useEffect(() => {
    saveData(data)
  }, [data])

  const addWorkout = useCallback((workout) => {
    const newWorkout = {
      ...workout,
      id: crypto.randomUUID(),
      createdAt: new Date().toISOString(),
    }
    setData(prev => ({
      ...prev,
      workouts: [newWorkout, ...prev.workouts],
    }))
    return newWorkout
  }, [])

  const deleteWorkout = useCallback((id) => {
    setData(prev => ({
      ...prev,
      workouts: prev.workouts.filter(w => w.id !== id),
    }))
  }, [])

  const getWorkoutsForDate = useCallback((dateStr) => {
    return data.workouts.filter(w => w.createdAt.startsWith(dateStr))
  }, [data.workouts])

  const getRecentWorkouts = useCallback((limit = 5) => {
    return data.workouts.slice(0, limit)
  }, [data.workouts])

  const getStats = useCallback(() => {
    const total = data.workouts.length
    const today = new Date().toISOString().split('T')[0]
    const todayCount = data.workouts.filter(w => w.createdAt.startsWith(today)).length

    // Calculate streak (consecutive days)
    let streak = 0
    const dates = [...new Set(data.workouts.map(w => w.createdAt.split('T')[0]))].sort().reverse()
    if (dates.length > 0) {
      const now = new Date()
      let checkDate = new Date(now.getFullYear(), now.getMonth(), now.getDate())
      // If no workout today, start checking from yesterday
      if (!dates.includes(checkDate.toISOString().split('T')[0])) {
        checkDate.setDate(checkDate.getDate() - 1)
      }
      for (const date of dates) {
        if (date === checkDate.toISOString().split('T')[0]) {
          streak++
          checkDate.setDate(checkDate.getDate() - 1)
        } else {
          break
        }
      }
    }

    return { total, todayCount, streak }
  }, [data.workouts])

  // Templates
  const saveTemplate = useCallback((template) => {
    const newTemplate = {
      ...template,
      id: crypto.randomUUID(),
    }
    setData(prev => ({
      ...prev,
      templates: [newTemplate, ...prev.templates],
    }))
    return newTemplate
  }, [])

  const deleteTemplate = useCallback((id) => {
    setData(prev => ({
      ...prev,
      templates: prev.templates.filter(t => t.id !== id),
    }))
  }, [])

  // Get last workout history for a specific exercise
  const getLastExerciseHistory = useCallback((exerciseId) => {
    for (const workout of data.workouts) {
      const found = workout.exercises?.find(e => e.id === exerciseId)
      if (found) {
        return { sets: found.sets, date: workout.createdAt, workoutTitle: workout.title }
      }
    }
    return null
  }, [data.workouts])

  return {
    workouts: data.workouts,
    templates: data.templates,
    addWorkout,
    deleteWorkout,
    getWorkoutsForDate,
    getRecentWorkouts,
    getStats,
    getLastExerciseHistory,
    saveTemplate,
    deleteTemplate,
  }
}

// Timer hook
export function useTimer() {
  const [seconds, setSeconds] = useState(0)
  const [isRunning, setIsRunning] = useState(false)
  const intervalRef = useRef(null)

  useEffect(() => {
    if (isRunning) {
      intervalRef.current = setInterval(() => {
        setSeconds(s => s + 1)
      }, 1000)
    } else {
      clearInterval(intervalRef.current)
    }
    return () => clearInterval(intervalRef.current)
  }, [isRunning])

  const start = () => setIsRunning(true)
  const pause = () => setIsRunning(false)
  const reset = () => { setIsRunning(false); setSeconds(0) }
  const toggle = () => setIsRunning(r => !r)

  const formatTime = (secs) => {
    const h = Math.floor(secs / 3600)
    const m = Math.floor((secs % 3600) / 60)
    const s = secs % 60
    if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
    return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  }

  return { seconds, isRunning, start, pause, reset, toggle, formatted: formatTime(seconds), formatTime }
}
