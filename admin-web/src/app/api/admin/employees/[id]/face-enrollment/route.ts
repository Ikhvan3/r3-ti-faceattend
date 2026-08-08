import { NextResponse } from "next/server";

import { SafeApiError } from "@/lib/auth/types";
import { resetEmployeeFaceEnrollmentWithSession } from "@/lib/server/admin-face-enrollment-bff";

export async function DELETE(
  _request: Request,
  ctx: { params: Promise<{ id: string }> },
): Promise<NextResponse> {
  const { id } = await ctx.params;

  try {
    await resetEmployeeFaceEnrollmentWithSession(id);
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
