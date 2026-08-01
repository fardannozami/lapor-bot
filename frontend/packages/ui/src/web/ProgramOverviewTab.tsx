import { Dumbbell, CalendarDays, Sparkles } from "lucide-react";
import type { TrainingProgram } from "@lapor-bot/shared";

interface ProgramOverviewTabProps {
  /** The user's active program. Null until the backend feature ships. */
  program: TrainingProgram | null;
  /** Placeholder hook for the upcoming "create program" flow. */
  onCreateProgram?: () => void;
}

/**
 * "Program Aktif" tab of the Program Latihan section.
 *
 * Backend does not exist yet: this renders a polished empty state that
 * explains the upcoming feature and prepares the layout for a real program.
 * Once the backend ships, pass a `TrainingProgram` and the section will
 * render the active program summary instead.
 */
export function ProgramOverviewTab({
  program,
  onCreateProgram,
}: ProgramOverviewTabProps) {
  if (program) {
    return (
      <div className="glass rounded-3xl p-4 sm:p-6">
        <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-3 mb-4">
          <div>
            <p className="text-[10px] text-system-blue font-mono uppercase tracking-[0.3em]">
              Program Latihan Aktif
            </p>
            <h3 className="mt-1 text-lg font-bold font-orbitron text-white">
              {program.title}
            </h3>
            {program.description && (
              <p className="text-xs text-gray-500 font-mono mt-1">
                {program.description}
              </p>
            )}
          </div>
          <div className="text-right shrink-0">
            <div className="text-2xl font-black font-orbitron text-system-blue">
              {program.progress_percent}%
            </div>
            <div className="text-[10px] text-gray-500 font-mono uppercase">
              Progress
            </div>
          </div>
        </div>

        <div className="h-3 rounded-full bg-gray-900 border border-gray-800 overflow-hidden p-[2px]">
          <div
            className="h-full rounded-full bg-gradient-to-r from-system-blue to-system-purple"
            style={{ width: `${program.progress_percent}%` }}
          />
        </div>

        <div className="mt-3 flex flex-wrap gap-3 text-[10px] text-gray-500 font-mono">
          <span className="inline-flex items-center gap-1.5">
            <CalendarDays size={12} />
            {program.start_date}
            {program.end_date ? ` → ${program.end_date}` : ""}
          </span>
          <span className="inline-flex items-center gap-1.5">
            <Dumbbell size={12} />
            {program.completed_sessions}/{program.total_sessions} sesi selesai
          </span>
          <span className="inline-flex items-center gap-1.5">
            <Sparkles size={12} />
            {program.sessions_per_week}x per minggu
          </span>
        </div>
      </div>
    );
  }

  return (
    <div className="glass rounded-3xl p-6 sm:p-10 text-center">
      <div className="mx-auto w-16 h-16 rounded-3xl bg-gray-950 border border-gray-800 flex items-center justify-center mb-4">
        <Dumbbell className="text-system-blue" size={28} />
      </div>
      <h3 className="text-lg font-bold font-orbitron text-white">
        Belum Ada Program Latihan
      </h3>
      <p className="text-xs text-gray-500 font-mono mt-2 max-w-sm mx-auto leading-relaxed">
        Program latihan pribadi akan segera hadir. Kamu bisa menyusun rencana
        mingguan (misal gym, lari, sepeda) dan melacak progress tiap sesi di
        sini.
      </p>

      <div className="mt-6 grid gap-3 sm:grid-cols-3 max-w-2xl mx-auto text-left">
        {[
          {
            icon: Dumbbell,
            title: "Rencana Mingguan",
            desc: "Susun sesi latihan per hari sesuai jadwalmu.",
          },
          {
            icon: Sparkles,
            title: "Progress Per Sesi",
            desc: "Tandai sesi selesai dan lihat progress program.",
          },
          {
            icon: CalendarDays,
            title: "Konsisten Setiap Minggu",
            desc: "Program mendukung target mingguanmu.",
          },
        ].map((f) => {
          const Icon = f.icon;
          return (
            <div
              key={f.title}
              className="rounded-2xl border border-gray-800/60 bg-gray-950/50 p-4"
            >
              <Icon className="text-system-purple mb-2" size={18} />
              <p className="text-xs font-bold font-mono text-gray-200">
                {f.title}
              </p>
              <p className="text-[10px] text-gray-500 font-mono mt-1 leading-relaxed">
                {f.desc}
              </p>
            </div>
          );
        })}
      </div>

      <button
        type="button"
        disabled
        onClick={onCreateProgram}
        className="mt-6 inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-system-blue text-gray-950 text-sm font-bold opacity-40 cursor-not-allowed"
        title="Fitur program latihan belum tersedia"
      >
        <Dumbbell size={14} />
        Buat Program Latihan
      </button>
      <p className="mt-2 text-[10px] text-gray-600 font-mono">
        Segera hadir - backend belum tersedia.
      </p>
    </div>
  );
}
