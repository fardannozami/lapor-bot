import { LayoutDashboard, Dumbbell, type LucideIcon } from "lucide-react";

export type PersonalPageTab = "summary" | "program";

const PERSONAL_TABS: {
  id: PersonalPageTab;
  label: string;
  icon: LucideIcon;
}[] = [
  { id: "summary", label: "Ringkasan", icon: LayoutDashboard },
  { id: "program", label: "Program Latihan", icon: Dumbbell },
];

interface PersonalPageTabsProps {
  activeTab: PersonalPageTab;
  onChange: (tab: PersonalPageTab) => void;
}

/**
 * Page-level navigation for the post-login personal page: splits the
 * long single-column profile into "Ringkasan" (stats/streaks/quests) and
 * "Program Latihan" (upcoming feature, backend not yet available).
 * Same pill/glass pattern as the leaderboard tab bar.
 */
export function PersonalPageTabs({
  activeTab,
  onChange,
}: PersonalPageTabsProps) {
  return (
    <div
      role="tablist"
      aria-label="Navigasi halaman profil"
      className="flex flex-nowrap overflow-x-auto scrollbar-none bg-gray-950/80 p-1.5 rounded-xl border border-gray-800/60 w-full lg:w-auto lg:max-w-fit"
    >
      {PERSONAL_TABS.map((tab) => {
        const Icon = tab.icon;
        const active = activeTab === tab.id;
        return (
          <button
            key={tab.id}
            role="tab"
            id={`page-tab-${tab.id}`}
            aria-selected={active}
            aria-controls={`page-panel-${tab.id}`}
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
