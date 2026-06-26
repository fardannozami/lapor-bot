import React from "react";
import type { EnrichedReport } from "@lapor-bot/shared";
import { getJobBadgeClass } from "@lapor-bot/shared";

interface HunterCellProps {
  hunter: EnrichedReport;
  isActiveToday?: boolean;
}

export const HunterCell: React.FC<HunterCellProps> = ({ hunter, isActiveToday }) => (
  <td className="py-4 pl-4 align-middle">
    <div className="flex items-center gap-3">
      <div className="relative shrink-0">
        <div className="w-10 h-10 rounded-full bg-gradient-to-br from-gray-800 to-gray-900 border border-gray-700 flex items-center justify-center text-gray-400">
          <span className="text-sm font-bold">{hunter.name.charAt(0).toUpperCase()}</span>
        </div>
        {isActiveToday && (
          <span className="absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full bg-system-green border-2 border-gray-950" title="Aktif hari ini" />
        )}
      </div>
      <div className="min-w-0">
        <div className="text-sm text-white font-medium truncate group-hover:text-system-blue transition-colors">
          {hunter.name}
        </div>
        <div className="text-[10px] text-gray-500 font-mono">
          {hunter.user_id}
        </div>
      </div>
    </div>
  </td>
);

interface JobCellProps {
  hunter: EnrichedReport;
}

export const JobCell: React.FC<JobCellProps> = ({ hunter }) => (
  <td className="py-4 pl-4 align-middle">
    <span className={`text-xs px-2.5 py-1 rounded-md border font-mono ${getJobBadgeClass(hunter.job_class)}`}>
      {hunter.job_icon} {hunter.job_name}
    </span>
  </td>
);
