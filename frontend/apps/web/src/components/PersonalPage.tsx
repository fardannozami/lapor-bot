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
} from "lucide-react";
import type {
  DailyActivity,
  EnrichedReport,
  QuestTask,
} from "@lapor-bot/shared";

interface PersonalPageProps {
  user: EnrichedReport;
  onLogout: () => void;
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

export function PersonalPage({ user, onLogout }: PersonalPageProps) {
  const glowClass = getRankGlow(user.rank_name);
  const sideQuests = user.today_side_quests ?? [];
  const dailyActivity = user.daily_activity ?? [];
  const activeGoal = user.active_goal;
  const allBadges = [...user.achievements, ...user.seasonal_achievements];
  const xpPercent = Math.min(
    100,
    Math.round(
      (user.xp_progress.CurrentXP / user.xp_progress.RequiredXP) * 100,
    ),
  );
  const completedSideQuests = sideQuests.filter(
    (quest) => quest.progress >= quest.target,
  ).length;

  return (
    <div className="min-h-[60vh] px-4 py-8">
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
          className={`relative overflow-hidden glass rounded-[2rem] p-6 md:p-8 mb-6 ${glowClass}`}
        >
          <div className="absolute right-0 top-0 h-36 w-36 rounded-full bg-system-green/10 blur-3xl" />
          <div className="relative grid gap-6 lg:grid-cols-[1.25fr_0.75fr]">
            <div className="flex items-start gap-4">
              <div className="w-16 h-16 rounded-3xl bg-gray-950 border border-gray-800 flex items-center justify-center text-3xl shrink-0 shadow-neon-purple">
                {user.job_icon}
              </div>
              <div className="min-w-0">
                <p className="text-[10px] text-system-green font-mono uppercase tracking-[0.3em]">
                  Personal Hunter Profile
                </p>
                <h2 className="mt-2 text-3xl md:text-4xl font-black font-orbitron text-white tracking-wide truncate">
                  {user.name}
                </h2>
                <p className="text-sm text-gray-400 font-mono mt-1">
                  {user.job_name} {user.level_icon} Lv.{user.level} •{" "}
                  {user.rank_icon} {user.rank_name}
                </p>
                <p className="text-[10px] text-gray-600 font-mono mt-1">
                  {user.user_id}
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
                    {user.level_icon} {user.level_name}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-xl font-black font-orbitron text-system-gold">
                    {user.total_points}
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
                <span>{user.xp_progress.CurrentXP} XP</span>
                <span className="text-system-blue font-bold">{xpPercent}%</span>
                <span>{user.xp_progress.RequiredXP} XP</span>
              </div>
            </div>
          </div>
        </section>

        <section className="grid grid-cols-2 lg:grid-cols-4 gap-3 mb-6">
          <StatCard
            label="Season Points"
            value={user.seasonal_points}
            tone="text-system-gold"
          />
          <StatCard
            label="Daily Streak"
            value={`${user.current_daily_streak ?? 0} hari`}
            tone="text-system-red"
          />
          <StatCard
            label="Weekly Streak"
            value={`${user.streak} minggu`}
            tone="text-system-green"
          />
          <StatCard
            label="Active Window"
            value={`${user.active_days_in_window ?? 0}/${dailyActivity.length || 35}`}
            tone="text-system-blue"
          />
        </section>

        <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <div className="space-y-6">
            <section className={`glass rounded-3xl p-6 ${glowClass}`}>
              <div className="flex items-start justify-between gap-4 mb-5">
                <div>
                  <h3 className="text-lg font-bold font-orbitron text-white flex items-center gap-2">
                    <Flame className="text-system-red" size={18} />
                    Daily Streak Map
                  </h3>
                  <p className="text-xs text-gray-500 font-mono mt-1 leading-relaxed">
                    Mini GitHub-style heatmap: baseline harian pribadi, terpisah
                    dari leaderboard mingguan.
                  </p>
                </div>
                <div className="text-right shrink-0">
                  <div className="text-sm font-bold font-orbitron text-white">
                    {user.longest_daily_streak ?? 0} hari
                  </div>
                  <div className="text-[10px] text-gray-500 font-mono uppercase">
                    Best Daily
                  </div>
                </div>
              </div>

              <div className="overflow-x-auto pb-1">
                <div
                  className="flex gap-1 min-w-fit"
                  aria-label="Kalender aktivitas harian"
                >
                  {chunkWeeks(dailyActivity).map((week, weekIdx) => (
                    <div
                      key={`week-${weekIdx}`}
                      className="grid grid-rows-7 gap-1"
                    >
                      {week.map((day) => (
                        <span
                          key={day.date}
                          className={`h-4 w-4 rounded-[4px] border transition-transform hover:scale-125 ${
                            day.active
                              ? "bg-system-green border-system-green/70 shadow-neon-purple"
                              : "bg-gray-950 border-gray-800"
                          }`}
                          title={`${formatDate(day.date)} — ${day.active ? "aktif" : "belum aktif"}`}
                        />
                      ))}
                    </div>
                  ))}
                </div>
              </div>
              <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-[10px] text-gray-500 font-mono uppercase tracking-wider">
                <span>
                  {dailyActivity[0] ? formatDate(dailyActivity[0].date) : "—"} →{" "}
                  {dailyActivity.at(-1)
                    ? formatDate(dailyActivity.at(-1)!.date)
                    : "—"}
                </span>
                <span className="flex items-center gap-2">
                  <span className="h-3 w-3 rounded bg-gray-950 border border-gray-800" />{" "}
                  Rest <span className="h-3 w-3 rounded bg-system-green" />{" "}
                  Active
                </span>
              </div>
            </section>

            <section className={`glass rounded-3xl p-6 ${glowClass}`}>
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
                  <div className="grid grid-cols-7 gap-2">
                    {activeGoal.days.map((day) => (
                      <div
                        key={day.date}
                        className={`rounded-xl border p-2 text-center ${day.active ? "border-system-green/50 bg-system-green/10" : "border-gray-800 bg-gray-950/50"}`}
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
            </section>

            <section className={`glass rounded-3xl p-6 ${glowClass}`}>
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

          <aside className="space-y-6">
            <section className={`glass rounded-3xl p-6 ${glowClass}`}>
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
                    value: user.str,
                    color: "text-system-red",
                  },
                  {
                    icon: Zap,
                    label: "STA",
                    hint: "Stamina / Run",
                    value: user.sta,
                    color: "text-system-blue",
                  },
                  {
                    icon: Shield,
                    label: "AGI",
                    hint: "Agility / Sport",
                    value: user.agi,
                    color: "text-system-purple",
                  },
                  {
                    icon: Heart,
                    label: "VIT",
                    hint: "Vitality / Recovery",
                    value: user.vit,
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

            <section className={`glass rounded-3xl p-6 ${glowClass}`}>
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
                      {user.rank_icon} {user.rank_name}
                    </p>
                  </div>
                  <p className="text-sm font-bold font-orbitron text-system-gold">
                    {user.seasonal_points} pts
                  </p>
                </div>
                <div className="h-3 rounded-full bg-gray-900 border border-gray-800 overflow-hidden p-[2px]">
                  <div
                    className="h-full rounded-full bg-gradient-to-r from-system-purple to-system-gold"
                    style={{ width: `${user.season_rank_progress.percent}%` }}
                  />
                </div>
                <p className="mt-2 text-[10px] text-gray-500 font-mono uppercase">
                  {user.season_rank_progress.is_max
                    ? "Max rank season ini"
                    : `Menuju ${user.season_rank_progress.next_icon} ${user.season_rank_progress.next_name}: ${user.season_rank_progress.remaining} pts lagi`}
                </p>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <StatCard
                  label="Goals Done"
                  value={user.goals_completed}
                  tone="text-system-gold"
                />
                <StatCard
                  label="Side Quests"
                  value={user.total_side_quests}
                  tone="text-system-green"
                />
                <StatCard
                  label="Season Days"
                  value={user.seasonal_activity_count}
                  tone="text-system-blue"
                />
                <StatCard
                  label="Lifetime Days"
                  value={user.activity_count}
                  tone="text-white"
                />
              </div>
            </section>

            <section className={`glass rounded-3xl p-6 ${glowClass}`}>
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
                  {user.achievements.map((badge) => (
                    <span
                      key={`life-${badge}`}
                      className="inline-flex items-center gap-1 rounded-lg border border-system-gold/30 bg-system-gold/10 px-2.5 py-1 text-[10px] font-bold font-mono text-system-gold"
                    >
                      <Award size={11} /> {badge}
                    </span>
                  ))}
                  {user.seasonal_achievements.map((badge) => (
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
    </div>
  );
}
