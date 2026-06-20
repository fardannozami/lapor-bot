import type {
	IReportRepository,
	EnrichedReport,
	GlobalSummary,
	JobInfo,
} from "@lapor-bot/shared";
import { HttpClient } from "../http/HttpClient";

export class HttpReportRepository
	extends HttpClient
	implements IReportRepository
{
	constructor(baseURL: string = "") {
		super(baseURL);
	}

	async getLeaderboard(): Promise<EnrichedReport[]> {
		return this.get<EnrichedReport[]>("/api/leaderboard");
	}

	async getSummary(): Promise<GlobalSummary> {
		return this.get<GlobalSummary>("/api/summary");
	}

	async updateName(
		phone: string,
		name: string,
	): Promise<{ success: boolean; message: string }> {
		return this.post<{ success: boolean; message: string }>("/api/user/name", {
			phone,
			name,
		});
	}

	async selectJob(
		phone: string,
		jobId: string,
	): Promise<{ success: boolean; message: string }> {
		return this.post<{ success: boolean; message: string }>("/api/user/job", {
			phone,
			jobId,
		});
	}

	async setGoal(
		phone: string,
		targetDays: number,
		activity: string,
		start?: { startAt?: string; startDate?: string; startHour?: number },
	): Promise<{ success: boolean; message: string }> {
		const payload: any = { phone, target_days: String(targetDays), activity };
		if (start) {
			if (start.startAt) payload.start_at = start.startAt;
			if (start.startDate) payload.start_date = start.startDate;
			if (typeof start.startHour === "number")
				payload.start_hour = start.startHour;
		}
		return this.post<{ success: boolean; message: string }>(
			"/api/user/goal",
			payload,
		);
	}

	async resetGoal(
		phone: string,
	): Promise<{ success: boolean; message: string }> {
		return this.post<{ success: boolean; message: string }>("/api/user/goal", {
			phone,
			action: "reset",
		});
	}

	async listJobs(): Promise<JobInfo[]> {
		return this.get<JobInfo[]>("/api/jobs");
	}

	async fetchUserByPhone(phone: string): Promise<EnrichedReport> {
		return this.post<EnrichedReport>("/api/user", { phone });
	}
}
