import React from "react";
import { Users, Flame, Activity, Shield } from "lucide-react";
import type { GlobalSummary } from "@lapor-bot/shared";

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

const TONE_STYLES: Record<
  StatTone,
  { text: string; glow: string; edge: string; chip: string }
> = {
  blue: {
    text: "text-system-blue",
    glow: "glass-glow-blue",
    edge: "var(--color-system-blue)",
    chip: "shadow-neon-blue",
  },
  red: {
    text: "text-system-red",
    glow: "glass-glow-red",
    edge: "var(--color-system-red)",
    chip: "shadow-neon-red",
  },
  green: {
    text: "text-system-green",
    glow: "glass-glow-green",
    edge: "var(--color-system-green)",
    chip: "shadow-neon-green",
  },
  gold: {
    text: "text-system-gold",
    glow: "glass-glow-gold",
    edge: "var(--color-system-gold)",
    chip: "shadow-neon-gold",
  },
};

export const StatsOverview: React.FC<StatsOverviewProps> = ({
  summary,
  loading,
}) => {
  const stats: StatItem[] = [
    {
      key: "hunters",
      label: "Total Hunters Active",
      value: loading ? "..." : (summary?.total_participants ?? 0),
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
      value: loading ? "..." : (summary?.total_workouts_logged ?? 0),
      icon: Activity,
      tone: "green",
      description: "Total workouts in lifetime history",
    },
    {
      key: "season",
      label: "System Current State",
      value: loading
        ? "..."
        : `Season ${summary?.current_season ?? 1} - Day ${summary?.current_day ?? 1}`,
      icon: Shield,
      tone: "gold",
      description: "Active campaign stage progress",
    },
  ];

  return (
    <div className="stagger-children grid grid-cols-2 lg:grid-cols-4 gap-3 md:gap-5 mb-8">
      {stats.map((stat, i) => {
        const Icon = stat.icon;
        const tone = TONE_STYLES[stat.tone];
        return (
          <div
            key={stat.key}
            style={
              {
                "--stagger-i": i,
                "--edge": tone.edge,
              } as React.CSSProperties
            }
            className={`glass stat-edge card-sheen animate-fade-in-up p-4 md:p-6 rounded-2xl transition-all duration-300 hover:scale-[1.02] hover:-translate-y-1 ${tone.glow}`}
          >
            <div className="flex items-center justify-between gap-2">
              <p className="text-gray-400 text-xs md:text-sm font-medium leading-tight">
                {stat.label}
              </p>
              <div
                className={`shrink-0 p-2 md:p-3 rounded-xl bg-gray-900/60 ${tone.text} ${tone.chip} ${stat.pulse ? "animate-pulse" : ""}`}
              >
                <Icon size={24} className="w-5 h-5 md:w-6 md:h-6" />
              </div>
            </div>
            <h3 className="text-2xl md:text-4xl font-bold font-orbitron tracking-wide mt-1.5 md:mt-2 text-white break-words tnum">
              {stat.value}
            </h3>
            <p className="text-[10px] md:text-xs text-gray-500 mt-3 md:mt-4 font-mono">
              {stat.description}
            </p>
          </div>
        );
      })}
    </div>
  );
};
