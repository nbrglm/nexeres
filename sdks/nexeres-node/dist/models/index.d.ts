import { type PublicFacingBrandingConfig, type PublicFacingJWTConfig, type PublicFacingMultitenancyConfig, type PublicFacingNotificationsConfig, type PublicFacingPublicEndpointConfig, type PublicFacingSecurityConfig } from './systemConfigPublicFacing/index.js';
import { type AdditionalDataHolder, type ApiError, type Guid, type Parsable, type ParseNode, type SerializationWriter } from '@microsoft/kiota-abstractions';
export interface AdminCreateDomainRequest extends Parsable {
    /**
     * Whether users from this domain can auto-join the organization.
     */
    autoJoin?: boolean | null;
    /**
     * Role ID to assign to users who auto-join via this domain.
     */
    autoJoinRoleId?: Guid | null;
    /**
     * Role name corresponding to the Role ID.
     */
    autoJoinRoleName?: string | null;
    /**
     * Domain name to be added.
     */
    domain?: string | null;
}
export interface AdminCreateDomainResponse extends Parsable {
    /**
     * The domain property
     */
    domain?: Domain | null;
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface AdminCreateRoleRequest extends Parsable {
    /**
     * List of permissions to assign to the role.
     */
    permissions?: string[] | null;
    /**
     * Description of the role.
     */
    roleDesc?: string | null;
    /**
     * Name of the role.
     */
    roleName?: string | null;
}
export interface AdminCreateRoleResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * The role property
     */
    role?: Role | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface AdminDeleteDomainResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface AdminDeleteRoleResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface AdminGetDomainVerificationCodeResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
    /**
     * Domain verification code.
     */
    verificationCode?: string | null;
}
export interface AdminUpdateDomainRequest extends Parsable {
    /**
     * Whether users from this domain can auto-join the organization.
     */
    autoJoin?: boolean | null;
    /**
     * Role ID to assign to users who auto-join via this domain.
     */
    autoJoinRoleId?: Guid | null;
    /**
     * Role name corresponding to the Role ID.
     */
    autoJoinRoleName?: string | null;
}
export interface AdminUpdateDomainResponse extends Parsable {
    /**
     * The domain property
     */
    domain?: Domain | null;
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface AdminUpdateOrgRequest extends Parsable {
    /**
     * URL of the organization's avatar image. If you wish to upload a new image, use the dedicated avatar upload endpoint. It will automatically update the avatarURL.
     */
    avatarURL?: string | null;
    /**
     * New description for the organization.
     */
    description?: string | null;
    /**
     * New name for the organization.
     */
    name?: string | null;
    /**
     * Organization settings schema.
     */
    settings?: OrgSettings | null;
}
export interface AdminUpdateOrgResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface AdminUpdateRoleRequest extends Parsable {
    /**
     * List of permissions to assign to the role.
     */
    permissions?: string[] | null;
    /**
     * Description of the role.
     */
    roleDesc?: string | null;
    /**
     * New name for the role.
     */
    roleName?: string | null;
}
export interface AdminUpdateRoleResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface AdminVerifyDomainResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
    /**
     * Indicates if the domain is verified.
     */
    verified?: boolean | null;
}
export interface AuthFlow extends Parsable {
    /**
     * Timestamp when the flow was created.
     */
    createdAt?: Date | null;
    /**
     * Email associated with the authentication flow.
     */
    email?: string | null;
    /**
     * Timestamp when the flow expires.
     */
    expiresAt?: Date | null;
    /**
     * Unique identifier for the authentication flow.
     */
    id?: Guid | null;
    /**
     * Indicates if MFA is required for this flow.
     */
    mfaRequired?: boolean | null;
    /**
     * Indicates if MFA has been verified in this flow.
     */
    mfaVerified?: boolean | null;
    /**
     * List of organization IDs associated with the flow, if applicable. Mostly used to handle multi-tenant login.
     */
    orgs?: Org[] | null;
    /**
     * URL to redirect after flow completion.
     */
    returnTo?: string | null;
    /**
     * Type of authentication flow (e.g., login, change-password).
     */
    type?: AuthFlow_type | null;
    /**
     * User ID associated with the flow, if applicable.
     */
    userId?: Guid | null;
}
export type AuthFlow_type = (typeof AuthFlow_typeObject)[keyof typeof AuthFlow_typeObject];
export interface AuthGetFlowResponse extends Parsable {
    /**
     * The flow property
     */
    flow?: AuthFlow | null;
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface AuthLoginRequest extends Parsable {
    /**
     * User's email address.
     */
    email?: string | null;
    /**
     * Optional field to store in the flow data which can be fetched by the client after loginThis can be used to redirect the user to a specific page after loginor to maintain the state of the application.It is recommended to validate this field on the client side to prevent open redirect vulnerabilities.
     */
    flowReturnTo?: string | null;
    /**
     * User's password.
     */
    password?: string | null;
    /**
     * User agent string of the client making the request.
     */
    userAgent?: string | null;
    /**
     * The userIp property
     */
    userIp?: string | null;
}
export type AuthLoginRequest_userIp = string;
export interface AuthLoginResponse extends Parsable {
    /**
     * ID of the initiated authentication flow. Tokens are not provided if this is present.
     */
    flowId?: Guid | null;
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if email verification is required. Tokens are not provided if this is true.
     */
    requireEmailVerification?: boolean | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
    /**
     * The tokens property
     */
    tokens?: AuthTokens | null;
}
export interface AuthLogoutRequest extends Parsable {
    /**
     * Refresh token of the user to be logged out.
     */
    refreshToken?: string | null;
    /**
     * Session token of the user to be logged out.
     */
    sessionToken?: string | null;
}
export interface AuthLogoutResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface AuthRefreshRequest extends Parsable {
    /**
     * Refresh token used to obtain new authentication tokens.
     */
    refreshToken?: string | null;
    /**
     * User agent string of the client making the request.
     */
    userAgent?: string | null;
    /**
     * The userIp property
     */
    userIp?: string | null;
}
export type AuthRefreshRequest_userIp = string;
export interface AuthRefreshResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
    /**
     * The tokens property
     */
    tokens?: AuthTokens | null;
}
export interface AuthSendVerificationEmailRequest extends Parsable {
    /**
     * User's email address to send the verification email to.
     */
    email?: string | null;
}
export interface AuthSendVerificationEmailResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface AuthSignupRequest extends Parsable {
    /**
     * Confirmation of the user's desired password.
     */
    confirmPassword?: string | null;
    /**
     * User's email address.
     */
    email?: string | null;
    /**
     * Invitation token for joining an organization (if applicable).
     */
    inviteToken?: string | null;
    /**
     * User's full name.
     */
    name?: string | null;
    /**
     * User's desired password.
     */
    password?: string | null;
}
export interface AuthSignupResponse extends Parsable {
    /**
     * List of backup codes for account recovery. Only provided if the organization has 2FA enabled.
     */
    backupCodes?: BackupCode[] | null;
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
    /**
     * ID of the newly created user.
     */
    userId?: Guid | null;
}
export interface AuthTokens extends Parsable {
    /**
     * Refresh token for obtaining new session tokens.This is base64.RawURLEncoding encoded.DO NOT DECODE IT WHILE RETRIEVING IT FROM THE COOKIES/CLIENT.
     */
    refreshToken?: string | null;
    /**
     * Expiration time of the refresh token.
     */
    refreshTokenExpiry?: Date | null;
    /**
     * Unique identifier for the session.
     */
    sessionId?: Guid | null;
    /**
     * Session token for maintaining user sessions.This is a jwt which is base64.RawURLEncoding encoded.YOU NEED TO DECODE IT WHILE RETRIEVING IT FROM THE COOKIES/CLIENT.DO NOT USE IT AS IS. VALIDATION WILL FAIL WITHOUT DECODING.ONLY WHEN DECODED, YOU SHOULD PASS IT TO THE THINGS THAT REQUIRE THE SESSION TOKEN.
     */
    sessionToken?: string | null;
    /**
     * Expiration time of the session token.
     */
    sessionTokenExpiry?: Date | null;
}
export interface AuthVerifyEmailRequest extends Parsable {
    /**
     * Token attached to the verification link in the email.
     */
    token?: string | null;
}
export interface AuthVerifyEmailResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface BackupCode extends Parsable {
    /**
     * The backup code string.
     */
    code?: string | null;
    /**
     * Indicates if the backup code has been used.
     */
    used?: boolean | null;
}
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminCreateDomainRequest}
 */
