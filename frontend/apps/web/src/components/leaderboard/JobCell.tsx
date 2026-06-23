import React from "react";
import type { EnrichedReport } from "@lapor-bot/shared";
import { getJobBadgeClass } from "@lapor-bot/shared";

export const JobCell: React.FC<{ hunter: EnrichedReport }> = ({ hunter }) => (
  <td className="py-4 pl-4 align-middle">
    <span
      className={`text-xs px-2.5 py-1 rounded-md border font-mono ${getJobBadgeClass(hunter.job_class)}`}
    >
      {hunter.job_icon} {hunter.job_name}
    </span>
  </td>
);
