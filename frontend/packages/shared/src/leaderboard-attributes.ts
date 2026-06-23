import type { EnrichedReport, AttributeTab, LeaderboardSortKey } from "./types";

export interface AttributeMeta {
  key: "str" | "sta" | "agi" | "vit";
  label: string;
  short: string;
  icon: string;
  color: string;
  sortKey: LeaderboardSortKey;
}

export const ATTRIBUTE_META: Record<Exclude<AttributeTab, "overall">, AttributeMeta> = {
  str: {
    key: "str",
    label: "Strength",
    short: "STR",
    icon: "⚔️",
    color: "text-system-red",
    sortKey: "attribute_str",
  },
  sta: {
    key: "sta",
    label: "Stamina",
    short: "STA",
    icon: "🛡️",
    color: "text-system-gold",
    sortKey: "attribute_sta",
  },
  agi: {
    key: "agi",
    label: "Agility",
    short: "AGI",
    icon: "🗡️",
    color: "text-system-blue",
    sortKey: "attribute_agi",
  },
  vit: {
    key: "vit",
    label: "Vitality",
    short: "VIT",
    icon: "💚",
    color: "text-system-green",
    sortKey: "attribute_vit",
  },
};

export const ATTRIBUTE_LIST = Object.values(ATTRIBUTE_META);

const ATTRIBUTE_BAR_MAX = 12;

export function getAttributeAverage(r: EnrichedReport): number {
  const total =
    Math.max(1, r.str) + Math.max(1, r.sta) +
    Math.max(1, r.agi) + Math.max(1, r.vit);
  return Math.floor(total / 4);
}

export function getAttributeValue(
  r: EnrichedReport,
  tab: AttributeTab,
): number {
  if (tab === "overall") return getAttributeAverage(r);
  return Math.max(1, r[tab]);
}

export function formatAttributeValue(value: number): string {
  return String(value);
}

export function getAttributeBarPercent(value: number): number {
  return Math.min(100, Math.max(8, value * (100 / ATTRIBUTE_BAR_MAX)));
}
