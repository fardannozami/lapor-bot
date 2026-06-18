import { useState, useEffect } from 'react';
import { LayoutGrid, Table2, RefreshCw, AlertCircle, ShieldAlert, Cpu } from 'lucide-react';
import type { EnrichedReport, GlobalSummary } from './types';
import { StatsOverview } from './components/StatsOverview';
import { LeaderboardTable } from './components/LeaderboardTable';
import { HunterCard } from './components/HunterCard';
import { ProfileModal } from './components/ProfileModal';

function App() {
  const [hunters, setHunters] = useState<EnrichedReport[]>([]);
  const [summary, setSummary] = useState<GlobalSummary | null>(null);
  const [selectedHunter, setSelectedHunter] = useState<EnrichedReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'table' | 'cards'>('table');

  const fetchData = async (showRefreshIndicator = false) => {
    if (showRefreshIndicator) {
      setRefreshing(true);
    } else {
      setLoading(true);
    }
    setError(null);

    try {
      // Fetch summary statistics
      const summaryRes = await fetch('/api/summary');
      if (!summaryRes.ok) {
        throw new Error(`Failed to fetch summary: ${summaryRes.statusText}`);
      }
      const summaryData = (await summaryRes.json()) as GlobalSummary;
      setSummary(summaryData);

      // Fetch leaderboard list
      const leaderboardRes = await fetch('/api/leaderboard');
      if (!leaderboardRes.ok) {
        throw new Error(`Failed to fetch leaderboard: ${leaderboardRes.statusText}`);
      }
      const leaderboardData = (await leaderboardRes.json()) as EnrichedReport[];
      setHunters(leaderboardData);
    } catch (err: any) {
      console.error(err);
      setError(
        err.message || 'System Link Failure. Check connection to the Go server.'
      );
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  return (
    <div className="min-h-screen pb-16 px-4 md:px-8">
      {/* HUD System Top Bar */}
      <header className="max-w-7xl mx-auto pt-8 pb-6 border-b border-gray-850 flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="w-3 h-3 bg-system-blue shadow-neon-blue rounded-full animate-pulse"></div>
            <h1 className="text-xl md:text-2xl font-black font-orbitron tracking-widest text-white uppercase flex items-center gap-2">
              <Cpu className="text-system-blue" size={20} />
              Hunter Status Monitor
            </h1>
          </div>
          <p className="text-xs text-gray-500 font-mono mt-1 tracking-wider uppercase">
            Active Group Workout & Leveling System — [CONGRUENT OVERLAY CONNECTED]
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
              onClick={() => setViewMode('cards')}
              className={`p-2 rounded-lg transition-colors ${
                viewMode === 'cards' ? 'bg-gray-800 text-system-blue' : 'text-gray-500 hover:text-gray-300'
              }`}
              title="Grid View"
            >
              <LayoutGrid size={16} />
            </button>
          </div>

          <button
            onClick={() => fetchData(true)}
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
        {/* Error Alert panel */}
        {error && (
          <div className="mb-6 p-4.5 rounded-2xl bg-system-red/10 border border-system-red/35 flex items-start gap-3 text-sm text-red-300">
            <AlertCircle className="text-system-red mt-0.5 shrink-0" size={18} />
            <div>
              <p className="font-bold font-mono uppercase text-xs tracking-wider text-system-red">
                System Link Disrupted
              </p>
              <p className="mt-1 text-xs font-mono">{error}</p>
              <button
                onClick={() => fetchData()}
                className="mt-3 text-xs font-bold text-white hover:underline uppercase tracking-wide block font-mono"
              >
                Reconnect System Core
              </button>
            </div>
          </div>
        )}

        {/* Stats Section */}
        <StatsOverview summary={summary} loading={loading} />

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
              The database does not contain any active reports yet. Have members submit a workout report using "#lapor" on WhatsApp!
            </p>
          </div>
        ) : viewMode === 'table' ? (
          <LeaderboardTable hunters={hunters} onSelectHunter={setSelectedHunter} />
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
            {hunters.map((hunter) => (
              <HunterCard
                key={hunter.user_id}
                hunter={hunter}
                onClick={() => setSelectedHunter(hunter)}
              />
            ))}
          </div>
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
