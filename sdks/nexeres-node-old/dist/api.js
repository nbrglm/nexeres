import { AdminAPI } from "./adminAPI.js";
import { AuthAPI } from "./authAPI.js";
import { NexeresClient } from "./client.js";
export class Nexeres {
    auth;
    admin;
    constructor(config) {
        const client = new NexeresClient(config);
        this.auth = new AuthAPI(client);
        this.admin = new AdminAPI(client);
    }
}
export * from "./client.js";
export * from "./config.js";
export * from "./models/index.js";
export * from "./authAPI.js";
export * from "./adminAPI.js";
//# sourceMappingURL=api.js.map