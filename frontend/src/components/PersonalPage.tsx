import { ArrowLeft, Shield, Swords, Zap, Heart, Trophy, LogOut, ScrollText, CheckCircle2, Circle } from 'lucide-react';
import type { EnrichedReport, QuestTask } from '../types';

interface PersonalPageProps {
  user: EnrichedReport;
  onLogout: () => void;
}

function getRankGlow(rankName: string) {
  if (rankName.includes('S-Rank') || rankName.includes('Monarch')) return 'glass-glow-gold';
  if (rankName.includes('A-Rank')) return 'glass-glow-purple';
  if (rankName.includes('B-Rank') || rankName.includes('C-Rank')) return 'glass-glow-blue';
  return 'glass-glow-blue';
}

function formatQuestDifficulty(difficulty: string) {
  if (difficulty === 'easy') return '🟢 Easy';
  if (difficulty === 'medium') return '🟡 Medium';
  if (difficulty === 'hard') return '🔴 Hard';
  return difficulty;
}

function formatQuestTarget(task: QuestTask) {
  if (task.id === 'easycardio') return 'jalan kaki 4000 langkah atau sepeda 5 km';
  if (task.unit === '100m') return `${(task.target / 10).toFixed(1)} km`;
  return `${task.target} ${task.unit}`;
}

function formatQuestProgress(task: QuestTask) {
  if (task.id === 'easycardio') return task.progress >= task.target ? 'selesai' : 'belum selesai';
  if (task.unit === '100m') return `${(task.progress / 10).toFixed(1)} / ${(task.target / 10).toFixed(1)} km`;
  return `${task.progress} / ${task.target} ${task.unit}`;
}

