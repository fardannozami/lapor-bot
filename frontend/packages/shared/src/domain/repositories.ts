import type { EnrichedReport, GlobalSummary, JobInfo } from '../types';

export interface IReportRepository {
  getLeaderboard(): Promise<EnrichedReport[]>;
  getSummary(): Promise<GlobalSummary>;
  updateName(phone: string, name: string): Promise<{success: boolean; message: string}>;
  selectJob(phone: string, jobId: string): Promise<{success: boolean; message: string}>;
  setGoal(phone: string, targetDays: number, activity: string): Promise<{success: boolean; message: string}>;
  listJobs(): Promise<JobInfo[]>;
  fetchUserByPhone(phone: string): Promise<EnrichedReport>;
}

export interface IAuthRepository {
  login(phone: string): Promise<EnrichedReport>;
}
