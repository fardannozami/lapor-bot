import { useState, useEffect } from 'react'
import { Search, X, Plus } from 'lucide-react'
import { searchExercises, categories } from '../data/exercises'

const CUSTOM_EXERCISES_KEY = 'custom-exercises'

function loadCustomExercises() {
  try {
    const raw = localStorage.getItem(CUSTOM_EXERCISES_KEY)
    return raw ? JSON.parse(raw) : []
  } catch {
    return []
  }
}

function saveCustomExercises(list) {
  localStorage.setItem(CUSTOM_EXERCISES_KEY, JSON.stringify(list))
}

export default function ExercisePicker({ onSelect, onClose }) {
  const [query, setQuery] = useState('')
  const [activeCategory, setActiveCategory] = useState(null)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [customExercises, setCustomExercises] = useState(loadCustomExercises)

  // New exercise form
  const [newName, setNewName] = useState('')
  const [newCategory, setNewCategory] = useState('Lainnya')
  const [newType, setNewType] = useState('strength')

  useEffect(() => {
    saveCustomExercises(customExercises)
  }, [customExercises])

  // Merge built-in + custom
  const allExercises = [...searchExercises(''), ...customExercises]

  const results = allExercises
    .filter(e => {
      const q = query.toLowerCase()
      if (q && !e.name.toLowerCase().includes(q) && !e.category.toLowerCase().includes(q)) return false
      if (activeCategory && e.category !== activeCategory) return false
      return true
    })

  const allCategories = [...new Set([...categories, ...customExercises.map(e => e.category)])]

  const handleCreate = () => {
    if (!newName.trim()) return
    const newExercise = {
      id: `custom_${Date.now()}`,
      name: newName.trim(),
      category: newCategory.trim() || 'Lainnya',
      type: newType,
      muscle: 'custom',
      isCustom: true,
    }
    setCustomExercises(prev => [...prev, newExercise])
    onSelect(newExercise)
    setShowCreateForm(false)
    setNewName('')
  }

  const handleDeleteCustom = (id, e) => {
    e.stopPropagation()
    setCustomExercises(prev => prev.filter(ex => ex.id !== id))
  }

  // Common padding for sections
  const sectionPadding = "px-6"

  return (
    <div className="fixed inset-0 z-50 flex justify-center" style={{ background: 'rgba(0,0,0,0.5)' }}>
      <div className="w-full max-w-[430px] flex flex-col pt-2" style={{ background: 'var(--bg-primary)' }}>
        {/* Header */}
        <div className={`flex items-center gap-3 ${sectionPadding} py-3 border-b border-[var(--border)]`}>
          <button onClick={onClose} className="p-1.5 rounded-lg hover:bg-[var(--bg-hover)] transition-colors">
            <X size={22} className="text-[var(--text-secondary)]" />
          </button>
          <h2 className="text-lg font-semibold flex-1">Pilih Latihan</h2>
          <button
            onClick={() => setShowCreateForm(!showCreateForm)}
            className="flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs font-medium bg-[var(--accent)] text-white hover:opacity-90 transition-opacity"
          >
            <Plus size={14} />
            Buat
          </button>
        </div>

        {/* Create form */}
        {showCreateForm && (
          <div className={`${sectionPadding} py-3 border-b border-[var(--border)] bg-[var(--bg-secondary)] animate-fade-in`}>
            <p className="text-xs text-[var(--text-muted)] font-medium mb-2.5 uppercase tracking-wider">Latihan Baru</p>

            <input
              type="text"
              placeholder="Nama latihan"
              value={newName}
              onChange={e => setNewName(e.target.value)}
              className="input mb-2 text-sm"
              autoFocus
            />

            <div className="flex gap-2 mb-2">
              <input
                type="text"
                placeholder="Kategori"
                value={newCategory}
                onChange={e => setNewCategory(e.target.value)}
                className="input text-sm flex-1"
                list="category-suggestions"
              />
              <datalist id="category-suggestions">
                {allCategories.map(c => <option key={c} value={c} />)}
              </datalist>

              <select
                value={newType}
                onChange={e => setNewType(e.target.value)}
                className="input text-sm w-[120px] !py-2"
              >
                <option value="strength">🏋️ Set/Rep</option>
                <option value="duration">⏱ Durasi</option>
              </select>
            </div>

            <div className="flex gap-2">
              <button
                onClick={() => setShowCreateForm(false)}
                className="btn-secondary flex-1 text-xs"
              >
                Batal
              </button>
              <button
                onClick={handleCreate}
                disabled={!newName.trim()}
                className="btn-primary flex-1 text-xs disabled:opacity-30"
              >
                Simpan & Pilih
              </button>
            </div>
          </div>
        )}

        {/* Search */}
        <div className={`${sectionPadding} py-3`}>
          <div className="relative">
            <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-[var(--text-muted)]" />
            <input
              type="text"
              placeholder="Cari latihan..."
              value={query}
              onChange={e => setQuery(e.target.value)}
              className="input pl-10"
            />
          </div>
        </div>

        {/* Category pills */}
        <div className={`${sectionPadding} pb-2 flex gap-2 overflow-x-auto no-scrollbar`}>
          <button
            onClick={() => setActiveCategory(null)}
            className={`px-3 py-1.5 rounded-full text-xs font-medium whitespace-nowrap transition-all shrink-0 ${!activeCategory
                ? 'bg-[var(--accent)] text-white'
                : 'bg-[var(--bg-hover)] text-[var(--text-secondary)] hover:bg-[var(--border)]'
              }`}
          >
            Semua
          </button>
          {allCategories.map(cat => (
            <button
              key={cat}
              onClick={() => setActiveCategory(activeCategory === cat ? null : cat)}
              className={`px-3 py-1.5 rounded-full text-xs font-medium whitespace-nowrap transition-all shrink-0 ${activeCategory === cat
                  ? 'bg-[var(--accent)] text-white'
                  : 'bg-[var(--bg-hover)] text-[var(--text-secondary)] hover:bg-[var(--border)]'
                }`}
            >
              {cat}
            </button>
          ))}
        </div>

        {/* Exercise list */}
        <div className={`${sectionPadding} flex-1 overflow-y-auto pb-4`}>
          {results.length === 0 ? (
            <div className="text-center py-12">
              <p className="text-[var(--text-muted)] text-sm mb-3">Latihan tidak ditemukan</p>
              <button
                onClick={() => { setShowCreateForm(true); setNewName(query) }}
                className="text-xs text-[var(--accent-light)] hover:underline"
              >
                + Buat "{query}" sebagai latihan baru
              </button>
            </div>
          ) : (
            <div className="space-y-1">
              {results.map(exercise => (
                <button
                  key={exercise.id}
                  onClick={() => onSelect(exercise)}
                  className="w-full text-left px-4 py-3 rounded-xl hover:bg-[var(--bg-hover)] transition-colors flex items-center justify-between group"
                >
                  <div>
                    <div className="font-medium text-[15px] flex items-center gap-1.5">
                      {exercise.name}
                      {exercise.isCustom && (
                        <span className="text-[9px] px-1.5 py-0.5 rounded bg-[var(--accent)] text-white font-semibold">CUSTOM</span>
                      )}
                    </div>
                    <div className="text-xs text-[var(--text-muted)] mt-0.5">{exercise.category}</div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs px-2 py-1 rounded-md bg-[var(--bg-secondary)] text-[var(--text-muted)] group-hover:text-[var(--accent-light)]">
                      {exercise.type === 'duration' ? '⏱ Durasi' : '🏋️ Set/Rep'}
                    </span>
                    {exercise.isCustom && (
                      <button
                        onClick={(e) => handleDeleteCustom(exercise.id, e)}
                        className="p-1 rounded hover:bg-[rgba(225,112,85,0.15)] text-[var(--text-muted)] hover:text-[var(--danger)] transition-colors opacity-0 group-hover:opacity-100"
                      >
                        <X size={14} />
                      </button>
                    )}
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
