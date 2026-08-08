import type {
  AdminAttendanceDetail,
  AdminAttendanceListItem,
  AdminAttendanceListQuery,
  AdminAttendanceListResponse,
  AdminAttendanceState,
  AdminAttendanceSummary,
} from "@/lib/attendance/types";
import { isRecord } from "@/lib/employee/utils";

const ATTENDANCE_STATES: AdminAttendanceState[] = [
  "NOT_CHECKED_IN",
  "CHECKED_IN",
  "CHECKED_OUT",
];

export type AttendanceSearchQuery = {
  date_from: string;
  date_to: string;
  search: string;
  attendance_state: AdminAttendanceState | "";
  is_late: "" | "true" | "false";
  page: number;
  page_size: number;
};

export function parseAttendanceSearchParams(
  params: Record<string, string | string[] | undefined>,
): AttendanceSearchQuery {
  const stateValue = first(params.attendance_state);
  const lateValue = first(params.is_late);
  return {
    date_from: first(params.date_from),
    date_to: first(params.date_to),
    search: first(params.search),
    attendance_state: isAttendanceState(stateValue) ? stateValue : "",
    is_late: lateValue === "true" || lateValue === "false" ? lateValue : "",
    page: positiveInt(first(params.page), 1),
    page_size: Math.min(positiveInt(first(params.page_size), 20), 100),
  };
}

export function buildAttendanceQueryParams(
  query: AdminAttendanceListQuery,
): string {
  const params = new URLSearchParams();
  if (query.date_from) params.set("date_from", query.date_from);
  if (query.date_to) params.set("date_to", query.date_to);
  if (query.employee_id) params.set("employee_id", query.employee_id);
  if (query.search) params.set("search", query.search);
  if (query.attendance_state)
    params.set("attendance_state", query.attendance_state);
  if (query.is_late !== undefined)
    params.set("is_late", String(query.is_late));
  if (query.page) params.set("page", String(query.page));
  if (query.page_size) params.set("page_size", String(query.page_size));
  return params.toString();
}

export function isAttendanceSummary(
  value: unknown,
): value is AdminAttendanceSummary {
  return (
    isRecord(value) &&
    typeof value.date === "string" &&
    nonNegativeNumber(value.active_employees) &&
    nonNegativeNumber(value.checked_in) &&
    nonNegativeNumber(value.checked_out) &&
    nonNegativeNumber(value.not_checked_in) &&
    nonNegativeNumber(value.late)
  );
}

export function isAttendanceListResponse(
  value: unknown,
): value is AdminAttendanceListResponse {
  return (
    isRecord(value) &&
    Array.isArray(value.items) &&
    value.items.every(isAttendanceListItem) &&
    nonNegativeNumber(value.page) &&
    nonNegativeNumber(value.page_size) &&
    nonNegativeNumber(value.total_items) &&
    nonNegativeNumber(value.total_pages)
  );
}

export function isAttendanceDetail(
  value: unknown,
): value is AdminAttendanceDetail {
  return (
    isAttendanceListItem(value) &&
    typeof value.id === "string" &&
    typeof value.check_in_at === "string" &&
    "check_in_location" in value &&
    "check_out_location" in value &&
    (value.check_in_location === null ||
      isLocationEvidence(value.check_in_location)) &&
    (value.check_out_location === null ||
      isLocationEvidence(value.check_out_location))
  );
}

export function attendanceStateLabel(state: AdminAttendanceState): string {
  switch (state) {
    case "NOT_CHECKED_IN":
      return "Belum check-in";
    case "CHECKED_IN":
      return "Sudah check-in";
    case "CHECKED_OUT":
      return "Sudah check-out";
  }
}

export function formatBusinessDate(date: string): string {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(date)) return date;
  const [year, month, day] = date.split("-").map(Number);
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeZone: "Asia/Jakarta",
  }).format(new Date(Date.UTC(year, month - 1, day, 12)));
}

export function formatBusinessTime(value: string | null): string {
  if (!value) return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone: "Asia/Jakarta",
  }).format(date);
}

export function formatBusinessTimeInput(value: string | null): string | null {
  if (!value) return null;
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;
  const parts = new Intl.DateTimeFormat("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone: "Asia/Jakarta",
  }).formatToParts(date);
  const hour = parts.find((part) => part.type === "hour")?.value;
  const minute = parts.find((part) => part.type === "minute")?.value;
  return hour && minute ? `${hour}:${minute}` : null;
}

function isAttendanceListItem(
  value: unknown,
): value is AdminAttendanceListItem {
  if (!isRecord(value) || !isRecord(value.employee) || !isRecord(value.schedule)) {
    return false;
  }
  return (
    (value.id === null || typeof value.id === "string") &&
    typeof value.attendance_date === "string" &&
    typeof value.employee.id === "string" &&
    typeof value.employee.employee_number === "string" &&
    typeof value.employee.name === "string" &&
    typeof value.employee.email === "string" &&
    (value.employee.position === null ||
      typeof value.employee.position === "string") &&
    typeof value.schedule.id === "string" &&
    typeof value.schedule.name === "string" &&
    typeof value.schedule.start_time === "string" &&
    typeof value.schedule.end_time === "string" &&
    typeof value.schedule.grace_minutes === "number" &&
    typeof value.schedule.is_active === "boolean" &&
    (value.check_in_at === null || typeof value.check_in_at === "string") &&
    (value.check_out_at === null || typeof value.check_out_at === "string") &&
    isAttendanceState(value.attendance_state) &&
    typeof value.is_late === "boolean"
  );
}

function isLocationEvidence(value: unknown): boolean {
  return (
    isRecord(value) &&
    typeof value.office_location_id === "string" &&
    typeof value.office_location_name === "string" &&
    typeof value.radius_meters === "number" &&
    typeof value.latitude === "number" &&
    typeof value.longitude === "number" &&
    typeof value.accuracy_meters === "number" &&
    typeof value.distance_meters === "number" &&
    typeof value.inside_geofence === "boolean"
  );
}

function isAttendanceState(value: unknown): value is AdminAttendanceState {
  return (
    typeof value === "string" &&
    ATTENDANCE_STATES.includes(value as AdminAttendanceState)
  );
}

function first(value: string | string[] | undefined): string {
  return Array.isArray(value) ? value[0] ?? "" : value ?? "";
}

function positiveInt(value: string, fallback: number): number {
  const parsed = Number.parseInt(value, 10);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}

function nonNegativeNumber(value: unknown): value is number {
  return typeof value === "number" && Number.isFinite(value) && value >= 0;
}
