import { NextResponse } from "next/server";

import { SafeApiError } from "@/lib/auth/types";
import { resetEmployeeFaceEnrollmentWithSession } from "@/lib/server/admin-face-enrollment-bff";

export async function DELETE(
  request: Request,
  ctx: { params: Promise<{ id: string }> },
): Promise<NextResponse> {
  const { id } = await ctx.params;
  const body = (await request.json().catch(() => null)) as unknown;
  const reason = readReason(body);

  if (!reason || reason.length < 5 || reason.length > 1000) {
    return NextResponse.json(
      {
        status: "error",
        message: "Alasan reset enrollment wajib diisi minimal 5 karakter.",
      },
      { status: 400 },
    );
  }

  try {
    await resetEmployeeFaceEnrollmentWithSession(id, reason);
    return NextResponse.json({
      status: "ok",
      message: "Enrollment wajah berhasil direset.",
    });
  } catch (error) {
    if (error instanceof SafeApiError) {
      return NextResponse.json(
        { status: "error", message: error.message },
        { status: error.status },
      );
    }

    return NextResponse.json(
      { status: "error", message: "Reset enrollment wajah gagal diproses." },
      { status: 500 },
    );
  }
}

function readReason(value: unknown): string | null {
  if (typeof value !== "object" || value === null) return null;
  const reason = (value as Record<string, unknown>).reason;
  return typeof reason === "string" ? reason.trim() : null;
}
