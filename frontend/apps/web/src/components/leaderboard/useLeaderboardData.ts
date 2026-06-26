import { useMemo, useState } from "react";
import type {
  EnrichedReport,
  LeaderboardTab,
  AttributeTab,
  StreakTab,
  LeaderboardSortKey,
} from "@lapor-bot/shared";
import {
  sortLeaderboard,
  hasSeasonActivity,
  hasAnyActivity,
  hasStreakActivity,
  hasAttributeActivity,
  currentDailyStreak,
  ATTRIBUTE_TAB_TO_SORT_KEY,
} from "@lapor-bot/shared";

const PAGE_SIZE = 15;

const TAB_SORT_KEY: Record<LeaderboardTab, LeaderboardSortKey> = {
  seasonal: "season_rank",
  lifetime: "lifetime_xp",
  streak: "weekly_streak",
  week: "weekly_activity",
  attributes: "attribute_overall",
};

const TAB_FILTER: Record<LeaderboardTab, (r: EnrichedReport) => boolean> = {
  seasonal: hasSeasonActivity,
  lifetime: hasAnyActivity,
  streak: hasStreakActivity,
  week: (r) => r.week_active_days > 0,
  attributes: hasAttributeActivity,
};

export interface LeaderboardData {
  search: string;
  setSearch: (v: string) => void;
  selectedJob: string;
  setSelectedJob: (v: string) => void;
  activeTab: LeaderboardTab;
  setActiveTab: (tab: LeaderboardTab) => void;
  attributeTab: AttributeTab;
  setAttributeTab: (tab: AttributeTab) => void;
  streakTab: StreakTab;
  setStreakTab: (tab: StreakTab) => void;
  safePage: number;
  totalPages: number;
  pageStartRank: number;
  totalCount: number;
  visibleHunters: EnrichedReport[];
  goToPage: (page: number) => void;
}

export function useLeaderboardData(hunters: EnrichedReport[]): LeaderboardData {
  const [search, setSearch] = useState("");
  const [selectedJob, setSelectedJob] = useState("all");
  const [activeTab, setActiveTab] = useState<LeaderboardTab>("seasonal");
  const [attributeTab, setAttributeTab] = useState<AttributeTab>("overall");
  const [streakTab, setStreakTab] = useState<StreakTab>("weekly");
  const [page, setPage] = useState(1);

  const filteredAndSorted = useMemo(() => {
    const q = search.trim().toLowerCase();
    const jobFilter = selectedJob.toLowerCase();

    const matchesSearch = (h: EnrichedReport) =>
      !q || h.name.toLowerCase().includes(q);

    const matchesJob = (h: EnrichedReport) =>
      jobFilter === "all" || h.job_class?.toLowerCase() === jobFilter;

    const sortKey =
      activeTab === "attributes"
        ? ATTRIBUTE_TAB_TO_SORT_KEY[attributeTab]
        : activeTab === "streak" && streakTab === "daily"
          ? "daily_streak"
          : TAB_SORT_KEY[activeTab];

    const passesTabFilter =
      activeTab === "streak" && streakTab === "daily"
        ? (h: EnrichedReport) => currentDailyStreak(h) > 0
        : TAB_FILTER[activeTab];
    const filtered = hunters.filter(
      (h) => passesTabFilter(h) && matchesSearch(h) && matchesJob(h),
    );

    return sortLeaderboard(filtered, sortKey);
  }, [hunters, activeTab, attributeTab, streakTab, search, selectedJob]);

  const totalPages = Math.max(
    1,
    Math.ceil(filteredAndSorted.length / PAGE_SIZE),
  );
  const safePage = Math.min(page, totalPages);

  const visibleHunters = useMemo(() => {
    const start = (safePage - 1) * PAGE_SIZE;
    return filteredAndSorted.slice(start, start + PAGE_SIZE);
  }, [filteredAndSorted, safePage]);

  const goToPage = (next: number) =>
    setPage(Math.min(Math.max(1, next), totalPages));

  return {
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
    totalCount: filteredAndSorted.length,
    pageStartRank: (safePage - 1) * PAGE_SIZE,
    visibleHunters,
    goToPage,
  };
}
