import "server-only";

import { getGoApiBaseUrl } from "@/lib/server/env";
import { buildGoApiUrl } from "@/lib/server/go-api-url";
import type {
  AuthTokenData,
  GoErrorResponse,
  GoSuccessResponse,
  LoginRequest,
  SafeUserProfile,
} from "@/lib/auth/types";
import { SafeApiError } from "@/lib/auth/types";
import type {
  CreateEmployeeRequest,
  Employee,
  EmployeeListQuery,
  EmployeeListResponse,
  UpdateEmployeeRequest,
  UpdateEmployeeStatusRequest,
} from "@/lib/employee/types";
import {
  buildEmployeeQueryParams,
  isEmployee,
  isRecord,
} from "@/lib/employee/utils";
import type {
  CreateLocationAssignmentRequest,
  CreateOfficeLocationRequest,
  EndLocationAssignmentRequest,
  LocationAssignment,
  LocationAssignmentListQuery,
  LocationAssignmentListResponse,
  OfficeLocation,
  OfficeLocationListQuery,
  OfficeLocationListResponse,
  UpdateOfficeLocationRequest,
  UpdateOfficeLocationStatusRequest,
} from "@/lib/location/types";
import {
  buildLocationAssignmentQueryParams,
  buildOfficeLocationQueryParams,
  isOfficeLocation,
  isOfficeLocationListResponse,
  normalizeLocationAssignment,
  normalizeLocationAssignmentListResponse,
} from "@/lib/location/utils";
import type {
  CreateScheduleAssignmentRequest,
  CreateWorkScheduleRequest,
  EndScheduleAssignmentRequest,
  ScheduleAssignment,
  ScheduleAssignmentListQuery,
  ScheduleAssignmentListResponse,
  UpdateWorkScheduleRequest,
  UpdateWorkScheduleStatusRequest,
  WorkSchedule,
  WorkScheduleListQuery,
  WorkScheduleListResponse,
} from "@/lib/schedule/types";
import {
  buildAssignmentQueryParams,
  buildWorkScheduleQueryParams,
  isWorkSchedule,
  isWorkScheduleListResponse,
  normalizeAssignment,
  normalizeAssignmentListResponse,
} from "@/lib/schedule/utils";
import {
  clearAuthCookies,
  readAccessToken,
  readRefreshToken,
  setAuthCookies,
} from "@/lib/server/auth-cookies";

type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "PATCH";
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

export async function getEmployees(
  query: EmployeeListQuery,
): Promise<EmployeeListResponse> {
  const queryString = buildEmployeeQueryParams(query);
  const response = await readWithRefresh((accessToken) =>
    goFetch<EmployeeListResponse>(`/admin/employees?${queryString}`, {
      method: "GET",
      accessToken,
    }),
  );

  if (!isEmployeeListResponse(response)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons daftar pegawai tidak sesuai.",
      502,
    );
  }

  return response;
}

export async function getEmployeeByID(id: string): Promise<Employee> {
  const response = await readWithRefresh((accessToken) =>
    goFetch<Employee>(`/admin/employees/${encodeURIComponent(id)}`, {
      method: "GET",
      accessToken,
    }),
  );

  if (!isEmployee(response)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons profil pegawai tidak sesuai.",
      502,
    );
  }

  return response;
}

export async function createEmployeeWithSession(
  request: CreateEmployeeRequest,
): Promise<Employee> {
  return mutateEmployeeWithRefresh((accessToken) =>
    goFetch<Employee>("/admin/employees", {
      method: "POST",
      body: request,
      accessToken,
    }),
  );
}

export async function updateEmployeeWithSession(
  id: string,
  request: UpdateEmployeeRequest,
): Promise<Employee> {
  return mutateEmployeeWithRefresh((accessToken) =>
    goFetch<Employee>(`/admin/employees/${encodeURIComponent(id)}`, {
      method: "PUT",
      body: request,
      accessToken,
    }),
  );
}

