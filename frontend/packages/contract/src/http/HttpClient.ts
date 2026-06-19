export class HttpClient {
  protected baseURL: string;

  constructor(baseURL: string = '') {
    this.baseURL = baseURL;
  }

  protected async get<T>(path: string): Promise<T> {
    const response = await fetch(`${this.baseURL}${path}`);
    if (!response.ok) {
      if (response.status === 404) {
        throw new Error('Not found');
      }
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    return response.json() as Promise<T>;
  }

  // Add POST, PUT, DELETE as needed in the future
}
