import React, { useState, useMemo } from 'react';
import { Search, SlidersHorizontal, Flame, Trophy, Award, Activity, User } from 'lucide-react';
import type { EnrichedReport } from '../types';

interface LeaderboardTableProps {
  hunters: EnrichedReport[];
  onSelectHunter: (hunter: EnrichedReport) => void;
}

type TabType = 'seasonal' | 'lifetime' | 'streak';
const PAGE_SIZE = 15;

export const LeaderboardTable: React.FC<LeaderboardTableProps> = ({ hunters, onSelectHunter }) => {
  const [search, setSearch] = useState('');
  const [selectedJob, setSelectedJob] = useState('all');
  const [activeTab, setActiveTab] = useState<TabType>('seasonal');
  const [page, setPage] = useState(1);

  const jobsList = [
    { id: 'all', name: 'All Jobs' },
    { id: 'fighter', name: 'Fighter ⚔️' },
    { id: 'tank', name: 'Tanker 🛡️' },
    { id: 'assassin', name: 'Assassin 🗡️' },
    { id: 'mage', name: 'Mage 🔥' },
    { id: 'ranger', name: 'Ranger 🏹' },
    { id: 'healer', name: 'Healer 💚' },
    { id: 'necromancer', name: 'Necromancer 🌑' },
  ];

  // Process data based on active tab, search, and job filters
  const filteredAndSorted = useMemo(() => {
    let result = [...hunters];

    // Filter by search name
    if (search.trim()) {
      const q = search.toLowerCase();
      result = result.filter(h => h.name.toLowerCase().includes(q));
    }

    // Filter by job
    if (selectedJob !== 'all') {
      result = result.filter(h => h.job_class?.toLowerCase() === selectedJob.toLowerCase());
    }

    // Sort based on active tab
    result.sort((a, b) => {
      if (activeTab === 'seasonal') {
        if (b.seasonal_points === a.seasonal_points) {
          return b.seasonal_activity_count - a.seasonal_activity_count;
        }
        return b.seasonal_points - a.seasonal_points;
      } else if (activeTab === 'lifetime') {
        if (b.total_points === a.total_points) {
          return b.activity_count - a.activity_count;
        }
        return b.total_points - a.total_points;
      } else { // activeTab === 'streak'
        if (b.streak === a.streak) {
          return b.seasonal_activity_count - a.seasonal_activity_count;
        }
        return b.streak - a.streak;
      }
    });

    return result;
  }, [hunters, search, selectedJob, activeTab]);

  const totalPages = Math.max(1, Math.ceil(filteredAndSorted.length / PAGE_SIZE));
  const safePage = Math.min(page, totalPages);
  const visibleHunters = useMemo(() => {
    const start = (safePage - 1) * PAGE_SIZE;
    return filteredAndSorted.slice(start, start + PAGE_SIZE);
  }, [filteredAndSorted, safePage]);
  const pageStartRank = (safePage - 1) * PAGE_SIZE;

  const goToPage = (nextPage: number) => {
    setPage(Math.min(Math.max(1, nextPage), totalPages));
  };

  const getRankBadge = (idx: number) => {
    switch (idx) {
      case 0: return <span className="flex items-center justify-center w-7 h-7 rounded-full bg-system-gold/20 text-system-gold border border-system-gold/50 shadow-neon-gold text-xs font-bold font-orbitron">1st</span>;
      case 1: return <span className="flex items-center justify-center w-7 h-7 rounded-full bg-slate-300/20 text-slate-300 border border-slate-300/40 text-xs font-bold font-orbitron">2nd</span>;
      case 2: return <span className="flex items-center justify-center w-7 h-7 rounded-full bg-amber-700/20 text-amber-600 border border-amber-600/40 text-xs font-bold font-orbitron">3rd</span>;
      default: return <span className="text-gray-400 font-mono text-xs">{idx + 1}</span>;
    }
  };

  const getJobBadgeClass = (jobId: string) => {
    switch (jobId?.toLowerCase()) {
      case 'fighter': return 'text-system-red bg-system-red/5 border-system-red/20';
      case 'tank': return 'text-system-gold bg-system-gold/5 border-system-gold/20';
      case 'assassin': return 'text-system-purple bg-system-purple/5 border-system-purple/20';
      case 'mage': return 'text-red-400 bg-red-400/5 border-red-400/20';
      case 'ranger': return 'text-system-blue bg-system-blue/5 border-system-blue/20';
      case 'healer': return 'text-system-green bg-system-green/5 border-system-green/20';
      case 'necromancer': return 'text-gray-400 bg-gray-800/40 border-gray-600/20';
      default: return 'text-gray-500 bg-gray-800/20 border-gray-700/10';
    }
  };

  return (
    <div className="glass rounded-3xl p-5 md:p-6 mb-8">
      {/* Top Controls: Tabs and Filters */}
      <div className="flex flex-col lg:flex-row gap-4 items-stretch lg:items-center justify-between mb-6 border-b border-gray-850 pb-5">
        {/* Leaderboard Mode Tabs */}
        <div className="flex bg-gray-950/80 p-1.5 rounded-xl border border-gray-800/60 max-w-fit">
          <button
            onClick={() => {
              setActiveTab('seasonal');
              setPage(1);
            }}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold font-orbitron tracking-wider transition-all uppercase ${
              activeTab === 'seasonal'
                ? 'bg-gradient-to-r from-system-purple to-system-blue text-white shadow-neon-blue'
                : 'text-gray-400 hover:text-white'
            }`}
          >
            <Trophy size={14} />
            Season Ranks
          </button>
          <button
            onClick={() => {
              setActiveTab('lifetime');
              setPage(1);
            }}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold font-orbitron tracking-wider transition-all uppercase ${
              activeTab === 'lifetime'
                ? 'bg-gradient-to-r from-system-purple to-system-blue text-white shadow-neon-blue'
                : 'text-gray-400 hover:text-white'
            }`}
          >
            <Award size={14} />
            Lifetime XP
          </button>
          <button
            onClick={() => {
              setActiveTab('streak');
              setPage(1);
            }}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold font-orbitron tracking-wider transition-all uppercase ${
              activeTab === 'streak'
                ? 'bg-gradient-to-r from-system-purple to-system-blue text-white shadow-neon-blue'
                : 'text-gray-400 hover:text-white'
            }`}
          >
            <Flame size={14} />
            Streak Masters
          </button>
        </div>

        {/* Filter Inputs */}
        <div className="flex flex-col sm:flex-row gap-3 sm:items-center">
          {/* Search bar */}
          <div className="relative">
            <input
              type="text"
              placeholder="Search hunter..."
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
              className="w-full sm:w-60 pl-10 pr-4 py-2 bg-gray-950/70 border border-gray-800 focus:border-system-blue focus:shadow-neon-blue rounded-xl text-xs text-white placeholder-gray-500 font-mono transition-all outline-none"
            />
            <Search className="absolute left-3.5 top-2.5 text-gray-500" size={14} />
          </div>

          {/* Job Filter */}
          <div className="relative flex items-center">
            <SlidersHorizontal className="absolute left-3.5 top-2.5 text-gray-500" size={14} />
            <select
              value={selectedJob}
              onChange={(e) => {
                setSelectedJob(e.target.value);
                setPage(1);
              }}
              className="w-full sm:w-44 pl-10 pr-8 py-2 bg-gray-950/70 border border-gray-800 focus:border-system-blue focus:shadow-neon-blue rounded-xl text-xs text-white font-mono transition-all outline-none appearance-none cursor-pointer"
            >
              {jobsList.map(job => (
                <option key={job.id} value={job.id} className="bg-dark-bg text-white">
                  {job.name}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Table Renders */}
      <div className="overflow-x-auto">
        <table className="w-full min-w-[700px] border-collapse">
          <thead>
            <tr className="border-b border-gray-850 text-left text-gray-500 text-[10px] font-mono font-bold tracking-widest uppercase">
              <th className="pb-3 pl-4 w-12 text-center">Rank</th>
              <th className="pb-3 pl-4">Hunter</th>
              <th className="pb-3 pl-4">Job Class</th>
              <th className="pb-3 pl-4 text-center">Level</th>
              <th className="pb-3 pl-4 text-center">Streak</th>
              <th className="pb-3 pl-4 text-center">Active Days</th>
              <th className="pb-3 pl-4 text-right pr-4">
                {activeTab === 'seasonal' ? 'Season Points' : activeTab === 'lifetime' ? 'Total XP' : 'Current Streak'}
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-900/60">
            {filteredAndSorted.length === 0 ? (
              <tr>
                <td colSpan={7} className="py-8 text-center text-xs text-gray-500 font-mono italic">
                  No active hunters match the filters.
                </td>
              </tr>
            ) : (
              visibleHunters.map((hunter, idx) => {
                const rank = pageStartRank + idx;
                return (
                <tr
                  key={hunter.user_id}
                  onClick={() => onSelectHunter(hunter)}
                  className={`group transition-all hover:bg-gray-800/20 cursor-pointer ${
                    rank < 3 ? 'bg-gradient-to-r from-gray-950/20 to-transparent' : ''
                  }`}
                >
                  {/* Rank */}
                  <td className="py-4 pl-4 text-center font-mono align-middle">
                    <div className="flex justify-center">{getRankBadge(rank)}</div>
                  </td>

                  {/* Name */}
                  <td className="py-4 pl-4 align-middle">
                    <div className="flex items-center gap-2">
                      <div className="relative">
                        <div className="w-8 h-8 rounded-full bg-gray-900 flex items-center justify-center border border-gray-800 group-hover:border-system-blue transition-colors">
                          <User size={14} className="text-gray-400" />
                        </div>
                        {hunter.is_active_today && (
                          <span className="absolute bottom-0 right-0 w-2.5 h-2.5 bg-system-green border-2 border-dark-bg rounded-full animate-pulse"></span>
                        )}
                      </div>
                      <div>
                        <div className="text-sm font-semibold text-white group-hover:text-system-blue transition-colors">
                          {hunter.name}
                        </div>
                        <div className="text-[10px] text-gray-500 font-mono">{hunter.user_id}</div>
                      </div>
                    </div>
                  </td>

                  {/* Job Class */}
                  <td className="py-4 pl-4 align-middle">
                    <span className={`text-xs px-2.5 py-1 rounded-md border font-mono ${getJobBadgeClass(hunter.job_class)}`}>
                      {hunter.job_icon} {hunter.job_name}
                    </span>
                  </td>

                  {/* Level */}
                  <td className="py-4 pl-4 text-center font-mono align-middle text-sm text-gray-300">
                    <span className="font-bold">Lv.{hunter.level}</span>
                  </td>

                  {/* Streak */}
                  <td className="py-4 pl-4 text-center align-middle font-mono text-sm">
                    <span className="flex items-center justify-center gap-1 text-system-red">
                      <Flame size={14} />
                      <span className="font-bold">{hunter.streak}</span>
                      <span className="text-[10px] text-gray-500">w</span>
                    </span>
                  </td>

                  {/* Active Days */}
                  <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
                    <span className="flex items-center justify-center gap-1">
                      <Activity size={12} className="text-system-blue" />
                      <span className="font-bold">{hunter.seasonal_activity_count}</span>
                      <span className="text-[10px] text-gray-500">d</span>
                    </span>
                  </td>

                  {/* Sort Target Value */}
                  <td className="py-4 pl-4 text-right pr-4 font-mono font-bold align-middle">
                    {activeTab === 'seasonal' ? (
                      <span className="text-system-gold">{hunter.seasonal_points} <span className="text-[9px] text-gray-500">PTS</span></span>
                    ) : activeTab === 'lifetime' ? (
                      <span className="text-system-blue">{hunter.total_points} <span className="text-[9px] text-gray-500">XP</span></span>
                    ) : (
                      <span className="text-system-red">{hunter.streak} <span className="text-[9px] text-gray-500">WEEKS</span></span>
                    )}
                  </td>
                </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
      <nav className="mt-5 flex flex-col sm:flex-row items-center justify-between gap-3 border-t border-gray-900/60 pt-5" aria-label="Leaderboard pagination">
        <p className="text-xs text-gray-500 font-mono uppercase tracking-wider">
          Showing {filteredAndSorted.length === 0 ? 0 : pageStartRank + 1}-{Math.min(pageStartRank + PAGE_SIZE, filteredAndSorted.length)} of {filteredAndSorted.length} athletes
        </p>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => goToPage(safePage - 1)}
            disabled={safePage <= 1}
            className="px-3 py-2 rounded-xl bg-gray-950 border border-gray-800 text-xs font-mono text-gray-300 disabled:opacity-40 hover:text-white transition-colors"
          >
            Previous
          </button>
          <span className="px-3 py-2 text-xs font-mono text-gray-500">
            Page {safePage} / {totalPages}
          </span>
          <button
            type="button"
            onClick={() => goToPage(safePage + 1)}
            disabled={safePage >= totalPages}
            className="px-3 py-2 rounded-xl bg-gray-950 border border-gray-800 text-xs font-mono text-gray-300 disabled:opacity-40 hover:text-white transition-colors"
          >
            Next
          </button>
        </div>
      </nav>
    </div>
  );
};
