import type { EnrichedReport, GlobalSummary, JobInfo } from "../types";

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
	resetGoal?(): Promise<{ success: boolean; message: string }>;
}

export interface IAuthRepository {
	login(phone: string): Promise<LoginResult>;
}