export function PersonalPage({ user, onLogout }: PersonalPageProps) {
  const glowClass = getRankGlow(user.rank_name);
  const sideQuests = user.today_side_quests ?? [];

  return (
    <div className="min-h-[60vh] flex flex-col items-center px-4 py-8">
      <div className="w-full max-w-2xl">
        {/* Top bar */}
        <div className="flex items-center justify-between mb-6">
          <button
            onClick={onLogout}
            className="flex items-center gap-2 px-3 py-2 rounded-xl bg-gray-950 hover:bg-gray-900 border border-gray-800 text-gray-400 hover:text-white font-mono text-xs transition-colors"
          >
            <ArrowLeft size={14} />
            Kembali
          </button>

          <button
            onClick={onLogout}
            className="flex items-center gap-2 px-3 py-2 rounded-xl bg-system-red/10 hover:bg-system-red/20 border border-system-red/20 text-system-red font-mono text-xs transition-colors"
          >
            <LogOut size={14} />
            Keluar
          </button>
        </div>

        {/* User identity card */}
        <div className={`glass rounded-3xl p-6 mb-6 ${glowClass}`}>
          <div className="flex items-start gap-4">
            <div className="w-14 h-14 rounded-2xl bg-gray-950 border border-gray-800 flex items-center justify-center text-2xl shrink-0">
              {user.job_icon}
            </div>
            <div className="flex-1 min-w-0">
              <h2 className="text-xl font-bold font-orbitron text-white tracking-wide truncate">
                {user.name}
              </h2>
              <p className="text-xs text-gray-500 font-mono">
                {user.job_name} {user.level_icon} Lv.{user.level}
              </p>
              <p className="text-[10px] text-gray-600 font-mono mt-1">
                {user.user_id}
              </p>
            </div>
            <div className="text-right shrink-0">
              <div className="text-xs text-gray-400 font-mono uppercase tracking-wider">Rank</div>
              <div className="text-sm font-bold font-orbitron text-white">
                {user.rank_icon} {user.rank_name}
              </div>
            </div>
          </div>

          {/* Quick stats row */}
          <div className="grid grid-cols-4 gap-3 mt-5 pt-4 border-t border-gray-800/50">
            {[
              { icon: Shield, label: 'Str', value: user.str },
              { icon: Zap, label: 'Sta', value: user.sta },
              { icon: Swords, label: 'Agi', value: user.agi },
              { icon: Heart, label: 'Vit', value: user.vit },
            ].map((stat) => {
              const Icon = stat.icon;
              return (
                <div key={stat.label} className="text-center">
                  <Icon className="text-gray-500 mx-auto mb-1" size={14} />
                  <div className="text-sm font-bold font-orbitron text-white">{stat.value}</div>
                  <div className="text-[10px] text-gray-500 font-mono uppercase">{stat.label}</div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Placeholder content */}
        <div className={`glass rounded-3xl p-10 text-center ${glowClass}`}>
          <Trophy className="text-system-gold mx-auto mb-4" size={40} />
          <h3 className="text-lg font-bold font-orbitron text-white">
            {user.rank_icon} Personal Dashboard — Coming Soon
          </h3>
          <p className="text-xs text-gray-500 font-mono mt-2 max-w-md mx-auto leading-relaxed">
            Data personal lengkap akan ditampilkan di sini — statistik detail,
            progress goals, history workout, dan achievements kamu.
          </p>

          <div className="grid grid-cols-2 gap-3 mt-6 max-w-sm mx-auto">
            {[
              { label: 'Total Points', value: user.total_points },
              { label: 'Streak', value: `${user.streak} hari` },
              { label: 'Max Streak', value: `${user.max_streak} hari` },
              { label: 'Season Points', value: user.seasonal_points },
            ].map((stat) => (
              <div
                key={stat.label}
                className="glass rounded-xl p-3 border border-gray-800/30"
              >
                <div className="text-sm font-bold font-orbitron text-white">
                  {stat.value}
                </div>
                <div className="text-[10px] text-gray-500 font-mono uppercase">
                  {stat.label}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Daily side quests */}
        <div className={`glass rounded-3xl p-6 mt-6 ${glowClass}`}>
          <div className="flex items-start justify-between gap-4 mb-5">
            <div>
              <h3 className="text-lg font-bold font-orbitron text-white flex items-center gap-2">
                <ScrollText className="text-system-green" size={18} />
                Side Quest Hari Ini
              </h3>
              <p className="text-xs text-gray-500 font-mono mt-1 leading-relaxed">
                Sama seperti <span className="text-gray-300">/lapor sidequest</span>. Selesaikan via WhatsApp dengan format <span className="text-gray-300">/lapor sidequest &lt;kegiatan&gt; &lt;jumlah&gt;</span>.
              </p>
            </div>
            <div className="text-right shrink-0">
              <div className="text-sm font-bold font-orbitron text-white">
                {sideQuests.filter((quest) => quest.progress >= quest.target).length}/{sideQuests.length}
              </div>
              <div className="text-[10px] text-gray-500 font-mono uppercase">Selesai</div>
            </div>
          </div>

          {sideQuests.length === 0 ? (
            <div className="rounded-2xl border border-gray-800/50 bg-gray-950/50 p-4 text-center">
              <p className="text-sm text-gray-400 font-mono">Side quest belum terbuka.</p>
              <p className="text-xs text-gray-600 font-mono mt-1">Side quest tersedia untuk profil yang sudah punya job.</p>
            </div>
          ) : (
            <div className="space-y-3">
              {sideQuests.map((quest) => {
                const done = quest.progress >= quest.target;
                return (
                  <div key={quest.id} className="rounded-2xl border border-gray-800/50 bg-gray-950/50 p-4">
                    <div className="flex items-start gap-3">
                      {done ? (
                        <CheckCircle2 className="text-system-green shrink-0 mt-0.5" size={18} />
                      ) : (
                        <Circle className="text-gray-600 shrink-0 mt-0.5" size={18} />
                      )}
                      <div className="flex-1 min-w-0">
                        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1">
                          <p className="text-sm font-bold text-white font-mono">{quest.name}</p>
                          <span className="text-[10px] text-gray-400 font-mono uppercase tracking-wider">
                            {formatQuestDifficulty(quest.difficulty)}
                          </span>
                        </div>
                        <p className="text-xs text-gray-500 font-mono mt-1">Target: {formatQuestTarget(quest)}</p>
                        <p className="text-xs text-gray-600 font-mono mt-1">Progress: {formatQuestProgress(quest)}</p>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
