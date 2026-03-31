import { AuthSendVerificationEmailRequest, AuthSendVerificationEmailResponse, AuthSignupRequest, AuthSignupResponse, ErrorResponse, NetworkOrParseError } from "./models.js";

export * from "./models.js";

export type NexeresAuth = {
  "X-NEXERES-SESSION-TOKEN"?: string | undefined,
  "X-NEXERES-REFRESH-TOKEN"?: string | undefined,
  "X-NEXERES-ADMIN-TOKEN"?: string | undefined,
};

type ApiResponse<T> = {
  result?: T | null;
  error?: ErrorResponse | NetworkOrParseError | null;
}

var c = null as NexeresApiClient | null;

export function client(baseUrl: string, apiKey: string): NexeresApiClient {
  if (!c) {
    c = new NexeresApiClient(baseUrl, apiKey);
  }
  return c;
}

export class NexeresApiClient {
  private baseUrl: string;
  private apiKey: string;

  constructor(baseUrl: string, apiKey: string) {
    this.baseUrl = baseUrl;
    this.apiKey = apiKey;
  }

  private async request<T>(params: any, path: string, method: string, auth: Partial<NexeresAuth>): Promise<ApiResponse<T>> {
    const headers = {
      "Content-Type": "application/json",
      "X-NEXERES-API-KEY": this.apiKey,
      ...auth,
    };
    if (auth) {
      if (!auth["X-NEXERES-SESSION-TOKEN"]) {
        delete headers["X-NEXERES-SESSION-TOKEN"];
      }

      if (!auth["X-NEXERES-REFRESH-TOKEN"]) {
        delete headers["X-NEXERES-REFRESH-TOKEN"];
      }

      if (!auth["X-NEXERES-ADMIN-TOKEN"]) {
        delete headers["X-NEXERES-ADMIN-TOKEN"];
      }
    }
    try {
      const response = await fetch(this.baseUrl + path, {
        method,
        headers,
        body: JSON.stringify(params),
      });

      const bodyJSON = await this.tryParseJSON(response);

      if (!response.ok) {
        const errorJson = bodyJSON as ErrorResponse | null;
        if (errorJson) {
          if (errorJson.message) {
            return {
              error: {
                message: errorJson.message,
                code: response.status,
                debug: errorJson.debug,
              }
            };
          } else {
            return {
              error: {
                errType: "UnknownError"
              }
            };
          }
        }

        return {
          error: {
            errType: "ParseError",
          }
        };
      }

      const json = bodyJSON as T | null;
      if (json) {
        return {
          result: json,
        };
      } else {
        return {
          error: {
            message: 'An error occurred while processing the request.',
            code: 0,
            debug: "Invalid response!",
          }
        };
      }
    } catch (e) {
      console.log(JSON.stringify(e));
      // return {
      //   error: {
      //     errType: "NetworkError"
      //   }
      // };
      return {
        error: {
          message: JSON.stringify(e),
          code: 0,
        }
      }
    }
  }

  /** 
   * Tries to parse the JSON response from the server.
   * @param response The response object from the fetch call.
   * @returns The parsed JSON object or null if parsing fails.
   */
  private async tryParseJSON(response: Response): Promise<any> {
    const text = await response.text();
    try {
      return JSON.parse(text);
    } catch (error) {
      return null;
    }
  }

  public signup(params: AuthSignupRequest) {
    return this.request<AuthSignupResponse>(params, "/api/v1/auth/signup", "POST", {})
  }

  public sendVerificationEmail(params: AuthSendVerificationEmailRequest) {
    return this.request<AuthSendVerificationEmailResponse>(params, "/api/v1/auth/send-verification-email", "POST", {})
  }
}