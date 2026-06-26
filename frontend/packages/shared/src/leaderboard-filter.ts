import type { EnrichedReport } from "./types";
import { currentDailyStreak, longestDailyStreak } from "./leaderboard-sort";

/**
 * Filter helpers — mirror domain.HasSeasonActivity / HasAnyActivity /
 * HasStreakActivity / HasAttributeActivity on the backend.
 *
 * Keeping these as named pure functions (rather than methods on a class)
 * lets tree-shaking drop the ones a given app doesn't use.
 */

/** User has any seasonal engagement (points or activity). */
export function hasSeasonActivity(h: EnrichedReport): boolean {
  return h.seasonal_points > 0 || h.seasonal_activity_count > 0;
}

/** User has any lifetime engagement (points or active days). */
export function hasAnyActivity(h: EnrichedReport): boolean {
  return h.total_points > 0 || h.total_active_days > 0;
}

/** User has a meaningful streak (current or historical best). */
export function hasStreakActivity(h: EnrichedReport): boolean {
  return (
    h.streak > 0 ||
    h.max_streak > 0 ||
    currentDailyStreak(h) > 0 ||
    longestDailyStreak(h) > 0
  );
}

/** User has at least one attribute above the minimum baseline (1). */
export function hasAttributeActivity(h: EnrichedReport): boolean {
  return h.str > 1 || h.sta > 1 || h.agi > 1 || h.vit > 1;
}

/** User has logged at least one day this week. */
export function hasWeekActivity(h: EnrichedReport): boolean {
  return h.week_active_days > 0;
}

/**
 * Streak status derived from report data.
 * Mirrors the WA bot's active/comeback/inactive classification.
 */
export type StreakStatus = "active" | "comeback" | "inactive";

export function getStreakStatus(h: EnrichedReport): StreakStatus {
  if (h.days_since_last_report <= 7 && h.streak > 0) return "active";
  if (h.comeback_streak > 0 || h.days_since_last_report <= 14)
    return "comeback";
  return "inactive";
}

/**
 * Centurion prefix label — mirrors the WA bot's `[S1-C{n}]` cycle marker.
 * Returns empty string when no cycles exist.
 */
export function getCenturionLabel(h: EnrichedReport): string {
  if (h.centurion_cycles <= 0) return "";
  return `[S1-C${h.centurion_cycles + 1}]`;
}
