import React from "react";

interface QuestBadgeProps {
  completed: number;
  sideQuests: number;
}

export const QuestBadge: React.FC<QuestBadgeProps> = ({ completed, sideQuests }) => {
  if (completed === 0 && sideQuests === 0) {
    return <span className="text-gray-600 text-xs font-mono">—</span>;
  }

  return (
    <div className="flex flex-col items-center gap-0.5">
      {completed > 0 && (
        <span className="inline-flex items-center gap-1 text-xs font-mono text-system-green">
          🎯 {completed}
          <span className="text-[9px] text-gray-500">goals</span>
        </span>
      )}
      {sideQuests > 0 && (
        <span className="inline-flex items-center gap-1 text-xs font-mono text-system-purple">
          ⚡ {sideQuests}
          <span className="text-[9px] text-gray-500">quests</span>
        </span>
      )}
    </div>
  );
};
