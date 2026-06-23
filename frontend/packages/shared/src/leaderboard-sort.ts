import type { EnrichedReport, LeaderboardSortKey, AttributeTab } from "./types";

export const LEADERBOARD_SORT_KEYS: LeaderboardSortKey[] = [
  "season_rank",
  "lifetime_xp",
  "weekly_streak",
  "daily_streak",
  "weekly_activity",
  "attribute_overall",
  "attribute_str",
  "attribute_sta",
  "attribute_agi",
  "attribute_vit",
];

export const ATTRIBUTE_SORT_KEYS: LeaderboardSortKey[] = [
  "attribute_overall",
  "attribute_str",
  "attribute_sta",
  "attribute_agi",
  "attribute_vit",
];

export const ATTRIBUTE_TAB_TO_SORT_KEY: Record<
  AttributeTab,
  LeaderboardSortKey
> = {
  overall: "attribute_overall",
  str: "attribute_str",
  sta: "attribute_sta",
  agi: "attribute_agi",
  vit: "attribute_vit",
};

export function getActiveAttributeTab(key: LeaderboardSortKey): AttributeTab {
  switch (key) {
    case "attribute_str":
      return "str";
    case "attribute_sta":
      return "sta";
    case "attribute_agi":
      return "agi";
    case "attribute_vit":
      return "vit";
    default:
      return "overall";
  }
}

export function totalActiveDays(r: EnrichedReport): number {
  return r.total_active_days ?? r.centurion_cycles * 100 + r.activity_count;
}

export function currentDailyStreak(r: EnrichedReport): number {
  return r.current_daily_streak ?? 0;
}

export function longestDailyStreak(r: EnrichedReport): number {
  return r.longest_daily_streak ?? currentDailyStreak(r);
}

function clampAttr(v: number): number {
  return Math.max(1, v);
}

export function attributeAverage(r: EnrichedReport): number {
  const total =
    clampAttr(r.str) + clampAttr(r.sta) + clampAttr(r.agi) + clampAttr(r.vit);
  return Math.floor(total / 4);
}

export function attributeValueByKey(
  r: EnrichedReport,
  key: LeaderboardSortKey,
): number {
  switch (key) {
    case "attribute_str":
      return clampAttr(r.str);
    case "attribute_sta":
      return clampAttr(r.sta);
    case "attribute_agi":
      return clampAttr(r.agi);
    case "attribute_vit":
      return clampAttr(r.vit);
    default:
      return attributeAverage(r);
  }
}

export function hasSeasonActivity(r: EnrichedReport): boolean {
  return r.seasonal_points > 0 || r.seasonal_activity_count > 0;
}

export function hasAnyActivity(r: EnrichedReport): boolean {
  return r.total_points > 0 || r.activity_count > 0 || r.centurion_cycles > 0;
}

export function hasStreakActivity(r: EnrichedReport): boolean {
  return (
    r.streak > 0 ||
    r.max_streak > 0 ||
    r.seasonal_max_streak > 0 ||
    currentDailyStreak(r) > 0 ||
    longestDailyStreak(r) > 0 ||
    r.activity_count > 0
  );
}

export function hasAttributeActivity(r: EnrichedReport): boolean {
  return r.str > 0 || r.sta > 0 || r.agi > 0 || r.vit > 0;
}

export function compareReports(
  a: EnrichedReport,
  b: EnrichedReport,
  key: LeaderboardSortKey,
): boolean {
  switch (key) {
    case "season_rank":
      return compareSeasonRank(a, b);
    case "lifetime_xp":
      return compareLifetimeXP(a, b);
    case "weekly_streak":
      return compareWeeklyStreak(a, b);
    case "daily_streak":
      return compareDailyStreak(a, b);
    case "weekly_activity":
      return compareWeeklyActivity(a, b);
    case "attribute_overall":
      return compareAttributeOverall(a, b);
    case "attribute_str":
    case "attribute_sta":
    case "attribute_agi":
    case "attribute_vit":
      return compareAttribute(a, b, key);
    default:
      return compareSeasonRank(a, b);
  }
}

function compareSeasonRank(a: EnrichedReport, b: EnrichedReport): boolean {
  if (a.seasonal_points !== b.seasonal_points)
    return a.seasonal_points > b.seasonal_points;
  if (a.seasonal_activity_count !== b.seasonal_activity_count)
    return a.seasonal_activity_count > b.seasonal_activity_count;
  if (a.streak !== b.streak) return a.streak > b.streak;
  if (totalActiveDays(a) !== totalActiveDays(b))
    return totalActiveDays(a) > totalActiveDays(b);
  return compareNameThenUserID(a, b);
}

function compareLifetimeXP(a: EnrichedReport, b: EnrichedReport): boolean {
  if (a.total_points !== b.total_points) return a.total_points > b.total_points;
  if (totalActiveDays(a) !== totalActiveDays(b))
    return totalActiveDays(a) > totalActiveDays(b);
  if (a.max_streak !== b.max_streak) return a.max_streak > b.max_streak;
  return compareNameThenUserID(a, b);
}

