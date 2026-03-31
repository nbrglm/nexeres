import { NewNexeresOrg, NexeresOrg } from "./org.js";

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

export function NewFlow(data: Flow): Flow {
  data.orgs = data.orgs.map(o => NewNexeresOrg(o));
  data.createdAt = data.createdAt instanceof Date ? data.createdAt : new Date(data.createdAt);
  data.expiresAt = data.expiresAt instanceof Date ? data.expiresAt : new Date(data.expiresAt);
  return data;
}