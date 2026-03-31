export class Tokens {
    sessionId;
    sessionToken;
    sessionTokenExpiry;
    refreshToken;
    refreshTokenExpiry;
    constructor(data) {
        this.sessionId = data.sessionId;
        this.sessionToken = data.sessionToken;
        this.sessionTokenExpiry = new Date(data.sessionTokenExpiry);
        this.refreshToken = data.refreshToken;
        this.refreshTokenExpiry = new Date(data.refreshTokenExpiry);
    }
}
//# sourceMappingURL=tokens.js.map