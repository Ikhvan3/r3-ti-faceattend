import type { Employee } from "../employee/types";
import { isRecord } from "../employee/utils";
import type { AssignmentStatus } from "../schedule/types";
import {
  assignmentStatusLabel,
  deriveAssignmentStatus,
  formatDateOnly,
  isAssignmentStatus,
  isDateOnly,
} from "../schedule/utils";
import type {
  LocationAssignment,
  LocationAssignmentListQuery,
  LocationAssignmentListResponse,
  OfficeLocation,
  OfficeLocationListQuery,
  OfficeLocationListResponse,
  OfficeLocationStatus,
} from "./types";

export const OFFICE_LOCATION_STATUSES: OfficeLocationStatus[] = [
  "ACTIVE",
  "INACTIVE",
];
export const LOCATION_ASSIGNMENT_STATUSES: AssignmentStatus[] = [
  "CURRENT",
  "UPCOMING",
  "ENDED",
];
export const DEFAULT_LOCATION_PAGE = 1;
export const DEFAULT_LOCATION_PAGE_SIZE = 10;
export const MAX_LOCATION_PAGE_SIZE = 100;

export type ParsedOfficeLocationSearchParams = {
  page: number;
  page_size: number;
  search: string;
  status: OfficeLocationStatus | "";
};

export type ParsedLocationAssignmentSearchParams = {
  page: number;
  page_size: number;
  search: string;
  user_id: string;
  office_location_id: string;
  status: AssignmentStatus | "";
};

export function parseOfficeLocationSearchParams(
  searchParams: Record<string, string | string[] | undefined>,
): ParsedOfficeLocationSearchParams {
  const status = firstParam(searchParams.status).trim().toUpperCase();
  return {
    page: parsePositiveInt(
      firstParam(searchParams.page),
      DEFAULT_LOCATION_PAGE,
      Number.MAX_SAFE_INTEGER,
    ),
    page_size: parsePositiveInt(
      firstParam(searchParams.page_size),
      DEFAULT_LOCATION_PAGE_SIZE,
      MAX_LOCATION_PAGE_SIZE,
    ),
    search: firstParam(searchParams.search).trim(),
    status: isOfficeLocationStatus(status) ? status : "",
  };
}

export function parseLocationAssignmentSearchParams(
  searchParams: Record<string, string | string[] | undefined>,
): ParsedLocationAssignmentSearchParams {
  const status = firstParam(searchParams.status).trim().toUpperCase();
  return {
    page: parsePositiveInt(
      firstParam(searchParams.page),
      DEFAULT_LOCATION_PAGE,
      Number.MAX_SAFE_INTEGER,
    ),
    page_size: parsePositiveInt(
      firstParam(searchParams.page_size),
      DEFAULT_LOCATION_PAGE_SIZE,
      MAX_LOCATION_PAGE_SIZE,
    ),
    search: firstParam(searchParams.search).trim(),
    user_id: firstParam(searchParams.user_id).trim(),
    office_location_id: firstParam(searchParams.office_location_id).trim(),
    status: isAssignmentStatus(status) ? status : "",
  };
}

export function buildOfficeLocationQueryParams(
  query: OfficeLocationListQuery,
): string {
  const params = new URLSearchParams();
  params.set(
    "page",
    String(clampPositiveInt(query.page, DEFAULT_LOCATION_PAGE)),
  );
  params.set(
    "page_size",
    String(
      clampPositiveInt(
        query.page_size,
        DEFAULT_LOCATION_PAGE_SIZE,
        MAX_LOCATION_PAGE_SIZE,
      ),
    ),
  );
  setTrimmed(params, "search", query.search);
  if (query.status && isOfficeLocationStatus(query.status)) {
    params.set("status", query.status);
  }
  return params.toString();
}

export function buildLocationAssignmentQueryParams(
  query: LocationAssignmentListQuery,
): string {
  const params = new URLSearchParams();
  params.set(
    "page",
    String(clampPositiveInt(query.page, DEFAULT_LOCATION_PAGE)),
  );
  params.set(
    "page_size",
    String(
      clampPositiveInt(
        query.page_size,
        DEFAULT_LOCATION_PAGE_SIZE,
        MAX_LOCATION_PAGE_SIZE,
      ),
    ),
  );
  setTrimmed(params, "search", query.search);
  setTrimmed(params, "user_id", query.user_id);
  setTrimmed(params, "office_location_id", query.office_location_id);
  if (query.status && isAssignmentStatus(query.status)) {
    params.set("status", query.status);
  }
  return params.toString();
}

