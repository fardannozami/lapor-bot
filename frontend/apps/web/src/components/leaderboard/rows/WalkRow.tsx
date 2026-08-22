import React from "react";
import type { EnrichedReport } from "@lapor-bot/shared";
import { RankBadge } from "../RankBadge";
import { HunterCell } from "../HunterCell";
import { JobCell } from "../JobCell";

interface WalkRowProps {
  hunter: EnrichedReport;
  rank: number;
  rowClass: string;
  onSelectHunter: (hunter: EnrichedReport) => void;
}

function formatStepsID(steps: number): string {
  return new Intl.NumberFormat("id-ID").format(steps);
}

export const WalkRow: React.FC<WalkRowProps> = ({
  hunter,
  rank,
  rowClass,
  onSelectHunter,
}) => (
  <tr
    key={hunter.user_id}
    tabIndex={0}
    role="button"
    onClick={() => onSelectHunter(hunter)}
    onKeyDown={(e) => {
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        onSelectHunter(hunter);
      }
    }}
    className={rowClass}
  >
    <td className="py-4 pl-4 text-center font-mono align-middle">
      <div className="flex justify-center">
        <RankBadge rank={rank} />
      </div>
    </td>
    <HunterCell hunter={hunter} />
    <JobCell hunter={hunter} />
    <td className="py-4 pl-4 text-center font-mono text-sm text-gray-300 align-middle">
      {formatStepsID(hunter.week_walk_steps ?? 0)}
      <span className="text-[10px] text-gray-500"> langkah</span>
    </td>
    <td className="py-4 pl-4 text-center font-mono text-sm text-gray-300 align-middle">
      {formatStepsID(hunter.seasonal_walk_steps ?? 0)}
      <span className="text-[10px] text-gray-500"> langkah</span>
    </td>
    <td className="py-4 pl-4 text-right pr-4 font-mono font-bold text-system-gold align-middle">
      {formatStepsID(hunter.walk_steps_total ?? 0)}
      <span className="text-[10px] text-gray-500"> langkah</span>
    </td>
  </tr>
);
