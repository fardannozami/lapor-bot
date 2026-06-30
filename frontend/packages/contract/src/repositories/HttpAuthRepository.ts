import type { IAuthRepository, LoginResult, EnrichedReport } from '@lapor-bot/shared';
import { HttpClient } from '../http/HttpClient';
import type { GetTokenFn, OnUnauthorizedFn } from '../http/HttpClient';

export class HttpAuthRepository extends HttpClient implements IAuthRepository {
  constructor(baseURL: string = '', getToken: GetTokenFn = () => null, onUnauthorized?: OnUnauthorizedFn) {
    super(baseURL, getToken, onUnauthorized);
  }

  async login(phone: string): Promise<LoginResult> {
    const res = await this.post<{ token: string; expires_at: string; user: { phone: string; name: string } }>(
      '/api/auth/login',
      { phone },
    );

    // Fetch full profile using the new token
    const profile = await this.getWithToken<EnrichedReport>('/api/user', res.token);

    return { ...res, profile };
  }

  private async getWithToken<T>(path: string, token: string): Promise<T> {
    const response = await fetch(`${this.baseURL}${path}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!response.ok) {
      let message: string;
      try {
        const body = await response.json();
        message = body?.error || body?.message || JSON.stringify(body);
      } catch {
        message = response.statusText || `HTTP ${response.status}`;
      }
      throw new Error(message);
    }
    return response.json() as Promise<T>;
  }
}
