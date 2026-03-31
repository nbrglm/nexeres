/**
 * Base response interface for all API responses
 */
export interface BaseResponse {
  success: boolean;
  message: string;
}

/**
 * Error response interface for API errors
 */
export interface ErrorResponse {
  /**
   * Error message describing what went wrong, safe to show to end users
   */
  message: string;

  /**
   * Debug information, won't be present when server is running in production mode
   */
  debug?: string | null;

  /**
   * Status code as received from the server
   */
  code: number;
}

export interface NetworkOrParseError {
  errType: "NetworkError" | "ParseError" | "UnknownError";
}

export interface AuthSignupRequest {
  email: string;
  password: string;
  name: string;
  confirmPassword: string;
  inviteToken?: string | null;
}

export interface AuthSignupResponse extends BaseResponse {
  backupCodes?: BackupCode[] | null;
  userId: string;
}

export interface AuthSendVerificationEmailRequest {
  email: string;
}

export interface AuthSendVerificationEmailResponse extends BaseResponse { }

export interface BackupCode {
  code: string;
  used: boolean;
}