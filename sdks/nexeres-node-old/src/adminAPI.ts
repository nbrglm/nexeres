import { NexeresAdminResponse, NexeresClient, NexeresResponse } from "./client.js";
import { NewNexeresOrg, NewNexeresOrgDomain, NewNexeresOrgRole, NexeresOrg, NexeresOrgDomain, NexeresOrgRole, NexeresOrgSettings } from "./models/index.js";

export class AdminAPI {
  constructor(private client: NexeresClient) {
  }

  login(params: AdminLoginParams): NexeresResponse<AdminLoginResponse> {
    return this.client.post<AdminLoginResponse>("/api/admin/login", params);
  }

  verifyLogin(params: AdminLoginVerifyParams): NexeresAdminResponse<AdminLoginVerifyResponse> {
    // The empty string is a placeholder for the admin token, which is not needed for this endpoint
    return this.client.adminPost<AdminLoginVerifyResponse>("/api/admin/login/verify", "", params);
  }

  getConfig(adminToken: string): NexeresAdminResponse<NexeresAdminViewableConfig> {
    return this.client.adminGet<NexeresAdminViewableConfig>("/api/admin/config", adminToken);
  }

  async getOrgs(adminToken: string, params: GetOrgsParams): NexeresAdminResponse<NexeresOrganizationsResponse> {
    let result = await this.client.adminPost<NexeresOrganizationsResponse>("/api/admin/orgs", adminToken, params);
    if (result.result) {
      result.result.orgs = result.result.orgs.map(o => NewNexeresOrg(o));
    }
    return result;
  }

  createOrg(adminToken: string, params: CreateOrgParams): NexeresAdminResponse<CreateOrgResponse> {
    return this.client.adminPut<CreateOrgResponse>("/api/admin/orgs", adminToken, params);
  }

  async getOrgDetails(adminToken: string, params: GetOrgDetailsParams): NexeresAdminResponse<GetOrgDetailsResponse> {
    let result = await this.client.adminGet<GetOrgDetailsResponse>(`/api/admin/orgs/${params.orgId}`, adminToken);
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

  isErrorAdminTokenExpired(err: any, code: number): boolean {
    return code === 401 && err && err.message && typeof err.message === "string" && err.message.toLowerCase().includes("admin token expired");
  }
}

/** Parameters for the admin login */
export interface AdminLoginParams {
  /** The admin's email address */
  email: string;
}

/** Response for the admin login */
export interface AdminLoginResponse {
  /** Indicates if the login was successful */
  emailSent: string;

  /** The flow ID for the login verification step */
  flowId: string;
}

/** Parameters for the admin login verification */
export interface AdminLoginVerifyParams {
  /** The flow ID returned from the initial login request */
  flowId: string;
  /** The verification code sent to the admin's email */
  code: string;
}

/** Response for the admin login verification */
export interface AdminLoginVerifyResponse {
  /** Indicates if the verification was successful */
  success: boolean;

  /** The ephemeral token for the admin user, expires in 15 minutes of inactivity */
  token: string;
}

/** Viewable config for the admin */
export interface NexeresAdminViewableConfig {
  /** Indicates if backend is in debug mode */
  debug: boolean;

  /** The public configuration for the backend. Affects the URL generation in backend. */
  publicEndpoint: NexeresAdminPublicConfig;

  /** Indicates if the backend is in multitenancy mode */
  multitenancy: {
    enabled: boolean;
  };

  /**
   * The configuration for JWTs.
   */
  jwt: NexeresAdminJWTConfig;

  /**
   * Configuration for Notifications like Emails, SMS, etc.
   */
  notifications: NexeresAdminNotificationsConfig;

  /**
   * Configuration for Branding in the backend, for emails, sms, etc.
   */
  branding: NexeresAdminBrandingConfig;

  /**
   * Configuration for Security Settings, like API Keys, etc.
   */
  security: NexeresAdminSecurityConfig;
}

/** Public config for the admin */
export interface NexeresAdminPublicConfig {
  /** The scheme (http or https) used by the Nexeres instance */
  scheme: string;

  /** The domain used by the Nexeres UI instance, as configured in the backend. */
  domain: string;

  /** The subdomain used by the Nexeres UI instance, as configured in the backend. */
  subDomain: string;

  /** The debug base URL for the Nexeres instance, as configured in the backend. */
  debugBaseURL: string;
}

/** Configuration for JWTs */
export interface NexeresAdminJWTConfig {
  /** The expiration time for session tokens */
  sessionTokenExpirySeconds: string;

