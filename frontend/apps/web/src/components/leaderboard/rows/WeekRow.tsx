import React from "react";
import type { EnrichedReport } from "@lapor-bot/shared";
import { RankBadge } from "../RankBadge";
import { HunterCell } from "../HunterCell";
import { JobCell } from "../JobCell";
import { StreakCell } from "../StreakCell";
import { WeekDots } from "../WeekDots";

interface WeekRowProps {
  hunter: EnrichedReport;
  rank: number;
  rowClass: string;
  onSelectHunter: (hunter: EnrichedReport) => void;
}

export const WeekRow: React.FC<WeekRowProps> = ({
  hunter,
  rank,
  rowClass,
  onSelectHunter,
}) => (
  <tr key={hunter.user_id} onClick={() => onSelectHunter(hunter)} className={rowClass}>
    <td className="py-4 pl-4 text-center font-mono align-middle">
      <div className="flex justify-center">
        <RankBadge rank={rank} />
      </div>
    </td>
    <HunterCell hunter={hunter} />
    <JobCell hunter={hunter} />
    <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
      <span className="font-bold">{hunter.week_active_days}</span>
      <span className="text-[10px] text-gray-500"> / 7 hari minggu ini</span>
    </td>
    <td className="py-4 pl-4 text-center align-middle">
      <WeekDots hunter={hunter} />
    </td>
    <td className="py-4 pl-4 text-center align-middle text-sm">
      <StreakCell weeks={hunter.streak} />
    </td>
    <td className="py-4 pl-4 text-right pr-4 font-mono font-bold align-middle text-system-gold">
      {hunter.estimated_weekly_points}{" "}
      <span className="text-[9px] text-gray-500">pts (est.)</span>
    </td>
  </tr>
);
