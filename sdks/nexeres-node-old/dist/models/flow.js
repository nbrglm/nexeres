import { NewNexeresOrg } from "./org.js";
export function NewFlow(data) {
    data.orgs = data.orgs.map(o => NewNexeresOrg(o));
    data.createdAt = data.createdAt instanceof Date ? data.createdAt : new Date(data.createdAt);
    data.expiresAt = data.expiresAt instanceof Date ? data.expiresAt : new Date(data.expiresAt);
    return data;
}
//# sourceMappingURL=flow.js.map