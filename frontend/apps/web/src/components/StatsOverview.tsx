import React from 'react';
import { Users, Flame, Activity, Shield } from 'lucide-react';
import type { GlobalSummary } from '@lapor-bot/shared';

interface StatsOverviewProps {
  summary: GlobalSummary | null;
  loading: boolean;
}

type StatTone = "blue" | "red" | "green" | "gold";

interface StatItem {
  key: string;
  label: string;
  value: string | number;
  icon: React.ElementType;
  tone: StatTone;
  description: string;
  pulse?: boolean;
}

const TONE_STYLES: Record<StatTone, { text: string; glow: string }> = {
  blue: { text: "text-system-blue", glow: "glass-glow-blue" },
  red: { text: "text-system-red", glow: "glass-glow-red" },
  green: { text: "text-system-green", glow: "glass-glow-blue" },
  gold: { text: "text-system-gold", glow: "glass-glow-gold" },
};

export const StatsOverview: React.FC<StatsOverviewProps> = ({ summary, loading }) => {
  const stats: StatItem[] = [
    {
      key: "hunters",
      label: "Total Hunters Active",
      value: loading ? "..." : summary?.total_participants ?? 0,
      icon: Users,
      tone: "blue",
      description: "Hunters who reported this season",
    },
    {
      key: "streak",
      label: "Keep Streak 🔥",
      value: loading ? "..." : `${summary?.active_streak_count ?? 0} Players`,
      icon: Flame,
      tone: "red",
      pulse: true,
      description: "Active streak holders this week",
    },
    {
      key: "workouts",
      label: "Total Workouts Logged",
      value: loading ? "..." : summary?.total_workouts_logged ?? 0,
      icon: Activity,
      tone: "green",
      description: "Total workouts in lifetime history",
    },
    {
      key: "season",
      label: "System Current State",
      value: loading ? "..." : `Season ${summary?.current_season ?? 1} - Day ${summary?.current_day ?? 1}`,
      icon: Shield,
      tone: "gold",
      description: "Active campaign stage progress",
    },
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 mb-8">
      {stats.map((stat) => {
        const Icon = stat.icon;
        const tone = TONE_STYLES[stat.tone];
        return (
          <div
            key={stat.key}
            className={`glass p-6 rounded-2xl transition-all duration-300 hover:scale-[1.02] hover:-translate-y-1 ${tone.glow}`}
          >
            <div className="flex items-start justify-between">
              <div>
                <p className="text-gray-400 text-sm font-medium">{stat.label}</p>
                <h3 className="text-3xl font-bold font-orbitron tracking-wide mt-2 text-white">
                  {stat.value}
                </h3>
              </div>
              <div className={`p-3 rounded-xl bg-gray-900/60 ${tone.text} ${stat.pulse ? "animate-pulse" : ""}`}>
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
