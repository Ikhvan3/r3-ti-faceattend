import "server-only";

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

export async function resetEmployeeFaceEnrollmentWithSession(
  employeeID: string,
  reason: string,
): Promise<void> {
  let accessToken = await readAccessToken();
  if (!accessToken) {
    const refreshedAccessToken = await refreshAdminAccessToken();
    accessToken = refreshedAccessToken ?? undefined;
  }
  if (!accessToken) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }

  try {
    await sendReset(employeeID, reason, accessToken);
    return;
  } catch (error) {
    if (!(error instanceof SafeApiError) || error.code !== "UNAUTHORIZED") {
      throw error;
    }
  }

  const refreshed = await refreshAdminAccessToken();
  if (!refreshed) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }

  await sendReset(employeeID, reason, refreshed);
}

async function refreshAdminAccessToken(): Promise<string | null> {
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
  } catch {
    await clearAuthCookies();
    return null;
  }
}

async function sendReset(
  employeeID: string,
  reason: string,
  accessToken: string,
): Promise<void> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);

  let response: Response;
  try {
    response = await fetch(
      buildGoApiUrl(
        getGoApiBaseUrl(),
        `/admin/face-enrollments/${encodeURIComponent(employeeID)}`,
      ),
      {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${accessToken}`,
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ reason }),
        cache: "no-store",
        signal: controller.signal,
      },
    );
  } catch {
    throw new SafeApiError(
      "GO_API_UNAVAILABLE",
      "Layanan backend belum tersedia.",
      503,
    );
  } finally {
    clearTimeout(timeout);
  }

  const payload = await readPayload(response);
  if (!response.ok) {
    throw mapError(response.status, payload?.message);
  }
  if (!payload || payload.status !== "ok") {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons reset enrollment wajah tidak sesuai.",
      502,
    );
  }
}

async function readPayload(
  response: Response,
): Promise<{ status: string; message?: string } | null> {
  try {
    const value = (await response.json()) as unknown;
    if (typeof value !== "object" || value === null) {
      return null;
    }
    const record = value as Record<string, unknown>;
    return {
      status: typeof record.status === "string" ? record.status : "",
      message: typeof record.message === "string" ? record.message : undefined,
    };
  } catch {
    return null;
  }
}

function mapError(status: number, message?: string): SafeApiError {
  if (status === 400) {
    return new SafeApiError(
      "BAD_REQUEST",
      message ?? "User atau alasan reset enrollment tidak valid.",
      400,
    );
  }
  if (status === 401) {
    return new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }
  if (status === 403) {
    return new SafeApiError("FORBIDDEN", "Akses admin diperlukan.", 403);
  }
  if (status === 404) {
    return new SafeApiError(
      "NOT_FOUND",
      "Enrollment wajah pegawai tidak ditemukan.",
      404,
    );
  }
  return new SafeApiError(
    "INTERNAL_ERROR",
    message ?? "Reset enrollment wajah gagal diproses.",
    500,
  );
}
