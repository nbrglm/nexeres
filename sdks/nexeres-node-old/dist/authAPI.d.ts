import { NexeresClient, NexeresResponse } from "./client.js";
import { Tokens } from "./models/tokens.js";
import { Flow } from "./models/flow.js";
export declare class AuthAPI {
    private client;
    constructor(client: NexeresClient);
    login(params: LoginParams): NexeresResponse<LoginResponse>;
    signup(params: SignupParams): NexeresResponse<SignupResponse>;
    sendVerificationEmail(params: SendVerificationEmailParams): NexeresResponse<SendVerificationEmailResponse>;
    verifyEmail(params: VerifyEmailParams): NexeresResponse<VerifyEmailResponse>;
    refreshToken(params: RefreshTokenParams): NexeresResponse<RefreshTokenResponse>;
    logout(params: LogoutParams): NexeresResponse<LogoutResponse>;
    getFlowData(params: {
        flowId: string;
    }): NexeresResponse<Flow>;
}
/** Parameters for the login API */
export type LoginParams = {
    /** The user's email address */
    email: string;
    /** The user's password */
    password: string;
    /** Optional return URL for the login flow, if Nexeres is set with multi-tenancy */
    flowReturnTo?: string | undefined;
    /** The user's IP address */
    userIp: string;
    /** The user's user agent string */
    userAgent: string;
};
/** Response from the login API */
export type LoginResponse = {
    /** The user's authentication tokens */
    tokens?: Tokens | undefined;
    /** Indicates if email verification is required */
    requireEmailVerification: boolean;
    /** A user-friendly message */
    message: string;
    /** The flow ID for the login process, only present if multi-tenancy is enabled */
    flowId?: string | undefined;
};
/** Parameters for the signup API */
export type SignupParams = {
    /** The user's email address */
    email: string;
    /** The user's password */
    password: string;
    /** Confirm Password */
    confirmPassword: string;
    /** The user's first name */
    firstName: string;
    /** The user's last name */
    lastName: string;
    /** Optional invite token for the signup process, if multi-tenancy is enabled */
    inviteToken?: string | undefined;
};
/** Response from the signup API */
export type SignupResponse = {
    /** A user-friendly message */
    message: string;
    /** The newly-created user's ID */
    userId: string;
};
/** Parameters for sending a verification email */
export type SendVerificationEmailParams = {
    /** The user's email address */
    email: string;
};
/** Response from sending a verification email */
export type SendVerificationEmailResponse = {
    /** A user-friendly message */
    message: string;
    /** Indicates if the email was sent successfully */
    success: boolean;
};
/** Parameters for verifying an email */
export type VerifyEmailParams = {
    /** The verification token sent to the user's email */
    token: string;
};
/** Response from verifying an email */
export type VerifyEmailResponse = {
    /** A user-friendly message */
    message: string;
    /** Indicates if the email was verified successfully */
    success: boolean;
};
/** Parameters for refreshing a token */
export type RefreshTokenParams = {
    /** The refresh token */
    refreshToken: string;
    /** The user's IP address */
    userIp: string;
    /** The user's user agent string */
    userAgent: string;
};
/** Response from refreshing a token */
export type RefreshTokenResponse = {
    /** The new authentication tokens */
    tokens: Tokens;
};
/** Parameters for logging out */
export type LogoutParams = {
    /** The refresh token */
    refreshToken?: string;
    /** The session token */
    sessionToken?: string;
};
/** Response from logging out */
export type LogoutResponse = {
    /** A user-friendly message */
    message: string;
    /** Indicates if the logout was successful */
    success: boolean;
};
//# sourceMappingURL=authAPI.d.ts.map