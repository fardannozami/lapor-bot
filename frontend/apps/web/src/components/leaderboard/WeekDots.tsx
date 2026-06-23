import React from "react";
import type { EnrichedReport } from "@lapor-bot/shared";

interface WeekDotsProps {
  hunter: EnrichedReport;
}

export const WeekDots: React.FC<WeekDotsProps> = ({ hunter }) => (
  <div
    className="flex justify-center gap-1"
    aria-label={`${hunter.week_active_days} dari 7 hari aktif minggu ini`}
  >
    {hunter.week_activity.map((active, idx) => (
      <span
        key={`${hunter.user_id}-week-${idx}`}
        className={`h-3 w-3 rounded-sm border ${
          active
            ? "bg-system-green border-system-green/60 shadow-neon-purple"
            : "bg-gray-900 border-gray-800"
        }`}
        title={active ? "Aktif" : "Belum aktif"}
      />
    ))}
  </div>
);
