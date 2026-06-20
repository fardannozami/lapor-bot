import { useState, useEffect, useMemo } from 'react';
import { LayoutGrid, Table2, RefreshCw, AlertCircle, ShieldAlert, HeartPulse, Sun, Moon, LogIn } from 'lucide-react';
import { useReports } from '@lapor-bot/shared';
import type { EnrichedReport } from '@lapor-bot/shared';
import { StatsOverview } from './components/StatsOverview';
import { LeaderboardTable } from './components/LeaderboardTable';
import { HunterCard } from './components/HunterCard';
import { ProfileModal } from './components/ProfileModal';
import { MotivationBanner } from './components/MotivationBanner';
import { LoginPage } from './components/LoginPage';
import { PersonalPage } from './components/PersonalPage';

const PAGE_SIZE = 15;
type Theme = 'light' | 'dark';
type Page = 'dashboard' | 'login' | 'personal';

function getInitialTheme(): Theme {
  if (typeof document === 'undefined') return 'dark';
  return document.documentElement.dataset.theme === 'light' ? 'light' : 'dark';
}

function App() {
  const { summary, hunters, loading, refreshing, error, refresh } = useReports();
  const [selectedHunter, setSelectedHunter] = useState<EnrichedReport | null>(null);
  const [viewMode, setViewMode] = useState<'table' | 'cards'>('table');
  const [theme, setTheme] = useState<Theme>(getInitialTheme);
  const [cardPage, setCardPage] = useState(1);
  const [page, setPage] = useState<Page>('dashboard');
  const [personalUser, setPersonalUser] = useState<EnrichedReport | null>(null);

  const seasonTitle = summary
    ? `SWEG Healthy Club - Season ${summary.current_season}`
    : 'SWEG Healthy Club';
  const cardTotalPages = Math.max(1, Math.ceil(hunters.length / PAGE_SIZE));
  const safeCardPage = Math.min(cardPage, cardTotalPages);
  const visibleCardHunters = useMemo(() => {
    const start = (safeCardPage - 1) * PAGE_SIZE;
    return hunters.slice(start, start + PAGE_SIZE);
  }, [hunters, safeCardPage]);



  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    try {
      localStorage.setItem('lapor-bot-theme', theme);
    } catch {
      return;
    }
  }, [theme]);

  useEffect(() => {
    document.title = seasonTitle;
  }, [seasonTitle]);

  const goToCardPage = (page: number) => {
    setCardPage(Math.min(Math.max(1, page), cardTotalPages));
  };

  return (
    <div className="min-h-screen pb-16 px-4 md:px-8">
      {/* HUD System Top Bar */}
      <header className="max-w-7xl mx-auto pt-8 pb-6 border-b border-gray-850 flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="w-3 h-3 bg-system-green shadow-neon-purple rounded-full animate-pulse"></div>
            <h1 className="text-xl md:text-2xl font-black font-orbitron tracking-widest text-white uppercase flex items-center gap-2">
              <HeartPulse className="text-system-green" size={22} />
              {seasonTitle}
            </h1>
          </div>
          <p className="text-xs text-gray-500 font-mono mt-1 tracking-wider uppercase">
            Healthy with sports consistency leaderboard — Where consistency becomes rank
          </p>
        </div>

        {/* Global actions and toggles */}
        <div className="flex items-center gap-3 w-full md:w-auto justify-between md:justify-end">
          <div className="flex items-center bg-gray-950 p-1 rounded-xl border border-gray-800">
            <button
              onClick={() => setViewMode('table')}
              className={`p-2 rounded-lg transition-colors ${
                viewMode === 'table' ? 'bg-gray-800 text-system-blue' : 'text-gray-500 hover:text-gray-300'
              }`}
              title="Table View"
            >
              <Table2 size={16} />
            </button>
            <button
              onClick={() => {
                setViewMode('cards');
                setCardPage(1);
              }}
              className={`p-2 rounded-lg transition-colors ${
                viewMode === 'cards' ? 'bg-gray-800 text-system-blue' : 'text-gray-500 hover:text-gray-300'
              }`}
              title="Grid View"
            >
              <LayoutGrid size={16} />
            </button>
          </div>

          <button
            type="button"
            aria-pressed={theme === 'dark'}
            onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
            className="flex items-center gap-2 px-3 py-2 rounded-xl bg-gray-950 hover:bg-gray-900 border border-gray-800 font-mono text-xs text-gray-300 hover:text-white transition-colors"
            title="Toggle light/dark theme"
          >
            {theme === 'dark' ? <Moon size={14} /> : <Sun size={14} />}
            {theme === 'dark' ? 'Dark' : 'Light'}
          </button>

          <button
            type="button"
            onClick={() => setPage('login')}
            className="flex items-center gap-2 px-3 py-2 rounded-xl bg-system-blue/10 hover:bg-system-blue/20 border border-system-blue/30 font-mono text-xs text-system-blue hover:text-white transition-colors"
          >
            <LogIn size={14} />
            Masuk
          </button>

          <button
            onClick={() => refresh()}
            disabled={loading || refreshing}
            className="flex items-center gap-2 px-4.5 py-2 rounded-xl bg-gray-950 hover:bg-gray-900 border border-gray-800 font-mono text-xs text-gray-300 hover:text-white transition-colors disabled:opacity-50"
          >
            <RefreshCw size={12} className={`${refreshing ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </div>
      </header>

      {/* Main Content Dashboard Frame */}
      <main className="max-w-7xl mx-auto">
        {page === 'login' && (
          <LoginPage
            onLoginSuccess={(user) => {
              setPersonalUser(user);
              setPage('personal');
            }}
            onBack={() => setPage('dashboard')}
          />
        )}

        {page === 'personal' && personalUser && (
          <PersonalPage
            user={personalUser}
            onLogout={() => {
              setPersonalUser(null);
              setPage('dashboard');
            }}
          />
        )}

        {page === 'dashboard' && (
          <>
            {/* Error Alert panel */}
            {error && (
              <div className="mb-6 p-4.5 rounded-2xl bg-system-red/10 border border-system-red/35 flex items-start gap-3 text-sm text-red-300">
                <AlertCircle className="text-system-red mt-0.5 shrink-0" size={18} />
                <div>
                  <p className="font-bold font-mono uppercase text-xs tracking-wider text-system-red">
                    System Link Disrupted
                  </p>
                  <p className="mt-1 text-xs font-mono">{error instanceof Error ? error.message : String(error)}</p>
                  <button
                    onClick={() => refresh()}
                    className="mt-3 text-xs font-bold text-white hover:underline uppercase tracking-wide block font-mono"
                  >
                    Reconnect System Core
                  </button>
                </div>
              </div>
            )}

            {/* Stats Section */}
            <StatsOverview summary={summary ?? null} loading={loading} />

            {!loading && !error && hunters.length > 0 && <MotivationBanner />}

            {/* Dynamic Display Area */}
            {loading ? (
              <div className="py-20 flex flex-col items-center justify-center">
                <div className="relative w-16 h-16">
                  <div className="absolute inset-0 rounded-full border-4 border-gray-800"></div>
                  <div className="absolute inset-0 rounded-full border-4 border-t-system-blue border-r-transparent border-b-transparent border-l-transparent animate-spin"></div>
                </div>
                <p className="text-xs text-gray-400 mt-6 font-mono tracking-widest uppercase animate-pulse">
                  Retrieving Hunter Status...
                </p>
              </div>
            ) : hunters.length === 0 ? (
              <div className="glass rounded-3xl p-16 text-center border border-gray-800">
                <ShieldAlert className="text-system-gold mx-auto mb-4 animate-bounce" size={40} />
                <h3 className="text-lg font-bold text-white font-orbitron">No Roster Registered</h3>
                <p className="text-xs text-gray-500 font-mono mt-2 max-w-sm mx-auto">
                  The database does not contain any active reports yet. Have members submit a workout report using "/lapor" on WhatsApp!
                </p>
              </div>
            ) : viewMode === 'table' ? (
              <LeaderboardTable hunters={hunters} onSelectHunter={setSelectedHunter} />
            ) : (
              <>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
                  {visibleCardHunters.map((hunter) => (
                    <HunterCard
                      key={hunter.user_id}
                      hunter={hunter}
                      onClick={() => setSelectedHunter(hunter)}
                    />
                  ))}
                </div>
                <nav className="mt-6 flex flex-col sm:flex-row items-center justify-between gap-3 glass rounded-2xl p-4" aria-label="Card pagination">
                  <p className="text-xs text-gray-500 font-mono uppercase tracking-wider">
                    Showing {visibleCardHunters.length === 0 ? 0 : (safeCardPage - 1) * PAGE_SIZE + 1}-{Math.min(safeCardPage * PAGE_SIZE, hunters.length)} of {hunters.length}
                  </p>
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={() => goToCardPage(safeCardPage - 1)}
                      disabled={safeCardPage <= 1}
                      className="px-3 py-2 rounded-xl bg-gray-950 border border-gray-800 text-xs font-mono text-gray-300 disabled:opacity-40 hover:text-white transition-colors"
                    >
                      Previous
                    </button>
                    <span className="px-3 py-2 text-xs font-mono text-gray-500">
                      Page {safeCardPage} / {cardTotalPages}
                    </span>
                    <button
                      type="button"
                      onClick={() => goToCardPage(safeCardPage + 1)}
                      disabled={safeCardPage >= cardTotalPages}
                      className="px-3 py-2 rounded-xl bg-gray-950 border border-gray-800 text-xs font-mono text-gray-300 disabled:opacity-40 hover:text-white transition-colors"
                    >
                      Next
                    </button>
                  </div>
                </nav>
              </>
            )}
          </>
        )}
      </main>

      {/* Profile/RPG Status Screen Modal */}
      {selectedHunter && (
        <ProfileModal
          hunter={selectedHunter}
          onClose={() => setSelectedHunter(null)}
        />
      )}

      {/* Futuristic footer HUD */}
      <footer className="max-w-7xl mx-auto mt-16 pt-6 border-t border-gray-900/60 flex flex-col sm:flex-row justify-between items-center gap-4 text-[10px] text-gray-600 font-mono tracking-widest uppercase">
        <p>SYSTEM CORE V1.5.0-ALPHA // SECURE CONNECTION</p>
        <p>© 2026 WHATSAPP ACTIVITY TRACKER. ALL RIGHTS RESERVED.</p>
      </footer>
    </div>
  );
}

export default App;
