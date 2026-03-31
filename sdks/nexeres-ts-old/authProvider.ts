import {
  AuthenticationProvider,
  RequestInformation
} from "@microsoft/kiota-abstractions";

export class NexeresAuthProvider implements AuthenticationProvider {
  constructor(
    private readonly apiKey: string,
    private readonly sessionToken?: string | undefined,
    private readonly refreshToken?: string | undefined,
    private readonly adminToken?: string | undefined,
  ) { }

  async authenticateRequest(request: RequestInformation): Promise<void> {
    request.headers.tryAdd("X-NEXERES-API-KEY", this.apiKey);

    if (this.sessionToken) {
      request.headers.tryAdd("X-NEXERES-SESSION-TOKEN", this.sessionToken);
    }

    if (this.refreshToken) {
      request.headers.tryAdd("X-NEXERES-REFRESH-TOKEN", this.refreshToken);
    }

    if (this.adminToken) {
      request.headers.tryAdd("X-NEXERES-ADMIN-TOKEN", this.adminToken);
    }
  }
}