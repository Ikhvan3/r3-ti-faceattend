import { isRecord } from "@/lib/employee/utils";
import type { Employee } from "@/lib/employee/types";
import {
  type AssignmentStatus,
  type ScheduleAssignment,
  type ScheduleAssignmentListQuery,
  type ScheduleAssignmentListResponse,
  type ScheduleStatus,
  type WorkSchedule,
  type WorkScheduleListQuery,
  type WorkScheduleListResponse,
} from "@/lib/schedule/types";

export const SCHEDULE_STATUSES: ScheduleStatus[] = ["ACTIVE", "INACTIVE"];
export const ASSIGNMENT_STATUSES: AssignmentStatus[] = [
  "CURRENT",
  "UPCOMING",
  "ENDED",
];
export const DEFAULT_SCHEDULE_PAGE = 1;
export const DEFAULT_SCHEDULE_PAGE_SIZE = 10;
export const MAX_SCHEDULE_PAGE_SIZE = 100;

export type ParsedWorkScheduleSearchParams = {
  page: number;
  page_size: number;
  search: string;
  status: ScheduleStatus | "";
};

export type ParsedAssignmentSearchParams = {
  page: number;
  page_size: number;
  search: string;
  user_id: string;
  schedule_id: string;
  status: AssignmentStatus | "";
};

export function parseWorkScheduleSearchParams(
  searchParams: Record<string, string | string[] | undefined>,
): ParsedWorkScheduleSearchParams {
  const status = firstParam(searchParams.status).trim().toUpperCase();

  return {
    page: parsePositiveInt(
      firstParam(searchParams.page),
      DEFAULT_SCHEDULE_PAGE,
      Number.MAX_SAFE_INTEGER,
    ),
    page_size: parsePositiveInt(
      firstParam(searchParams.page_size),
      DEFAULT_SCHEDULE_PAGE_SIZE,
      MAX_SCHEDULE_PAGE_SIZE,
    ),
    search: firstParam(searchParams.search).trim(),
    status: isScheduleStatus(status) ? status : "",
  };
}

export function parseAssignmentSearchParams(
  searchParams: Record<string, string | string[] | undefined>,
): ParsedAssignmentSearchParams {
  const status = firstParam(searchParams.status).trim().toUpperCase();

  return {
    page: parsePositiveInt(
      firstParam(searchParams.page),
      DEFAULT_SCHEDULE_PAGE,
      Number.MAX_SAFE_INTEGER,
    ),
    page_size: parsePositiveInt(
      firstParam(searchParams.page_size),
      DEFAULT_SCHEDULE_PAGE_SIZE,
      MAX_SCHEDULE_PAGE_SIZE,
    ),
    search: firstParam(searchParams.search).trim(),
    user_id: firstParam(searchParams.user_id).trim(),
    schedule_id: firstParam(searchParams.schedule_id).trim(),
    status: isAssignmentStatus(status) ? status : "",
  };
}

export function buildWorkScheduleQueryParams(
  query: WorkScheduleListQuery,
): string {
  const params = new URLSearchParams();
  params.set(
    "page",
    String(clampPositiveInt(query.page, DEFAULT_SCHEDULE_PAGE)),
  );
  params.set(
    "page_size",
    String(
      clampPositiveInt(
        query.page_size,
        DEFAULT_SCHEDULE_PAGE_SIZE,
        MAX_SCHEDULE_PAGE_SIZE,
      ),
    ),
  );
  const search = query.search?.trim();
  if (search) {
    params.set("search", search);
  }
  if (query.status && isScheduleStatus(query.status)) {
    params.set("status", query.status);
  }
  return params.toString();
}

export function buildAssignmentQueryParams(
  query: ScheduleAssignmentListQuery,
): string {
  const params = new URLSearchParams();
  params.set(
    "page",
    String(clampPositiveInt(query.page, DEFAULT_SCHEDULE_PAGE)),
  );
  params.set(
    "page_size",
    String(
      clampPositiveInt(
        query.page_size,
        DEFAULT_SCHEDULE_PAGE_SIZE,
        MAX_SCHEDULE_PAGE_SIZE,
      ),
    ),
  );
  setTrimmed(params, "search", query.search);
  setTrimmed(params, "user_id", query.user_id);
  setTrimmed(params, "schedule_id", query.schedule_id);
  if (query.status && isAssignmentStatus(query.status)) {
    params.set("status", query.status);
  }
  return params.toString();
}

export function workScheduleListHref(
  query: ParsedWorkScheduleSearchParams,
  page: number,
): string {
  const params = new URLSearchParams();
  params.set("page", String(Math.max(1, page)));
  params.set("page_size", String(query.page_size));
  setTrimmed(params, "search", query.search);
  if (query.status) {
    params.set("status", query.status);
  }
  return `/work-schedules?${params.toString()}`;
}

export function assignmentListHref(
  query: ParsedAssignmentSearchParams,
  page: number,
): string {
  const params = new URLSearchParams();
  params.set("page", String(Math.max(1, page)));
  params.set("page_size", String(query.page_size));
  setTrimmed(params, "search", query.search);
  setTrimmed(params, "user_id", query.user_id);
  setTrimmed(params, "schedule_id", query.schedule_id);
  if (query.status) {
    params.set("status", query.status);
  }
  return `/schedule-assignments?${params.toString()}`;
}

