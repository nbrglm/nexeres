export interface NexeresOrg {
  /** The unique identifier for the organization */
  id: string;

  /** The unique slug for the organization */
  slug: string;

  /** The name of the organization */
  name: string;

  /** The description of the organization */
  description?: string;

  /** The avatar URL of the organization, if set */
  avatarURL?: string;

  settings: NexeresOrgSettings;

  /** The creation date of the organization in ISO 8601 format */
  createdAt: string | Date;

  /** The last updated date of the organization in ISO 8601 format */
  updatedAt: string | Date;
};

export function NewNexeresOrg(
  data: NexeresOrg
): NexeresOrg {
  data.createdAt = data.createdAt instanceof Date ? data.createdAt : new Date(data.createdAt);
  data.updatedAt = data.updatedAt instanceof Date ? data.updatedAt : new Date(data.updatedAt);
  return data;
};

export interface NexeresOrgSettings {
  mfa: {
    /** Whether MFA is required for users in this organization */
    required: boolean;
    roles: { [key: string]: string }; // role name to role-id mapping
  };
};

export interface NexeresOrgRole {
  id: string;
  roleName: string;
  roleDesc?: string;
  createdAt: string | Date;
  updatedAt: string | Date;
};

export function NewNexeresOrgRole(
  data: NexeresOrgRole
): NexeresOrgRole {
  data.createdAt = data.createdAt instanceof Date ? data.createdAt : new Date(data.createdAt);
  data.updatedAt = data.updatedAt instanceof Date ? data.updatedAt : new Date(data.updatedAt);
  return data;
};

export interface NexeresOrgDomain {
  /** The domain name */
  domain: string;

  /** The ID of the organization this domain belongs to */
  orgId: string;

  /** Whether the domain has been verified */
  verified: boolean;

  /** The date the domain was verified in ISO 8601 format, if verified */
  verifiedAt?: string | Date;

  /** Whether users with this email domain should be automatically joined to the organization */
  autoJoin: boolean;

  /** The role ID to assign to users who join via this domain, if autoJoin is true */
  autoJoinRoleId?: string;

  /** The creation date of the organization domain in ISO 8601 format */
  createdAt: string | Date;

  /** The last updated date of the organization domain in ISO 8601 format */
  updatedAt: string | Date;
}

export function NewNexeresOrgDomain(
  data: NexeresOrgDomain
): NexeresOrgDomain {
  data.createdAt = data.createdAt instanceof Date ? data.createdAt : new Date(data.createdAt);
  data.updatedAt = data.updatedAt instanceof Date ? data.updatedAt : new Date(data.updatedAt);
  data.verifiedAt = data.verifiedAt ? (data.verifiedAt instanceof Date ? data.verifiedAt : new Date(data.verifiedAt)) : undefined;
  return data;
}