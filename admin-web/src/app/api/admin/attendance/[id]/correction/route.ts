import { NextResponse } from "next/server";

import type { AdminAttendanceCorrectionInput } from "@/lib/attendance/types";
import { SafeApiError } from "@/lib/auth/types";
import { correctAdminAttendance } from "@/lib/server/admin-attendance-bff";

export async function PATCH(
  request: Request,
  ctx: { params: Promise<{ id: string }> },
): Promise<NextResponse> {
  const { id } = await ctx.params;
  const body = (await request.json().catch(() => null)) as unknown;
  const input = parseCorrection(body);

  if (!input) {
    return NextResponse.json(
      { status: "error", message: "Data koreksi presensi tidak valid." },
      { status: 400 },
    );
  }

  try {
    const attendance = await correctAdminAttendance(id, input);
    return NextResponse.json({
      status: "ok",
      message: "Koreksi presensi berhasil disimpan.",
      data: attendance,
    });
  } catch (error) {
    if (error instanceof SafeApiError) {
      return NextResponse.json(
        { status: "error", message: error.message },
        { status: error.status },
      );
    }
    return NextResponse.json(
      { status: "error", message: "Koreksi presensi gagal diproses." },
      { status: 500 },
    );
  }
}

function parseCorrection(value: unknown): AdminAttendanceCorrectionInput | null {
  if (typeof value !== "object" || value === null) return null;
  const record = value as Record<string, unknown>;
  if (typeof record.check_in_time !== "string") return null;
  if (
    record.check_out_time !== null &&
    typeof record.check_out_time !== "string"
  ) {
    return null;
  }
  if (typeof record.reason !== "string") return null;

  const checkIn = record.check_in_time.trim();
  const checkOut =
    typeof record.check_out_time === "string"
      ? record.check_out_time.trim() || null
      : null;
  const reason = record.reason.trim();

  if (!/^\d{2}:\d{2}$/.test(checkIn)) return null;
  if (checkOut !== null && !/^\d{2}:\d{2}$/.test(checkOut)) return null;
  if (reason.length < 5 || reason.length > 1000) return null;

  return {
    check_in_time: checkIn,
    check_out_time: checkOut,
    reason,
  };
}
