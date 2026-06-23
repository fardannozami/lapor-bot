import React from "react";

export const ActiveTodayBadge: React.FC = () => (
  <span className="inline-flex items-center gap-1 text-[10px] font-mono text-system-green">
    <span className="h-1.5 w-1.5 rounded-full bg-system-green animate-pulse" />
    hari ini
  </span>
);