export async function updateEmployeeStatusWithSession(
  id: string,
  request: UpdateEmployeeStatusRequest,
): Promise<Employee> {
  return mutateEmployeeWithRefresh((accessToken) =>
    goFetch<Employee>(`/admin/employees/${encodeURIComponent(id)}/status`, {
      method: "PATCH",
      body: request,
      accessToken,
    }),
  );
}

export async function getWorkSchedules(
  query: WorkScheduleListQuery,
): Promise<WorkScheduleListResponse> {
  const queryString = buildWorkScheduleQueryParams(query);
  const response = await readWithRefresh((accessToken) =>
    goFetch<WorkScheduleListResponse>(`/admin/work-schedules?${queryString}`, {
      method: "GET",
      accessToken,
    }),
  );

  if (!isWorkScheduleListResponse(response)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons daftar jadwal kerja tidak sesuai.",
      502,
    );
  }

  return response;
}

export async function getActiveWorkSchedules(): Promise<WorkSchedule[]> {
  const result = await getWorkSchedules({
    page: 1,
    page_size: 100,
    status: "ACTIVE",
  });
  return result.items;
}

export async function getWorkScheduleByID(id: string): Promise<WorkSchedule> {
  const response = await readWithRefresh((accessToken) =>
    goFetch<WorkSchedule>(`/admin/work-schedules/${encodeURIComponent(id)}`, {
      method: "GET",
      accessToken,
    }),
  );

  if (!isWorkSchedule(response)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons jadwal kerja tidak sesuai.",
      502,
    );
  }

  return response;
}

export async function createWorkScheduleWithSession(
  request: CreateWorkScheduleRequest,
): Promise<WorkSchedule> {
  return mutateWorkScheduleWithRefresh((accessToken) =>
    goFetch<WorkSchedule>("/admin/work-schedules", {
      method: "POST",
      body: request,
      accessToken,
    }),
  );
}

export async function updateWorkScheduleWithSession(
  id: string,
  request: UpdateWorkScheduleRequest,
): Promise<WorkSchedule> {
  return mutateWorkScheduleWithRefresh((accessToken) =>
    goFetch<WorkSchedule>(`/admin/work-schedules/${encodeURIComponent(id)}`, {
      method: "PUT",
      body: request,
      accessToken,
    }),
  );
}

export async function updateWorkScheduleStatusWithSession(
  id: string,
  request: UpdateWorkScheduleStatusRequest,
): Promise<WorkSchedule> {
  return mutateWorkScheduleWithRefresh((accessToken) =>
    goFetch<WorkSchedule>(
      `/admin/work-schedules/${encodeURIComponent(id)}/status`,
      {
        method: "PATCH",
        body: request,
        accessToken,
      },
    ),
  );
}

export async function getScheduleAssignments(
  query: ScheduleAssignmentListQuery,
): Promise<ScheduleAssignmentListResponse> {
  const queryString = buildAssignmentQueryParams(query);
  const response = await readWithRefresh((accessToken) =>
    goFetch<ScheduleAssignmentListResponse>(
      `/admin/schedule-assignments?${queryString}`,
      {
        method: "GET",
        accessToken,
      },
    ),
  );

  const normalized = normalizeAssignmentListResponse(response);
  if (!normalized) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons daftar penugasan jadwal tidak sesuai.",
      502,
    );
  }

  return normalized;
}

export async function getScheduleAssignmentByID(
  id: string,
): Promise<ScheduleAssignment> {
  const response = await readWithRefresh((accessToken) =>
    goFetch<ScheduleAssignment>(
      `/admin/schedule-assignments/${encodeURIComponent(id)}`,
      {
        method: "GET",
        accessToken,
      },
    ),
  );

  const normalized = normalizeAssignment(response);
  if (!normalized) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons penugasan jadwal tidak sesuai.",
      502,
    );
  }

  return normalized;
}

export async function createScheduleAssignmentWithSession(
  request: CreateScheduleAssignmentRequest,
): Promise<ScheduleAssignment> {
  return mutateAssignmentWithRefresh((accessToken) =>
    goFetch<ScheduleAssignment>("/admin/schedule-assignments", {
      method: "POST",
      body: request,
      accessToken,
    }),
  );
}

