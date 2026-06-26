import React, { useCallback } from "react";
import {
  Search,
  SlidersHorizontal,
  Flame,
  Trophy,
  Award,
  CalendarDays,
  Swords,
} from "lucide-react";
import type {
  EnrichedReport,
  LeaderboardTab,
  AttributeTab,
  StreakTab,
} from "@lapor-bot/shared";

import { SeasonalRow } from "./leaderboard/rows/SeasonalRow";
import { LifetimeRow } from "./leaderboard/rows/LifetimeRow";
import { StreakRow } from "./leaderboard/rows/StreakRow";
import { WeekRow } from "./leaderboard/rows/WeekRow";
import { AttributeRow } from "./leaderboard/rows/AttributeRow";
import { useLeaderboardData } from "./leaderboard/useLeaderboardData";

interface LeaderboardTableProps {
  hunters: EnrichedReport[];
  onSelectHunter: (hunter: EnrichedReport) => void;
}

const JOB_FILTER_OPTIONS = [
  { id: "all", name: "All Jobs" },
  { id: "fighter", name: "Fighter ⚔️" },
  { id: "tank", name: "Tanker 🛡️" },
  { id: "assassin", name: "Assassin 🗡️" },
  { id: "mage", name: "Mage 🔥" },
  { id: "ranger", name: "Ranger 🏹" },
  { id: "healer", name: "Healer 💚" },
  { id: "necromancer", name: "Necromancer 🌑" },
] as const;

const LEADERBOARD_TABS: {
  id: LeaderboardTab;
  label: string;
  icon: typeof Trophy;
}[] = [
  { id: "seasonal", label: "Season Rank", icon: Trophy },
  { id: "lifetime", label: "Lifetime XP", icon: Award },
  { id: "streak", label: "Streak Masters", icon: Flame },
  { id: "week", label: "Minggu Ini", icon: CalendarDays },
  { id: "attributes", label: "Attributes", icon: Swords },
];

const ATTRIBUTE_SUB_TABS: { id: AttributeTab; label: string }[] = [
  { id: "overall", label: "Overall" },
  { id: "str", label: "STR" },
  { id: "sta", label: "STA" },
  { id: "agi", label: "AGI" },
  { id: "vit", label: "VIT" },
];

const STREAK_SUB_TABS: { id: StreakTab; label: string }[] = [
  { id: "weekly", label: "Mingguan" },
  { id: "daily", label: "Harian" },
];

const TABLE_HEADERS: Record<LeaderboardTab, string[]> = {
  seasonal: [
    "Rank",
    "Hunter",
    "Job",
    "Season Rank",
    "Streak",
    "Active Days",
    "Best Season Streak",
    "Freezes",
    "Season Points + Progress",
  ],
  lifetime: [
    "Rank",
    "Hunter",
    "Job",
    "Level Tier",
    "Level",
    "XP Progress",
    "Lifetime Days",
    "Best Streak",
    "Quests Done",
  ],
  streak: [
    "Rank",
    "Hunter",
    "Job",
    "Current Streak",
    "Best Streak",
    "Lifetime Days",
    "Comeback",
    "Centurion",
    "Status",
  ],
  week: [
    "Rank",
    "Hunter",
    "Job",
    "Active Days",
    "Week Dots",
    "Streak",
    "Est. Points",
  ],
  attributes: [
    "Rank",
    "Hunter",
    "Job",
    "Attribute",
    "STR",
    "STA",
    "AGI",
    "VIT",
    "Level",
  ],
};

const rowClass = (rank: number) =>
  `group transition-all hover:bg-gray-800/20 cursor-pointer ${rank < 3 ? "bg-gradient-to-r from-gray-950/20 to-transparent" : ""}`;

