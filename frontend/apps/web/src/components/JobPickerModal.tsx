import React from "react";
import { X, Loader2 } from "lucide-react";
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
		<div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-fade-in">
			{/* Backdrop click dismiss */}
			<div className="absolute inset-0" onClick={onClose} />

			<div className="relative w-full max-w-lg overflow-hidden glass rounded-3xl p-6 md:p-8 z-10 max-h-[90vh] flex flex-col">
				{/* Header */}
				<div className="flex items-start justify-between mb-5 border-b border-gray-800 pb-4 shrink-0">
					<div>
						<h2 className="text-xl font-bold font-orbitron text-white">
							Pilih Hunter Job
						</h2>
						<p className="text-xs text-gray-500 font-mono mt-1">
							Job menentukan peran dan side quest hunter kamu.
						</p>
					</div>
					<button
						onClick={onClose}
						disabled={selecting}
						className="p-2 text-gray-400 hover:text-white rounded-lg bg-gray-800/40 hover:bg-gray-800 transition-colors"
					>
						<X size={20} />
					</button>
				</div>

				{/* Error banner */}
				{error && (
					<div className="mb-4 p-3 rounded-xl bg-system-red/10 border border-system-red/30 text-sm text-system-red font-mono shrink-0">
						{error}
					</div>
				)}

				{/* Scrollable job list */}
				<div className="overflow-y-auto -mx-2 px-2 flex-1">
					{loading ? (
						<div className="flex flex-col items-center justify-center py-16 gap-3">
							<Loader2 size={28} className="animate-spin text-system-blue" />
							<span className="text-sm text-gray-400 font-mono">
								Memuat daftar job...
							</span>
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
										className={`relative text-left p-4 rounded-xl border transition-all duration-200 ${
											isCurrent
												? "border-system-green/50 bg-system-green/5 ring-1 ring-system-green/20"
												: "border-gray-800 bg-gray-900/50 hover:border-gray-700 hover:bg-gray-800/50"
										} ${selecting ? "opacity-50 cursor-not-allowed" : ""}`}
									>
										{/* Selecting spinner overlay */}
										{selecting && (
											<div className="absolute inset-0 flex items-center justify-center bg-gray-950/60 rounded-xl">
												<Loader2
													size={20}
													className="animate-spin text-system-blue"
												/>
											</div>
										)}
										<div className="flex items-center gap-2 mb-1">
											<span className="text-2xl">{j.icon}</span>
											<span
												className={`text-sm font-bold font-orbitron ${colorClass.split(" ")[0] || "text-white"}`}
											>
												{j.name}
											</span>
										</div>
										<div className="text-[11px] text-gray-400 font-mono leading-relaxed mb-2">
											{j.description}
										</div>
										{j.trait && (
											<div className="text-[10px] text-system-gold font-mono inline-flex items-center gap-1.5 bg-gray-950/50 px-2.5 py-1 rounded-md border border-gray-800">
												<span className="uppercase tracking-wider text-[9px]">
													Trait:
												</span>
												{j.trait}
											</div>
										)}
										{isCurrent && (
											<div className="mt-2 text-[10px] text-system-green font-mono font-bold uppercase tracking-wider">
												✦ Job Aktif
											</div>
										)}
									</button>
								);
							})}
						</div>
					)}
				</div>
			</div>
		</div>
	);
};
