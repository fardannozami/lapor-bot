import {
	ArrowLeft,
	Shield,
	Swords,
	Zap,
	Heart,
	LogOut,
	ScrollText,
	CheckCircle2,
	Circle,
	Flame,
	Target,
	Award,
	TrendingUp,
	Activity,
	CalendarDays,
	Edit2,
} from "lucide-react";
import { useState, useEffect } from "react";
import { useRepositories, getJobColor } from "@lapor-bot/shared";
import type {
	DailyActivity,
	EnrichedReport,
	QuestTask,
	JobInfo,
} from "@lapor-bot/shared";
import { JobPickerModal } from "./JobPickerModal";
interface PersonalPageProps {
	user: EnrichedReport;
	onLogout: () => void;
	onUserRefresh?: (user: EnrichedReport) => void;
}

function getRankGlow(rankName: string) {
	if (rankName.includes("S-Rank") || rankName.includes("Monarch"))
		return "glass-glow-gold";
	if (rankName.includes("A-Rank")) return "glass-glow-purple";
	if (rankName.includes("B-Rank") || rankName.includes("C-Rank"))
		return "glass-glow-blue";
	return "glass-glow-blue";
}

function formatQuestDifficulty(difficulty: string) {
	if (difficulty === "easy") return "🟢 Easy";
	if (difficulty === "medium") return "🟡 Medium";
	if (difficulty === "hard") return "🔴 Hard";
	return difficulty;
}

function formatQuestTarget(task: QuestTask) {
	if (task.id === "easycardio")
		return "jalan kaki 4000 langkah atau sepeda 5 km";
	if (task.unit === "100m") return `${(task.target / 10).toFixed(1)} km`;
	return `${task.target} ${task.unit}`;
}

function formatQuestProgress(task: QuestTask) {
	if (task.id === "easycardio")
		return task.progress >= task.target ? "selesai" : "belum selesai";
	if (task.unit === "100m")
		return `${(task.progress / 10).toFixed(1)} / ${(task.target / 10).toFixed(1)} km`;
	return `${task.progress} / ${task.target} ${task.unit}`;
}

function formatDate(date: string) {
	return new Intl.DateTimeFormat("id-ID", {
		day: "2-digit",
		month: "short",
	}).format(new Date(`${date}T00:00:00`));
}

function chunkWeeks(days: DailyActivity[]) {
	const weeks: DailyActivity[][] = [];
	for (let i = 0; i < days.length; i += 7) {
		weeks.push(days.slice(i, i + 7));
	}
	return weeks;
}

function StatCard({
	label,
	value,
	tone = "text-white",
}: {
	label: string;
	value: string | number;
	tone?: string;
}) {
	return (
		<div className="rounded-2xl border border-gray-800/50 bg-gray-950/50 p-4">
			<div className={`text-xl font-black font-orbitron ${tone}`}>{value}</div>
			<div className="mt-1 text-[10px] text-gray-500 font-mono uppercase tracking-wider">
				{label}
			</div>
		</div>
	);
}

