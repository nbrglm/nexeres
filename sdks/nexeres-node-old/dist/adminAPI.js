import { NewNexeresOrg, NewNexeresOrgDomain, NewNexeresOrgRole } from "./models/index.js";
export class AdminAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    login(params) {
        return this.client.post("/api/admin/login", params);
    }
    verifyLogin(params) {
        // The empty string is a placeholder for the admin token, which is not needed for this endpoint
        return this.client.adminPost("/api/admin/login/verify", "", params);
    }
    getConfig(adminToken) {
        return this.client.adminGet("/api/admin/config", adminToken);
    }
    async getOrgs(adminToken, params) {
        let result = await this.client.adminPost("/api/admin/orgs", adminToken, params);
        if (result.result) {
            result.result.orgs = result.result.orgs.map(o => NewNexeresOrg(o));
        }
        return result;
    }
    createOrg(adminToken, params) {
        return this.client.adminPut("/api/admin/orgs", adminToken, params);
    }
    async getOrgDetails(adminToken, params) {
        let result = await this.client.adminGet(`/api/admin/orgs/${params.orgId}`, adminToken);
        if (result.result) {
            if (result.result.org) {
                result.result.org = NewNexeresOrg(result.result.org);
            }
            if (result.result.roles) {
                result.result.roles = result.result.roles.map(role => NewNexeresOrgRole(role));
            }
            if (result.result.domains) {
                result.result.domains = result.result.domains.map(domain => NewNexeresOrgDomain(domain));
            }
        }
        return result;
    }
    isErrorAdminTokenExpired(err, code) {
        return code === 401 && err && err.message && typeof err.message === "string" && err.message.toLowerCase().includes("admin token expired");
    }
}
export function isValidOrgField(field) {
    return ["name", "slug", "created_at"].includes(field);
}
export function isValidOrgFilterOp(op) {
    return ["contains", "equals", "lte", "gte", "lt", "gt"].includes(op);
}
export function isValidOrgSortDir(dir) {
    return ["ASC", "DESC"].includes(dir);
}
export function isValidOrgSortField(field) {
    return ["name", "slug", "created_at"].includes(field);
}
;
//# sourceMappingURL=adminAPI.js.map