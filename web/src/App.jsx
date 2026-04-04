import { BrowserRouter, Routes, Route } from 'react-router-dom'
import BottomNav from './components/BottomNav'
import Dashboard from './pages/Dashboard'
import LogWorkout from './pages/LogWorkout'
import History from './pages/History'

export default function App() {
  return (
    <BrowserRouter>
      <div className="min-h-dvh flex flex-col">
        <main className="flex-1 flex flex-col">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/workout/new" element={<LogWorkout />} />
            <Route path="/history" element={<History />} />
          </Routes>
        </main>
        <BottomNav />
      </div>
    </BrowserRouter>
  )
}
