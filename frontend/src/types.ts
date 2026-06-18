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
  rank_name: string;
  rank_icon: string;
  season_rank_progress: TierProgress;
  week_active_days: number;
  week_activity: boolean[];
  estimated_weekly_points: number;
  is_active_today: boolean;
}

export interface GlobalSummary {
  total_participants: number;
  active_streak_count: number;
  total_workouts_logged: number;
  active_jobs: Record<string, number>;
  current_season: number;
  current_day: number;
}