export function officeLocationListHref(
  query: ParsedOfficeLocationSearchParams,
  page: number,
): string {
  const params = new URLSearchParams();
  params.set("page", String(Math.max(1, page)));
  params.set("page_size", String(query.page_size));
  setTrimmed(params, "search", query.search);
  if (query.status) {
    params.set("status", query.status);
  }
  return `/office-locations?${params.toString()}`;
}

export function locationAssignmentListHref(
  query: ParsedLocationAssignmentSearchParams,
  page: number,
): string {
  const params = new URLSearchParams();
  params.set("page", String(Math.max(1, page)));
  params.set("page_size", String(query.page_size));
  setTrimmed(params, "search", query.search);
  setTrimmed(params, "user_id", query.user_id);
  setTrimmed(params, "office_location_id", query.office_location_id);
  if (query.status) {
    params.set("status", query.status);
  }
  return `/location-assignments?${params.toString()}`;
}

export function isOfficeLocationStatus(
  value: unknown,
): value is OfficeLocationStatus {
  return value === "ACTIVE" || value === "INACTIVE";
}

export function officeLocationStatusFromActive(
  isActive: boolean,
): OfficeLocationStatus {
  return isActive ? "ACTIVE" : "INACTIVE";
}

export function officeLocationStatusLabel(
  status: OfficeLocationStatus,
): string {
  return status === "ACTIVE" ? "Aktif" : "Nonaktif";
}

export function locationBadgeClass(
  status: OfficeLocationStatus | AssignmentStatus,
): string {
  switch (status) {
    case "ACTIVE":
    case "CURRENT":
      return "border-emerald-200 bg-emerald-50 text-emerald-700";
    case "UPCOMING":
      return "border-sky-200 bg-sky-50 text-sky-700";
    case "ENDED":
    case "INACTIVE":
      return "border-slate-200 bg-slate-100 text-slate-700";
  }
}

export function locationStatusLabel(
  status: OfficeLocationStatus | AssignmentStatus,
): string {
  if (status === "ACTIVE" || status === "INACTIVE") {
    return officeLocationStatusLabel(status);
  }
  return assignmentStatusLabel(status);
}

export function parseLatitude(value: string): number | null {
  return parseBoundedNumber(value, -90, 90);
}

export function parseLongitude(value: string): number | null {
  return parseBoundedNumber(value, -180, 180);
}

export function parseRadiusMeters(value: string): number | null {
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || !Number.isInteger(parsed)) {
    return null;
  }
  if (parsed < 10 || parsed > 2000) {
    return null;
  }
  return parsed;
}

export function formatCoordinate(value: number): string {
  if (!Number.isFinite(value)) {
    return "-";
  }
  return new Intl.NumberFormat("id-ID", {
    maximumFractionDigits: 6,
  }).format(value);
}

export function formatRadius(value: number): string {
  return `${value} meter`;
}

export { formatDateOnly, isDateOnly };

export function safeLocationApiMessage(
  status: number,
  fallback: string,
): string {
  if (status === 400) {
    return "Request tidak valid. Periksa kembali data lokasi.";
  }
  if (status === 401) {
    return "Session tidak valid. Silakan login ulang.";
  }
  if (status === 403) {
    return "Akses admin diperlukan untuk mengelola lokasi.";
  }
  if (status === 404) {
    return "Data lokasi tidak ditemukan.";
  }
  if (status === 409) {
    return fallback;
  }
  if (status === 503) {
    return "Layanan backend belum tersedia. Coba lagi nanti.";
  }
  return "Terjadi kesalahan. Coba lagi nanti.";
}

export function isOfficeLocation(value: unknown): value is OfficeLocation {
  return (
    isRecord(value) &&
    typeof value.id === "string" &&
    typeof value.name === "string" &&
    (typeof value.address === "string" || value.address === null) &&
    typeof value.latitude === "number" &&
    Number.isFinite(value.latitude) &&
    typeof value.longitude === "number" &&
    Number.isFinite(value.longitude) &&
    typeof value.radius_meters === "number" &&
    Number.isInteger(value.radius_meters) &&
    typeof value.is_active === "boolean" &&
    typeof value.created_at === "string" &&
    typeof value.updated_at === "string"
  );
}

