import { NexeresConfig } from "./config.js";

export class NexeresClient {
  constructor(
    private config: NexeresConfig
  ) {
  }

  /** Make a request to the Nexeres Admin API
   * @param endpoint The API endpoint to call, e.g. "/api/admin/login"
   * @param adminToken The admin's ephemeral token for authentication
   * @param options Fetch options, e.g. method, headers, body, etc.
   * @returns A promise that resolves to the response data or an error object.
  */
  private async adminRequest<T>(endpoint: string, adminToken: string, options: RequestInit = {}): NexeresAdminResponse<T> {
    const res = await fetch(`${this.config.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        'X-NEXERES-API-Key': this.config.apiKey,
        'X-NEXERES-Admin-Token': adminToken,
        ...(options.headers || {}),
      }
    });
    const adminTokenExpiry = res.headers.get('X-NEXERES-Admin-Token-Expiry') || (new Date()).toString();

    if (!res.ok) {
      const error: NexeresError = await res.json();
      return { error, code: res.status, adminTokenExpiry };
    }
    return { result: await res.json() as T, code: res.status, adminTokenExpiry };
  }

  adminGet<T>(endpoint: string, adminToken: string, options: RequestInit = {}): NexeresAdminResponse<T> {
    return this.adminRequest<T>(endpoint, adminToken, { method: 'GET', ...options });
  }

  adminPost<T>(endpoint: string, adminToken: string, body: any, options: RequestInit = {}): NexeresAdminResponse<T> {
    return this.adminRequest<T>(endpoint, adminToken, { method: 'POST', body: JSON.stringify(body), ...options });
  }

  adminPut<T>(endpoint: string, adminToken: string, body: any, options: RequestInit = {}): NexeresAdminResponse<T> {
    return this.adminRequest<T>(endpoint, adminToken, { method: 'PUT', body: JSON.stringify(body), ...options });
  }

  adminDelete<T>(endpoint: string, adminToken: string, options: RequestInit = {}): NexeresAdminResponse<T> {
    return this.adminRequest<T>(endpoint, adminToken, { method: 'DELETE', ...options });
  }

  /** Make a request to the Nexeres API.
   * @param endpoint The API endpoint to call, e.g. "/auth/login"
   * @param options Fetch options, e.g. method, headers, body, etc.
   * @returns A promise that resolves to the response data or an error object.
   */
  private async request<T>(endpoint: string, options: RequestInit = {}): NexeresResponse<T> {
    const res = await fetch(`${this.config.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        'X-NEXERES-API-Key': this.config.apiKey,
        ...(options.headers || {}),
      }
    });

    if (!res.ok) {
      const error: NexeresError = await res.json();
      return { error, code: res.status };
    }

    return { result: await res.json() as T, code: res.status };
  }

  get<T>(endpoint: string, options: RequestInit = {}): NexeresResponse<T> {
    return this.request<T>(endpoint, { method: 'GET', ...options });
  }

  post<T>(endpoint: string, body: any, options: RequestInit = {}): NexeresResponse<T> {
    return this.request<T>(endpoint, { method: 'POST', body: JSON.stringify(body), ...options });
  }

  put<T>(endpoint: string, body: any, options: RequestInit = {}): NexeresResponse<T> {
    return this.request<T>(endpoint, { method: 'PUT', body: JSON.stringify(body), ...options });
  }

  delete<T>(endpoint: string, options: RequestInit = {}): NexeresResponse<T> {
    return this.request<T>(endpoint, { method: 'DELETE', ...options });
  }
}

/** Response for general API calls */
export type NexeresResponse<T> = Promise<{ code: number, result?: T, error?: NexeresError }>;

/** Response for admin-related API calls */
export type NexeresAdminResponse<T> = Promise<{
  /** HTTP Status Code returned by the server: Non 200 code represents error, while a 200 code represents a successfull result */
  code: number, result?: T, error?: NexeresError,
  /** The expiry, in seconds from now, of the admin token. */
  adminTokenExpiry: string
}>;

/** Error response from the Nexeres API */
export type NexeresError = {
  /** Error message 
   * 
   * You can safely display this message to end users.
  */
  message: string;

  /** Debug information
   * 
   * This is meant for debugging purposes only and should not be displayed to end users.
   * 
   * This won't be present unless the `debug` flag is enabled in your Nexeres backend config.
  */
  debug?: string | undefined;
}