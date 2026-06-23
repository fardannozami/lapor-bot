import React from "react";
import type { EnrichedReport } from "@lapor-bot/shared";
import { RankBadge } from "../RankBadge";
import { HunterCell } from "../HunterCell";
import { JobCell } from "../JobCell";
import { StreakCell } from "../StreakCell";
import { getStreakStatus } from "@lapor-bot/shared";

interface StreakRowProps {
  hunter: EnrichedReport;
  rank: number;
  rowClass: string;
  onSelectHunter: (hunter: EnrichedReport) => void;
}

export const StreakRow: React.FC<StreakRowProps> = ({
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
    <td className="py-4 pl-4 text-center align-middle text-sm">
      <StreakCell weeks={hunter.streak} />
    </td>
    <td className="py-4 pl-4 text-center align-middle text-sm">
      <StreakCell weeks={hunter.max_streak} />
    </td>
    <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
      <span className="font-bold">{hunter.total_active_days}</span>
      <span className="text-[10px] text-gray-500"> hari lifetime</span>
    </td>
    <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
      <span className="font-bold text-system-purple">{hunter.comeback_streak}</span>
      <span className="text-[10px] text-gray-500"> comeback</span>
    </td>
    <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
      {hunter.centurion_cycles > 0 && (
        <span className="font-bold text-system-gold">
          {hunter.centurion_cycles}x 💯
        </span>
      )}
      {hunter.centurion_cycles === 0 && (
        <span className="text-gray-600">—</span>
      )}
    </td>
    <td className="py-4 pl-4 pr-4 text-center align-middle font-mono text-sm">
      {getStreakStatus(hunter)}
    </td>
  </tr>
);
