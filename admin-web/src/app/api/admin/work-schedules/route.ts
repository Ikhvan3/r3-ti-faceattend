import { NextResponse } from "next/server";

import {
  readCreateWorkScheduleBody,
  scheduleBffError,
} from "@/lib/server/schedule-bff";
import { createWorkScheduleWithSession } from "@/lib/server/go-api";

export async function POST(request: Request): Promise<NextResponse> {
  const body = await readCreateWorkScheduleBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const schedule = await createWorkScheduleWithSession(body.value);
    return NextResponse.json(
      {
        status: "ok",
        message: "Jadwal kerja berhasil dibuat.",
        data: schedule,
      },
      { status: 201 },
    );
  } catch (error) {
    return scheduleBffError(error);
  }
}
