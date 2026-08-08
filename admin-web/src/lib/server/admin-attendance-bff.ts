import "server-only";

import { SafeApiError } from "@/lib/auth/types";
import type {
  AdminAttendanceCorrectionInput,
  AdminAttendanceDetail,
  AdminAttendanceListQuery,
  AdminAttendanceListResponse,
  AdminAttendanceSummary,
} from "@/lib/attendance/types";
import {
  buildAttendanceQueryParams,
  isAttendanceDetail,
  isAttendanceListResponse,
  isAttendanceSummary,
} from "@/lib/attendance/utils";
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

type AttendanceRequestOptions = {
  method?: "GET" | "PATCH";
  body?: Record<string, unknown>;
};

export async function getAdminAttendanceSummary(
  date?: string,
): Promise<AdminAttendanceSummary> {
  const params = new URLSearchParams();
  if (date) params.set("date", date);
  const suffix = params.size > 0 ? `?${params.toString()}` : "";
  const result = await attendanceRequestWithSession<AdminAttendanceSummary>(
    `/admin/attendance/summary${suffix}`,
  );
  if (!isAttendanceSummary(result)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons ringkasan presensi tidak sesuai.",
      502,
    );
  }
  return result;
}

export async function getAdminAttendance(
  query: AdminAttendanceListQuery,
): Promise<AdminAttendanceListResponse> {
  const queryString = buildAttendanceQueryParams(query);
  const suffix = queryString ? `?${queryString}` : "";
  const result = await attendanceRequestWithSession<AdminAttendanceListResponse>(
    `/admin/attendance${suffix}`,
  );
  if (!isAttendanceListResponse(result)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons daftar presensi tidak sesuai.",
      502,
    );
  }
  return result;
}

export async function getAdminAttendanceDetail(
  id: string,
): Promise<AdminAttendanceDetail> {
  const result = await attendanceRequestWithSession<AdminAttendanceDetail>(
    `/admin/attendance/${encodeURIComponent(id)}`,
  );
  if (!isAttendanceDetail(result)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons detail presensi tidak sesuai.",
      502,
    );
  }
  return result;
}

export async function correctAdminAttendance(
  id: string,
  input: AdminAttendanceCorrectionInput,
): Promise<AdminAttendanceDetail> {
  const result = await attendanceRequestWithSession<AdminAttendanceDetail>(
    `/admin/attendance/${encodeURIComponent(id)}/correction`,
    {
      method: "PATCH",
      body: {
        check_in_time: input.check_in_time,
        check_out_time: input.check_out_time,
        reason: input.reason,
      },
    },
  );
  if (!isAttendanceDetail(result)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons koreksi presensi tidak sesuai.",
      502,
    );
  }
  return result;
}

async function attendanceRequestWithSession<T>(
  path: string,
  options: AttendanceRequestOptions = {},
): Promise<T> {
  let accessToken: string | undefined = await readAccessToken();
  if (!accessToken) {
    accessToken = (await refreshAccessToken()) ?? undefined;
  }
  if (!accessToken) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }

  try {
    return await request<T>(path, accessToken, options);
  } catch (error) {
    if (!(error instanceof SafeApiError) || error.code !== "UNAUTHORIZED") {
      throw error;
    }
  }

  const refreshed = await refreshAccessToken();
  if (!refreshed) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }
  return request<T>(path, refreshed, options);
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

async function request<T>(
  path: string,
  accessToken: string,
  options: AttendanceRequestOptions,
): Promise<T> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
  try {
    const headers: Record<string, string> = {
      Accept: "application/json",
      Authorization: `Bearer ${accessToken}`,
    };
    if (options.body) {
      headers["Content-Type"] = "application/json";
    }

    const response = await fetch(buildGoApiUrl(getGoApiBaseUrl(), path), {
      method: options.method ?? "GET",
      headers,
      body: options.body ? JSON.stringify(options.body) : undefined,
      cache: "no-store",
      signal: controller.signal,
    });
    const payload: unknown = await response.json().catch(() => null);
    if (!response.ok) {
      const message = readMessage(payload);
      if (response.status === 401) {
        throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
      }
      if (response.status === 403) {
        throw new SafeApiError("FORBIDDEN", "Akses tidak diizinkan.", 403);
      }
      if (response.status === 404) {
        throw new SafeApiError("NOT_FOUND", "Presensi tidak ditemukan.", 404);
      }
      if (response.status === 400) {
        throw new SafeApiError(
          "BAD_REQUEST",
          message ?? "Data presensi tidak valid.",
          400,
        );
      }
      throw new SafeApiError(
        "INTERNAL_ERROR",
        message ?? "Layanan backend mengalami gangguan.",
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

function readMessage(value: unknown): string | null {
  if (typeof value !== "object" || value === null) return null;
  const message = (value as Record<string, unknown>).message;
  return typeof message === "string" ? message : null;
}
