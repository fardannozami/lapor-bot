import React from 'react';
import { Flame, Trophy, Activity, ArrowUpRight } from 'lucide-react';
import type { EnrichedReport } from '@lapor-bot/shared';
import { getJobColor } from '@lapor-bot/shared';

interface HunterCardProps {
  hunter: EnrichedReport;
  onClick: () => void;
}

export const HunterCard: React.FC<HunterCardProps> = ({ hunter, onClick }) => {

  const getRankGlow = (rankName: string) => {
    if (rankName.includes('S-Rank') || rankName.includes('Monarch')) return 'glass-glow-gold hover:shadow-neon-gold';
    if (rankName.includes('A-Rank')) return 'glass-glow-purple hover:shadow-neon-purple';
    if (rankName.includes('B-Rank')) return 'glass-glow-blue hover:shadow-neon-blue';
    return 'hover:border-gray-700';
  };

  return (
    <div
      onClick={onClick}
      className={`relative glass p-5 rounded-2xl cursor-pointer transition-all duration-300 hover:scale-[1.03] hover:-translate-y-1 flex flex-col justify-between overflow-hidden group ${getRankGlow(
        hunter.rank_name
      )}`}
    >
      {/* Active today pulse indicator */}
      {hunter.is_active_today && (
        <span className="absolute top-3 right-3 flex h-2.5 w-2.5">
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-system-green opacity-75"></span>
          <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-system-green"></span>
        </span>
      )}

      {/* Card Content */}
      <div>
        {/* Level and Title */}
        <div className="flex justify-between items-center mb-3">
          <span className="text-[10px] text-gray-500 font-mono tracking-widest uppercase">
            {hunter.rank_name}
          </span>
          <span className="text-xs font-bold font-orbitron text-system-blue">
            Lv.{hunter.level}
          </span>
        </div>

        {/* Profile Info */}
        <h3 className="text-lg font-bold text-white tracking-wide truncate group-hover:text-system-blue transition-colors mb-1 pr-6">
          {hunter.name}
        </h3>
        <p className="text-[10px] text-gray-500 font-mono mb-4">{hunter.user_id}</p>

        {/* Job Class Badge */}
        <div className="mb-4">
          <span className={`inline-flex items-center gap-1.5 text-[10px] font-mono px-2 py-0.5 rounded-full border ${getJobColor(hunter.job_class)}`}>
            {hunter.job_icon} {hunter.job_name}
          </span>
        </div>
      </div>

      {/* Stats footer */}
      <div className="border-t border-gray-900/60 pt-3 flex justify-between items-center text-xs font-mono">
        {/* Streak */}
        <div className="flex items-center gap-1 text-system-red">
          <Flame size={12} className={hunter.streak > 0 ? "animate-pulse" : ""} />
          <span className="font-bold">{hunter.streak}w</span>
        </div>

        {/* Active Days */}
        <div className="flex items-center gap-1 text-system-blue">
          <Activity size={12} />
          <span className="font-bold">{hunter.seasonal_activity_count}d</span>
        </div>

        {/* Points */}
        <div className="flex items-center gap-1 text-system-gold">
          <Trophy size={12} />
          <span className="font-bold">{hunter.seasonal_points}p</span>
        </div>
      </div>

      {/* Hover arrow indicator */}
      <div className="absolute bottom-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity text-system-blue">
        <ArrowUpRight size={14} />
      </div>
    </div>
  );
};
