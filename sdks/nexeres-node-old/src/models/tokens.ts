type TokensJSON = {
  sessionId: string;
  sessionToken: string;
  sessionTokenExpiry: string | Date;
  refreshToken: string;
  refreshTokenExpiry: string | Date;
};

export class Tokens {
  sessionId: string;
  sessionToken: string;
  sessionTokenExpiry: Date;
  refreshToken: string;
  refreshTokenExpiry: Date;
  constructor(data: TokensJSON) {
    this.sessionId = data.sessionId;
    this.sessionToken = data.sessionToken;
    this.sessionTokenExpiry = new Date(data.sessionTokenExpiry);
    this.refreshToken = data.refreshToken;
    this.refreshTokenExpiry = new Date(data.refreshTokenExpiry);
  }
}
