import { NavLink } from 'react-router-dom'
import { Home, Dumbbell, Clock, User } from 'lucide-react'

const navItems = [
  { to: '/', icon: Home, label: 'Beranda' },
  { to: '/workout/new', icon: Dumbbell, label: 'Workout' },
  { to: '/history', icon: Clock, label: 'Riwayat' },
]

export default function BottomNav() {
  return (
    <nav className="bottom-nav glass">
      <div className="flex justify-around items-center px-4">
        {navItems.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            className={({ isActive }) =>
              `flex flex-col items-center gap-1 py-1.5 px-4 rounded-xl transition-all duration-200 ${
                isActive
                  ? 'text-[var(--accent-light)]'
                  : 'text-[var(--text-muted)] hover:text-[var(--text-secondary)]'
              }`
            }
          >
            <Icon size={22} strokeWidth={isActive => isActive ? 2.5 : 1.8} />
            <span className="text-[11px] font-medium">{label}</span>
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
