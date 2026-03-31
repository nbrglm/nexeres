import { ApiClient, createApiClient } from "./apiClient.js";
import { FetchRequestAdapter } from "@microsoft/kiota-http-fetchlibrary";
import { NexeresAuthProvider } from "./authProvider.js";
import { AuthSignupRequest } from "./models/index.js";

export type NexeresAuth = {
  apiKey: string,
  sessionToken?: string | undefined,
  refreshToken?: string | undefined,
  adminToken?: string | undefined,
};

export function createNexeresApiClient(
  baseUrl: string,
): NexeresApiClient {
  return new NexeresApiClient(baseUrl);
}

export class NexeresApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private createClient(auth: NexeresAuth): ApiClient {
    const requestAdapter = new FetchRequestAdapter(new NexeresAuthProvider(
      auth.apiKey,
      auth.sessionToken,
      auth.refreshToken,
      auth.adminToken,
    ));
    requestAdapter.baseUrl = this.baseUrl;
    return createApiClient(requestAdapter);
  }

  public signup(params: StrictAuthSignupRequest, auth: Pick<NexeresAuth, "apiKey">) {
    const apiClient = this.createClient(auth);
    return apiClient.api.auth.signup.post(params);
  }
}

// ------------- STRICT TYPES SECTION ---------------------

type RequireFields<T, K extends keyof T> = Required<Pick<T, K>> & Omit<T, K>;

type StrictAuthSignupRequest = RequireFields<AuthSignupRequest, "email" | "password" | "name" | "confirmPassword">;