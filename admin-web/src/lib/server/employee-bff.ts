import { NextResponse } from "next/server";

import { SafeApiError } from "@/lib/auth/types";
import type { AccountStatus } from "@/lib/auth/types";
import type {
  ApiErrorResponse,
  CreateEmployeeRequest,
  UpdateEmployeeRequest,
  UpdateEmployeeStatusRequest,
} from "@/lib/employee/types";
import { isAccountStatus, isRecord } from "@/lib/employee/utils";

const MAX_EMPLOYEE_BODY_BYTES = 20_000;

type BodyResult<T> =
  | {
      ok: true;
      value: T;
    }
  | {
      ok: false;
      response: NextResponse<ApiErrorResponse>;
    };

export async function readCreateEmployeeBody(
  request: Request,
): Promise<BodyResult<CreateEmployeeRequest>> {
  const payload = await readJsonBody(request);
  if (!payload.ok) {
    return payload;
  }

  if (!isCreatePayload(payload.value)) {
    return invalidBody("Data pegawai wajib diisi dengan lengkap.");
  }

  return {
    ok: true,
    value: {
      employee_number: payload.value.employee_number.trim(),
      name: payload.value.name.trim(),
      email: payload.value.email.trim(),
      initial_password: payload.value.initial_password,
      phone: optionalString(payload.value.phone),
      position: optionalString(payload.value.position),
    },
  };
}

export async function readUpdateEmployeeBody(
  request: Request,
): Promise<BodyResult<UpdateEmployeeRequest>> {
  const payload = await readJsonBody(request);
  if (!payload.ok) {
    return payload;
  }

  if (!isUpdatePayload(payload.value)) {
    return invalidBody("Data pegawai wajib diisi dengan lengkap.");
  }

  return {
    ok: true,
    value: {
      employee_number: payload.value.employee_number.trim(),
      name: payload.value.name.trim(),
      email: payload.value.email.trim(),
      phone: optionalString(payload.value.phone),
      position: optionalString(payload.value.position),
    },
  };
}

export async function readUpdateStatusBody(
  request: Request,
): Promise<BodyResult<UpdateEmployeeStatusRequest>> {
  const payload = await readJsonBody(request);
  if (!payload.ok) {
    return payload;
  }

  if (!isStatusPayload(payload.value)) {
    return invalidBody("Status pegawai tidak valid.");
  }

  return {
    ok: true,
    value: {
      account_status: payload.value.account_status,
    },
  };
}

export function employeeBffError(error: unknown): NextResponse<ApiErrorResponse> {
  if (error instanceof SafeApiError) {
    if ([400, 401, 403, 404, 409].includes(error.status)) {
      return jsonError(error.message, error.status);
    }

    return jsonError("Layanan pegawai belum tersedia. Coba lagi nanti.", 503);
  }

  return jsonError("Terjadi kesalahan. Coba lagi nanti.", 500);
}

export function jsonError(
  message: string,
  status: number,
): NextResponse<ApiErrorResponse> {
  return NextResponse.json({ status: "error", message }, { status });
}

async function readJsonBody(request: Request): Promise<BodyResult<unknown>> {
  const contentLength = Number(request.headers.get("content-length") ?? "0");
  if (contentLength > MAX_EMPLOYEE_BODY_BYTES) {
    return invalidBody("Request terlalu besar.");
  }

  let rawBody = "";
  try {
    rawBody = await request.text();
  } catch {
    return invalidBody("Request tidak valid.");
  }

  if (rawBody.length > MAX_EMPLOYEE_BODY_BYTES) {
    return invalidBody("Request terlalu besar.");
  }

  try {
    return { ok: true, value: JSON.parse(rawBody) as unknown };
  } catch {
    return invalidBody("Request tidak valid.");
  }
}

function invalidBody(message: string): BodyResult<never> {
  return {
    ok: false,
    response: jsonError(message, 400),
  };
}

function isCreatePayload(
  value: unknown,
): value is CreateEmployeeRequest {
  if (!isRecord(value) || !hasOnlyKeys(value, [
    "employee_number",
    "name",
    "email",
    "initial_password",
    "phone",
    "position",
  ])) {
    return false;
  }

  return (
    requiredString(value.employee_number) &&
    requiredString(value.name) &&
    requiredString(value.email) &&
    typeof value.initial_password === "string" &&
    value.initial_password.length >= 8 &&
    optionalStringLike(value.phone) &&
    optionalStringLike(value.position)
  );
}

function isUpdatePayload(
  value: unknown,
): value is UpdateEmployeeRequest {
  if (!isRecord(value) || !hasOnlyKeys(value, [
    "employee_number",
    "name",
    "email",
    "phone",
    "position",
  ])) {
    return false;
  }

  return (
    requiredString(value.employee_number) &&
    requiredString(value.name) &&
    requiredString(value.email) &&
    optionalStringLike(value.phone) &&
    optionalStringLike(value.position)
  );
}

function isStatusPayload(
  value: unknown,
): value is { account_status: AccountStatus } {
  return (
    isRecord(value) &&
    hasOnlyKeys(value, ["account_status"]) &&
    isAccountStatus(value.account_status)
  );
}

function hasOnlyKeys(
  value: Record<string, unknown>,
  keys: string[],
): boolean {
  return Object.keys(value).every((key) => keys.includes(key));
}

function requiredString(value: unknown): value is string {
  return typeof value === "string" && value.trim() !== "";
}

function optionalStringLike(value: unknown): value is string | null {
  return typeof value === "string" || value === null;
}

function optionalString(value: string | null): string | null {
  const normalized = value?.trim() ?? "";
  return normalized === "" ? null : normalized;
}