function compareWeeklyStreak(a: EnrichedReport, b: EnrichedReport): boolean {
  if (a.streak !== b.streak) return a.streak > b.streak;

  const aDaily = currentDailyStreak(a);
  const bDaily = currentDailyStreak(b);
  if (aDaily !== bDaily) return aDaily > bDaily;

  if (a.seasonal_points !== b.seasonal_points)
    return a.seasonal_points > b.seasonal_points;
  if (a.max_streak !== b.max_streak) return a.max_streak > b.max_streak;
  return compareNameThenUserID(a, b);
}

function compareDailyStreak(a: EnrichedReport, b: EnrichedReport): boolean {
  const aDaily = currentDailyStreak(a);
  const bDaily = currentDailyStreak(b);
  if (aDaily !== bDaily) return aDaily > bDaily;
  const aLongest = longestDailyStreak(a);
  const bLongest = longestDailyStreak(b);
  if (aLongest !== bLongest) return aLongest > bLongest;
  if (a.total_points !== b.total_points) return a.total_points > b.total_points;
  return compareNameThenUserID(a, b);
}

function compareWeeklyActivity(a: EnrichedReport, b: EnrichedReport): boolean {
  if (a.week_active_days !== b.week_active_days)
    return a.week_active_days > b.week_active_days;
  if (a.streak !== b.streak) return a.streak > b.streak;
  if (a.seasonal_points !== b.seasonal_points)
    return a.seasonal_points > b.seasonal_points;
  return compareNameThenUserID(a, b);
}

function compareAttributeOverall(
  a: EnrichedReport,
  b: EnrichedReport,
): boolean {
  const aAvg = attributeAverage(a);
  const bAvg = attributeAverage(b);
  if (aAvg !== bAvg) return aAvg > bAvg;
  if (a.total_points !== b.total_points) return a.total_points > b.total_points;
  return compareNameThenUserID(a, b);
}

function compareAttribute(
  a: EnrichedReport,
  b: EnrichedReport,
  key: LeaderboardSortKey,
): boolean {
  const aVal = attributeValueByKey(a, key);
  const bVal = attributeValueByKey(b, key);
  if (aVal !== bVal) return aVal > bVal;
  if (a.total_points !== b.total_points) return a.total_points > b.total_points;
  return compareNameThenUserID(a, b);
}

function compareNameThenUserID(a: EnrichedReport, b: EnrichedReport): boolean {
  if (a.name !== b.name) return a.name < b.name;
  return a.user_id < b.user_id;
}

export function sortLeaderboard(
  reports: readonly EnrichedReport[],
  key: LeaderboardSortKey,
): EnrichedReport[] {
  return [...reports].sort((a, b) => {
    if (compareReports(a, b, key)) return -1;
    if (compareReports(b, a, key)) return 1;
    return 0;
  });
}

export interface FilterOptions {
  search?: string;
  jobClass?: string;
}

export function filterLeaderboard(
  reports: readonly EnrichedReport[],
  keep: (r: EnrichedReport) => boolean,
): EnrichedReport[] {
  return reports.filter(keep);
}

export function filterLeaderboardByTab(
  reports: readonly EnrichedReport[],
  sortKey: LeaderboardSortKey,
  options: FilterOptions = {},
): EnrichedReport[] {
  let result = [...reports];

  if (options.search?.trim()) {
    const q = options.search.toLowerCase();
    result = result.filter((r) => r.name.toLowerCase().includes(q));
  }

  if (options.jobClass && options.jobClass !== "all") {
    result = result.filter(
      (r) => r.job_class?.toLowerCase() === options.jobClass!.toLowerCase(),
    );
  }

  const keepFn = filterPredicateForSortKey(sortKey);
  return result.filter(keepFn);
}

function filterPredicateForSortKey(
  key: LeaderboardSortKey,
): (r: EnrichedReport) => boolean {
  switch (key) {
    case "season_rank":
      return hasSeasonActivity;
    case "weekly_activity":
      return (r) => r.week_active_days > 0;
    case "lifetime_xp":
      return hasAnyActivity;
    case "weekly_streak":
    case "daily_streak":
      return hasStreakActivity;
    case "attribute_overall":
    case "attribute_str":
    case "attribute_sta":
    case "attribute_agi":
    case "attribute_vit":
      return hasAttributeActivity;
    default:
      return hasSeasonActivity;
  }
}

export function getStreakStatus(hunter: EnrichedReport): string {
  if (hunter.days_since_last_report <= 7 && hunter.streak > 0)
    return "🔥 Active";
  if (hunter.comeback_streak > 0 || hunter.days_since_last_report <= 14)
    return "🗡️ Comeback";
  return "💔 Inactive";
}
