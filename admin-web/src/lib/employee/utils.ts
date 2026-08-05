import type { AccountStatus } from "@/lib/auth/types";
import type { Employee, EmployeeListQuery } from "@/lib/employee/types";

export const EMPLOYEE_STATUSES: AccountStatus[] = [
  "ACTIVE",
  "INACTIVE",
  "SUSPENDED",
];

export const DEFAULT_EMPLOYEE_PAGE = 1;
export const DEFAULT_EMPLOYEE_PAGE_SIZE = 10;
export const MAX_EMPLOYEE_PAGE_SIZE = 100;

export type ParsedEmployeeSearchParams = {
  page: number;
  page_size: number;
  search: string;
  status: AccountStatus | "";
};

export function isAccountStatus(value: unknown): value is AccountStatus {
  return (
    value === "ACTIVE" || value === "INACTIVE" || value === "SUSPENDED"
  );
}

export function parseEmployeeSearchParams(
  searchParams: Record<string, string | string[] | undefined>,
): ParsedEmployeeSearchParams {
  const statusValue = firstParam(searchParams.status).trim().toUpperCase();

  return {
    page: parsePositiveInt(
      firstParam(searchParams.page),
      DEFAULT_EMPLOYEE_PAGE,
      Number.MAX_SAFE_INTEGER,
    ),
    page_size: parsePositiveInt(
      firstParam(searchParams.page_size),
      DEFAULT_EMPLOYEE_PAGE_SIZE,
      MAX_EMPLOYEE_PAGE_SIZE,
    ),
    search: firstParam(searchParams.search).trim(),
    status: isAccountStatus(statusValue) ? statusValue : "",
  };
}

export function buildEmployeeQueryParams(query: EmployeeListQuery): string {
  const params = new URLSearchParams();
  const page = clampPositiveInt(query.page, DEFAULT_EMPLOYEE_PAGE);
  const pageSize = clampPositiveInt(
    query.page_size,
    DEFAULT_EMPLOYEE_PAGE_SIZE,
    MAX_EMPLOYEE_PAGE_SIZE,
  );

  params.set("page", String(page));
  params.set("page_size", String(pageSize));

  const search = query.search?.trim();
  if (search) {
    params.set("search", search);
  }

  if (query.status && isAccountStatus(query.status)) {
    params.set("status", query.status);
  }

  return params.toString();
}

export function employeeListHref(
  query: ParsedEmployeeSearchParams,
  page: number,
): string {
  const params = new URLSearchParams();
  params.set("page", String(Math.max(1, page)));
  params.set("page_size", String(query.page_size));

  if (query.search) {
    params.set("search", query.search);
  }
  if (query.status) {
    params.set("status", query.status);
  }

  return `/employees?${params.toString()}`;
}

export function statusLabel(status: AccountStatus): string {
  switch (status) {
    case "ACTIVE":
      return "Aktif";
    case "INACTIVE":
      return "Nonaktif";
    case "SUSPENDED":
      return "Ditangguhkan";
  }
}

export function statusDescription(status: AccountStatus): string {
  switch (status) {
    case "ACTIVE":
      return "Pegawai dapat menggunakan akun kembali setelah status aktif.";
    case "INACTIVE":
      return "Pegawai tidak dapat menggunakan akun sampai diaktifkan lagi.";
    case "SUSPENDED":
      return "Akun pegawai ditangguhkan untuk pemeriksaan admin.";
  }
}

export function statusBadgeClass(status: AccountStatus): string {
  switch (status) {
    case "ACTIVE":
      return "border-emerald-200 bg-emerald-50 text-emerald-700";
    case "INACTIVE":
      return "border-slate-200 bg-slate-100 text-slate-700";
    case "SUSPENDED":
      return "border-amber-200 bg-amber-50 text-amber-800";
  }
}

export function safeEmployeeApiMessage(status: number, fallback: string): string {
  if (status === 400) {
    return "Request tidak valid. Periksa kembali data pegawai.";
  }
  if (status === 401) {
    return "Session tidak valid. Silakan login ulang.";
  }
  if (status === 403) {
    return "Akses admin diperlukan untuk mengelola pegawai.";
  }
  if (status === 404) {
    return "Pegawai tidak ditemukan.";
  }
  if (status === 409) {
    return fallback;
  }
  if (status === 503) {
    return "Layanan backend belum tersedia. Coba lagi nanti.";
  }

  return "Terjadi kesalahan. Coba lagi nanti.";
}

export function isEmployee(value: unknown): value is Employee {
  if (!isRecord(value)) {
    return false;
  }

  return (
    typeof value.id === "string" &&
    typeof value.employee_number === "string" &&
    typeof value.name === "string" &&
    typeof value.email === "string" &&
    (typeof value.phone === "string" || value.phone === null) &&
    (typeof value.position === "string" || value.position === null) &&
    (value.role === "ADMIN" || value.role === "USER") &&
    isAccountStatus(value.account_status) &&
    typeof value.created_at === "string" &&
    typeof value.updated_at === "string"
  );
}

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function firstParam(value: string | string[] | undefined): string {
  if (Array.isArray(value)) {
    return value[0] ?? "";
  }

  return value ?? "";
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
  if (!Number.isInteger(value) || value === undefined || value < 1) {
    return fallback;
  }

  return Math.min(value, max);
}