export async function endScheduleAssignmentWithSession(
  id: string,
  request: EndScheduleAssignmentRequest,
): Promise<ScheduleAssignment> {
  return mutateAssignmentWithRefresh((accessToken) =>
    goFetch<ScheduleAssignment>(
      `/admin/schedule-assignments/${encodeURIComponent(id)}/end`,
      {
        method: "PATCH",
        body: request,
        accessToken,
      },
    ),
  );
}

export async function getEmployeeOptions(): Promise<Employee[]> {
  const result = await getEmployees({ page: 1, page_size: 100, status: "ACTIVE" });
  return result.items.filter((employee) => employee.role === "USER");
}

export async function getOfficeLocations(
  query: OfficeLocationListQuery,
): Promise<OfficeLocationListResponse> {
  const queryString = buildOfficeLocationQueryParams(query);
  const response = await readWithRefresh((accessToken) =>
    goFetch<OfficeLocationListResponse>(
      `/admin/office-locations?${queryString}`,
      {
        method: "GET",
        accessToken,
      },
    ),
  );

  if (!isOfficeLocationListResponse(response)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons daftar lokasi kantor tidak sesuai.",
      502,
    );
  }

  return response;
}

export async function getActiveOfficeLocations(): Promise<OfficeLocation[]> {
  const result = await getOfficeLocations({
    page: 1,
    page_size: 100,
    status: "ACTIVE",
  });
  return result.items;
}

export async function getOfficeLocationByID(
  id: string,
): Promise<OfficeLocation> {
  const response = await readWithRefresh((accessToken) =>
    goFetch<OfficeLocation>(
      `/admin/office-locations/${encodeURIComponent(id)}`,
      {
        method: "GET",
        accessToken,
      },
    ),
  );

  if (!isOfficeLocation(response)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons lokasi kantor tidak sesuai.",
      502,
    );
  }

  return response;
}

export async function createOfficeLocationWithSession(
  request: CreateOfficeLocationRequest,
): Promise<OfficeLocation> {
  return mutateOfficeLocationWithRefresh((accessToken) =>
    goFetch<OfficeLocation>("/admin/office-locations", {
      method: "POST",
      body: request,
      accessToken,
    }),
  );
}

export async function updateOfficeLocationWithSession(
  id: string,
  request: UpdateOfficeLocationRequest,
): Promise<OfficeLocation> {
  return mutateOfficeLocationWithRefresh((accessToken) =>
    goFetch<OfficeLocation>(
      `/admin/office-locations/${encodeURIComponent(id)}`,
      {
        method: "PUT",
        body: request,
        accessToken,
      },
    ),
  );
}

export async function updateOfficeLocationStatusWithSession(
  id: string,
  request: UpdateOfficeLocationStatusRequest,
): Promise<OfficeLocation> {
  return mutateOfficeLocationWithRefresh((accessToken) =>
    goFetch<OfficeLocation>(
      `/admin/office-locations/${encodeURIComponent(id)}/status`,
      {
        method: "PATCH",
        body: request,
        accessToken,
      },
    ),
  );
}

export async function getLocationAssignments(
  query: LocationAssignmentListQuery,
): Promise<LocationAssignmentListResponse> {
  const queryString = buildLocationAssignmentQueryParams(query);
  const response = await readWithRefresh((accessToken) =>
    goFetch<LocationAssignmentListResponse>(
      `/admin/location-assignments?${queryString}`,
      {
        method: "GET",
        accessToken,
      },
    ),
  );

  const normalized = normalizeLocationAssignmentListResponse(response);
  if (!normalized) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons daftar penugasan lokasi tidak sesuai.",
      502,
    );
  }

  return normalized;
}

export async function getLocationAssignmentByID(
  id: string,
): Promise<LocationAssignment> {
  const response = await readWithRefresh((accessToken) =>
    goFetch<LocationAssignment>(
      `/admin/location-assignments/${encodeURIComponent(id)}`,
      {
        method: "GET",
        accessToken,
      },
    ),
  );

  const normalized = normalizeLocationAssignment(response);
  if (!normalized) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons penugasan lokasi tidak sesuai.",
      502,
    );
  }

  return normalized;
}

