// Exercise database - Indonesian fitness exercises
const exercises = [
  // Strength - Upper Body
  { id: 'bench_press', name: 'Bench Press', category: 'Dada', type: 'strength', muscle: 'chest' },
  { id: 'incline_bench', name: 'Incline Bench Press', category: 'Dada', type: 'strength', muscle: 'chest' },
  { id: 'dumbbell_fly', name: 'Dumbbell Fly', category: 'Dada', type: 'strength', muscle: 'chest' },
  { id: 'push_up', name: 'Push Up', category: 'Dada', type: 'strength', muscle: 'chest' },
  { id: 'overhead_press', name: 'Overhead Press', category: 'Bahu', type: 'strength', muscle: 'shoulders' },
  { id: 'lateral_raise', name: 'Lateral Raise', category: 'Bahu', type: 'strength', muscle: 'shoulders' },
  { id: 'front_raise', name: 'Front Raise', category: 'Bahu', type: 'strength', muscle: 'shoulders' },
  { id: 'bicep_curl', name: 'Bicep Curl', category: 'Lengan', type: 'strength', muscle: 'arms' },
  { id: 'hammer_curl', name: 'Hammer Curl', category: 'Lengan', type: 'strength', muscle: 'arms' },
  { id: 'tricep_pushdown', name: 'Tricep Pushdown', category: 'Lengan', type: 'strength', muscle: 'arms' },
  { id: 'tricep_dip', name: 'Tricep Dip', category: 'Lengan', type: 'strength', muscle: 'arms' },
  { id: 'pull_up', name: 'Pull Up', category: 'Punggung', type: 'strength', muscle: 'back' },
  { id: 'lat_pulldown', name: 'Lat Pulldown', category: 'Punggung', type: 'strength', muscle: 'back' },
  { id: 'barbell_row', name: 'Barbell Row', category: 'Punggung', type: 'strength', muscle: 'back' },
  { id: 'cable_row', name: 'Cable Row', category: 'Punggung', type: 'strength', muscle: 'back' },

  // Strength - Lower Body
  { id: 'squat', name: 'Squat', category: 'Kaki', type: 'strength', muscle: 'legs' },
  { id: 'leg_press', name: 'Leg Press', category: 'Kaki', type: 'strength', muscle: 'legs' },
  { id: 'deadlift', name: 'Deadlift', category: 'Kaki', type: 'strength', muscle: 'legs' },
  { id: 'lunges', name: 'Lunges', category: 'Kaki', type: 'strength', muscle: 'legs' },
  { id: 'leg_extension', name: 'Leg Extension', category: 'Kaki', type: 'strength', muscle: 'legs' },
  { id: 'leg_curl', name: 'Leg Curl', category: 'Kaki', type: 'strength', muscle: 'legs' },
  { id: 'calf_raise', name: 'Calf Raise', category: 'Kaki', type: 'strength', muscle: 'legs' },

  // Core
  { id: 'plank', name: 'Plank', category: 'Perut', type: 'duration', muscle: 'core' },
  { id: 'crunch', name: 'Crunch', category: 'Perut', type: 'strength', muscle: 'core' },
  { id: 'russian_twist', name: 'Russian Twist', category: 'Perut', type: 'strength', muscle: 'core' },
  { id: 'leg_raise', name: 'Leg Raise', category: 'Perut', type: 'strength', muscle: 'core' },
  { id: 'mountain_climber', name: 'Mountain Climber', category: 'Perut', type: 'duration', muscle: 'core' },

  // Flexibility
  { id: 'warm_up', name: 'Warm Up', category: 'Pemanasan', type: 'duration', muscle: 'full' },
  { id: 'butterfly_stretch', name: 'Butterfly Stretch', category: 'Peregangan', type: 'duration', muscle: 'legs' },
  { id: 'leg_stretch', name: 'Leg Stretch', category: 'Peregangan', type: 'duration', muscle: 'legs' },
  { id: 'forward_fold', name: 'Forward Fold', category: 'Peregangan', type: 'duration', muscle: 'back' },
  { id: 'front_wide_stretch', name: 'Front Wide Stretch', category: 'Peregangan', type: 'duration', muscle: 'legs' },
  { id: 'happy_baby_pose', name: 'Happy Baby Pose', category: 'Peregangan', type: 'duration', muscle: 'full' },
  { id: 'supine_twist', name: 'Supine Twist', category: 'Peregangan', type: 'duration', muscle: 'core' },

  // Cardio
  { id: 'running', name: 'Lari', category: 'Kardio', type: 'duration', muscle: 'cardio' },
  { id: 'cycling', name: 'Bersepeda', category: 'Kardio', type: 'duration', muscle: 'cardio' },
  { id: 'jump_rope', name: 'Lompat Tali', category: 'Kardio', type: 'duration', muscle: 'cardio' },
  { id: 'burpee', name: 'Burpee', category: 'Kardio', type: 'strength', muscle: 'full' },
  { id: 'jumping_jack', name: 'Jumping Jack', category: 'Kardio', type: 'strength', muscle: 'full' },
]

export const categories = [...new Set(exercises.map(e => e.category))]

export function searchExercises(query) {
  if (!query) return exercises
  const q = query.toLowerCase()
  return exercises.filter(
    e => e.name.toLowerCase().includes(q) || e.category.toLowerCase().includes(q)
  )
}

export function getExerciseById(id) {
  return exercises.find(e => e.id === id)
}

export default exercises
