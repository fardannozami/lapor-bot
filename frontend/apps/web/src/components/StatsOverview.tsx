import React from 'react';
import { Users, Flame, Activity, Shield } from 'lucide-react';
import type { GlobalSummary } from '@lapor-bot/shared';

interface StatsOverviewProps {
  summary: GlobalSummary | null;
  loading: boolean;
}

export const StatsOverview: React.FC<StatsOverviewProps> = ({ summary, loading }) => {
  const stats = [
    {
      label: "Total Hunters Active",
      value: loading ? "..." : summary?.total_participants ?? 0,
      icon: Users,
      color: "text-system-blue",
      glow: "glass-glow-blue",
      description: "Hunters who reported this season",
    },
    {
      label: "Keep Streak 🔥",
      value: loading ? "..." : `${summary?.active_streak_count ?? 0} Players`,
      icon: Flame,
      color: "text-system-red animate-pulse",
      glow: "glass-glow-red",
      description: "Active streak holders this week",
    },
    {
      label: "Total Workouts Logged",
      value: loading ? "..." : summary?.total_workouts_logged ?? 0,
      icon: Activity,
      color: "text-system-green",
      glow: "glass-glow-blue",
      description: "Total workouts in lifetime history",
    },
    {
      label: "System Current State",
      value: loading ? "..." : `Season ${summary?.current_season ?? 1} - Day ${summary?.current_day ?? 1}`,
      icon: Shield,
      color: "text-system-gold",
      glow: "glass-glow-gold",
      description: "Active campaign stage progress",
    },
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 mb-8">
      {stats.map((stat, idx) => {
        const Icon = stat.icon;
        return (
          <div
            key={idx}
            className={`glass p-6 rounded-2xl transition-all duration-300 hover:scale-[1.02] hover:-translate-y-1 ${stat.glow}`}
          >
            <div className="flex items-start justify-between">
              <div>
                <p className="text-gray-400 text-sm font-medium">{stat.label}</p>
                <h3 className="text-3xl font-bold font-orbitron tracking-wide mt-2 text-white">
                  {stat.value}
                </h3>
              </div>
              <div className={`p-3 rounded-xl bg-gray-900/60 ${stat.color}`}>
                <Icon size={24} />
              </div>
            </div>
            <p className="text-xs text-gray-500 mt-4 font-mono">{stat.description}</p>
          </div>
        );
      })}
    </div>
  );
};