export declare function createAdminCreateDomainRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminCreateDomainResponse}
 */
export declare function createAdminCreateDomainResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminCreateRoleRequest}
 */
export declare function createAdminCreateRoleRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminCreateRoleResponse}
 */
export declare function createAdminCreateRoleResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminDeleteDomainResponse}
 */
export declare function createAdminDeleteDomainResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminDeleteRoleResponse}
 */
export declare function createAdminDeleteRoleResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminGetDomainVerificationCodeResponse}
 */
export declare function createAdminGetDomainVerificationCodeResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminUpdateDomainRequest}
 */
export declare function createAdminUpdateDomainRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminUpdateDomainResponse}
 */
export declare function createAdminUpdateDomainResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminUpdateOrgRequest}
 */
export declare function createAdminUpdateOrgRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminUpdateOrgResponse}
 */
export declare function createAdminUpdateOrgResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminUpdateRoleRequest}
 */
export declare function createAdminUpdateRoleRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminUpdateRoleResponse}
 */
export declare function createAdminUpdateRoleResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AdminVerifyDomainResponse}
 */
export declare function createAdminVerifyDomainResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthFlow}
 */
export declare function createAuthFlowFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthGetFlowResponse}
 */
export declare function createAuthGetFlowResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {string}
 */
