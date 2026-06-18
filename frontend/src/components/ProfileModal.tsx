import React from 'react';
import { X, Award, Shield, Heart, Zap, Swords, Flame, Trophy, Compass } from 'lucide-react';
import type { EnrichedReport } from '../types';

interface ProfileModalProps {
  hunter: EnrichedReport | null;
  onClose: () => void;
}

export const ProfileModal: React.FC<ProfileModalProps> = ({ hunter, onClose }) => {
  if (!hunter) return null;

  const xpPercent = Math.min(
    100,
    Math.round((hunter.xp_progress.CurrentXP / hunter.xp_progress.RequiredXP) * 100)
  );

  // Map job class ID to style colors
  const getJobColor = (jobId: string) => {
    switch (jobId?.toLowerCase()) {
      case 'fighter': return 'text-system-red bg-system-red/10 border-system-red/30';
      case 'tank': return 'text-system-gold bg-system-gold/10 border-system-gold/30';
      case 'assassin': return 'text-system-purple bg-system-purple/10 border-system-purple/30';
      case 'mage': return 'text-red-400 bg-red-400/10 border-red-400/30';
      case 'ranger': return 'text-system-blue bg-system-blue/10 border-system-blue/30';
      case 'healer': return 'text-system-green bg-system-green/10 border-system-green/30';
      case 'necromancer': return 'text-gray-400 bg-gray-800/40 border-gray-600/30';
      default: return 'text-gray-400 bg-gray-800/30 border-gray-700/30';
    }
  };

  const getRankGlow = (rankName: string) => {
    if (rankName.includes('S-Rank') || rankName.includes('Monarch')) return 'glass-glow-gold';
    if (rankName.includes('A-Rank')) return 'glass-glow-purple';
    if (rankName.includes('B-Rank')) return 'glass-glow-blue';
    return '';
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-fade-in">
      {/* Background click dismiss */}
      <div className="absolute inset-0" onClick={onClose}></div>

      {/* Modal Container */}
      <div className={`relative w-full max-w-2xl overflow-hidden glass rounded-3xl p-6 md:p-8 z-10 max-h-[90vh] overflow-y-auto ${getRankGlow(hunter.rank_name)}`}>
        {/* Header */}
        <div className="flex items-start justify-between mb-6 border-b border-gray-800 pb-4">
          <div>
            <div className="flex items-center gap-3">
              <h2 className="text-2xl md:text-3xl font-extrabold font-orbitron text-white tracking-wide">
                {hunter.name}
              </h2>
              <span className={`text-xs px-2.5 py-1 rounded-full border font-mono ${getJobColor(hunter.job_class)}`}>
                {hunter.job_icon} {hunter.job_name}
              </span>
            </div>
            <p className="text-xs text-gray-500 font-mono mt-1">ID: {hunter.user_id}</p>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-gray-400 hover:text-white rounded-lg bg-gray-800/40 hover:bg-gray-800 transition-colors"
          >
            <X size={20} />
          </button>
        </div>

        {/* Content Layout */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Left Side: RPG Level & Stats */}
          <div className="space-y-6">
            {/* Level Panel */}
            <div className="bg-gray-900/50 border border-gray-800 p-5 rounded-2xl">
              <div className="flex justify-between items-center mb-3">
                <div>
                  <span className="text-xs text-system-blue font-mono font-bold uppercase tracking-widest">
                    Level Progress
                  </span>
                  <div className="flex items-baseline gap-1 mt-1">
                    <span className="text-3xl font-black font-orbitron text-white">
                      Lv.{hunter.level}
                    </span>
                    <span className="text-xs text-gray-400 font-mono">({hunter.level_name} {hunter.level_icon})</span>
                  </div>
                </div>
                <div className="text-right">
                  <span className="text-xs text-gray-400 font-mono block">Lifetime XP</span>
                  <span className="text-sm font-bold font-orbitron text-white mt-1 block">
                    {hunter.total_points} PTS
                  </span>
                </div>
              </div>

              {/* Progress Bar */}
              <div className="w-full bg-gray-800 h-3.5 rounded-full overflow-hidden p-[2px]">
                <div 
                  className="bg-gradient-to-r from-system-blue to-system-purple h-full rounded-full transition-all duration-500 ease-out shadow-neon-blue"
                  style={{ width: `${xpPercent}%` }}
                ></div>
              </div>
              <div className="flex justify-between items-center text-[10px] text-gray-400 font-mono mt-2">
                <span>{hunter.xp_progress.CurrentXP} XP</span>
                <span className="text-system-blue font-bold">{xpPercent}%</span>
                <span>{hunter.xp_progress.RequiredXP} XP</span>
              </div>
            </div>

            {/* Attributes Section */}
            <div className="bg-gray-900/50 border border-gray-800 p-5 rounded-2xl">
              <h3 className="text-xs text-system-purple font-mono font-bold uppercase tracking-widest mb-4">
                Hunter Attributes
              </h3>
              <div className="space-y-3">
                {/* STR */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-system-red">
                    <Swords size={16} />
                    <span className="text-xs font-mono font-semibold">STR (Strength/Gym)</span>
                  </div>
                  <span className="text-sm font-bold font-orbitron text-white">{hunter.str}</span>
                </div>
                {/* STA */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-system-blue">
                    <Zap size={16} />
                    <span className="text-xs font-mono font-semibold">STA (Stamina/Run)</span>
                  </div>
                  <span className="text-sm font-bold font-orbitron text-white">{hunter.sta}</span>
                </div>
                {/* AGI */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-system-purple">
                    <Shield size={16} />
                    <span className="text-xs font-mono font-semibold">AGI (Agility/Sport)</span>
                  </div>
                  <span className="text-sm font-bold font-orbitron text-white">{hunter.agi}</span>
                </div>
                {/* VIT */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-system-green">
                    <Heart size={16} />
                    <span className="text-xs font-mono font-semibold">VIT (Vitality/Yoga)</span>
                  </div>
                  <span className="text-sm font-bold font-orbitron text-white">{hunter.vit}</span>
                </div>
              </div>
            </div>

            {/* Campaign Metrics */}
            <div className="bg-gray-900/50 border border-gray-800 p-4 rounded-2xl grid grid-cols-2 gap-4">
              <div className="text-center p-3 bg-gray-950/40 rounded-xl">
                <span className="text-[10px] text-gray-500 font-mono block">Side Quests</span>
                <span className="text-xl font-bold font-orbitron text-system-blue mt-1 block">
                  {hunter.total_side_quests}
                </span>
                <span className="text-[9px] text-gray-500 font-mono block">({hunter.seasonal_side_quests} this season)</span>
              </div>
              <div className="text-center p-3 bg-gray-950/40 rounded-xl">
                <span className="text-[10px] text-gray-500 font-mono block">Goals Done</span>
                <span className="text-xl font-bold font-orbitron text-system-gold mt-1 block">
                  {hunter.goals_completed}
                </span>
                <span className="text-[9px] text-gray-500 font-mono block">Weekly Goals</span>
              </div>
            </div>
          </div>

          {/* Right Side: Achievements, Streaks & Class Traits */}
          <div className="space-y-6">
            {/* Job Details */}
            <div className="bg-gray-900/50 border border-gray-800 p-5 rounded-2xl">
              <h3 className="text-xs text-system-gold font-mono font-bold uppercase tracking-widest mb-2">
                Job Profile
              </h3>
              <p className="text-sm text-gray-300 mt-1 leading-relaxed">
                {hunter.job_description}
              </p>
              {hunter.job_trait && (
                <div className="mt-3 p-2.5 rounded-lg bg-gray-950/50 border border-gray-800 text-xs text-gray-400 font-mono">
                  <span className="text-system-gold font-bold">Trait:</span> {hunter.job_trait}
                </div>
              )}
            </div>

            {/* Streak Highlights */}
            <div className="bg-gray-900/50 border border-gray-800 p-5 rounded-2xl">
              <h3 className="text-xs text-system-red font-mono font-bold uppercase tracking-widest mb-4">
                Streak Records
              </h3>
              <div className="grid grid-cols-3 gap-3">
                <div className="flex flex-col items-center p-3 bg-gray-950/50 rounded-xl border border-gray-800/80">
                  <Flame size={18} className="text-system-red mb-1 animate-pulse" />
                  <span className="text-[9px] text-gray-500 font-mono uppercase">Streak</span>
                  <span className="text-lg font-bold font-orbitron text-white mt-1">
                    {hunter.streak}w
                  </span>
                </div>
                <div className="flex flex-col items-center p-3 bg-gray-950/50 rounded-xl border border-gray-800/80">
                  <Trophy size={18} className="text-system-gold mb-1" />
                  <span className="text-[9px] text-gray-500 font-mono uppercase">Max</span>
                  <span className="text-lg font-bold font-orbitron text-white mt-1">
                    {hunter.max_streak}w
                  </span>
                </div>
                <div className="flex flex-col items-center p-3 bg-gray-950/50 rounded-xl border border-gray-800/80">
                  <Compass size={18} className="text-system-blue mb-1" />
                  <span className="text-[9px] text-gray-500 font-mono uppercase">Active Days</span>
                  <span className="text-lg font-bold font-orbitron text-white mt-1">
                    {hunter.seasonal_activity_count}d
                  </span>
                </div>
              </div>
            </div>

            {/* Achievements/Badges Grid */}
            <div className="bg-gray-900/50 border border-gray-800 p-5 rounded-2xl flex-1 flex flex-col">
              <h3 className="text-xs text-system-blue font-mono font-bold uppercase tracking-widest mb-3">
                Rank & Badge Collection
              </h3>
              
              {/* Season Rank Box */}
              <div className="flex items-center gap-3 p-3 bg-gray-950/60 rounded-xl border border-gray-800 mb-4">
                <div className="text-2xl">{hunter.rank_icon}</div>
                <div>
                  <span className="text-[10px] text-gray-500 font-mono">SEASON RANK</span>
                  <p className="text-sm font-bold font-orbitron text-white mt-0.5">{hunter.rank_name}</p>
                </div>
                <div className="ml-auto text-right">
                  <span className="text-[10px] text-gray-500 font-mono block">Seasonal Points</span>
                  <span className="text-sm font-bold font-orbitron text-system-gold">{hunter.seasonal_points} PTS</span>
                </div>
              </div>

              {/* Achievements Badges */}
              <span className="text-[10px] text-gray-400 font-mono font-bold mb-2 block">Badges Earned:</span>
              <div className="flex flex-wrap gap-2 max-h-[120px] overflow-y-auto pr-1">
                {hunter.achievements.length === 0 && hunter.seasonal_achievements.length === 0 ? (
                  <p className="text-xs text-gray-600 font-mono italic">No badges acquired yet.</p>
                ) : (
                  <>
                    {/* Lifetime Badges */}
                    {hunter.achievements.map((badge, idx) => (
                      <div
                        key={`life-${idx}`}
                        className="flex items-center gap-1 text-[10px] font-mono font-semibold px-2 py-1 rounded-md bg-gray-900 border border-gray-700 text-gray-300"
                        title="Lifetime Badge"
                      >
                        <Award size={10} className="text-system-gold" />
                        {badge}
                      </div>
                    ))}
                    {/* Seasonal Badges */}
                    {hunter.seasonal_achievements.map((badge, idx) => (
                      <div
                        key={`season-${idx}`}
                        className="flex items-center gap-1 text-[10px] font-mono font-semibold px-2 py-1 rounded-md bg-system-purple/10 border border-system-purple/30 text-system-purple"
                        title="Seasonal Badge"
                      >
                        <Award size={10} />
                        {badge}
                      </div>
                    ))}
                  </>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
