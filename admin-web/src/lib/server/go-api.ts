import "server-only";

import { getGoApiBaseUrl } from "@/lib/server/env";
import type {
  AuthTokenData,
  GoErrorResponse,
  GoSuccessResponse,
  LoginRequest,
  SafeUserProfile,
} from "@/lib/auth/types";
import { SafeApiError } from "@/lib/auth/types";

type RequestOptions = {
  method?: "GET" | "POST";
  body?: unknown;
  accessToken?: string;
};

const REQUEST_TIMEOUT_MS = 8000;

export async function loginToGoApi(
  request: LoginRequest,
): Promise<AuthTokenData> {
  const response = await goFetch<AuthTokenData>("/auth/login", {
    method: "POST",
    body: request,
  });

  if (!isAuthTokenData(response)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons autentikasi tidak sesuai.",
      502,
    );
  }

  return response;
}

export async function refreshToGoApi(
  refreshToken: string,
): Promise<AuthTokenData> {
  const response = await goFetch<AuthTokenData>("/auth/refresh", {
    method: "POST",
    body: { refresh_token: refreshToken },
  });

  if (!isAuthTokenData(response)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons refresh tidak sesuai.",
      502,
    );
  }

  return response;
}

export async function logoutToGoApi(refreshToken: string): Promise<void> {
  await goFetch<null>("/auth/logout", {
    method: "POST",
    body: { refresh_token: refreshToken },
  });
}

export async function getMeFromGoApi(
  accessToken: string,
): Promise<SafeUserProfile> {
  const response = await goFetch<SafeUserProfile>("/auth/me", {
    method: "GET",
    accessToken,
  });

  if (!isSafeUserProfile(response)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons profil tidak sesuai.",
      502,
    );
  }

  return response;
}

async function goFetch<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);

  try {
    const response = await fetch(`${getGoApiBaseUrl()}${path}`, {
      method: options.method ?? "GET",
      headers: buildHeaders(options),
      body:
        options.body === undefined ? undefined : JSON.stringify(options.body),
      cache: "no-store",
      signal: controller.signal,
    });

    const payload = await readJson(response);

    if (!response.ok) {
      throw mapGoApiError(response.status, payload);
    }

    if (!isGoSuccessResponse<T>(payload)) {
      throw new SafeApiError(
        "INVALID_RESPONSE",
        "Respons backend tidak sesuai.",
        502,
      );
    }

    return payload.data;
  } catch (error) {
    if (error instanceof SafeApiError) {
      throw error;
    }

    throw new SafeApiError(
      "GO_API_UNAVAILABLE",
      "Layanan autentikasi belum tersedia. Coba lagi nanti.",
      503,
    );
  } finally {
    clearTimeout(timeout);
  }
}

function buildHeaders(options: RequestOptions): HeadersInit {
  const headers: Record<string, string> = {
    Accept: "application/json",
  };

  if (options.body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  if (options.accessToken) {
    headers.Authorization = `Bearer ${options.accessToken}`;
  }

  return headers;
}

async function readJson(response: Response): Promise<unknown> {
  try {
    return await response.json();
  } catch {
    return null;
  }
}

function mapGoApiError(status: number, payload: unknown): SafeApiError {
  const message = isGoErrorResponse(payload)
    ? payload.message
    : "Request autentikasi gagal.";

  if (status === 400) {
    return new SafeApiError("BAD_REQUEST", "Request tidak valid.", status);
  }

  if (status === 401) {
    return new SafeApiError(
      message.toLowerCase().includes("password")
        ? "INVALID_CREDENTIALS"
        : "UNAUTHORIZED",
      "Email atau password tidak valid.",
      status,
    );
  }

  if (status === 403) {
    return new SafeApiError("FORBIDDEN", "Akses tidak diizinkan.", status);
  }

  return new SafeApiError(
    "INTERNAL_ERROR",
    "Layanan autentikasi mengalami gangguan.",
    status,
  );
}

function isGoSuccessResponse<T>(
  payload: unknown,
): payload is GoSuccessResponse<T> {
  if (!isRecord(payload)) {
    return false;
  }

  return payload.status === "ok" && "data" in payload;
}

function isGoErrorResponse(payload: unknown): payload is GoErrorResponse {
  return (
    isRecord(payload) &&
    payload.status === "error" &&
    typeof payload.message === "string"
  );
}

function isAuthTokenData(value: unknown): value is AuthTokenData {
  return (
    isRecord(value) &&
    typeof value.access_token === "string" &&
    typeof value.refresh_token === "string" &&
    value.token_type === "Bearer" &&
    typeof value.expires_in === "number" &&
    isSafeUserProfile(value.user)
  );
}

function isSafeUserProfile(value: unknown): value is SafeUserProfile {
  return (
    isRecord(value) &&
    typeof value.id === "string" &&
    typeof value.employee_number === "string" &&
    typeof value.name === "string" &&
    typeof value.email === "string" &&
    (typeof value.phone === "string" || value.phone === null) &&
    (typeof value.position === "string" || value.position === null) &&
    (value.role === "ADMIN" || value.role === "USER") &&
    ["ACTIVE", "INACTIVE", "SUSPENDED"].includes(
      String(value.account_status),
    )
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
