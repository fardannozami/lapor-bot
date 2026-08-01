export interface NumericLevelProgress {
  Level: number;
  CurrentXP: number;
  RequiredXP: number;
  TotalPoints: number;
}

export interface TierProgress {
  current_min: number;
  next_min: number;
  value: number;
  percent: number;
  remaining: number;
  next_name: string;
  next_icon: string;
  is_max: boolean;
}

export interface EnrichedReport {
  user_id: string;
  name: string;
  job_class: string;
  job_name: string;
  job_icon: string;
  job_description: string;
  job_trait: string;
  streak: number;
  activity_count: number;
  total_active_days: number;
  last_report_date: string;
  max_streak: number;
  total_points: number;
  level: number;
  level_name: string;
  level_icon: string;
  xp_progress: NumericLevelProgress;
  level_tier_progress: TierProgress;
  achievements: string[];
  comeback_streak: number;
  inactive_days: number;
  days_since_last_report: number;
  centurion_cycles: number;
  seasonal_points: number;
  seasonal_activity_count: number;
  seasonal_max_streak: number;
  seasonal_achievements: string[];
  streak_freezes: number;
  goals_completed: number;
  total_side_quests: number;
  seasonal_side_quests: number;
  str: number;
  sta: number;
  agi: number;
  vit: number;
  end: number;
  luk: number;
  rank_name: string;
  rank_icon: string;
  season_rank_progress: TierProgress;
  week_active_days: number;
  week_activity: boolean[];
  estimated_weekly_points: number;
  is_active_today: boolean;
  daily_activity?: DailyActivity[];
  current_daily_streak?: number;
  longest_daily_streak?: number;
  active_days_in_window?: number;
  active_goal?: PersonalGoal;
  today_side_quests?: QuestTask[];
}

export interface DailyActivity {
  date: string;
  count: number;
  active: boolean;
}

export interface GoalDay {
  date: string;
  day_label: string;
  activity: string;
  active: boolean;
}

export interface PersonalGoal {
  target_days: number;
  activity: string;
  start_at: string;
  end_at: string;
  completed_days: number;
  remaining_days: number;
  percent: number;
  is_completed: boolean;
  completed_at?: string;
  days: GoalDay[];
}

/**
 * A structured training program (program latihan) assigned to a user.
 * Backend does not exist yet - this is the forward contract for the
 * upcoming program feature. The UI renders from this shape once wired.
 */
export interface TrainingProgram {
  id: string;
  title: string;
  description?: string;
  /** ISO date the program starts (YYYY-MM-DD). */
  start_date: string;
  /** ISO date the program ends (YYYY-MM-DD). */
  end_date?: string;
  /** Program-level completion 0-100. */
  progress_percent: number;
  /** Total sessions in the program. */
  total_sessions: number;
  /** Sessions the user has completed. */
  completed_sessions: number;
  /** Weekly cadence target (e.g. 3 = 3x per week). */
  sessions_per_week: number;
  /** Optional structured weeks of the program. */
  weeks?: TrainingWeek[];
  /** Marked active when the program is the user's current one. */
  is_active?: boolean;
}

export interface TrainingWeek {
  week_number: number;
  title?: string;
  sessions: TrainingSession[];
}

export interface TrainingSession {
  id: string;
  /** Session title, e.g. "Push Day" or "Lari Interval". */
  title: string;
  /** ISO date the session is scheduled (YYYY-MM-DD). */
  scheduled_date: string;
  /** Session-level completion 0-100. */
  progress_percent: number;
  is_completed: boolean;
  /** Optional exercises / moves inside the session. */
  exercises?: TrainingExercise[];
}

export interface TrainingExercise {
  id: string;
  name: string;
  /** Target value, e.g. "10", "5 km", "30 menit". */
  target: string;
  /** Best-effort result logged by the user, when reported. */
  result?: string;
  is_completed: boolean;
}

export interface QuestTask {
  id: string;
  name: string;
  difficulty: "easy" | "medium" | "hard" | string;
  target: number;
  progress: number;
  unit: string;
  reward_points: number;
}

export interface StravaStatus {
  linked: boolean;
  athlete_id?: number;
  name?: string;
  expires_at?: string;
}

export interface GlobalSummary {
  total_participants: number;
  active_streak_count: number;
  total_workouts_logged: number;
  active_jobs: Record<string, number>;
  current_season: number;
  current_day: number;
}

export interface JobInfo {
  id: string;
  name: string;
  icon: string;
  description: string;
  trait: string;
}

// Mirror of backend LeaderboardSortKey in internal/domain/leaderboard_comparator.go.
// Changing these values breaks the FE↔BE sort contract.
export type LeaderboardTab =
  | "seasonal"
  | "lifetime"
  | "streak"
  | "week"
  | "attributes";

export type AttributeTab =
  | "overall"
  | "str"
  | "sta"
  | "agi"
  | "vit"
  | "end"
  | "luk";

export type StreakTab = "weekly" | "daily";

export type LeaderboardSortKey =
  | "season_rank"
  | "lifetime_xp"
  | "weekly_streak"
  | "daily_streak"
  | "weekly_activity"
  | "attribute_overall"
  | "attribute_str"
  | "attribute_sta"
  | "attribute_agi"
  | "attribute_vit"
  | "attribute_end"
  | "attribute_luk";