export async function createLocationAssignmentWithSession(
  request: CreateLocationAssignmentRequest,
): Promise<LocationAssignment> {
  return mutateLocationAssignmentWithRefresh((accessToken) =>
    goFetch<LocationAssignment>("/admin/location-assignments", {
      method: "POST",
      body: request,
      accessToken,
    }),
  );
}

export async function endLocationAssignmentWithSession(
  id: string,
  request: EndLocationAssignmentRequest,
): Promise<LocationAssignment> {
  return mutateLocationAssignmentWithRefresh((accessToken) =>
    goFetch<LocationAssignment>(
      `/admin/location-assignments/${encodeURIComponent(id)}/end`,
      {
        method: "PATCH",
        body: request,
        accessToken,
      },
    ),
  );
}

export async function readAccessTokenWithRefresh(): Promise<string | null> {
  const accessToken = await readAccessToken();
  if (accessToken) {
    return accessToken;
  }

  return refreshAccessTokenFromCookie();
}

async function refreshAccessTokenFromCookie(): Promise<string | null> {
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

async function mutateEmployeeWithRefresh(
  operation: (accessToken: string) => Promise<Employee>,
): Promise<Employee> {
  let accessToken = await readAccessToken();
  if (!accessToken) {
    const refreshedAccessToken = await refreshAccessTokenFromCookie();
    accessToken = refreshedAccessToken ?? undefined;
  }
  if (!accessToken) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }

  try {
    const employee = await operation(accessToken);
    if (!isEmployee(employee)) {
      throw new SafeApiError(
        "INVALID_RESPONSE",
        "Respons pegawai tidak sesuai.",
        502,
      );
    }
    return employee;
  } catch (error) {
    if (!(error instanceof SafeApiError) || error.code !== "UNAUTHORIZED") {
      throw error;
    }
  }

  const refreshedAccessToken = await refreshAccessTokenFromCookie();
  if (!refreshedAccessToken) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }

  const employee = await operation(refreshedAccessToken);
  if (!isEmployee(employee)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons pegawai tidak sesuai.",
      502,
    );
  }

  return employee;
}

async function mutateWorkScheduleWithRefresh(
  operation: (accessToken: string) => Promise<WorkSchedule>,
): Promise<WorkSchedule> {
  const schedule = await mutateWithRefresh(operation);
  if (!isWorkSchedule(schedule)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons jadwal kerja tidak sesuai.",
      502,
    );
  }
  return schedule;
}

async function mutateAssignmentWithRefresh(
  operation: (accessToken: string) => Promise<ScheduleAssignment>,
): Promise<ScheduleAssignment> {
  const assignment = await mutateWithRefresh(operation);
  const normalized = normalizeAssignment(assignment);
  if (!normalized) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons penugasan jadwal tidak sesuai.",
      502,
    );
  }
  return normalized;
}

async function mutateOfficeLocationWithRefresh(
  operation: (accessToken: string) => Promise<OfficeLocation>,
): Promise<OfficeLocation> {
  const office = await mutateWithRefresh(operation);
  if (!isOfficeLocation(office)) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons lokasi kantor tidak sesuai.",
      502,
    );
  }
  return office;
}

async function mutateLocationAssignmentWithRefresh(
  operation: (accessToken: string) => Promise<LocationAssignment>,
): Promise<LocationAssignment> {
  const assignment = await mutateWithRefresh(operation);
  const normalized = normalizeLocationAssignment(assignment);
  if (!normalized) {
    throw new SafeApiError(
      "INVALID_RESPONSE",
      "Respons penugasan lokasi tidak sesuai.",
      502,
    );
  }
  return normalized;
}

