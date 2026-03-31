import { AdminAPI } from "./adminAPI.js";
import { AuthAPI } from "./authAPI.js";
import { NexeresConfig } from "./config.js";
export declare class Nexeres {
    auth: AuthAPI;
    admin: AdminAPI;
    constructor(config: NexeresConfig);
}
export * from "./client.js";
export * from "./config.js";
export * from "./models/index.js";
export * from "./authAPI.js";
export * from "./adminAPI.js";
//# sourceMappingURL=api.d.ts.map