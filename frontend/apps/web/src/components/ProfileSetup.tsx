import React, { useState, useEffect } from "react";
import {
	User,
	Compass,
	Target,
	ChevronRight,
	Check,
	ArrowLeft,
	Loader2,
} from "lucide-react";
import { useRepositories } from "@lapor-bot/shared";
import type { EnrichedReport } from "@lapor-bot/shared";

interface JobInfo {
	id: string;
	name: string;
	icon: string;
	description: string;
	trait: string;
}

interface ProfileSetupProps {
	user: EnrichedReport;
	onComplete: (user: EnrichedReport) => void;
	onBack: () => void;
}

const STEPS = [
	{ step: 1, label: "Nama", icon: User },
	{ step: 2, label: "Job", icon: Compass },
	{ step: 3, label: "Goal", icon: Target },
] as const;

export function ProfileSetup({ user, onComplete, onBack }: ProfileSetupProps) {
	const { reports: repo } = useRepositories();

	const [step, setStep] = useState<1 | 2 | 3>(1);
	const [name, setName] = useState(user.name ?? "");
	const [selectedJobId, setSelectedJobId] = useState<string | null>(
		user.job_class || null,
	);
	const [targetDays, setTargetDays] = useState(3);
	const [activity, setActivity] = useState("Olahraga");
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const [jobs, setJobs] = useState<JobInfo[]>([]);
	const [jobsLoading, setJobsLoading] = useState(true);

	useEffect(() => {
		let cancelled = false;
		repo
			.listJobs()
			.then((data) => {
				if (!cancelled) setJobs(data);
			})
			.catch((err) => {
				if (!cancelled)
					setError(
						err instanceof Error ? err.message : "Gagal memuat daftar job",
					);
			})
			.finally(() => {
				if (!cancelled) setJobsLoading(false);
			});
		return () => {
			cancelled = true;
		};
	}, [repo]);

	const handleStep1 = async () => {
		const trimmed = name.trim();
		if (!trimmed) return;
		setError(null);
		setLoading(true);
		try {
			await repo.updateName(trimmed);
			setStep(2);
		} catch (err) {
			setError(err instanceof Error ? err.message : "Gagal memperbarui nama");
		} finally {
			setLoading(false);
		}
	};

	const handleStep2 = async () => {
		setError(null);
		setLoading(true);
		try {
			if (selectedJobId) {
				await repo.selectJob(selectedJobId);
			}
			setStep(3);
		} catch (err) {
			setError(err instanceof Error ? err.message : "Gagal memilih job");
		} finally {
			setLoading(false);
		}
	};

	const handleStep2Skip = () => {
		setStep(3);
	};

	const handleStep3 = async () => {
		setError(null);
		setLoading(true);
		try {
			await repo.setGoal(targetDays, activity.trim() || "Olahraga");
			const refreshedUser = await repo.fetchUser();
			onComplete(refreshedUser);
		} catch (err) {
			setError(err instanceof Error ? err.message : "Gagal menyimpan goal");
		} finally {
			setLoading(false);
		}
	};

	const handleBack = () => {
		if (step === 1) {
			onBack();
		} else {
			setStep((prev) => (prev - 1) as 1 | 2 | 3);
			setError(null);
		}
	};

	return (
		<div className="min-h-[60vh] px-4 py-8">
			<div className="max-w-lg mx-auto">
				{/* Back button */}
				<div className="mb-6">
					<button
						onClick={handleBack}
						disabled={loading}
						className="flex items-center gap-2 px-3 py-2 rounded-xl bg-gray-950 hover:bg-gray-900 border border-gray-800 text-gray-400 hover:text-white font-mono text-xs transition-colors disabled:opacity-50"
					>
						<ArrowLeft size={14} />
						{step === 1 ? "Kembali ke Login" : "Sebelumnya"}
					</button>
				</div>

				{/* Progress indicator */}
				<div className="flex items-center justify-center gap-3 mb-8">
					{STEPS.map((s, idx) => (
						<React.Fragment key={s.step}>
							<div className="flex flex-col items-center gap-2">
								<div
									className={`w-10 h-10 rounded-full flex items-center justify-center border-2 transition-all duration-300 ${
										step >= s.step
											? "border-system-green bg-system-green/20 text-system-green shadow-[0_0_12px_rgb(16,185,129,0.3)]"
											: "border-gray-700 bg-gray-950 text-gray-600"
									}`}
								>
									{step > s.step ? <Check size={18} /> : <s.icon size={18} />}
								</div>
								<span
									className={`text-[10px] font-mono uppercase tracking-wider ${
										step >= s.step ? "text-system-green" : "text-gray-600"
									}`}
								>
									{s.label}
								</span>
							</div>
							{idx < STEPS.length - 1 && (
								<div
									className={`w-12 h-0.5 rounded-full mt-[-18px] transition-colors duration-300 ${
										step > s.step ? "bg-system-green" : "bg-gray-800"
									}`}
								/>
							)}
						</React.Fragment>
					))}
				</div>

				{/* Error display */}
				{error && (
					<div className="mb-6 p-4 rounded-2xl bg-system-red/10 border border-system-red/35 flex items-start gap-3 text-sm text-red-300">
						<div>
							<p className="font-bold font-mono uppercase text-xs tracking-wider text-system-red">
								Error
							</p>
							<p className="mt-1 text-xs font-mono">{error}</p>
						</div>
					</div>
				)}

				{/* Step 1: Name */}
				{step === 1 && (
					<div className="glass rounded-[2rem] p-6 md:p-8 border border-gray-800/50">
						<div className="absolute right-0 top-0 h-36 w-36 rounded-full bg-system-green/10 blur-3xl" />
						<div className="flex items-center gap-3 mb-6">
							<div className="w-10 h-10 rounded-xl bg-system-green/15 border border-system-green/30 flex items-center justify-center">
								<User size={20} className="text-system-green" />
							</div>
							<h2 className="text-xl font-black font-orbitron text-white">
								Nama Hunter
							</h2>
						</div>
						<p className="text-xs text-gray-500 font-mono mb-6">
							Nama ini akan ditampilkan di leaderboard dan profil personal kamu.
						</p>
						<div className="space-y-4">
							<div>
								<label className="block text-[10px] text-gray-500 font-mono uppercase tracking-wider mb-2">
									Nama Kamu
								</label>
								<input
									type="text"
									value={name}
									onChange={(e) => setName(e.target.value)}
									onKeyDown={(e) => {
										if (e.key === "Enter" && name.trim()) handleStep1();
									}}
									placeholder="Masukkan nama hunter..."
									className="w-full px-4 py-3 rounded-2xl bg-gray-950 border border-gray-800 text-white font-mono text-sm placeholder:text-gray-600 focus:outline-none focus:border-system-green/50 focus:ring-1 focus:ring-system-green/20 transition-all"
									autoFocus
								/>
							</div>
							<button
								onClick={handleStep1}
								disabled={loading || !name.trim()}
								className="w-full flex items-center justify-center gap-2 px-6 py-3 rounded-2xl bg-system-green hover:bg-system-green/90 text-gray-950 font-bold font-orbitron text-sm transition-all disabled:opacity-40 disabled:cursor-not-allowed"
							>
								{loading ? (
									<Loader2 size={18} className="animate-spin" />
								) : (
									<>
										Lanjut
										<ChevronRight size={18} />
									</>
								)}
							</button>
						</div>
					</div>
				)}

				{/* Step 2: Job Selection */}
				{step === 2 && (
					<div className="glass rounded-[2rem] p-6 md:p-8 border border-gray-800/50">
						<div className="flex items-center gap-3 mb-6">
							<div className="w-10 h-10 rounded-xl bg-system-gold/15 border border-system-gold/30 flex items-center justify-center">
								<Compass size={20} className="text-system-gold" />
							</div>
							<h2 className="text-xl font-black font-orbitron text-white">
								Pilih Job
							</h2>
						</div>
						<p className="text-xs text-gray-500 font-mono mb-6">
							Job menentukan peran hunter kamu. Pilih satu, atau skip untuk
							melanjutkan.
						</p>

						{jobsLoading ? (
							<div className="flex items-center justify-center py-12">
								<Loader2 size={32} className="animate-spin text-system-green" />
							</div>
						) : (
							<div className="grid grid-cols-2 gap-3 mb-6">
								{jobs.map((job) => {
									const isSelected = selectedJobId === job.id;
									return (
										<button
											key={job.id}
											onClick={() => setSelectedJobId(job.id)}
											className={`text-left p-4 rounded-2xl border transition-all duration-200 ${
												isSelected
													? "border-system-gold bg-system-gold/10 shadow-[0_0_16px_rgb(234,179,8,0.15)]"
													: "border-gray-800 bg-gray-950/50 hover:border-gray-700"
											}`}
										>
											<div className="text-2xl mb-2">{job.icon}</div>
											<div className="text-sm font-bold font-orbitron text-white mb-1">
												{job.name}
											</div>
											<div className="text-[10px] text-gray-500 font-mono leading-relaxed mb-2">
												{job.description}
											</div>
											<div className="text-[10px] text-system-gold font-mono uppercase tracking-wider">
												{job.trait}
											</div>
										</button>
									);
								})}
							</div>
						)}

						<div className="flex gap-3">
							<button
								onClick={handleStep2Skip}
								disabled={loading}
								className="flex-1 px-6 py-3 rounded-2xl bg-gray-950 border border-gray-800 text-gray-400 hover:text-white font-mono text-xs transition-colors disabled:opacity-50"
							>
								Skip
							</button>
							<button
								onClick={handleStep2}
								disabled={loading || jobsLoading}
								className="flex-1 flex items-center justify-center gap-2 px-6 py-3 rounded-2xl bg-system-gold hover:bg-system-gold/90 text-gray-950 font-bold font-orbitron text-sm transition-all disabled:opacity-40 disabled:cursor-not-allowed"
							>
								{loading ? (
									<Loader2 size={18} className="animate-spin" />
								) : (
									<>
										Lanjut
										<ChevronRight size={18} />
									</>
								)}
							</button>
						</div>
					</div>
				)}

				{/* Step 3: Goal Setup */}
				{step === 3 && (
					<div className="glass rounded-[2rem] p-6 md:p-8 border border-gray-800/50">
						<div className="flex items-center gap-3 mb-6">
							<div className="w-10 h-10 rounded-xl bg-system-purple/15 border border-system-purple/30 flex items-center justify-center">
								<Target size={20} className="text-system-purple" />
							</div>
							<h2 className="text-xl font-black font-orbitron text-white">
								Setup Goal
							</h2>
						</div>
						<p className="text-xs text-gray-500 font-mono mb-6">
							Tetapkan target mingguan untuk konsistensi latihan kamu.
						</p>

						<div className="space-y-5 mb-6">
							<div>
								<label className="block text-[10px] text-gray-500 font-mono uppercase tracking-wider mb-2">
									Target Hari / Minggu
								</label>
								<select
									value={targetDays}
									onChange={(e) => setTargetDays(Number(e.target.value))}
									className="w-full px-4 py-3 rounded-2xl bg-gray-950 border border-gray-800 text-white font-mono text-sm focus:outline-none focus:border-system-purple/50 focus:ring-1 focus:ring-system-purple/20 transition-all appearance-none"
								>
									{[1, 2, 3, 4, 5, 6, 7].map((d) => (
										<option key={d} value={d}>
											{d} hari / minggu
										</option>
									))}
								</select>
							</div>

							<div>
								<label className="block text-[10px] text-gray-500 font-mono uppercase tracking-wider mb-2">
									Aktivitas
								</label>
								<input
									type="text"
									value={activity}
									onChange={(e) => setActivity(e.target.value)}
									onKeyDown={(e) => {
										if (e.key === "Enter") handleStep3();
									}}
									placeholder="Contoh: Olahraga, Lari, Gym..."
									className="w-full px-4 py-3 rounded-2xl bg-gray-950 border border-gray-800 text-white font-mono text-sm placeholder:text-gray-600 focus:outline-none focus:border-system-purple/50 focus:ring-1 focus:ring-system-purple/20 transition-all"
								/>
							</div>
						</div>

						<button
							onClick={handleStep3}
							disabled={loading}
							className="w-full flex items-center justify-center gap-2 px-6 py-3 rounded-2xl bg-system-purple hover:bg-system-purple/90 text-white font-bold font-orbitron text-sm transition-all disabled:opacity-40 disabled:cursor-not-allowed"
						>
							{loading ? (
								<Loader2 size={18} className="animate-spin" />
							) : (
								<>
									<Check size={18} />
									Selesai
								</>
							)}
						</button>
					</div>
				)}
			</div>
		</div>
	);
}
