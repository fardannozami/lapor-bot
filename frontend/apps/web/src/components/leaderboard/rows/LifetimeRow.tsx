import React from "react";
import type { EnrichedReport } from "@lapor-bot/shared";
import { RankBadge } from "../RankBadge";
import { HunterCell } from "../HunterCell";
import { JobCell } from "../JobCell";
import { ProgressBar } from "../ProgressBar";
import { StreakCell } from "../StreakCell";
import { QuestBadge } from "../QuestBadge";

interface LifetimeRowProps {
  hunter: EnrichedReport;
  rank: number;
  rowClass: string;
  onSelectHunter: (hunter: EnrichedReport) => void;
}

export const LifetimeRow: React.FC<LifetimeRowProps> = ({
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
      <span className="font-bold">
        {hunter.level_icon} {hunter.level_name}
      </span>
    </td>
    <td className="py-4 pl-4 text-center font-mono align-middle text-sm text-gray-300">
      <span className="font-bold">Lv.{hunter.level}</span>
    </td>
    <td className="py-4 pl-4 align-middle">
      <ProgressBar
        progress={hunter.level_tier_progress}
        valueLabel={`${hunter.total_points} lifetime XP`}
        tone="blue"
      />
    </td>
    <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
      <span className="font-bold">{hunter.total_active_days}</span>
      <span className="text-[10px] text-gray-500"> hari</span>
    </td>
    <td className="py-4 pl-4 text-center align-middle text-sm">
      <StreakCell weeks={hunter.max_streak} />
    </td>
    <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300 pr-4">
      <QuestBadge
        completed={hunter.goals_completed}
        sideQuests={hunter.total_side_quests}
      />
    </td>
  </tr>
);
