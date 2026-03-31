import { AdminAPI } from "./adminAPI.js";
import { AuthAPI } from "./authAPI.js";
import { NexeresClient } from "./client.js";
import { NexeresConfig } from "./config.js";

export class Nexeres {
  auth: AuthAPI;
  admin: AdminAPI;

  constructor(config: NexeresConfig) {
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