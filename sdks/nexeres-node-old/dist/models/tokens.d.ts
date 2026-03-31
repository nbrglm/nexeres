type TokensJSON = {
    sessionId: string;
    sessionToken: string;
    sessionTokenExpiry: string | Date;
    refreshToken: string;
    refreshTokenExpiry: string | Date;
};
export declare class Tokens {
    sessionId: string;
    sessionToken: string;
    sessionTokenExpiry: Date;
    refreshToken: string;
    refreshTokenExpiry: Date;
    constructor(data: TokensJSON);
}
export {};
//# sourceMappingURL=tokens.d.ts.map