export declare function createAuthLoginRequest_userIpFromDiscriminatorValue(parseNode: ParseNode | undefined): string | undefined;
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthLoginRequest}
 */
export declare function createAuthLoginRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthLoginResponse}
 */
export declare function createAuthLoginResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthLogoutRequest}
 */
export declare function createAuthLogoutRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthLogoutResponse}
 */
export declare function createAuthLogoutResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {string}
 */
export declare function createAuthRefreshRequest_userIpFromDiscriminatorValue(parseNode: ParseNode | undefined): string | undefined;
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthRefreshRequest}
 */
export declare function createAuthRefreshRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthRefreshResponse}
 */
export declare function createAuthRefreshResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthSendVerificationEmailRequest}
 */
export declare function createAuthSendVerificationEmailRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthSendVerificationEmailResponse}
 */
export declare function createAuthSendVerificationEmailResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthSignupRequest}
 */
export declare function createAuthSignupRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthSignupResponse}
 */
export declare function createAuthSignupResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthTokens}
 */
export declare function createAuthTokensFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthVerifyEmailRequest}
 */
export declare function createAuthVerifyEmailRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {AuthVerifyEmailResponse}
 */
export declare function createAuthVerifyEmailResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {BackupCode}
 */
export declare function createBackupCodeFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {Domain}
 */
export declare function createDomainFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {ErrorResponse}
 */
export declare function createErrorResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {MinimalRole}
 */
export declare function createMinimalRoleFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {Org}
 */
export declare function createOrgFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {OrgSettings}
 */
export declare function createOrgSettingsFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {OrgSettingsMFA_roles}
 */
export declare function createOrgSettingsMFA_rolesFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {OrgSettingsMFA}
 */
export declare function createOrgSettingsMFAFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {QueryFilterOption}
 */
export declare function createQueryFilterOptionFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {QueryFilters}
 */
export declare function createQueryFiltersFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {QueryPagination}
 */
export declare function createQueryPaginationFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {QuerySortOption}
 */
export declare function createQuerySortOptionFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {Role}
 */
export declare function createRoleFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {SysAdminCreateOrgRequest}
 */
export declare function createSysAdminCreateOrgRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {SysAdminCreateOrgResponse}
 */
export declare function createSysAdminCreateOrgResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {SysAdminDeleteOrgResponse}
 */
export declare function createSysAdminDeleteOrgResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {SysAdminGetConfigResponse}
 */
export declare function createSysAdminGetConfigResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {SysAdminGetOrgResponse}
 */
export declare function createSysAdminGetOrgResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {SysAdminListOrgsRequest}
 */
export declare function createSysAdminListOrgsRequestFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {SysAdminListOrgsResponse}
 */
export declare function createSysAdminListOrgsResponseFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * Creates a new instance of the appropriate class based on discriminator value
 * @param parseNode The parse node to use to read the discriminator value and create the object
 * @returns {SystemConfigPublicFacing}
 */
