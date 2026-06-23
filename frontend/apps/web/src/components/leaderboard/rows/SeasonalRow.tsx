import React from "react";
import type { EnrichedReport } from "@lapor-bot/shared";
import { RankBadge } from "../RankBadge";
import { HunterCell } from "../HunterCell";
import { JobCell } from "../JobCell";
import { StreakCell } from "../StreakCell";
import { ActiveTodayBadge } from "../ActiveTodayBadge";
import { ProgressBar } from "../ProgressBar";
import { Snowflake } from "lucide-react";

interface SeasonalRowProps {
  hunter: EnrichedReport;
  rank: number;
  rowClass: string;
  onSelectHunter: (hunter: EnrichedReport) => void;
}

export const SeasonalRow: React.FC<SeasonalRowProps> = ({
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
    <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-white">
      <div className="flex flex-col items-center gap-1">
        <span className="font-bold">
          {hunter.rank_icon} {hunter.rank_name}
        </span>
        {hunter.is_active_today && <ActiveTodayBadge />}
      </div>
    </td>
    <td className="py-4 pl-4 text-center align-middle text-sm">
      <StreakCell weeks={hunter.streak} />
    </td>
    <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
      <span className="font-bold">{hunter.seasonal_activity_count}</span>
      <span className="text-[10px] text-gray-500"> hari / season</span>
    </td>
    <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-400">
      <div className="flex flex-col items-center gap-1">
        <span className="font-bold text-gray-300">{hunter.seasonal_max_streak}</span>
        <span className="text-[10px] text-gray-500">best streak</span>
      </div>
    </td>
    <td className="py-4 pl-4 text-center align-middle">
      {hunter.streak_freezes > 0 && (
        <span
          className="inline-flex items-center gap-1 text-xs font-mono text-system-blue"
          title={`${hunter.streak_freezes} streak freeze tersisa`}
        >
          <Snowflake size={12} />
          {hunter.streak_freezes}
        </span>
      )}
    </td>
    <td className="py-4 pl-4 pr-4 align-middle">
      <ProgressBar
        progress={hunter.season_rank_progress}
        valueLabel={`${hunter.seasonal_points} season pts`}
        tone="gold"
      />
    </td>
  </tr>
);
