import type {
  AuditAction,
  AuditEntityType,
  AuditLogItem,
  AuditLogListResponse,
  AuditLogQuery,
} from "@/lib/audit/types";

const ACTIONS: AuditAction[] = [
  "ATTENDANCE_CORRECTED",
  "FACE_ENROLLMENT_RESET",
];
const ENTITY_TYPES: AuditEntityType[] = ["ATTENDANCE_RECORD", "FACE_PROFILE"];

export function buildAuditQueryParams(query: AuditLogQuery): string {
  const params = new URLSearchParams();
  if (query.action) params.set("action", query.action);
  if (query.entity_type) params.set("entity_type", query.entity_type);
  if (query.entity_id) params.set("entity_id", query.entity_id);
  if (query.date_from) params.set("date_from", query.date_from);
  if (query.date_to) params.set("date_to", query.date_to);
  if (query.page) params.set("page", String(query.page));
  if (query.page_size) params.set("page_size", String(query.page_size));
  return params.toString();
}

export function isAuditLogListResponse(
  value: unknown,
): value is AuditLogListResponse {
  if (!isRecord(value) || !Array.isArray(value.items)) return false;
  return (
    value.items.every(isAuditLogItem) &&
    isNonNegativeNumber(value.page) &&
    isNonNegativeNumber(value.page_size) &&
    isNonNegativeNumber(value.total_items) &&
    isNonNegativeNumber(value.total_pages)
  );
}

export function auditActionLabel(action: AuditAction): string {
  switch (action) {
    case "ATTENDANCE_CORRECTED":
      return "Koreksi Presensi";
    case "FACE_ENROLLMENT_RESET":
      return "Reset Enrollment Wajah";
  }
}

export function auditEntityLabel(entity: AuditEntityType): string {
  switch (entity) {
    case "ATTENDANCE_RECORD":
      return "Presensi";
    case "FACE_PROFILE":
      return "Enrollment Wajah";
  }
}

export function formatAuditDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "Asia/Jakarta",
  }).format(date);
}

export function auditSnapshotSummary(
  item: AuditLogItem,
): { before: string; after: string } {
  if (item.action === "ATTENDANCE_CORRECTED") {
    return {
      before: attendanceSnapshot(item.before_data),
      after: attendanceSnapshot(item.after_data),
    };
  }
  return {
    before: stringField(item.before_data.status) ?? "-",
    after: stringField(item.after_data.status) ?? "-",
  };
}

function attendanceSnapshot(value: Record<string, unknown>): string {
  const checkIn = formatSnapshotTime(value.check_in_at);
  const checkOut = formatSnapshotTime(value.check_out_at);
  return `Masuk ${checkIn} · Pulang ${checkOut}`;
}

function formatSnapshotTime(value: unknown): string {
  if (typeof value !== "string") return "-";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat("id-ID", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone: "Asia/Jakarta",
  }).format(date);
}

function isAuditLogItem(value: unknown): value is AuditLogItem {
  if (!isRecord(value)) return false;
  return (
    typeof value.id === "string" &&
    (value.actor_user_id === null || typeof value.actor_user_id === "string") &&
    typeof value.actor_email === "string" &&
    typeof value.actor_role === "string" &&
    typeof value.action === "string" &&
    ACTIONS.includes(value.action as AuditAction) &&
    typeof value.entity_type === "string" &&
    ENTITY_TYPES.includes(value.entity_type as AuditEntityType) &&
    (value.entity_id === null || typeof value.entity_id === "string") &&
    (value.target_user_id === null || typeof value.target_user_id === "string") &&
    (value.target_label === null || typeof value.target_label === "string") &&
    typeof value.reason === "string" &&
    isRecord(value.before_data) &&
    isRecord(value.after_data) &&
    typeof value.created_at === "string"
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isNonNegativeNumber(value: unknown): value is number {
  return typeof value === "number" && Number.isFinite(value) && value >= 0;
}

function stringField(value: unknown): string | null {
  return typeof value === "string" ? value : null;
}
