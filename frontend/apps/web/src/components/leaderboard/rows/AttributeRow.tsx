import React from "react";
import type { EnrichedReport, AttributeTab } from "@lapor-bot/shared";
import { getAttributeAverage, getAttributeValue } from "@lapor-bot/shared";
import { RankBadge } from "../RankBadge";
import { HunterCell } from "../HunterCell";
import { JobCell } from "../JobCell";
import { AttributeBar } from "../AttributeBar";

interface AttributeRowProps {
  hunter: EnrichedReport;
  rank: number;
  rowClass: string;
  attributeTab: AttributeTab;
  onSelectHunter: (hunter: EnrichedReport) => void;
}

export const AttributeRow: React.FC<AttributeRowProps> = ({
  hunter,
  rank,
  rowClass,
  attributeTab,
  onSelectHunter,
}) => {
  const isOverall = attributeTab === "overall";
  const primaryValue = isOverall
    ? getAttributeAverage(hunter)
    : getAttributeValue(hunter, attributeTab);

  return (
    <tr
      key={hunter.user_id}
      onClick={() => onSelectHunter(hunter)}
      className={rowClass}
    >
      <td className="py-4 pl-4 text-center font-mono align-middle">
        <div className="flex justify-center">
          <RankBadge rank={rank} />
        </div>
      </td>
      <HunterCell hunter={hunter} />
      <JobCell hunter={hunter} />

      <td className="py-4 pl-4 text-center align-middle">
        <AttributeBar
          value={primaryValue}
          label={isOverall ? "AVG" : attributeTab.toUpperCase()}
          tone={isOverall ? undefined : (attributeTab as "str" | "sta" | "agi" | "vit")}
        />
      </td>

      {isOverall ? (
        <>
          <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
            <span className="font-bold">{hunter.str}</span>
          </td>
          <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
            <span className="font-bold">{hunter.sta}</span>
          </td>
          <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
            <span className="font-bold">{hunter.agi}</span>
          </td>
          <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
            <span className="font-bold">{hunter.vit}</span>
          </td>
        </>
      ) : (
        <td className="py-4 pl-4 text-center align-middle font-mono text-sm text-gray-300">
          <span className="font-bold">{hunter.total_points}</span>
          <span className="text-[10px] text-gray-500"> XP</span>
        </td>
      )}

      <td className="py-4 pl-4 pr-4 text-center align-middle font-mono text-sm text-gray-300">
        <span className="font-bold">Lv.{hunter.level}</span>
      </td>
    </tr>
  );
};
