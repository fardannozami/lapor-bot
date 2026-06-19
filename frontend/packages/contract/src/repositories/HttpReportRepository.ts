import type { IReportRepository, EnrichedReport, GlobalSummary } from '@lapor-bot/shared';
import { HttpClient } from '../http/HttpClient';

export class HttpReportRepository extends HttpClient implements IReportRepository {
  constructor(baseURL: string = '') {
    super(baseURL);
  }

  async getLeaderboard(): Promise<EnrichedReport[]> {
    return this.get<EnrichedReport[]>('/api/leaderboard');
  }

  async getSummary(): Promise<GlobalSummary> {
    return this.get<GlobalSummary>('/api/summary');
  }
}