export declare function createSystemConfigPublicFacingFromDiscriminatorValue(parseNode: ParseNode | undefined): ((instance?: Parsable) => Record<string, (node: ParseNode) => void>);
/**
 * The deserialization information for the current model
 * @param AdminCreateDomainRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminCreateDomainRequest(adminCreateDomainRequest?: Partial<AdminCreateDomainRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminCreateDomainResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminCreateDomainResponse(adminCreateDomainResponse?: Partial<AdminCreateDomainResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminCreateRoleRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminCreateRoleRequest(adminCreateRoleRequest?: Partial<AdminCreateRoleRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminCreateRoleResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminCreateRoleResponse(adminCreateRoleResponse?: Partial<AdminCreateRoleResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminDeleteDomainResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminDeleteDomainResponse(adminDeleteDomainResponse?: Partial<AdminDeleteDomainResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminDeleteRoleResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminDeleteRoleResponse(adminDeleteRoleResponse?: Partial<AdminDeleteRoleResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminGetDomainVerificationCodeResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminGetDomainVerificationCodeResponse(adminGetDomainVerificationCodeResponse?: Partial<AdminGetDomainVerificationCodeResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminUpdateDomainRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminUpdateDomainRequest(adminUpdateDomainRequest?: Partial<AdminUpdateDomainRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminUpdateDomainResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminUpdateDomainResponse(adminUpdateDomainResponse?: Partial<AdminUpdateDomainResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminUpdateOrgRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminUpdateOrgRequest(adminUpdateOrgRequest?: Partial<AdminUpdateOrgRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminUpdateOrgResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminUpdateOrgResponse(adminUpdateOrgResponse?: Partial<AdminUpdateOrgResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminUpdateRoleRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminUpdateRoleRequest(adminUpdateRoleRequest?: Partial<AdminUpdateRoleRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminUpdateRoleResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminUpdateRoleResponse(adminUpdateRoleResponse?: Partial<AdminUpdateRoleResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AdminVerifyDomainResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAdminVerifyDomainResponse(adminVerifyDomainResponse?: Partial<AdminVerifyDomainResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthFlow The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthFlow(authFlow?: Partial<AuthFlow> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthGetFlowResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthGetFlowResponse(authGetFlowResponse?: Partial<AuthGetFlowResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthLoginRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthLoginRequest(authLoginRequest?: Partial<AuthLoginRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthLoginResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthLoginResponse(authLoginResponse?: Partial<AuthLoginResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthLogoutRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthLogoutRequest(authLogoutRequest?: Partial<AuthLogoutRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthLogoutResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthLogoutResponse(authLogoutResponse?: Partial<AuthLogoutResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthRefreshRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthRefreshRequest(authRefreshRequest?: Partial<AuthRefreshRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthRefreshResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthRefreshResponse(authRefreshResponse?: Partial<AuthRefreshResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthSendVerificationEmailRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthSendVerificationEmailRequest(authSendVerificationEmailRequest?: Partial<AuthSendVerificationEmailRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthSendVerificationEmailResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthSendVerificationEmailResponse(authSendVerificationEmailResponse?: Partial<AuthSendVerificationEmailResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthSignupRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthSignupRequest(authSignupRequest?: Partial<AuthSignupRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthSignupResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthSignupResponse(authSignupResponse?: Partial<AuthSignupResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthTokens The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthTokens(authTokens?: Partial<AuthTokens> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthVerifyEmailRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthVerifyEmailRequest(authVerifyEmailRequest?: Partial<AuthVerifyEmailRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param AuthVerifyEmailResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoAuthVerifyEmailResponse(authVerifyEmailResponse?: Partial<AuthVerifyEmailResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param BackupCode The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoBackupCode(backupCode?: Partial<BackupCode> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param Domain The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoDomain(domain?: Partial<Domain> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param ErrorResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoErrorResponse(errorResponse?: Partial<ErrorResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param MinimalRole The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoMinimalRole(minimalRole?: Partial<MinimalRole> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param Org The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoOrg(org?: Partial<Org> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param OrgSettings The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoOrgSettings(orgSettings?: Partial<OrgSettings> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param OrgSettingsMFA The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoOrgSettingsMFA(orgSettingsMFA?: Partial<OrgSettingsMFA> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param OrgSettingsMFA_roles The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoOrgSettingsMFA_roles(orgSettingsMFA_roles?: Partial<OrgSettingsMFA_roles> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param QueryFilterOption The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoQueryFilterOption(queryFilterOption?: Partial<QueryFilterOption> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param QueryFilters The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoQueryFilters(queryFilters?: Partial<QueryFilters> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param QueryPagination The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoQueryPagination(queryPagination?: Partial<QueryPagination> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param QuerySortOption The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoQuerySortOption(querySortOption?: Partial<QuerySortOption> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param Role The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoRole(role?: Partial<Role> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param SysAdminCreateOrgRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoSysAdminCreateOrgRequest(sysAdminCreateOrgRequest?: Partial<SysAdminCreateOrgRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param SysAdminCreateOrgResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoSysAdminCreateOrgResponse(sysAdminCreateOrgResponse?: Partial<SysAdminCreateOrgResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param SysAdminDeleteOrgResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoSysAdminDeleteOrgResponse(sysAdminDeleteOrgResponse?: Partial<SysAdminDeleteOrgResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param SysAdminGetConfigResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoSysAdminGetConfigResponse(sysAdminGetConfigResponse?: Partial<SysAdminGetConfigResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param SysAdminGetOrgResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoSysAdminGetOrgResponse(sysAdminGetOrgResponse?: Partial<SysAdminGetOrgResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param SysAdminListOrgsRequest The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoSysAdminListOrgsRequest(sysAdminListOrgsRequest?: Partial<SysAdminListOrgsRequest> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param SysAdminListOrgsResponse The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoSysAdminListOrgsResponse(sysAdminListOrgsResponse?: Partial<SysAdminListOrgsResponse> | undefined): Record<string, (node: ParseNode) => void>;
/**
 * The deserialization information for the current model
 * @param SystemConfigPublicFacing The instance to deserialize into.
 * @returns {Record<string, (node: ParseNode) => void>}
 */
