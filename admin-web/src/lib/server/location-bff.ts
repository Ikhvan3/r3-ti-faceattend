import { NextResponse } from "next/server";

import { SafeApiError } from "@/lib/auth/types";
import type { ApiErrorResponse } from "@/lib/employee/types";
import { isRecord } from "@/lib/employee/utils";
import type {
  CreateLocationAssignmentRequest,
  CreateOfficeLocationRequest,
  EndLocationAssignmentRequest,
  UpdateOfficeLocationRequest,
  UpdateOfficeLocationStatusRequest,
} from "@/lib/location/types";
import {
  isDateOnly,
  parseLatitude,
  parseLongitude,
  parseRadiusMeters,
} from "@/lib/location/utils";

const MAX_LOCATION_BODY_BYTES = 20_000;

type BodyResult<T> =
  | {
      ok: true;
      value: T;
    }
  | {
      ok: false;
      response: NextResponse<ApiErrorResponse>;
    };

export async function readCreateOfficeLocationBody(
  request: Request,
): Promise<BodyResult<CreateOfficeLocationRequest>> {
  const payload = await readJsonBody(request);
  if (!payload.ok) {
    return payload;
  }
  if (!isOfficeLocationPayload(payload.value)) {
    return invalidBody("Data lokasi kantor wajib diisi dengan lengkap.");
  }

  return {
    ok: true,
    value: {
      name: normalizeRequiredName(payload.value.name),
      address: normalizeOptionalString(payload.value.address),
      latitude: payload.value.latitude,
      longitude: payload.value.longitude,
      radius_meters: payload.value.radius_meters,
    },
  };
}

export async function readUpdateOfficeLocationBody(
  request: Request,
): Promise<BodyResult<UpdateOfficeLocationRequest>> {
  return readCreateOfficeLocationBody(request);
}

export async function readUpdateOfficeLocationStatusBody(
  request: Request,
): Promise<BodyResult<UpdateOfficeLocationStatusRequest>> {
  const payload = await readJsonBody(request);
  if (!payload.ok) {
    return payload;
  }
  if (
    !isRecord(payload.value) ||
    !hasOnlyKeys(payload.value, ["is_active"]) ||
    typeof payload.value.is_active !== "boolean"
  ) {
    return invalidBody("Status lokasi kantor tidak valid.");
  }
  return { ok: true, value: { is_active: payload.value.is_active } };
}

export async function readCreateLocationAssignmentBody(
  request: Request,
): Promise<BodyResult<CreateLocationAssignmentRequest>> {
  const payload = await readJsonBody(request);
  if (!payload.ok) {
    return payload;
  }
  if (!isLocationAssignmentPayload(payload.value)) {
    return invalidBody("Data penugasan lokasi wajib diisi dengan lengkap.");
  }
  return {
    ok: true,
    value: {
      user_id: payload.value.user_id.trim(),
      office_location_id: payload.value.office_location_id.trim(),
      effective_from: payload.value.effective_from.trim(),
      effective_to:
        payload.value.effective_to === null
          ? null
          : payload.value.effective_to.trim(),
    },
  };
}

export async function readEndLocationAssignmentBody(
  request: Request,
): Promise<BodyResult<EndLocationAssignmentRequest>> {
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

export function locationBffError(
  error: unknown,
): NextResponse<ApiErrorResponse> {
  if (error instanceof SafeApiError) {
    if ([400, 401, 403, 404, 409].includes(error.status)) {
      return jsonError(error.message, error.status);
    }
    return jsonError("Layanan lokasi belum tersedia. Coba lagi nanti.", 503);
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
  if (contentLength > MAX_LOCATION_BODY_BYTES) {
    return invalidBody("Request terlalu besar.");
  }

  let rawBody = "";
  try {
    rawBody = await request.text();
  } catch {
    return invalidBody("Request tidak valid.");
  }
  if (rawBody.length > MAX_LOCATION_BODY_BYTES) {
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

function isOfficeLocationPayload(
  value: unknown,
): value is CreateOfficeLocationRequest {
  return (
    isRecord(value) &&
    hasOnlyKeys(value, [
      "name",
      "address",
      "latitude",
      "longitude",
      "radius_meters",
    ]) &&
    typeof value.name === "string" &&
    normalizeRequiredName(value.name) !== "" &&
    (typeof value.address === "string" || value.address === null) &&
    typeof value.latitude === "number" &&
    parseLatitude(String(value.latitude)) !== null &&
    typeof value.longitude === "number" &&
    parseLongitude(String(value.longitude)) !== null &&
    typeof value.radius_meters === "number" &&
    parseRadiusMeters(String(value.radius_meters)) !== null
  );
}

function isLocationAssignmentPayload(
  value: unknown,
): value is CreateLocationAssignmentRequest {
  return (
    isRecord(value) &&
    hasOnlyKeys(value, [
      "user_id",
      "office_location_id",
      "effective_from",
      "effective_to",
    ]) &&
    requiredString(value.user_id) &&
    requiredString(value.office_location_id) &&
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

function normalizeRequiredName(value: string): string {
  return value.trim().split(/\s+/).filter(Boolean).join(" ");
}

function normalizeOptionalString(value: string | null): string | null {
  const normalized = value?.trim() ?? "";
  return normalized === "" ? null : normalized;
}
