import type { IAuthRepository, EnrichedReport } from '@lapor-bot/shared';
import { HttpClient } from '../http/HttpClient';

export class HttpAuthRepository extends HttpClient implements IAuthRepository {
  constructor(baseURL: string = '') {
    super(baseURL);
  }

  async login(phone: string): Promise<EnrichedReport> {
    return this.get<EnrichedReport>(`/api/user?phone=${phone}`);
  }
}
