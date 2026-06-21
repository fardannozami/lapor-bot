import React from "react";
import { X } from "lucide-react";
import { getJobColor } from "@lapor-bot/shared";
import type { JobInfo } from "@lapor-bot/shared";

interface JobPickerModalProps {
  jobs: JobInfo[];
  loading: boolean;
  error: string | null;
  currentJobId: string;
  onSelect: (jobId: string) => Promise<void>;
  onClose: () => void;
  selecting: boolean;
}

export const JobPickerModal: React.FC<JobPickerModalProps> = ({
  jobs,
  loading,
  error,
  currentJobId,
  onSelect,
  onClose,
  selecting,
}) => {
  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
      {/* Backdrop click dismiss */}
      <div className="absolute inset-0" onClick={onClose} />

      <div className="relative w-full h-full sm:h-auto sm:max-w-lg rounded-3xl border border-gray-800 bg-gray-950/95 p-6 z-10 overflow-y-auto max-h-[90vh]">
        {/* Header */}
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-xl font-bold font-orbitron text-white">
            Pilih Hunter Job
          </h2>
          <button
            onClick={onClose}
            className="p-2 text-gray-400 hover:text-white rounded-lg bg-gray-800/40 hover:bg-gray-800 transition-colors"
          >
            <X size={18} />
          </button>
        </div>

        {/* Error banner */}
        {error && (
          <div className="mb-4 p-3 rounded-xl bg-system-red/10 border border-system-red/30 text-sm text-system-red font-mono">
            {error}
          </div>
        )}

        {/* Loading state */}
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-sm text-gray-400">Loading jobs...</div>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {jobs.map((j) => {
              const isCurrent = j.id === currentJobId;
              const colorClass = getJobColor(j.id);
              return (
                <button
                  key={j.id}
                  disabled={selecting}
                  onClick={() => onSelect(j.id)}
                  className={`relative text-left p-4 rounded-xl border transition-colors ${
                    isCurrent
                      ? "border-system-green/50 bg-system-green/5"
                      : "border-gray-800 bg-gray-950/50 hover:border-gray-700"
                  } ${selecting ? "opacity-50 cursor-not-allowed" : ""}`}
                >
                  {/* Selecting spinner overlay */}
                  {selecting && (
                    <div className="absolute inset-0 flex items-center justify-center bg-gray-950/60 rounded-xl">
                      <div className="w-5 h-5 border-2 border-t-transparent border-system-blue rounded-full animate-spin" />
                    </div>
                  )}
                  <div className="text-2xl mb-1">{j.icon}</div>
                  <div className={`text-sm font-bold ${colorClass.split(" ")[0]}`}>
                    {j.name}
                  </div>
                  <div className="text-[11px] text-gray-500 mt-1 leading-relaxed">
                    {j.description}
                  </div>
                  {j.trait && (
                    <div className="mt-2 text-[10px] text-gray-400 font-mono bg-gray-900/50 px-2 py-0.5 rounded-md inline-block">
                      {j.trait}
                    </div>
                  )}
                </button>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};