  /** The expiration time for refresh tokens */
  refreshTokenExpirySeconds: string;
}

/** Configuration for Notifications */
export interface NexeresAdminNotificationsConfig {
  /** Configuration for Email notifications */
  email: NexeresAdminEmailConfig;

  /** Configuration for SMS notifications */
  sms?: NexeresAdminSMSConfig;
}

/** Configuration for Email notifications */
export interface NexeresAdminEmailConfig {
  /** The email provider used for sending emails */
  provider: string;

  /**
   * The endpoints (full-URLs) used in emails for various actions like verification, password reset, etc.
   */
  endpoints: {
    verification: string;
    passwordReset: string;
  }
}

/** Configuration for SMS notifications */
export interface NexeresAdminSMSConfig {
  /** The SMS provider used for sending SMS */
  provider: string;
}

/** Configuration for Branding */
export interface NexeresAdminBrandingConfig {
  /** The name of the application, used in emails and other communications */
  appName: string;

  /** The full name of the Company */
  companyName: string;

  /** The short name of the Company */
  companyNameShort: string;

  /**
   * The SupportURL for the Company.
   *
   * Eg. "https://example.com/support" or "mailto:support@example.com"
   */
  supportURL: string;
}

/** Configuration for Security Settings */
export interface NexeresAdminSecurityConfig {
  /** Configuration for Audit Logs */
  auditLogs: {
    /** Indicates if audit logging is enabled */
    enable: boolean;
  }

  /** Configuration for API Keys */
  apiKeys: APIKeyAdminConfig[];

  /** Configuration for Rate Limiting */
  rateLimit: {
    /** The rate limit for API requests
     * 
     * Format: "R-U", where R is the number of requests and U is the time unit (e.g., "100-h" for 100 requests per hour)
     * 
     * Supported time units:
     * - s: Second
     * - m: Minute
     * - h: Hour
     * - d: Day
     */
    rate: string;
  }
}

/** Configuration for an API Key */
export interface APIKeyAdminConfig {
  /** The name of the API Key */
  name: string;

  /** The description of the API Key */
  description: string;
}

export function isValidOrgField(field: string): field is "name" | "slug" | "created_at" {
  return ["name", "slug", "created_at"].includes(field);
}

export function isValidOrgFilterOp(op: string): op is "contains" | "equals" | "lte" | "gte" | "lt" | "gt" {
  return ["contains", "equals", "lte", "gte", "lt", "gt"].includes(op);
}

export function isValidOrgSortDir(dir: string): dir is "ASC" | "DESC" {
  return ["ASC", "DESC"].includes(dir);
}

export function isValidOrgSortField(field: string): field is "name" | "slug" | "created_at" {
  return ["name", "slug", "created_at"].includes(field);
}

/** Parameters for getting organizations */
export interface GetOrgsParams {
  /**
   * Filters for querying organizations
   */
  filters?: {
    /**
     * The filter options
     */
    options: {
      field: "name" | "slug" | "created_at";
      op: "contains" | "equals" | "lte" | "gte" | "lt" | "gt";
      value: string;
    }[];
    /**
     * The logical operator to use for combining filters
     */
    mode: "AND" | "OR";
  }
  /**
   * Pagination settings
   */
  pagination?: {
    page: number;
    pageSize: number;
  }
  sort?: {
    field: "name" | "slug" | "created_at";
    direction: "ASC" | "DESC";
  }[]
};

/** Response for the organizations list */
export interface NexeresOrganizationsResponse {
  /** The list of organizations */
  orgs: NexeresOrg[];

  /** The total number of organizations */
  total: number;

  /**
   * Pagination settings
   */
  pagination: {
    page: number;
    pageSize: number;
  }
}

/** Parameters for creating a new organization */
export interface CreateOrgParams {
  /** The name of the organization */
  name: string;

  /** The slug for the organization, used in URLs */
  slug: string;

  /** The description of the organization */
  description?: string;

  /** The avatar URL for the organization */
  avatarURL?: string;

  settings: NexeresOrgSettings;
}

/** Response for creating a new organization */
export interface CreateOrgResponse {
  orgId: string;
}

/** Get Org Details Params */
export interface GetOrgDetailsParams {
  /** The ID of the organization to retrieve details for */
  orgId: string;
}

/** Get Org Details Response */
export interface GetOrgDetailsResponse {
  /** The organization details */
  org: NexeresOrg;

  /** The roles associated with the organization */
  roles: NexeresOrgRole[];

  /** The domains associated with the organization */
  domains: NexeresOrgDomain[];
}