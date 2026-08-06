import { NextResponse } from "next/server";

import { SafeApiError } from "@/lib/auth/types";
import type { ApiErrorResponse } from "@/lib/employee/types";
import type {
  CreateScheduleAssignmentRequest,
  CreateWorkScheduleRequest,
  EndScheduleAssignmentRequest,
  UpdateWorkScheduleRequest,
  UpdateWorkScheduleStatusRequest,
} from "@/lib/schedule/types";
import { isDateOnly } from "@/lib/schedule/utils";
import { isRecord } from "@/lib/employee/utils";

const MAX_SCHEDULE_BODY_BYTES = 20_000;

type BodyResult<T> =
  | {
      ok: true;
      value: T;
    }
  | {
      ok: false;
      response: NextResponse<ApiErrorResponse>;
    };

export async function readCreateWorkScheduleBody(
  request: Request,
): Promise<BodyResult<CreateWorkScheduleRequest>> {
  const payload = await readJsonBody(request);
  if (!payload.ok) {
    return payload;
  }
  if (!isWorkSchedulePayload(payload.value)) {
    return invalidBody("Data jadwal kerja wajib diisi dengan lengkap.");
  }

  return {
    ok: true,
    value: normalizeWorkSchedulePayload(payload.value),
  };
}

export async function readUpdateWorkScheduleBody(
  request: Request,
): Promise<BodyResult<UpdateWorkScheduleRequest>> {
  return readCreateWorkScheduleBody(request);
}

export async function readUpdateWorkScheduleStatusBody(
  request: Request,
): Promise<BodyResult<UpdateWorkScheduleStatusRequest>> {
  const payload = await readJsonBody(request);
  if (!payload.ok) {
    return payload;
  }
  if (
    !isRecord(payload.value) ||
    !hasOnlyKeys(payload.value, ["is_active"]) ||
    typeof payload.value.is_active !== "boolean"
  ) {
    return invalidBody("Status jadwal kerja tidak valid.");
  }
  return { ok: true, value: { is_active: payload.value.is_active } };
}

export async function readCreateAssignmentBody(
  request: Request,
): Promise<BodyResult<CreateScheduleAssignmentRequest>> {
  const payload = await readJsonBody(request);
  if (!payload.ok) {
    return payload;
  }
  if (!isAssignmentPayload(payload.value)) {
    return invalidBody("Data penugasan jadwal wajib diisi dengan lengkap.");
  }

  return {
    ok: true,
    value: {
      user_id: payload.value.user_id.trim(),
      schedule_id: payload.value.schedule_id.trim(),
      effective_from: payload.value.effective_from.trim(),
      effective_to:
        payload.value.effective_to === null
          ? null
          : payload.value.effective_to.trim(),
    },
  };
}

export async function readEndAssignmentBody(
  request: Request,
): Promise<BodyResult<EndScheduleAssignmentRequest>> {
  const payload = await readJsonBody(request);
  if (!payload.ok) {
    return payload;
  }
  if (
    !isRecord(payload.value) ||
    !hasOnlyKeys(payload.value, ["effective_to"]) ||
    typeof payload.value.effective_to !== "string" ||
    !isDateOnly(payload.value.effective_to.trim())
  ) {
    return invalidBody("Tanggal akhir penugasan tidak valid.");
  }

  return {
    ok: true,
    value: { effective_to: payload.value.effective_to.trim() },
  };
}

export function scheduleBffError(
  error: unknown,
): NextResponse<ApiErrorResponse> {
  if (error instanceof SafeApiError) {
    if ([400, 401, 403, 404, 409].includes(error.status)) {
      return jsonError(error.message, error.status);
    }
    return jsonError("Layanan jadwal belum tersedia. Coba lagi nanti.", 503);
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
  if (contentLength > MAX_SCHEDULE_BODY_BYTES) {
    return invalidBody("Request terlalu besar.");
  }

  let rawBody = "";
  try {
    rawBody = await request.text();
  } catch {
    return invalidBody("Request tidak valid.");
  }

  if (rawBody.length > MAX_SCHEDULE_BODY_BYTES) {
    return invalidBody("Request terlalu besar.");
  }

  try {
    return { ok: true, value: JSON.parse(rawBody) as unknown };
  } catch {
    return invalidBody("Request tidak valid.");
  }
}

function invalidBody(message: string): BodyResult<never> {
  return { ok: false, response: jsonError(message, 400) };
}

function isWorkSchedulePayload(
  value: unknown,
): value is CreateWorkScheduleRequest {
  return (
    isRecord(value) &&
    hasOnlyKeys(value, ["name", "start_time", "end_time", "grace_minutes"]) &&
    requiredString(value.name) &&
    isTime(value.start_time) &&
    isTime(value.end_time) &&
    typeof value.grace_minutes === "number" &&
    Number.isInteger(value.grace_minutes) &&
    value.grace_minutes >= 0 &&
    value.grace_minutes <= 240 &&
    value.end_time > value.start_time
  );
}

function normalizeWorkSchedulePayload(
  value: CreateWorkScheduleRequest,
): CreateWorkScheduleRequest {
  return {
    name: value.name.trim(),
    start_time: value.start_time.trim(),
    end_time: value.end_time.trim(),
    grace_minutes: value.grace_minutes,
  };
}

function isAssignmentPayload(
  value: unknown,
): value is CreateScheduleAssignmentRequest {
  return (
    isRecord(value) &&
    hasOnlyKeys(value, [
      "user_id",
      "schedule_id",
      "effective_from",
      "effective_to",
    ]) &&
    requiredString(value.user_id) &&
    requiredString(value.schedule_id) &&
    typeof value.effective_from === "string" &&
    isDateOnly(value.effective_from.trim()) &&
    (value.effective_to === null ||
      (typeof value.effective_to === "string" &&
        isDateOnly(value.effective_to.trim()))) &&
    (value.effective_to === null ||
      value.effective_to.trim() >= value.effective_from.trim())
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

function isTime(value: unknown): value is string {
  return typeof value === "string" && /^\d{2}:\d{2}$/.test(value.trim());
}
