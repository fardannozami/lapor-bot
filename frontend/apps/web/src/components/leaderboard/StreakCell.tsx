import React from "react";
import { Flame } from "lucide-react";

interface StreakCellProps {
  weeks: number;
  label?: string;
}

export const StreakCell: React.FC<StreakCellProps> = ({ weeks, label = "minggu" }) => (
  <span className="flex items-center justify-center gap-1 text-system-red font-mono">
    <Flame size={14} />
    <span className="font-bold">{weeks}</span>
    <span className="text-[10px] text-gray-500">{label}</span>
  </span>
);