export declare function deserializeIntoSystemConfigPublicFacing(systemConfigPublicFacing?: Partial<SystemConfigPublicFacing> | undefined): Record<string, (node: ParseNode) => void>;
export interface Domain extends Parsable {
    /**
     * Indicates if users from this domain can auto-join the organization.
     */
    autoJoin?: boolean | null;
    /**
     * Role ID assigned to users who auto-join via this domain.
     */
    autoJoinRoleId?: Guid | null;
    /**
     * Role name corresponding to the autoJoinRoleId.
     */
    autoJoinRoleName?: string | null;
    /**
     * Timestamp when the domain was created.
     */
    createdAt?: Date | null;
    /**
     * Domain name.
     */
    domain?: string | null;
    /**
     * ID of the organization the domain belongs to.
     */
    orgId?: Guid | null;
    /**
     * Timestamp when the domain was last updated.
     */
    updatedAt?: Date | null;
    /**
     * Indicates if the domain is verified.
     */
    verified?: boolean | null;
    /**
     * Timestamp when the domain was verified.
     */
    verifiedAt?: Date | null;
}
export interface ErrorResponse extends ApiError, Parsable {
    /**
     * Debug information for developers. This may contain sensitive information and should not be exposed to end users.This field will not be populated in production environments.
     */
    debug?: string | null;
    /**
     * Error message describing the issue. This is user-friendly and can be shown to the end user.
     */
    messageEscaped?: string | null;
}
export interface MinimalRole extends Parsable {
    /**
     * Timestamp when the role was created.
     */
    createdAt?: Date | null;
    /**
     * Unique identifier for the role.
     */
    id?: Guid | null;
    /**
     * Description of the role.
     */
    roleDesc?: string | null;
    /**
     * Name of the role.
     */
    roleName?: string | null;
    /**
     * Timestamp when the role was last updated.
     */
    updatedAt?: Date | null;
}
export interface Org extends Parsable {
    /**
     * URL of the organization's avatar image.
     */
    avatarURL?: string | null;
    /**
     * Timestamp when the organization was created.
     */
    createdAt?: Date | null;
    /**
     * Description of the organization.
     */
    description?: string | null;
    /**
     * Unique identifier for the organization.
     */
    id?: Guid | null;
    /**
     * Name of the organization.
     */
    name?: string | null;
    /**
     * Organization settings schema.
     */
    settings?: OrgSettings | null;
    /**
     * Unique URL-friendly slug for the organization.
     */
    slug?: string | null;
    /**
     * Timestamp when the organization was last updated.
     */
    updatedAt?: Date | null;
}
/**
 * Organization settings schema.
 */
export interface OrgSettings extends Parsable {
    /**
     * The mfa property
     */
    mfa?: OrgSettingsMFA | null;
}
export interface OrgSettingsMFA extends Parsable {
    /**
     * Indicates if MFA is required for the organization.
     */
    enabled?: boolean | null;
    /**
     * role name to role-id mapping for which MFA is enabled. If empty, MFA is enabled for all roles.
     */
    roles?: OrgSettingsMFA_roles | null;
}
/**
 * role name to role-id mapping for which MFA is enabled. If empty, MFA is enabled for all roles.
 */
