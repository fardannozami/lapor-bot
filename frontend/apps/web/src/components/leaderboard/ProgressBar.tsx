import React from "react";
import type { TierProgress } from "@lapor-bot/shared";

type Tone = "gold" | "blue" | "red" | "green";

const TONE_GRADIENTS: Record<Tone, string> = {
  gold: "from-system-gold to-system-purple",
  blue: "from-system-blue to-system-purple",
  red: "from-system-red to-system-gold",
  green: "from-system-green to-system-blue",
};

interface ProgressBarProps {
  progress: TierProgress;
  valueLabel: string;
  tone?: Tone;
}

export const ProgressBar: React.FC<ProgressBarProps> = ({
  progress,
  valueLabel,
  tone = "blue",
}) => (
  <div className="min-w-[180px]">
    <div className="h-2.5 w-full rounded-full bg-gray-900 border border-gray-800 overflow-hidden">
      <div
        className={`h-full rounded-full bg-gradient-to-r ${TONE_GRADIENTS[tone]} transition-all duration-500`}
        style={{ width: `${progress.percent}%` }}
      />
    </div>
    <div className="mt-1 flex justify-between text-[9px] font-mono text-gray-500">
      <span>{valueLabel}</span>
      {progress.is_max ? (
        <span className="text-system-gold">MAX</span>
      ) : (
        <span>
          → {progress.next_icon} {progress.next_name} ({progress.remaining} pts)
        </span>
      )}
    </div>
  </div>
);
