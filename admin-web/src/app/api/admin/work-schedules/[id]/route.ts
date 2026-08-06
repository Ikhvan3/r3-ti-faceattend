import { NextResponse } from "next/server";

import {
  readUpdateWorkScheduleBody,
  scheduleBffError,
} from "@/lib/server/schedule-bff";
import { updateWorkScheduleWithSession } from "@/lib/server/go-api";

export async function PUT(
  request: Request,
  ctx: { params: Promise<{ id: string }> },
): Promise<NextResponse> {
  const { id } = await ctx.params;
  const body = await readUpdateWorkScheduleBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const schedule = await updateWorkScheduleWithSession(id, body.value);
    return NextResponse.json({
      status: "ok",
      message: "Jadwal kerja berhasil diperbarui.",
      data: schedule,
    });
  } catch (error) {
    return scheduleBffError(error);
  }
}