export interface OrgSettingsMFA_roles extends AdditionalDataHolder, Parsable {
}
export type QueryFilterMode = (typeof QueryFilterModeObject)[keyof typeof QueryFilterModeObject];
export type QueryFilterOp = (typeof QueryFilterOpObject)[keyof typeof QueryFilterOpObject];
export interface QueryFilterOption extends Parsable {
    /**
     * Field name to filter on.
     */
    field?: string | null;
    /**
     * Enum representing query filter operations.
     */
    op?: QueryFilterOp | null;
}
export interface QueryFilters extends Parsable {
    /**
     * Enum representing query filter modes.
     */
    mode?: QueryFilterMode | null;
    /**
     * The options property
     */
    options?: QueryFilterOption[] | null;
}
export interface QueryPagination extends Parsable {
    /**
     * Page number (starting from 0).
     */
    page?: number | null;
    /**
     * Number of items per page.
     */
    pageSize?: number | null;
}
export interface QuerySortOption extends Parsable {
    /**
     * Sort order. True for ascending, false for descending.
     */
    ascending?: boolean | null;
    /**
     * Field name to sort by.
     */
    field?: string | null;
}
export interface Role extends Parsable {
    /**
     * Timestamp when the role was created.
     */
    createdAt?: Date | null;
    /**
     * Unique identifier for the role.
     */
    id?: Guid | null;
    /**
     * ID of the organization the role belongs to.
     */
    orgId?: Guid | null;
    /**
     * List of permissions assigned to the role.
     */
    permissions?: string[] | null;
    /**
     * Description of the role.
     */
    roleDesc?: string | null;
    /**
     * Name of the role.
     */
    roleName?: string | null;
    /**
     * Timestamp when the role was last updated.
     */
    updatedAt?: Date | null;
}
/**
 * Serializes information the current object
 * @param AdminCreateDomainRequest The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminCreateDomainRequest(writer: SerializationWriter, adminCreateDomainRequest?: Partial<AdminCreateDomainRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminCreateDomainResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminCreateDomainResponse(writer: SerializationWriter, adminCreateDomainResponse?: Partial<AdminCreateDomainResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminCreateRoleRequest The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminCreateRoleRequest(writer: SerializationWriter, adminCreateRoleRequest?: Partial<AdminCreateRoleRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminCreateRoleResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminCreateRoleResponse(writer: SerializationWriter, adminCreateRoleResponse?: Partial<AdminCreateRoleResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminDeleteDomainResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminDeleteDomainResponse(writer: SerializationWriter, adminDeleteDomainResponse?: Partial<AdminDeleteDomainResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminDeleteRoleResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminDeleteRoleResponse(writer: SerializationWriter, adminDeleteRoleResponse?: Partial<AdminDeleteRoleResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminGetDomainVerificationCodeResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminGetDomainVerificationCodeResponse(writer: SerializationWriter, adminGetDomainVerificationCodeResponse?: Partial<AdminGetDomainVerificationCodeResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminUpdateDomainRequest The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminUpdateDomainRequest(writer: SerializationWriter, adminUpdateDomainRequest?: Partial<AdminUpdateDomainRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminUpdateDomainResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminUpdateDomainResponse(writer: SerializationWriter, adminUpdateDomainResponse?: Partial<AdminUpdateDomainResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminUpdateOrgRequest The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminUpdateOrgRequest(writer: SerializationWriter, adminUpdateOrgRequest?: Partial<AdminUpdateOrgRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminUpdateOrgResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminUpdateOrgResponse(writer: SerializationWriter, adminUpdateOrgResponse?: Partial<AdminUpdateOrgResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminUpdateRoleRequest The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminUpdateRoleRequest(writer: SerializationWriter, adminUpdateRoleRequest?: Partial<AdminUpdateRoleRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminUpdateRoleResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminUpdateRoleResponse(writer: SerializationWriter, adminUpdateRoleResponse?: Partial<AdminUpdateRoleResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AdminVerifyDomainResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAdminVerifyDomainResponse(writer: SerializationWriter, adminVerifyDomainResponse?: Partial<AdminVerifyDomainResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthFlow The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthFlow(writer: SerializationWriter, authFlow?: Partial<AuthFlow> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthGetFlowResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthGetFlowResponse(writer: SerializationWriter, authGetFlowResponse?: Partial<AuthGetFlowResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthLoginRequest The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthLoginRequest(writer: SerializationWriter, authLoginRequest?: Partial<AuthLoginRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthLoginRequest_userIp The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param key The name of the property to write in the serialization.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthLoginRequest_userIp(writer: SerializationWriter, key: string, authLoginRequest_userIp: Parsable | AuthLoginRequest_userIp | undefined, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthLoginResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthLoginResponse(writer: SerializationWriter, authLoginResponse?: Partial<AuthLoginResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthLogoutRequest The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthLogoutRequest(writer: SerializationWriter, authLogoutRequest?: Partial<AuthLogoutRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthLogoutResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthLogoutResponse(writer: SerializationWriter, authLogoutResponse?: Partial<AuthLogoutResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthRefreshRequest The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthRefreshRequest(writer: SerializationWriter, authRefreshRequest?: Partial<AuthRefreshRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthRefreshRequest_userIp The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param key The name of the property to write in the serialization.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthRefreshRequest_userIp(writer: SerializationWriter, key: string, authRefreshRequest_userIp: Parsable | AuthRefreshRequest_userIp | undefined, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthRefreshResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthRefreshResponse(writer: SerializationWriter, authRefreshResponse?: Partial<AuthRefreshResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthSendVerificationEmailRequest The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthSendVerificationEmailRequest(writer: SerializationWriter, authSendVerificationEmailRequest?: Partial<AuthSendVerificationEmailRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthSendVerificationEmailResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthSendVerificationEmailResponse(writer: SerializationWriter, authSendVerificationEmailResponse?: Partial<AuthSendVerificationEmailResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthSignupRequest The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthSignupRequest(writer: SerializationWriter, authSignupRequest?: Partial<AuthSignupRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthSignupResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthSignupResponse(writer: SerializationWriter, authSignupResponse?: Partial<AuthSignupResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthTokens The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthTokens(writer: SerializationWriter, authTokens?: Partial<AuthTokens> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthVerifyEmailRequest The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthVerifyEmailRequest(writer: SerializationWriter, authVerifyEmailRequest?: Partial<AuthVerifyEmailRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param AuthVerifyEmailResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeAuthVerifyEmailResponse(writer: SerializationWriter, authVerifyEmailResponse?: Partial<AuthVerifyEmailResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param BackupCode The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeBackupCode(writer: SerializationWriter, backupCode?: Partial<BackupCode> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param Domain The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeDomain(writer: SerializationWriter, domain?: Partial<Domain> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param ErrorResponse The instance to serialize from.
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeErrorResponse(writer: SerializationWriter, errorResponse?: Partial<ErrorResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param MinimalRole The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeMinimalRole(writer: SerializationWriter, minimalRole?: Partial<MinimalRole> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param Org The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeOrg(writer: SerializationWriter, org?: Partial<Org> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param OrgSettings The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeOrgSettings(writer: SerializationWriter, orgSettings?: Partial<OrgSettings> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param OrgSettingsMFA The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeOrgSettingsMFA(writer: SerializationWriter, orgSettingsMFA?: Partial<OrgSettingsMFA> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param OrgSettingsMFA_roles The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeOrgSettingsMFA_roles(writer: SerializationWriter, orgSettingsMFA_roles?: Partial<OrgSettingsMFA_roles> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param QueryFilterOption The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeQueryFilterOption(writer: SerializationWriter, queryFilterOption?: Partial<QueryFilterOption> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param QueryFilters The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeQueryFilters(writer: SerializationWriter, queryFilters?: Partial<QueryFilters> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param QueryPagination The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeQueryPagination(writer: SerializationWriter, queryPagination?: Partial<QueryPagination> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param QuerySortOption The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeQuerySortOption(writer: SerializationWriter, querySortOption?: Partial<QuerySortOption> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param Role The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeRole(writer: SerializationWriter, role?: Partial<Role> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param SysAdminCreateOrgRequest The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeSysAdminCreateOrgRequest(writer: SerializationWriter, sysAdminCreateOrgRequest?: Partial<SysAdminCreateOrgRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param SysAdminCreateOrgResponse The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeSysAdminCreateOrgResponse(writer: SerializationWriter, sysAdminCreateOrgResponse?: Partial<SysAdminCreateOrgResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param SysAdminDeleteOrgResponse The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeSysAdminDeleteOrgResponse(writer: SerializationWriter, sysAdminDeleteOrgResponse?: Partial<SysAdminDeleteOrgResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param SysAdminGetConfigResponse The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeSysAdminGetConfigResponse(writer: SerializationWriter, sysAdminGetConfigResponse?: Partial<SysAdminGetConfigResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param SysAdminGetOrgResponse The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeSysAdminGetOrgResponse(writer: SerializationWriter, sysAdminGetOrgResponse?: Partial<SysAdminGetOrgResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param SysAdminListOrgsRequest The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeSysAdminListOrgsRequest(writer: SerializationWriter, sysAdminListOrgsRequest?: Partial<SysAdminListOrgsRequest> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param SysAdminListOrgsResponse The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeSysAdminListOrgsResponse(writer: SerializationWriter, sysAdminListOrgsResponse?: Partial<SysAdminListOrgsResponse> | undefined | null, isSerializingDerivedType?: boolean): void;
/**
 * Serializes information the current object
 * @param isSerializingDerivedType A boolean indicating whether the serialization is for a derived type.
 * @param SystemConfigPublicFacing The instance to serialize from.
 * @param writer Serialization writer to use to serialize this model
 */