export function isOfficeLocationListResponse(
  value: unknown,
): value is OfficeLocationListResponse {
  return (
    isRecord(value) &&
    Array.isArray(value.items) &&
    value.items.every(isOfficeLocation) &&
    typeof value.page === "number" &&
    typeof value.page_size === "number" &&
    typeof value.total_items === "number" &&
    typeof value.total_pages === "number"
  );
}

export function normalizeLocationAssignment(
  value: unknown,
): LocationAssignment | null {
  if (
    !isRecord(value) ||
    !isEmployeeLike(value.user) ||
    !isOfficeLocation(value.office_location)
  ) {
    return null;
  }
  if (
    typeof value.id !== "string" ||
    typeof value.effective_from !== "string" ||
    !(typeof value.effective_to === "string" || value.effective_to === null) ||
    typeof value.created_at !== "string" ||
    typeof value.updated_at !== "string"
  ) {
    return null;
  }

  const status = isAssignmentStatus(value.status)
    ? value.status
    : deriveAssignmentStatus(value.effective_from, value.effective_to);

  return {
    id: value.id,
    user: value.user,
    office_location: value.office_location,
    effective_from: value.effective_from,
    effective_to: value.effective_to,
    status,
    created_at: value.created_at,
    updated_at: value.updated_at,
  };
}

export function isLocationAssignment(
  value: unknown,
): value is LocationAssignment {
  return normalizeLocationAssignment(value) !== null;
}

export function normalizeLocationAssignmentListResponse(
  value: unknown,
): LocationAssignmentListResponse | null {
  if (!isRecord(value) || !Array.isArray(value.items)) {
    return null;
  }
  const items = value.items.map(normalizeLocationAssignment);
  if (
    !items.every((item): item is LocationAssignment => item !== null) ||
    typeof value.page !== "number" ||
    typeof value.page_size !== "number" ||
    typeof value.total_items !== "number" ||
    typeof value.total_pages !== "number"
  ) {
    return null;
  }
  return {
    items,
    page: value.page,
    page_size: value.page_size,
    total_items: value.total_items,
    total_pages: value.total_pages,
  };
}

export function isLocationAssignmentListResponse(
  value: unknown,
): value is LocationAssignmentListResponse {
  return normalizeLocationAssignmentListResponse(value) !== null;
}

function parseBoundedNumber(
  value: string,
  min: number,
  max: number,
): number | null {
  const parsed = Number(value);
  if (!Number.isFinite(parsed) || parsed < min || parsed > max) {
    return null;
  }
  return parsed;
}

function isEmployeeLike(value: unknown): value is Employee {
  return (
    isRecord(value) &&
    typeof value.id === "string" &&
    typeof value.employee_number === "string" &&
    typeof value.name === "string" &&
    typeof value.email === "string" &&
    (typeof value.phone === "string" || value.phone === null) &&
    (typeof value.position === "string" || value.position === null) &&
    value.role === "USER" &&
    ["ACTIVE", "INACTIVE", "SUSPENDED"].includes(String(value.account_status)) &&
    typeof value.created_at === "string" &&
    typeof value.updated_at === "string"
  );
}

function firstParam(value: string | string[] | undefined): string {
  return Array.isArray(value) ? value[0] ?? "" : value ?? "";
}

function parsePositiveInt(value: string, fallback: number, max: number): number {
  const parsed = Number(value);
  if (!Number.isInteger(parsed) || parsed < 1) {
    return fallback;
  }
  return Math.min(parsed, max);
}

function clampPositiveInt(
  value: number | undefined,
  fallback: number,
  max = Number.MAX_SAFE_INTEGER,
): number {
  if (value === undefined || !Number.isInteger(value) || value < 1) {
    return fallback;
  }
  return Math.min(value, max);
}

function setTrimmed(
  params: URLSearchParams,
  key: string,
  value: string | undefined,
): void {
  const normalized = value?.trim();
  if (normalized) {
    params.set(key, normalized);
  }
}
