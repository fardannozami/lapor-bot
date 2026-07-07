import type {
  EnrichedReport,
  GlobalSummary,
  JobInfo,
  StravaStatus,
} from "../types";

export interface ReportResult {
  success: boolean;
  message: string;
}

export interface LoginResult {
  token: string;
  expires_at: string;
  user: { phone: string; name: string };
  profile: EnrichedReport;
}

export interface IReportRepository {
  getLeaderboard(): Promise<EnrichedReport[]>;
  getSummary(): Promise<GlobalSummary>;
  updateName(name: string): Promise<{ success: boolean; message: string }>;
  selectJob(jobId: string): Promise<{ success: boolean; message: string }>;
  setGoal(
    targetDays: number,
    activity: string,
    start?: { startAt?: string; startDate?: string; startHour?: number },
  ): Promise<{ success: boolean; message: string }>;
  listJobs(): Promise<JobInfo[]>;
  fetchUser(): Promise<EnrichedReport>;
  getStravaStatus(): Promise<StravaStatus>;
  createStravaLink(): Promise<{ auth_url: string }>;
  disconnectStrava(): Promise<{ success: boolean; message: string }>;
  resetGoal?(): Promise<{ success: boolean; message: string }>;
  createReport(message: string): Promise<ReportResult>;
  createSidequestReport(activities: string[]): Promise<ReportResult>;
}

export interface IAuthRepository {
  login(phone: string): Promise<LoginResult>;
}
