import React from "react";

interface RankBadgeProps {
  rank: number;
}

export const RankBadge: React.FC<RankBadgeProps> = ({ rank }) => {
  switch (rank) {
    case 0:
      return (
        <span className="flex items-center justify-center w-7 h-7 rounded-full bg-system-gold/20 text-system-gold border border-system-gold/50 shadow-neon-gold text-xs font-bold font-orbitron">
          1st
        </span>
      );
    case 1:
      return (
        <span className="flex items-center justify-center w-7 h-7 rounded-full bg-slate-300/20 text-slate-300 border border-slate-300/40 text-xs font-bold font-orbitron">
          2nd
        </span>
      );
    case 2:
      return (
        <span className="flex items-center justify-center w-7 h-7 rounded-full bg-amber-700/20 text-amber-600 border border-amber-600/40 text-xs font-bold font-orbitron">
          3rd
        </span>
      );
    default:
      return (
        <span className="text-gray-400 font-mono text-xs">{rank + 1}</span>
      );
  }
};