export function isScheduleStatus(value: unknown): value is ScheduleStatus {
  return value === "ACTIVE" || value === "INACTIVE";
}

export function isAssignmentStatus(value: unknown): value is AssignmentStatus {
  return value === "CURRENT" || value === "UPCOMING" || value === "ENDED";
}

export function scheduleStatusFromActive(isActive: boolean): ScheduleStatus {
  return isActive ? "ACTIVE" : "INACTIVE";
}

export function scheduleStatusLabel(status: ScheduleStatus): string {
  return status === "ACTIVE" ? "Aktif" : "Nonaktif";
}

export function assignmentStatusLabel(status: AssignmentStatus): string {
  switch (status) {
    case "CURRENT":
      return "Berjalan";
    case "UPCOMING":
      return "Akan Datang";
    case "ENDED":
      return "Berakhir";
  }
}

export function badgeClass(status: ScheduleStatus | AssignmentStatus): string {
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

export function formatScheduleTime(value: string): string {
  return /^\d{2}:\d{2}/.test(value) ? value.slice(0, 5).replace(":", ".") : "-";
}

export function isDateOnly(value: string): boolean {
  return /^\d{4}-\d{2}-\d{2}$/.test(value) && parseDateOnly(value) !== null;
}

export function formatDateOnly(value: string | null): string {
  if (value === null) {
    return "Tanpa batas";
  }
  const parsed = parseDateOnly(value);
  if (!parsed) {
    return "-";
  }
  return new Intl.DateTimeFormat("id-ID", {
    day: "2-digit",
    month: "long",
    year: "numeric",
  }).format(new Date(parsed.year, parsed.month - 1, parsed.day));
}

export function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function safeScheduleApiMessage(
  status: number,
  fallback: string,
): string {
  if (status === 400) {
    return "Request tidak valid. Periksa kembali data jadwal.";
  }
  if (status === 401) {
    return "Session tidak valid. Silakan login ulang.";
  }
  if (status === 403) {
    return "Akses admin diperlukan untuk mengelola jadwal.";
  }
  if (status === 404) {
    return "Data jadwal tidak ditemukan.";
  }
  if (status === 409) {
    return fallback;
  }
  if (status === 503) {
    return "Layanan backend belum tersedia. Coba lagi nanti.";
  }
  return "Terjadi kesalahan. Coba lagi nanti.";
}

export function isWorkSchedule(value: unknown): value is WorkSchedule {
  return (
    isRecord(value) &&
    typeof value.id === "string" &&
    typeof value.name === "string" &&
    typeof value.start_time === "string" &&
    typeof value.end_time === "string" &&
    typeof value.grace_minutes === "number" &&
    typeof value.is_active === "boolean" &&
    typeof value.created_at === "string" &&
    typeof value.updated_at === "string"
  );
}

export function isWorkScheduleListResponse(
  value: unknown,
): value is WorkScheduleListResponse {
  return (
    isRecord(value) &&
    Array.isArray(value.items) &&
    value.items.every(isWorkSchedule) &&
    typeof value.page === "number" &&
    typeof value.page_size === "number" &&
    typeof value.total_items === "number" &&
    typeof value.total_pages === "number"
  );
}

export function normalizeAssignment(
  value: unknown,
): ScheduleAssignment | null {
  if (!isRecord(value) || !isEmployeeLike(value.user) || !isWorkSchedule(value.schedule)) {
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
    schedule: value.schedule,
    effective_from: value.effective_from,
    effective_to: value.effective_to,
    status,
    created_at: value.created_at,
    updated_at: value.updated_at,
  };
}

export function isScheduleAssignmentListResponse(
  value: unknown,
): value is ScheduleAssignmentListResponse {
  if (!isRecord(value) || !Array.isArray(value.items)) {
    return false;
  }
  const items = value.items.map(normalizeAssignment);
  return (
    items.every((item): item is ScheduleAssignment => item !== null) &&
    typeof value.page === "number" &&
    typeof value.page_size === "number" &&
    typeof value.total_items === "number" &&
    typeof value.total_pages === "number"
  );
}

export function normalizeAssignmentListResponse(
  value: unknown,
): ScheduleAssignmentListResponse | null {
  if (!isRecord(value) || !Array.isArray(value.items)) {
    return null;
  }
  const items = value.items.map(normalizeAssignment);
  if (
    !items.every((item): item is ScheduleAssignment => item !== null) ||
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

export function deriveAssignmentStatus(
  effectiveFrom: string,
  effectiveTo: string | null,
): AssignmentStatus {
  const today = localDateKey(new Date());
  if (effectiveFrom > today) {
    return "UPCOMING";
  }
  if (effectiveTo !== null && effectiveTo < today) {
    return "ENDED";
  }
  return "CURRENT";
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

function parseDateOnly(value: string): { year: number; month: number; day: number } | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(value);
  if (!match) {
    return null;
  }
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const date = new Date(year, month - 1, day);
  if (
    date.getFullYear() !== year ||
    date.getMonth() !== month - 1 ||
    date.getDate() !== day
  ) {
    return null;
  }
  return { year, month, day };
}

function localDateKey(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
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

function clampPositiveInt(value: number | undefined, fallback: number, max = Number.MAX_SAFE_INTEGER): number {
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
