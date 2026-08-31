import { History, Dumbbell, type LucideIcon } from "lucide-react";

export type ProgramTab = "program" | "history";

const PROGRAM_TABS: { id: ProgramTab; label: string; icon: LucideIcon }[] = [
  { id: "program", label: "Program Aktif", icon: Dumbbell },
  { id: "history", label: "Riwayat Program", icon: History },
];

interface ProgramTabBarProps {
  activeTab: ProgramTab;
  onChange: (tab: ProgramTab) => void;
}

/**
 * Sub-navigation for the Program Latihan section of the personal page.
 * Mirrors the leaderboard tab-bar pattern (glass pill container, gradient
 * active state) so the page feels like one coherent system.
 */
export function ProgramTabBar({ activeTab, onChange }: ProgramTabBarProps) {
  return (
    <div
      role="tablist"
      aria-label="Navigasi program latihan"
      className="flex flex-nowrap overflow-x-auto scrollbar-none bg-gray-950/80 p-1.5 rounded-xl border border-gray-800/60 w-full lg:w-auto lg:max-w-fit"
    >
      {PROGRAM_TABS.map((tab) => {
        const Icon = tab.icon;
        const active = activeTab === tab.id;
        return (
          <button
            key={tab.id}
            role="tab"
            id={`program-tab-${tab.id}`}
            aria-selected={active}
            aria-controls={`program-panel-${tab.id}`}
            onClick={() => onChange(tab.id)}
            className={`flex shrink-0 whitespace-nowrap items-center gap-2 px-4 py-2 min-h-[40px] rounded-lg text-xs font-bold font-orbitron tracking-wider transition-all uppercase ${
              active
                ? "bg-gradient-to-r from-system-blue to-system-purple text-white shadow-neon-purple"
                : "text-gray-400 hover:text-white hover:bg-gray-800/60"
            }`}
          >
            <Icon size={14} />
            {tab.label}
          </button>
        );
      })}
    </div>
  );
}
