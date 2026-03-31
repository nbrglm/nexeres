import { NexeresOrg } from "./org.js";
export type FlowType = "login" | "change-password" | "sso";
export interface Flow {
    id: string;
    type: FlowType;
    userId: string;
    email: string;
    orgs: Array<NexeresOrg>;
    mfaRequired: boolean;
    mfaVerified: boolean;
    ssoProvider?: string | undefined;
    ssoUserId?: string | undefined;
    returnTo?: string | undefined;
    createdAt: string | Date;
    expiresAt: string | Date;
}
export declare function NewFlow(data: Flow): Flow;
//# sourceMappingURL=flow.d.ts.map