import { History, Clock } from "lucide-react";
import type { TrainingProgram } from "@lapor-bot/shared";

interface ProgramHistoryTabProps {
  /** Past programs. Empty until the backend feature ships. */
  programs: TrainingProgram[];
}

/**
 * "Riwayat Program" tab of the Program Latihan section.
 * Empty state only for now - renders a list once the backend ships history.
 */
export function ProgramHistoryTab({ programs }: ProgramHistoryTabProps) {
  if (programs.length > 0) {
    return (
      <div className="space-y-3">
        {programs.map((program) => (
          <div
            key={program.id}
            className="glass rounded-2xl p-4 flex items-center justify-between gap-3"
          >
            <div className="min-w-0">
              <p className="text-sm font-bold font-mono text-white truncate">
                {program.title}
              </p>
              <p className="text-[10px] text-gray-500 font-mono mt-0.5">
                {program.start_date}
                {program.end_date ? ` → ${program.end_date}` : ""}
              </p>
            </div>
            <div className="text-right shrink-0">
              <div className="text-sm font-bold font-orbitron text-system-gold">
                {program.progress_percent}%
              </div>
              <div className="text-[10px] text-gray-500 font-mono">
                {program.completed_sessions}/{program.total_sessions} sesi
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="glass rounded-3xl p-6 sm:p-10 text-center">
      <div className="mx-auto w-16 h-16 rounded-3xl bg-gray-950 border border-gray-800 flex items-center justify-center mb-4">
        <History className="text-system-gold" size={28} />
      </div>
      <h3 className="text-lg font-bold font-orbitron text-white">
        Riwayat Program Kosong
      </h3>
      <p className="text-xs text-gray-500 font-mono mt-2 max-w-sm mx-auto leading-relaxed">
        Program yang sudah kamu selesaikan akan muncul di sini. Mulai dari tab
        Program Aktif untuk menyusun program pertamamu.
      </p>
      <p className="mt-4 inline-flex items-center gap-1.5 text-[10px] text-gray-600 font-mono">
        <Clock size={12} />
        Segera hadir - backend belum tersedia.
      </p>
    </div>
  );
}
