import type {
	IReportRepository,
	EnrichedReport,
	GlobalSummary,
	JobInfo,
	StravaStatus,
} from "@lapor-bot/shared";
import { HttpClient } from "../http/HttpClient";
import type { GetTokenFn, OnUnauthorizedFn } from "../http/HttpClient";

export class HttpReportRepository
	extends HttpClient
	implements IReportRepository
{
	constructor(baseURL: string = "", getToken: GetTokenFn = () => null, onUnauthorized?: OnUnauthorizedFn) {
		super(baseURL, getToken, onUnauthorized);
	}

	async getLeaderboard(): Promise<EnrichedReport[]> {
		return this.get<EnrichedReport[]>("/api/leaderboard");
	}

	async getSummary(): Promise<GlobalSummary> {
		return this.get<GlobalSummary>("/api/summary");
	}

	async updateName(
		name: string,
	): Promise<{ success: boolean; message: string }> {
		return this.patch<{ success: boolean; message: string }>("/api/user/name", {
			name,
		});
	}

	async selectJob(
		jobId: string,
	): Promise<{ success: boolean; message: string }> {
		return this.patch<{ success: boolean; message: string }>("/api/user/job", {
			job_id: jobId,
		});
	}

	async setGoal(
		targetDays: number,
		activity: string,
		start?: { startAt?: string; startDate?: string; startHour?: number },
	): Promise<{ success: boolean; message: string }> {
		const payload: Record<string, unknown> = {
			target_days: String(targetDays),
			activity,
		};
		if (start) {
			if (start.startAt) payload.start_at = start.startAt;
			if (start.startDate) payload.start_date = start.startDate;
			if (typeof start.startHour === "number")
				payload.start_hour = start.startHour;
		}
		return this.patch<{ success: boolean; message: string }>(
			"/api/user/goal",
			payload,
		);
	}

	async resetGoal(): Promise<{ success: boolean; message: string }> {
		return this.patch<{ success: boolean; message: string }>("/api/user/goal", {
			action: "reset",
		});
	}

	async listJobs(): Promise<JobInfo[]> {
		return this.get<JobInfo[]>("/api/jobs");
	}

	async fetchUser(): Promise<EnrichedReport> {
		return this.get<EnrichedReport>("/api/user");
	}

	async getStravaStatus(): Promise<StravaStatus> {
		return this.get<StravaStatus>("/api/user/strava");
	}

	async createStravaLink(): Promise<{ auth_url: string }> {
		return this.post<{ auth_url: string }>("/api/user/strava/link", {});
	}

	async disconnectStrava(): Promise<{ success: boolean; message: string }> {
		return this.delete<{ success: boolean; message: string }>("/api/user/strava");
	}
}
