import { NextResponse } from "next/server";

import {
  readUpdateWorkScheduleStatusBody,
  scheduleBffError,
} from "@/lib/server/schedule-bff";
import { updateWorkScheduleStatusWithSession } from "@/lib/server/go-api";

export async function PATCH(
  request: Request,
  ctx: { params: Promise<{ id: string }> },
): Promise<NextResponse> {
  const { id } = await ctx.params;
  const body = await readUpdateWorkScheduleStatusBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const schedule = await updateWorkScheduleStatusWithSession(id, body.value);
    return NextResponse.json({
      status: "ok",
      message: "Status jadwal kerja berhasil diperbarui.",
      data: schedule,
    });
  } catch (error) {
    return scheduleBffError(error);
  }
}
