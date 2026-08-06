import { NextResponse } from "next/server";

import {
  readCreateAssignmentBody,
  scheduleBffError,
} from "@/lib/server/schedule-bff";
import { createScheduleAssignmentWithSession } from "@/lib/server/go-api";

export async function POST(request: Request): Promise<NextResponse> {
  const body = await readCreateAssignmentBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const assignment = await createScheduleAssignmentWithSession(body.value);
    return NextResponse.json(
      {
        status: "ok",
        message: "Penugasan jadwal berhasil dibuat.",
        data: assignment,
      },
      { status: 201 },
    );
  } catch (error) {
    return scheduleBffError(error);
  }
}
