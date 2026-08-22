import React from "react";
import { Trophy, Medal } from "lucide-react";

interface RankBadgeProps {
  rank: number;
}

/**
 * Podium treatment: top 3 get medal hardware (trophy for the leader,
 * medals for silver/bronze) plus a visible ordinal — rank is never
 * conveyed by color alone, so color-impaired users can still tell
 * silver from bronze.
 */
export const RankBadge: React.FC<RankBadgeProps> = ({ rank }) => {
  switch (rank) {
    case 0:
      return (
        <span
          title="Peringkat 1"
          className="relative inline-flex items-center justify-center gap-1 h-8 min-w-8 px-1.5 rounded-full bg-system-gold/20 text-system-gold border border-system-gold/50 shadow-neon-gold animate-pop"
        >
          <Trophy size={13} strokeWidth={2.5} />
          <span aria-hidden="true" className="text-[11px] font-black font-orbitron tnum leading-none">
            1
          </span>
          <span className="sr-only">Peringkat 1</span>
          {/* Live-leader ring */}
          <span
            aria-hidden="true"
            className="absolute inset-0 rounded-full animate-pulse-glow"
          ></span>
        </span>
      );
    case 1:
      return (
        <span
          title="Peringkat 2"
          className="inline-flex items-center justify-center gap-1 h-8 min-w-8 px-1.5 rounded-full bg-slate-300/15 text-slate-300 border border-slate-300/40 animate-pop"
        >
          <Medal size={13} strokeWidth={2.5} />
          <span aria-hidden="true" className="text-[11px] font-black font-orbitron tnum leading-none">
            2
          </span>
          <span className="sr-only">Peringkat 2</span>
        </span>
      );
    case 2:
      return (
        <span
          title="Peringkat 3"
          className="inline-flex items-center justify-center gap-1 h-8 min-w-8 px-1.5 rounded-full bg-amber-700/15 text-amber-600 border border-amber-600/40 animate-pop"
        >
          <Medal size={13} strokeWidth={2.5} />
          <span aria-hidden="true" className="text-[11px] font-black font-orbitron tnum leading-none">
            3
          </span>
          <span className="sr-only">Peringkat 3</span>
        </span>
      );
    default:
      return (
        <span className="text-gray-400 font-mono text-xs tnum">{rank + 1}</span>
      );
  }
};
