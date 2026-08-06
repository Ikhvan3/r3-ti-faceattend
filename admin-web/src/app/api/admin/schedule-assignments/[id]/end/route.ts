import { NextResponse } from "next/server";

import {
  readEndAssignmentBody,
  scheduleBffError,
} from "@/lib/server/schedule-bff";
import { endScheduleAssignmentWithSession } from "@/lib/server/go-api";

export async function PATCH(
  request: Request,
  ctx: { params: Promise<{ id: string }> },
): Promise<NextResponse> {
  const { id } = await ctx.params;
  const body = await readEndAssignmentBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const assignment = await endScheduleAssignmentWithSession(id, body.value);
    return NextResponse.json({
      status: "ok",
      message: "Penugasan jadwal berhasil diakhiri.",
      data: assignment,
    });
  } catch (error) {
    return scheduleBffError(error);
  }
}
