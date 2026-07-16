/**
 * Attribute display helpers for the six RPG attributes: STR, STA, AGI, VIT, END, LUK.
 *
 * These mirror the backend `domain.AttributeType` / `domain.ClampedAttribute`
 * logic so the frontend and backend always agree on baseline and average.
 */

import type { EnrichedReport } from "./types";

/** Attribute identifiers — mirrors backend `domain.AttributeType`. */
export type AttributeType = "STR" | "STA" | "AGI" | "VIT" | "END" | "LUK";

/** "Overall" pseudo-attribute used for the combined attribute leaderboard. */
export const ATTRIBUTE_OVERALL = "OVERALL" as const;

export type AttributeSortKey = AttributeType | typeof ATTRIBUTE_OVERALL;

export const ATTRIBUTE_LIST: readonly AttributeType[] = [
  "STR",
  "STA",
  "AGI",
  "VIT",
  "END",
  "LUK",
] as const;

/** Display metadata for each attribute — single source of truth for labels & icons. */
export const ATTRIBUTE_META: Record<
  AttributeType,
  { label: string; icon: string; color: string; description: string }
> = {
  STR: {
    label: "Strength",
    icon: "💪",
    color: "text-system-red",
    description: "Power & raw force",
  },
  STA: {
    label: "Stamina",
    icon: "🛡️",
    color: "text-system-gold",
    description: "Cardio & low-intensity aerobic",
  },
  AGI: {
    label: "Agility",
    icon: "⚡",
    color: "text-system-green",
    description: "Speed & reflexes",
  },
  VIT: {
    label: "Vitality",
    icon: "❤️",
    color: "text-system-purple",
    description: "Health & life force",
  },
  END: {
    label: "Endurance",
    icon: "🏔️",
    color: "text-system-gold",
    description: "Sustained effort & long-duration",
  },
  LUK: {
    label: "Luck",
    icon: "🌑",
    color: "text-gray-400",
    description: "Resilience & comeback",
  },
};

/**
 * Returns the raw attribute value for the given type from an EnrichedReport.
 * Backend already clamps to MinAttributeValue (1) before serialization,
 * so no additional clamping is needed here.
 */
export function getAttributeValue(
  hunter: EnrichedReport,
  attr: AttributeType,
): number {
  switch (attr) {
    case "STR":
      return hunter.str;
    case "STA":
      return hunter.sta;
    case "AGI":
      return hunter.agi;
    case "VIT":
      return hunter.vit;
    case "END":
      return hunter.end;
    case "LUK":
      return hunter.luk;
  }
}

/**
 * Computes the rounded average of all six RPG attributes.
 * Mirrors backend `Report.AttributeAverage()`.
 */
export function getAttributeAverage(hunter: EnrichedReport): number {
  return Math.floor(
    (hunter.str +
      hunter.sta +
      hunter.agi +
      hunter.vit +
      hunter.end +
      hunter.luk) /
      6,
  );
}

/**
 * Returns the attribute value (or average for OVERALL) for sort/compare logic.
 * Mirrors backend `Report.AttributeValue(attr)`.
 */
export function getAttributeScore(
  hunter: EnrichedReport,
  key: AttributeSortKey,
): number {
  if (key === ATTRIBUTE_OVERALL) {
    return getAttributeAverage(hunter);
  }
  return getAttributeValue(hunter, key);
}

/**
 * Returns true if the hunter has at least one attribute point.
 * Mirrors backend `HasAttributeActivity()`.
 */
export function hasAttributeActivity(hunter: EnrichedReport): boolean {
  return (
    hunter.str > 0 ||
    hunter.sta > 0 ||
    hunter.agi > 0 ||
    hunter.vit > 0 ||
    hunter.end > 0 ||
    hunter.luk > 0
  );
}
