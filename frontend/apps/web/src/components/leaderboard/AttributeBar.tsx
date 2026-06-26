import React from "react";
import { attributeBarWidth } from "@lapor-bot/shared";

interface AttributeBarProps {
  value: number;
  label: string;
  tone?: "str" | "sta" | "agi" | "vit";
}

const TONE_GRADIENTS: Record<NonNullable<AttributeBarProps["tone"]>, string> = {
  str: "from-system-red to-amber-500",
  sta: "from-system-gold to-system-red",
  agi: "from-system-blue to-system-green",
  vit: "from-system-green to-system-blue",
};

export const AttributeBar: React.FC<AttributeBarProps> = ({
  value,
  label,
  tone,
}) => {
  const width = attributeBarWidth(value);
  const gradient = tone ? TONE_GRADIENTS[tone] : "from-system-blue to-system-purple";

  return (
    <div className="min-w-[120px]">
      <div className="flex items-center gap-2">
        <span className="text-[10px] font-mono text-gray-500 w-6">{label}</span>
        <div className="flex-1 h-2 rounded-full bg-gray-900 border border-gray-800 overflow-hidden">
          <div
            className={`h-full rounded-full bg-gradient-to-r ${gradient} transition-all duration-500`}
            style={{ width: `${width}%` }}
          />
        </div>
        <span className="text-xs font-mono font-bold text-white w-6 text-right">
          {value}
        </span>
      </div>
    </div>
  );
};