export declare function serializeSystemConfigPublicFacing(writer: SerializationWriter, systemConfigPublicFacing?: Partial<SystemConfigPublicFacing> | undefined | null, isSerializingDerivedType?: boolean): void;
export interface SysAdminCreateOrgRequest extends Parsable {
    /**
     * URL of the organization's avatar image. If you wish to upload a new image, use the dedicated avatar upload endpoint. It will automatically update the avatarURL. Uploading an avatar during org creation is not possible, you need to upload it after you create the org.
     */
    avatarURL?: string | null;
    /**
     * Description of the organization.
     */
    description?: string | null;
    /**
     * Name of the organization to be created.
     */
    name?: string | null;
    /**
     * Organization settings schema.
     */
    settings?: OrgSettings | null;
    /**
     * Unique slug for the organization.
     */
    slug?: string | null;
}
export interface SysAdminCreateOrgResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Created organization ID.
     */
    orgId?: Guid | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface SysAdminDeleteOrgResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface SysAdminGetConfigResponse extends Parsable {
    /**
     * The config property
     */
    config?: SystemConfigPublicFacing | null;
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface SysAdminGetOrgResponse extends Parsable {
    /**
     * List of domains associated with the organization.
     */
    domains?: Domain[] | null;
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * The org property
     */
    org?: Org | null;
    /**
     * List of roles associated with the organization.
     */
    roles?: MinimalRole[] | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
}
export interface SysAdminListOrgsRequest extends Parsable {
    /**
     * The filters property
     */
    filters?: QueryFilters | null;
    /**
     * The pagination property
     */
    pagination?: QueryPagination | null;
    /**
     * Sorting options.
     */
    sort?: QuerySortOption[] | null;
}
export interface SysAdminListOrgsResponse extends Parsable {
    /**
     * Response message.
     */
    message?: string | null;
    /**
     * List of organizations.
     */
    orgs?: Org[] | null;
    /**
     * Pagination details.
     */
    pagination?: QueryPagination | null;
    /**
     * Indicates if the request was successful.
     */
    success?: boolean | null;
    /**
     * Total number of organizations matching the query.
     */
    total?: number | null;
}
export interface SystemConfigPublicFacing extends Parsable {
    /**
     * Branding configuration for the public-facing Nexeres system.
     */
    branding?: PublicFacingBrandingConfig | null;
    /**
     * Indicates if the system is in debug mode.
     */
    debug?: boolean | null;
    /**
     * JWT configuration for the Nexeres system.
     */
    jwt?: PublicFacingJWTConfig | null;
    /**
     * Multitenancy configuration for the public-facing Nexeres system.
     */
    multitenancy?: PublicFacingMultitenancyConfig | null;
    /**
     * Notifications configuration for the Nexeres system.
     */
    notifications?: PublicFacingNotificationsConfig | null;
    /**
     * Public endpoint configuration for the Nexeres system.
     */
    publicEndpoint?: PublicFacingPublicEndpointConfig | null;
    /**
     * Security configuration for the Nexeres system.
     */
    security?: PublicFacingSecurityConfig | null;
}
/**
 * Type of authentication flow (e.g., login, change-password).
 */
export declare const AuthFlow_typeObject: {
    readonly Login: "login";
    readonly ChangePassword: "change-password";
};
/**
 * Enum representing query filter modes.
 */
export declare const QueryFilterModeObject: {
    readonly AND: "AND";
    readonly OR: "OR";
};
/**
 * Enum representing query filter operations.
 */
export declare const QueryFilterOpObject: {
    readonly EQ: "EQ";
    readonly NEQ: "NEQ";
    readonly GT: "GT";
    readonly GTE: "GTE";
    readonly LT: "LT";
    readonly LTE: "LTE";
    readonly CONTAINS: "CONTAINS";
};
//# sourceMappingURL=index.d.ts.map