export const LeaderboardTable: React.FC<LeaderboardTableProps> = ({
  hunters,
  onSelectHunter,
}) => {
  const {
    search,
    setSearch,
    selectedJob,
    setSelectedJob,
    activeTab,
    setActiveTab,
    attributeTab,
    setAttributeTab,
    streakTab,
    setStreakTab,
    safePage,
    totalPages,
    pageStartRank,
    totalCount,
    visibleHunters,
    goToPage,
  } = useLeaderboardData(hunters);

  const handleTabChange = useCallback(
    (tab: LeaderboardTab) => {
      setActiveTab(tab);
      setSearch("");
      setSelectedJob("all");
      goToPage(1);
    },
    [setActiveTab, setSearch, setSelectedJob, goToPage],
  );

  const handleAttributeSubTabChange = useCallback(
    (sub: AttributeTab) => {
      setAttributeTab(sub);
      goToPage(1);
    },
    [setAttributeTab, goToPage],
  );

  const handleStreakSubTabChange = useCallback(
    (sub: StreakTab) => {
      setStreakTab(sub);
      goToPage(1);
    },
    [setStreakTab, goToPage],
  );

  const renderRows = () =>
    visibleHunters.map((hunter, idx) => {
      const rank = pageStartRank + idx;
      const rc = rowClass(rank);
      switch (activeTab) {
        case "seasonal":
          return (
            <SeasonalRow
              key={hunter.user_id}
              hunter={hunter}
              rank={rank}
              rowClass={rc}
              onSelectHunter={onSelectHunter}
            />
          );
        case "lifetime":
          return (
            <LifetimeRow
              key={hunter.user_id}
              hunter={hunter}
              rank={rank}
              rowClass={rc}
              onSelectHunter={onSelectHunter}
            />
          );
        case "streak":
          return (
            <StreakRow
              key={hunter.user_id}
              hunter={hunter}
              rank={rank}
              rowClass={rc}
              streakTab={streakTab}
              onSelectHunter={onSelectHunter}
            />
          );
        case "week":
          return (
            <WeekRow
              key={hunter.user_id}
              hunter={hunter}
              rank={rank}
              rowClass={rc}
              onSelectHunter={onSelectHunter}
            />
          );
        case "attributes":
          return (
            <AttributeRow
              key={hunter.user_id}
              hunter={hunter}
              rank={rank}
              rowClass={rc}
              attributeTab={attributeTab}
              onSelectHunter={onSelectHunter}
            />
          );
        default:
          return null;
      }
    });

  const tableHeaders = TABLE_HEADERS[activeTab];

  return (
    <div className="glass rounded-3xl p-5 md:p-6 mb-8">
      <div className="flex flex-col lg:flex-row gap-4 items-stretch lg:items-center justify-between mb-6 border-b border-gray-850 pb-5">
        <div className="flex flex-wrap bg-gray-950/80 p-1.5 rounded-xl border border-gray-800/60 max-w-fit">
          {LEADERBOARD_TABS.map((tab) => {
            const Icon = tab.icon;
            const active = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => handleTabChange(tab.id)}
                className={`flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-bold font-orbitron tracking-wider transition-all uppercase ${
                  active
                    ? "bg-gradient-to-r from-system-blue to-system-purple text-white shadow-neon-purple"
                    : "text-gray-400 hover:text-white hover:bg-gray-800/60"
                }`}
              >
                <Icon size={14} />
                {tab.label}
              </button>
            );
          })}
        </div>

        <div className="flex gap-2 items-center">
          <div className="relative">
            <Search
              size={14}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500"
            />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Cari hunter…"
              className="pl-9 pr-3 py-2 rounded-xl bg-gray-950 border border-gray-800 text-xs font-mono text-gray-200 placeholder:text-gray-600 focus:outline-none focus:border-system-blue/50 transition-colors w-40"
            />
          </div>
          <div className="relative">
            <SlidersHorizontal
              size={14}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500"
            />
            <select
              value={selectedJob}
              onChange={(e) => setSelectedJob(e.target.value)}
              className="pl-9 pr-3 py-2 rounded-xl bg-gray-950 border border-gray-800 text-xs font-mono text-gray-200 focus:outline-none focus:border-system-blue/50 transition-colors appearance-none cursor-pointer"
            >
              {JOB_FILTER_OPTIONS.map((opt) => (
                <option key={opt.id} value={opt.id}>
                  {opt.name}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {activeTab === "attributes" && (
        <div className="flex flex-wrap gap-1.5 mb-4 p-1.5 bg-gray-950/60 rounded-xl border border-gray-800/40 max-w-fit">
          {ATTRIBUTE_SUB_TABS.map((sub) => {
            const active = attributeTab === sub.id;
            return (
              <button
                key={sub.id}
                onClick={() => handleAttributeSubTabChange(sub.id)}
                className={`px-3 py-1.5 rounded-lg text-[10px] font-bold font-mono tracking-wider transition-all uppercase ${
                  active
                    ? "bg-gradient-to-r from-system-red to-system-gold text-white"
                    : "text-gray-400 hover:text-white hover:bg-gray-800/60"
                }`}
              >
                {sub.label}
              </button>
            );
          })}
        </div>
      )}

      {activeTab === "streak" && (
        <div className="flex flex-wrap gap-1.5 mb-4 p-1.5 bg-gray-950/60 rounded-xl border border-gray-800/40 max-w-fit">
          {STREAK_SUB_TABS.map((sub) => {
            const active = streakTab === sub.id;
            return (
              <button
                key={sub.id}
                onClick={() => handleStreakSubTabChange(sub.id)}
                className={`px-3 py-1.5 rounded-lg text-[10px] font-bold font-mono tracking-wider transition-all uppercase ${
                  active
                    ? "bg-gradient-to-r from-system-gold to-system-red text-white"
                    : "text-gray-400 hover:text-white hover:bg-gray-800/60"
                }`}
              >
                {sub.label}
              </button>
            );
          })}
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="w-full">
          <thead>
            <tr className="border-b border-gray-800/80">
              {tableHeaders.map((header, i) => (
                <th
                  key={`${activeTab}-${i}`}
                  className={`py-3 px-4 text-[10px] font-mono font-bold uppercase tracking-wider text-gray-500 ${
                    i === 0
                      ? "text-center"
                      : i === 1
                        ? "text-left"
                        : "text-center"
                  }`}
                >
                  {header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>{renderRows()}</tbody>
        </table>
      </div>

      {totalCount === 0 && (
        <div className="text-center py-16 text-gray-600 font-mono text-sm">
          {search || selectedJob !== "all"
            ? "Tidak ada hunter yang cocok dengan filter ini."
            : "Belum ada hunter aktif untuk kategori ini."}
        </div>
      )}

      {totalPages > 1 && (
        <nav className="flex items-center justify-between mt-6 pt-4 border-t border-gray-800/60">
          <p className="text-xs font-mono text-gray-500">
            Menampilkan {pageStartRank + 1}–
            {Math.min(pageStartRank + visibleHunters.length, totalCount)} dari{" "}
            {totalCount} hunters
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
      )}
    </div>
  );
};
