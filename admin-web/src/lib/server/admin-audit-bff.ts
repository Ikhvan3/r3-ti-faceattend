import "server-only";

import type { AuditLogListResponse, AuditLogQuery } from "@/lib/audit/types";
import { buildAuditQueryParams, isAuditLogListResponse } from "@/lib/audit/utils";
import { SafeApiError } from "@/lib/auth/types";
import {
  clearAuthCookies,
  readAccessToken,
  readRefreshToken,
  setAuthCookies,
} from "@/lib/server/auth-cookies";
import { getGoApiBaseUrl } from "@/lib/server/env";
import { buildGoApiUrl } from "@/lib/server/go-api-url";
import { refreshToGoApi } from "@/lib/server/go-api";

const REQUEST_TIMEOUT_MS = 8000;

type ApiEnvelope<T> = {
  status: "ok";
  data: T;
};

export async function getAdminAuditLogs(
  query: AuditLogQuery,
): Promise<AuditLogListResponse> {
  const queryString = buildAuditQueryParams(query);
  const suffix = queryString ? `?${queryString}` : "";
  const result = await auditRequestWithSession<AuditLogListResponse>(
    `/admin/audit-logs${suffix}`,
  );
  if (!isAuditLogListResponse(result)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons audit log tidak sesuai.",
      502,
    );
  }
  return result;
}

async function auditRequestWithSession<T>(path: string): Promise<T> {
  let accessToken: string | undefined = await readAccessToken();
  if (!accessToken) {
    accessToken = (await refreshAccessToken()) ?? undefined;
  }
  if (!accessToken) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }

  try {
    return await request<T>(path, accessToken);
  } catch (error) {
    if (!(error instanceof SafeApiError) || error.code !== "UNAUTHORIZED") {
      throw error;
    }
  }

  const refreshed = await refreshAccessToken();
  if (!refreshed) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }
  return request<T>(path, refreshed);
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = await readRefreshToken();
  if (!refreshToken) {
    await clearAuthCookies();
    return null;
  }

  try {
    const refreshed = await refreshToGoApi(refreshToken);
    if (refreshed.user.role !== "ADMIN") {
      await clearAuthCookies();
      return null;
    }
    await setAuthCookies({
      accessToken: refreshed.access_token,
      refreshToken: refreshed.refresh_token,
      accessMaxAge: refreshed.expires_in,
    });
    return refreshed.access_token;
  } catch (error) {
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      await clearAuthCookies();
      return null;
    }
    throw error;
  }
}

async function request<T>(path: string, accessToken: string): Promise<T> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
  try {
    const response = await fetch(buildGoApiUrl(getGoApiBaseUrl(), path), {
      method: "GET",
      headers: {
        Accept: "application/json",
        Authorization: `Bearer ${accessToken}`,
      },
      cache: "no-store",
      signal: controller.signal,
    });
    const payload: unknown = await response.json().catch(() => null);
    if (!response.ok) {
      if (response.status === 401) {
        throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
      }
      if (response.status === 403) {
        throw new SafeApiError("FORBIDDEN", "Akses tidak diizinkan.", 403);
      }
      if (response.status === 400) {
        throw new SafeApiError("BAD_REQUEST", "Filter audit log tidak valid.", 400);
      }
      throw new SafeApiError(
        "INTERNAL_ERROR",
        "Layanan audit log mengalami gangguan.",
        response.status,
      );
    }
    if (!isEnvelope<T>(payload)) {
      throw new SafeApiError(
        "INVALID_RESPONSE",
        "Respons backend tidak sesuai.",
        502,
      );
    }
    return payload.data;
  } catch (error) {
    if (error instanceof SafeApiError) throw error;
    throw new SafeApiError(
      "GO_API_UNAVAILABLE",
      "Layanan backend belum tersedia. Coba lagi nanti.",
      503,
    );
  } finally {
    clearTimeout(timeout);
  }
}

function isEnvelope<T>(value: unknown): value is ApiEnvelope<T> {
  return (
    typeof value === "object" &&
    value !== null &&
    "status" in value &&
    value.status === "ok" &&
    "data" in value
  );
}