export function PersonalPage({
	user,
	onLogout,
	onUserRefresh,
}: PersonalPageProps) {
	const { reports: repo } = useRepositories();
	const phone = user.user_id;

	// inline edit states
	const [editingName, setEditingName] = useState(false);
	const [nameValue, setNameValue] = useState(user.name || "");
	const [nameLoading, setNameLoading] = useState(false);
	const [nameError, setNameError] = useState<string | null>(null);

	const [jobs, setJobs] = useState<JobInfo[]>([]);
	const [jobsLoading, setJobsLoading] = useState(false);
	const [showJobModal, setShowJobModal] = useState(false);
	const [jobLoading, setJobLoading] = useState(false);
	const [jobError, setJobError] = useState<string | null>(null);

	const [showGoalForm, setShowGoalForm] = useState(false);
	const [targetDays, setTargetDays] = useState<number>(
		user.active_goal?.target_days || 3,
	);
	const [activity, setActivity] = useState<string>(
		user.active_goal?.activity || "Olahraga",
	);
	const [startDate, setStartDate] = useState<string>(""); // YYYY-MM-DD
	const [startHour, setStartHour] = useState<number>(0);
	const [goalLoading, setGoalLoading] = useState(false);
	const [goalError, setGoalError] = useState<string | null>(null);
	const [resetLoading, setResetLoading] = useState(false);

	const [localUser, setLocalUser] = useState<EnrichedReport>(user);

	// keep local in sync if parent pushes a new user object
	useEffect(() => {
		setLocalUser(user);
		setNameValue(user.name || "");
	}, [user]);

	const refreshUser = async () => {
		try {
			const refreshed = await repo.fetchUserByPhone(phone);
			setLocalUser(refreshed);
			if (onUserRefresh) onUserRefresh(refreshed);
		} catch {
			// ignore; parent will still have old snapshot
		}
	};


	const glowClass = getRankGlow(localUser.rank_name);
	const sideQuests = localUser.today_side_quests ?? [];
	const dailyActivity = localUser.daily_activity ?? [];
	const activeGoal = localUser.active_goal;
	const allBadges = [
		...localUser.achievements,
		...localUser.seasonal_achievements,
	];
	const xpPercent = Math.min(
		100,
		Math.round(
			(localUser.xp_progress.CurrentXP / localUser.xp_progress.RequiredXP) *
				100,
		),
	);
	const completedSideQuests = sideQuests.filter(
		(quest) => quest.progress >= quest.target,
	).length;

	return (
		<div className="min-h-[60vh] px-3 sm:px-4 py-6 sm:py-8 overflow-hidden">
			<div className="max-w-6xl mx-auto">
				<div className="flex items-center justify-between mb-6">
					<button
						onClick={onLogout}
						className="flex items-center gap-2 px-3 py-2 rounded-xl bg-gray-950 hover:bg-gray-900 border border-gray-800 text-gray-400 hover:text-white font-mono text-xs transition-colors"
					>
						<ArrowLeft size={14} />
						Kembali
					</button>

					<button
						onClick={onLogout}
						className="flex items-center gap-2 px-3 py-2 rounded-xl bg-system-red/10 hover:bg-system-red/20 border border-system-red/20 text-system-red font-mono text-xs transition-colors"
					>
						<LogOut size={14} />
						Keluar
					</button>
				</div>

				<section
					className={`relative overflow-hidden glass rounded-[2rem] p-4 sm:p-6 md:p-8 mb-6 ${glowClass}`}
				>
					<div className="absolute right-0 top-0 h-36 w-36 rounded-full bg-system-green/10 blur-3xl" />
					<div className="relative grid gap-6 lg:grid-cols-[1.25fr_0.75fr]">
						<div className="flex items-start gap-4">
							<div className="w-12 h-12 sm:w-16 sm:h-16 rounded-3xl bg-gray-950 border border-gray-800 flex items-center justify-center text-3xl shrink-0 shadow-neon-purple">
								{localUser.job_icon}
							</div>
							<div className="min-w-0">
								<p className="text-[10px] text-system-green font-mono uppercase tracking-[0.3em]">
									Personal Hunter Profile
								</p>
								<div className="flex items-center gap-2">
									<h2 className="mt-2 text-2xl sm:text-3xl md:text-4xl font-black font-orbitron text-white tracking-wide truncate">
										{localUser.name}
									</h2>
									<button
										onClick={() => {
											setEditingName(true);
											setNameValue(localUser.name || "");
											setNameError(null);
										}}
										className="mt-2 text-xs px-2 py-1 rounded-lg border border-gray-700 text-gray-400 hover:text-white hover:border-gray-500 font-mono"
										title="Ubah nama"
									>
										<Edit2 size={14} />
									</button>
								</div>
								{editingName && (
									<div className="mt-2 flex items-center gap-2">
										<input
											value={nameValue}
											onChange={(e) => setNameValue(e.target.value)}
											className="px-3 py-1.5 rounded-xl bg-gray-950 border border-gray-800 text-sm font-mono text-white"
											placeholder="Nama hunter"
										/>
										<button
											disabled={nameLoading}
											onClick={async () => {
												const v = nameValue.trim();
												if (!v) return;
												setNameLoading(true);
												setNameError(null);
												try {
													await repo.updateName(phone, v);
													await refreshUser();
													setEditingName(false);
												} catch (e: any) {
													setNameError(e?.message || "Gagal update nama");
												} finally {
													setNameLoading(false);
												}
											}}
											className="px-3 py-1.5 rounded-xl bg-system-green text-gray-950 text-xs font-bold"
										>
											{nameLoading ? "..." : "Simpan"}
										</button>
										<button
											onClick={() => setEditingName(false)}
											className="px-2 py-1.5 text-xs text-gray-400"
										>
											Batal
										</button>
									</div>
								)}
								{nameError && (
									<p className="text-[10px] text-system-red mt-1">
										{nameError}
									</p>
								)}
								<p className="text-sm text-gray-400 font-mono mt-1">
									{localUser.job_name} {localUser.level_icon} Lv.
									{localUser.level} • {localUser.rank_icon}{" "}
									{localUser.rank_name}
								</p>
								<p className="text-[10px] text-gray-600 font-mono mt-1">
									{localUser.user_id}
								</p>
							</div>
						</div>

						<div className="rounded-3xl border border-gray-800 bg-gray-950/50 p-4">
							<div className="flex items-center justify-between gap-4 mb-3">
								<div>
									<p className="text-[10px] text-gray-500 font-mono uppercase tracking-wider">
										Level Progress
									</p>
									<p className="text-lg font-bold font-orbitron text-white">
										{localUser.level_icon} {localUser.level_name}
									</p>
								</div>
								<div className="text-right">
									<p className="text-xl font-black font-orbitron text-system-gold">
										{localUser.total_points}
									</p>
									<p className="text-[10px] text-gray-500 font-mono uppercase">
										Lifetime XP
									</p>
								</div>
							</div>
							<div className="h-3 rounded-full bg-gray-900 border border-gray-800 overflow-hidden p-[2px]">
								<div
									className="h-full rounded-full bg-gradient-to-r from-system-blue to-system-green"
									style={{ width: `${xpPercent}%` }}
								/>
							</div>
							<div className="mt-2 flex justify-between text-[10px] text-gray-500 font-mono">
								<span>{localUser.xp_progress.CurrentXP} XP</span>
								<span className="text-system-blue font-bold">{xpPercent}%</span>
								<span>{localUser.xp_progress.RequiredXP} XP</span>
							</div>
						</div>
					</div>
				</section>

				<section className="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-6">
					<StatCard
						label="Season Points"
						value={localUser.seasonal_points}
						tone="text-system-gold"
					/>
					<StatCard
						label="Daily Streak"
						value={`${localUser.current_daily_streak ?? 0} hari`}
						tone="text-system-red"
					/>
					<StatCard
						label="Weekly Streak"
						value={`${localUser.streak} minggu`}
						tone="text-system-green"
					/>
					<StatCard
						label="Active Window"
						value={`${localUser.active_days_in_window ?? 0}/${dailyActivity.length || 35}`}
						tone="text-system-blue"
					/>
				</section>

				<div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
					<div className="space-y-6">
						{/* Daily Streak Map — improved looks + on-theme wording */}
						<div className="glass rounded-3xl p-4 sm:p-6">
							<div className="flex items-start justify-between gap-4 mb-5">
								<div>
									<h3 className="text-lg font-bold font-orbitron text-white flex items-center gap-2">
										<Flame className="text-system-red" size={18} />
										Daily Streak Map
									</h3>
									<p className="text-xs text-gray-500 font-mono mt-1 leading-relaxed">
										Peta konsistensi harian pribadi. Data hanya untuk melacak
										kebiasaan rutin kamu sendiri.
									</p>
								</div>
								<div className="text-right shrink-0">
									<div className="text-sm font-bold font-orbitron text-white">
										{localUser.longest_daily_streak ?? 0} hari
									</div>
									<div className="text-[10px] text-gray-500 font-mono uppercase">
										Best Daily
									</div>
								</div>
							</div>

							{/* Improved heatmap */}
							<div className="overflow-x-auto pb-2">
								<div
									className="inline-flex gap-1.5 p-3 rounded-2xl bg-gray-950/70 border border-gray-800/60"
									aria-label="Peta aktivitas harian konsistensi pribadi"
								>
									{chunkWeeks(dailyActivity).map((week, weekIdx) => (
										<div
											key={`week-${weekIdx}`}
											className="grid grid-rows-7 gap-[5px]"
										>
											{week.map((day) => {
												const intensity =
													!day.active || day.count <= 0
														? "empty"
														: day.count >= 3
															? "high"
															: day.count === 2
																? "mid"
																: "low";

												const cellClass =
													intensity === "high"
														? "bg-system-green border-system-green shadow-[0_0_8px_rgb(16,185,129,0.65)]"
														: intensity === "mid"
															? "bg-system-green/90 border-system-green/80 shadow-[0_0_4px_rgb(16,185,129,0.5)]"
															: intensity === "low"
																? "bg-system-green border-system-green/70"
																: "bg-[#111418] border-gray-800/70";

												return (
													<span
														key={day.date}
														className={`h-[22px] w-[22px] rounded-[5px] border transition-all hover:scale-110 hover:brightness-110 ${cellClass}`}
													/>
												);
											})}
										</div>
									))}
								</div>
							</div>

							<div className="mt-5 flex flex-wrap items-center justify-between gap-4 text-[10px] text-gray-500 font-mono uppercase tracking-wider">
								<div className="text-gray-400">
									{dailyActivity[0] ? formatDate(dailyActivity[0].date) : "—"} →{" "}
									{dailyActivity.at(-1)
										? formatDate(dailyActivity.at(-1)!.date)
										: "—"}
								</div>
								<div className="flex items-center gap-3">
									<div className="flex items-center gap-1.5">
										<span className="h-3.5 w-3.5 rounded-[3px] bg-[#111418] border border-gray-800/70" />
										<span>Rest</span>
									</div>
									<div className="flex items-center gap-1.5">
										<span className="h-3.5 w-3.5 rounded-[3px] bg-system-green border border-system-green/60" />
										<span>Active</span>
									</div>
								</div>
							</div>
						</div>

						{/* Weekly Goal (kept intact) */}
						<section className={`glass rounded-3xl p-4 sm:p-6 ${glowClass}`}>
							<div className="flex items-start justify-between gap-4 mb-5">
								<div>
									<h3 className="text-lg font-bold font-orbitron text-white flex items-center gap-2">
										<Target className="text-system-gold" size={18} />
										Weekly Goal
									</h3>
									<p className="text-xs text-gray-500 font-mono mt-1">
										Data dari #goal, khusus progress pribadi minggu ini.
									</p>
								</div>
								{activeGoal && (
									<div className="text-right shrink-0">
										<div className="text-sm font-bold font-orbitron text-white">
											{activeGoal.completed_days}/{activeGoal.target_days}
										</div>
										<div className="text-[10px] text-gray-500 font-mono uppercase">
											Selesai
										</div>
									</div>
								)}
							</div>

							{!activeGoal ? (
								<div className="rounded-2xl border border-gray-800/50 bg-gray-950/50 p-4 text-center">
									<CalendarDays
										className="mx-auto mb-2 text-gray-600"
										size={22}
									/>
									<p className="text-sm text-gray-400 font-mono">
										Belum ada goal aktif.
									</p>
									<p className="text-xs text-gray-600 font-mono mt-1">
										Buat dari WhatsApp: #goal set 3 Olahraga
									</p>
								</div>
							) : (
								<div>
									<div className="mb-4 rounded-2xl border border-gray-800 bg-gray-950/50 p-4">
										<div className="flex items-center justify-between gap-3 mb-2">
											<p className="text-sm text-white font-bold font-mono">
												{activeGoal.target_days}x {activeGoal.activity}
											</p>
											<p className="text-xs text-system-gold font-mono">
												{activeGoal.percent}%
											</p>
										</div>
										<div className="h-3 rounded-full bg-gray-900 border border-gray-800 overflow-hidden p-[2px]">
											<div
												className="h-full rounded-full bg-gradient-to-r from-system-gold to-system-green"
												style={{ width: `${activeGoal.percent}%` }}
											/>
										</div>
										<p className="mt-2 text-[10px] text-gray-500 font-mono uppercase">
											Sisa {activeGoal.remaining_days} hari untuk mencapai goal.
										</p>
									</div>
									<div className="grid grid-cols-7 gap-1 sm:gap-2">
										{activeGoal.days.map((day) => (
											<div
												key={day.date}
												className={`rounded-xl border p-1 sm:p-2 text-center ${day.active ? "border-system-green/50 bg-system-green/10" : "border-gray-800 bg-gray-950/50"}`}
												title={day.activity}
											>
												<div className="text-[10px] text-gray-500 font-mono uppercase">
													{day.day_label}
												</div>
												<div className="mt-1 flex justify-center">
													{day.active ? (
														<CheckCircle2
															size={16}
															className="text-system-green"
														/>
													) : (
														<Circle size={16} className="text-gray-700" />
													)}
												</div>
											</div>
										))}
									</div>
								</div>
							)}

							{/* Goal setter controls in Personal Dashboard */}
							<div className="mt-4 pt-4 border-t border-gray-800/60">
								<div className="flex items-center gap-2 mb-2 flex-wrap">
									<button
										onClick={() => {
											setShowGoalForm((v) => !v);
											setGoalError(null);
											// prefill from current if any
											if (localUser.active_goal) {
												setTargetDays(localUser.active_goal.target_days || 3);
												setActivity(
													localUser.active_goal.activity || "Olahraga",
												);
											}
										}}
										className="text-xs px-3 py-1.5 rounded-xl border border-gray-700 text-gray-300 hover:text-white"
									>
										{showGoalForm ? "Tutup Form Goal" : "Atur / Ubah Goal"}
									</button>
									{localUser.active_goal && (
										<button
											disabled={resetLoading}
											onClick={async () => {
												setResetLoading(true);
												setGoalError(null);
												try {
													// prefer dedicated if present, else use action=reset
													if ((repo as any).resetGoal) {
														await (repo as any).resetGoal(phone);
													} else {
														await repo.setGoal(phone, 0, "", {
															/* triggers reset via backend action */
														} as any);
														// send reset via post with action if supported
														await fetch("/api/user/goal", {
															method: "POST",
															headers: { "Content-Type": "application/json" },
															body: JSON.stringify({ phone, action: "reset" }),
														});
													}
													await refreshUser();
												} catch (e: any) {
													setGoalError(e?.message || "Gagal reset goal");
												} finally {
													setResetLoading(false);
												}
											}}
											className="text-xs px-3 py-1.5 rounded-xl border border-system-red/30 text-system-red hover:bg-system-red/10"
										>
											{resetLoading ? "Reset..." : "Reset Goal"}
										</button>
									)}
								</div>

								{showGoalForm && (
									<div className="rounded-2xl border border-gray-800 bg-gray-950/50 p-4 space-y-3">
										<div>
											<label className="block text-[10px] text-gray-500 font-mono mb-1">
												Target Hari (1-7)
											</label>
											<select
												value={targetDays}
												onChange={(e) =>
													setTargetDays(parseInt(e.target.value))
												}
												className="w-full px-3 py-2 rounded-xl bg-gray-950 border border-gray-800 text-sm font-mono"
											>
												{[1, 2, 3, 4, 5, 6, 7].map((d) => (
													<option key={d} value={d}>
														{d} hari / minggu
													</option>
												))}
											</select>
										</div>
										<div>
											<label className="block text-[10px] text-gray-500 font-mono mb-1">
												Aktivitas
											</label>
											<input
												value={activity}
												onChange={(e) => setActivity(e.target.value)}
												className="w-full px-3 py-2 rounded-xl bg-gray-950 border border-gray-800 text-sm font-mono"
												placeholder="Olahraga"
											/>
										</div>
										<div className="grid grid-cols-2 gap-3">
											<div>
												<label className="block text-[10px] text-gray-500 font-mono mb-1">
													Mulai Tanggal (opsional)
												</label>
												<input
													type="date"
													value={startDate}
													onChange={(e) => setStartDate(e.target.value)}
													className="w-full px-3 py-2 rounded-xl bg-gray-950 border border-gray-800 text-sm font-mono"
												/>
											</div>
											<div>
												<label className="block text-[10px] text-gray-500 font-mono mb-1">
													Jam Mulai (0-23)
												</label>
												<input
													type="number"
													min={0}
													max={23}
													value={startHour}
													onChange={(e) =>
														setStartHour(parseInt(e.target.value || "0"))
													}
													className="w-full px-3 py-2 rounded-xl bg-gray-950 border border-gray-800 text-sm font-mono"
												/>
											</div>
										</div>

										{goalError && (
											<p className="text-[10px] text-system-red">{goalError}</p>
										)}

										<div className="flex gap-2">
											<button
												disabled={goalLoading}
												onClick={async () => {
													setGoalLoading(true);
													setGoalError(null);
													try {
														const startPayload: any = {};
														if (startDate) startPayload.startDate = startDate;
														if (startHour || startHour === 0)
															startPayload.startHour = startHour;
														// if user left startDate empty we can still pass hour-only (backend resolves to today)
														await repo.setGoal(
															phone,
															targetDays,
															activity || "Olahraga",
															startPayload,
														);
														await refreshUser();
														setShowGoalForm(false);
													} catch (e: any) {
														setGoalError(
															e?.message ||
																"Gagal menyimpan goal (mungkin sudah ada goal aktif)",
														);
													} finally {
														setGoalLoading(false);
													}
												}}
												className="flex-1 px-4 py-2 rounded-xl bg-system-gold text-gray-950 text-sm font-bold"
											>
												{goalLoading ? "Menyimpan..." : "Simpan Goal"}
											</button>
											<button
												onClick={() => setShowGoalForm(false)}
												className="px-4 py-2 rounded-xl border border-gray-700 text-xs"
											>
												Batal
											</button>
										</div>
										<p className="text-[10px] text-gray-500">
											Goal berlaku 7 hari dari tanggal+jam yang dipilih.
											Kosongkan tanggal untuk mulai hari ini.
										</p>
									</div>
								)}
							</div>
						</section>

						{/* Side Quest Hari Ini (kept intact) */}
						<section className={`glass rounded-3xl p-4 sm:p-6 ${glowClass}`}>
							<div className="flex items-start justify-between gap-4 mb-5">
								<div>
									<h3 className="text-lg font-bold font-orbitron text-white flex items-center gap-2">
										<ScrollText className="text-system-green" size={18} />
										Side Quest Hari Ini
									</h3>
									<p className="text-xs text-gray-500 font-mono mt-1 leading-relaxed">
										Selesaikan via WhatsApp:{" "}
										<span className="text-gray-300">
											/lapor sidequest &lt;kegiatan&gt; &lt;jumlah&gt;
										</span>
										.
									</p>
								</div>
								<div className="text-right shrink-0">
									<div className="text-sm font-bold font-orbitron text-white">
										{completedSideQuests}/{sideQuests.length}
									</div>
									<div className="text-[10px] text-gray-500 font-mono uppercase">
										Selesai
									</div>
								</div>
							</div>

							{sideQuests.length === 0 ? (
								<div className="rounded-2xl border border-gray-800/50 bg-gray-950/50 p-4 text-center">
									<p className="text-sm text-gray-400 font-mono">
										Side quest belum terbuka.
									</p>
									<p className="text-xs text-gray-600 font-mono mt-1">
										Side quest tersedia untuk profil yang sudah punya job.
									</p>
								</div>
							) : (
								<div className="space-y-3">
									{sideQuests.map((quest) => {
										const done = quest.progress >= quest.target;
										return (
											<div
												key={quest.id}
												className="rounded-2xl border border-gray-800/50 bg-gray-950/50 p-4"
											>
												<div className="flex items-start gap-3">
													{done ? (
														<CheckCircle2
															className="text-system-green shrink-0 mt-0.5"
															size={18}
														/>
													) : (
														<Circle
															className="text-gray-600 shrink-0 mt-0.5"
															size={18}
														/>
													)}
													<div className="flex-1 min-w-0">
														<div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1">
															<p className="text-sm font-bold text-white font-mono">
																{quest.name}
															</p>
															<span className="text-[10px] text-gray-400 font-mono uppercase tracking-wider">
																{formatQuestDifficulty(quest.difficulty)}
															</span>
														</div>
														<p className="text-xs text-gray-500 font-mono mt-1">
															Target: {formatQuestTarget(quest)}
														</p>
														<p className="text-xs text-gray-600 font-mono mt-1">
															Progress: {formatQuestProgress(quest)}
														</p>
													</div>
												</div>
											</div>
										);
									})}
								</div>
							)}
						</section>
					</div>

					{/* Right aside — attributes and achievements post incoming */}
					<aside className="space-y-6">
						{/* Job Profile — same layout as public leaderboard click (ProfileModal) */}
						<section className={`glass rounded-3xl p-4 sm:p-6 ${glowClass}`}>
							<h3 className="text-xs text-system-gold font-mono font-bold uppercase tracking-widest mb-3">
								Job Profile
							</h3>
							<div className="flex flex-col sm:flex-row items-center gap-3 mb-3">
								<span className="text-2xl">{localUser.job_icon}</span>
								<span
									className={`text-xs px-2.5 py-1 rounded-full border font-mono ${getJobColor(localUser.job_class)}`}
								>
									{localUser.job_name}
								</span>
								<button
									onClick={async () => {
										setJobError(null);
										setShowJobModal(true);
										if (jobs.length === 0) {
											setJobsLoading(true);
											try {
												const list = await repo.listJobs();
												setJobs(list);
											} catch (e: any) {
												setJobError(e?.message || "Gagal memuat jobs");
											} finally {
												setJobsLoading(false);
											}
										}
									}}
									className="ml-auto text-xs px-2 py-1 rounded-lg border border-gray-700 text-gray-400 hover:text-white"
								>
									Ganti
								</button>
							</div>
							<p className="text-sm text-gray-300 leading-relaxed">
								{localUser.job_description}
							</p>
							{localUser.job_trait && (
								<div className="mt-3 p-2.5 rounded-lg bg-gray-950/50 border border-gray-800 text-xs text-gray-400 font-mono">
									<span className="text-system-gold font-bold">Trait:</span>{" "}
									{localUser.job_trait}
								</div>
							)}

						</section>

						<section className={`glass rounded-3xl p-4 sm:p-6 ${glowClass}`}>
							<h3 className="text-lg font-bold font-orbitron text-white flex items-center gap-2 mb-4">
								<Activity className="text-system-blue" size={18} />
								Attributes
							</h3>
							<div className="space-y-4">
								{[
									{
										icon: Swords,
										label: "STR",
										hint: "Strength / Gym",
										value: localUser.str,
										color: "text-system-red",
									},
									{
										icon: Zap,
										label: "STA",
										hint: "Stamina / Run",
										value: localUser.sta,
										color: "text-system-blue",
									},
									{
										icon: Shield,
										label: "AGI",
										hint: "Agility / Sport",
										value: localUser.agi,
										color: "text-system-purple",
									},
									{
										icon: Heart,
										label: "VIT",
										hint: "Vitality / Recovery",
										value: localUser.vit,
										color: "text-system-green",
									},
								].map((stat) => {
									const Icon = stat.icon;
									const width = Math.min(100, Math.max(8, stat.value * 8));
									return (
										<div key={stat.label}>
											<div className="flex items-center justify-between gap-3 mb-1.5">
												<div
													className={`flex items-center gap-2 ${stat.color}`}
												>
													<Icon size={16} />
													<span className="text-xs font-bold font-mono">
														{stat.label}
													</span>
													<span className="text-[10px] text-gray-500 font-mono">
														{stat.hint}
													</span>
												</div>
												<span className="text-sm font-bold font-orbitron text-white">
													{stat.value}
												</span>
											</div>
											<div className="h-2 rounded-full bg-gray-950 border border-gray-800 overflow-hidden">
												<div
													className="h-full rounded-full bg-current opacity-80"
													style={{ width: `${width}%` }}
												/>
											</div>
										</div>
									);
								})}
							</div>
						</section>

						<section className={`glass rounded-3xl p-4 sm:p-6 ${glowClass}`}>
							<h3 className="text-lg font-bold font-orbitron text-white flex items-center gap-2 mb-4">
								<TrendingUp className="text-system-gold" size={18} />
								Rank Baseline
							</h3>
							<div className="rounded-2xl border border-gray-800 bg-gray-950/50 p-4 mb-4">
								<div className="flex items-center justify-between gap-3 mb-2">
									<div>
										<p className="text-[10px] text-gray-500 font-mono uppercase">
											Season Rank
										</p>
										<p className="text-lg font-bold font-orbitron text-white">
											{localUser.rank_icon} {localUser.rank_name}
										</p>
									</div>
									<p className="text-sm font-bold font-orbitron text-system-gold">
										{localUser.seasonal_points} pts
									</p>
								</div>
								<div className="h-3 rounded-full bg-gray-900 border border-gray-800 overflow-hidden p-[2px]">
									<div
										className="h-full rounded-full bg-gradient-to-r from-system-purple to-system-gold"
										style={{ width: `${localUser.season_rank_progress.percent}%` }}
									/>
								</div>
								<p className="mt-2 text-[10px] text-gray-500 font-mono uppercase">
									{localUser.season_rank_progress.is_max
										? "Max rank season ini"
										: `Menuju ${localUser.season_rank_progress.next_icon} ${localUser.season_rank_progress.next_name}: ${localUser.season_rank_progress.remaining} pts lagi`}
								</p>
							</div>
							<div className="grid grid-cols-2 gap-3">
								<StatCard
									label="Goals Done"
									value={localUser.goals_completed}
									tone="text-system-gold"
								/>
								<StatCard
									label="Side Quests"
									value={localUser.total_side_quests}
									tone="text-system-green"
								/>
								<StatCard
									label="Season Days"
									value={localUser.seasonal_activity_count}
									tone="text-system-blue"
								/>
								<StatCard
									label="Lifetime Days"
									value={localUser.total_active_days}
									tone="text-white"
								/>
							</div>
						</section>

						<section className={`glass rounded-3xl p-4 sm:p-6 ${glowClass}`}>
							<h3 className="text-lg font-bold font-orbitron text-white flex items-center gap-2 mb-4">
								<Award className="text-system-gold" size={18} />
								Achievements
							</h3>
							{allBadges.length === 0 ? (
								<p className="text-xs text-gray-600 font-mono italic">
									Belum ada badge. Fokus ke streak, goal, dan side quest dulu.
								</p>
							) : (
								<div className="flex flex-wrap gap-2">
									{localUser.achievements.map((badge) => (
										<span
											key={`life-${badge}`}
											className="inline-flex items-center gap-1 rounded-lg border border-system-gold/30 bg-system-gold/10 px-2.5 py-1 text-[10px] font-bold font-mono text-system-gold"
										>
											<Award size={11} /> {badge}
										</span>
									))}
									{localUser.seasonal_achievements.map((badge) => (
										<span
											key={`season-${badge}`}
											className="inline-flex items-center gap-1 rounded-lg border border-system-purple/30 bg-system-purple/10 px-2.5 py-1 text-[10px] font-bold font-mono text-system-purple"
										>
											<Award size={11} /> {badge}
										</span>
									))}
								</div>
							)}
						</section>
					</aside>
				</div>
			</div>

		{showJobModal && (
			<JobPickerModal
				jobs={jobs}
				loading={jobsLoading}
				error={jobError}
				currentJobId={localUser.job_class}
				selecting={jobLoading}
				onSelect={async (jobId: string) => {
					setJobLoading(true);
					setJobError(null);
					try {
						await repo.selectJob(phone, jobId);
						await refreshUser();
						setShowJobModal(false);
					} catch (e: any) {
						setJobError(e?.message || "Gagal memilih job (mungkin butuh >=50 poin)");
					} finally {
						setJobLoading(false);
					}
				}}
				onClose={() => {
					setShowJobModal(false);
					setJobError(null);
				}}
			/>
		)}
		</div>
	);
}
