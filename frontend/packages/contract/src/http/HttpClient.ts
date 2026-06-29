export type GetTokenFn = () => string | null;
export type OnUnauthorizedFn = () => void;

export class HttpClient {
  protected baseURL: string;
  private getToken: GetTokenFn;
  private onUnauthorized?: OnUnauthorizedFn;

  constructor(baseURL: string = '', getToken: GetTokenFn = () => null, onUnauthorized?: OnUnauthorizedFn) {
    this.baseURL = baseURL;
    this.getToken = getToken;
    this.onUnauthorized = onUnauthorized;
  }

  private authHeaders(): Record<string, string> {
    const token = this.getToken();
    if (token) return { Authorization: `Bearer ${token}` };
    return {};
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (response.ok) return response.json() as Promise<T>;

    if (response.status === 401) {
      this.onUnauthorized?.();
    }

    let message: string;
    try {
      const body = await response.json();
      message = body?.error || body?.message || JSON.stringify(body);
    } catch {
      message = response.statusText || `HTTP ${response.status}`;
    }
    throw new Error(message);
  }

  protected async get<T>(path: string): Promise<T> {
    const response = await fetch(`${this.baseURL}${path}`, {
      headers: { ...this.authHeaders() },
    });
    return this.handleResponse<T>(response);
  }

  protected async post<T>(path: string, body: unknown): Promise<T> {
    const response = await fetch(`${this.baseURL}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...this.authHeaders() },
      body: JSON.stringify(body),
    });
    return this.handleResponse<T>(response);
  }

  protected async patch<T>(path: string, body: unknown): Promise<T> {
    const response = await fetch(`${this.baseURL}${path}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json', ...this.authHeaders() },
      body: JSON.stringify(body),
    });
    return this.handleResponse<T>(response);
  }
}
