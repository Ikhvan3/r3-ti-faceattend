export type UserRole = "ADMIN" | "USER";

export type AccountStatus = "ACTIVE" | "INACTIVE" | "SUSPENDED";

export type SafeUserProfile = {
  id: string;
  employee_number: string;
  name: string;
  email: string;
  phone: string | null;
  position: string | null;
  role: UserRole;
  account_status: AccountStatus;
};

export type LoginRequest = {
  email: string;
  password: string;
};

export type AuthTokenData = {
  access_token: string;
  refresh_token: string;
  token_type: "Bearer";
  expires_in: number;
  user: SafeUserProfile;
};

export type GoSuccessResponse<T> = {
  status: "ok";
  message: string;
  data: T;
};

export type GoErrorResponse = {
  status: "error";
  message: string;
};

export type ApiErrorCode =
  | "BAD_REQUEST"
  | "INVALID_CREDENTIALS"
  | "FORBIDDEN"
  | "UNAUTHORIZED"
  | "GO_API_UNAVAILABLE"
  | "INVALID_RESPONSE"
  | "INTERNAL_ERROR";

export class SafeApiError extends Error {
  readonly code: ApiErrorCode;
  readonly status: number;

  constructor(code: ApiErrorCode, message: string, status = 500) {
    super(message);
    this.name = "SafeApiError";
    this.code = code;
    this.status = status;
  }
}