async function mutateWithRefresh<T>(
  operation: (accessToken: string) => Promise<T>,
): Promise<T> {
  let accessToken = await readAccessToken();
  if (!accessToken) {
    const refreshedAccessToken = await refreshAccessTokenFromCookie();
    accessToken = refreshedAccessToken ?? undefined;
  }
  if (!accessToken) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }

  try {
    return await operation(accessToken);
  } catch (error) {
    if (!(error instanceof SafeApiError) || error.code !== "UNAUTHORIZED") {
      throw error;
    }
  }

  const refreshedAccessToken = await refreshAccessTokenFromCookie();
  if (!refreshedAccessToken) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }

  return operation(refreshedAccessToken);
}

async function readWithRefresh<T>(
  operation: (accessToken: string) => Promise<T>,
): Promise<T> {
  let accessToken = await readAccessToken();
  if (!accessToken) {
    const refreshedAccessToken = await refreshAccessTokenFromCookie();
    accessToken = refreshedAccessToken ?? undefined;
  }
  if (!accessToken) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }

  try {
    return await operation(accessToken);
  } catch (error) {
    if (!(error instanceof SafeApiError) || error.code !== "UNAUTHORIZED") {
      throw error;
    }
  }

  const refreshedAccessToken = await refreshAccessTokenFromCookie();
  if (!refreshedAccessToken) {
    throw new SafeApiError("UNAUTHORIZED", "Session tidak valid.", 401);
  }

  return operation(refreshedAccessToken);
}

async function goFetch<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);

  try {
    const response = await fetch(buildGoApiUrl(getGoApiBaseUrl(), path), {
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
    : "Request backend gagal.";

  if (status === 400) {
    return new SafeApiError("BAD_REQUEST", "Request tidak valid.", status);
  }

  if (status === 401) {
    return new SafeApiError(
      message.toLowerCase().includes("password")
        ? "INVALID_CREDENTIALS"
        : "UNAUTHORIZED",
      message.toLowerCase().includes("password")
        ? "Email atau password tidak valid."
        : "Session tidak valid.",
      status,
    );
  }

  if (status === 403) {
    return new SafeApiError("FORBIDDEN", "Akses tidak diizinkan.", status);
  }

  if (status === 404) {
    return new SafeApiError("NOT_FOUND", "Data tidak ditemukan.", status);
  }

  if (status === 409) {
    return new SafeApiError(
      "CONFLICT",
      safeConflictMessage(message),
      status,
    );
  }

  return new SafeApiError(
    "INTERNAL_ERROR",
    "Layanan backend mengalami gangguan.",
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

function isEmployeeListResponse(value: unknown): value is EmployeeListResponse {
  if (!isRecord(value) || !Array.isArray(value.items)) {
    return false;
  }

  return (
    value.items.every(isEmployee) &&
    typeof value.page === "number" &&
    typeof value.page_size === "number" &&
    typeof value.total_items === "number" &&
    typeof value.total_pages === "number"
  );
}

function safeConflictMessage(message: string): string {
  const normalized = message.toLowerCase();
  if (normalized.includes("assignment lokasi")) {
    return "Periode penugasan lokasi bertumpang tindih dengan penugasan lain.";
  }
  if (normalized.includes("lokasi kantor") || normalized.includes("lokasi")) {
    return "Lokasi tidak dapat dinonaktifkan karena masih ditugaskan kepada pegawai.";
  }
  if (normalized.includes("bertumpang") || normalized.includes("overlap")) {
    return "Periode penugasan jadwal bertumpang tindih dengan penugasan lain.";
  }
  if (normalized.includes("assignment aktif") || normalized.includes("masa depan")) {
    return "Jadwal tidak dapat dinonaktifkan karena masih digunakan oleh pegawai.";
  }
  if (normalized.includes("jadwal")) {
    return "Nama jadwal kerja sudah digunakan.";
  }
  if (normalized.includes("email")) {
    return "Email sudah digunakan oleh pegawai lain.";
  }
  if (normalized.includes("nomor") || normalized.includes("employee")) {
    return "Nomor pegawai sudah digunakan oleh pegawai lain.";
  }

  return "Data pegawai sudah digunakan.";
}
