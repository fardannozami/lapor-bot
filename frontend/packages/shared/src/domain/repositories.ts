import type { EnrichedReport, GlobalSummary } from '../types';

export interface IReportRepository {
  getLeaderboard(): Promise<EnrichedReport[]>;
  getSummary(): Promise<GlobalSummary>;
}

export interface IAuthRepository {
  login(phone: string): Promise<EnrichedReport>;
}
