import { Tokens } from "./models/tokens.js";
export class AuthAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async login(params) {
        const result = await this.client.post("/api/auth/login", params);
        if (result.result?.tokens) {
            // Convert tokens to class instances
            result.result.tokens = new Tokens(result.result.tokens);
        }
        return result;
    }
    signup(params) {
        return this.client.post("/api/auth/signup", params);
    }
    sendVerificationEmail(params) {
        return this.client.post("/api/auth/verify-email/send", params);
    }
    verifyEmail(params) {
        return this.client.post("/api/auth/verify-email/verify", params);
    }
    async refreshToken(params) {
        const result = await this.client.post("/api/auth/refresh", {
            userIp: params.userIp,
            userAgent: params.userAgent
        }, {
            headers: {
                'X-NEXERES-Refresh-Token': params.refreshToken
            }
        });
        if (result.result?.tokens) {
            // Convert tokens to class instances
            result.result.tokens = new Tokens(result.result.tokens);
        }
        return result;
    }
    logout(params) {
        return this.client.post("/api/auth/logout", null, {
            headers: {
                'X-NEXERES-Refresh-Token': params.refreshToken ?? '',
                'X-NEXERES-Session-Token': params.sessionToken ?? '',
            }
        });
    }
    async getFlowData(params) {
        if (!params.flowId || params.flowId.trim() === '') {
            return {
                code: 400,
                error: { message: 'Invalid flow ID!', debug: 'Flow ID is empty or not provided' },
            };
        }
        return this.client.get(`/api/auth/flow/${params.flowId}`);
    }
}
//# sourceMappingURL=authAPI.js.map