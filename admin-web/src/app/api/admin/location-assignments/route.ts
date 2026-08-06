import { NextResponse } from "next/server";

import {
  locationBffError,
  readCreateLocationAssignmentBody,
} from "@/lib/server/location-bff";
import { createLocationAssignmentWithSession } from "@/lib/server/go-api";

export async function POST(request: Request): Promise<NextResponse> {
  const body = await readCreateLocationAssignmentBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const assignment = await createLocationAssignmentWithSession(body.value);
    return NextResponse.json(
      {
        status: "ok",
        message: "Penugasan lokasi berhasil dibuat.",
        data: assignment,
      },
      { status: 201 },
    );
  } catch (error) {
    return locationBffError(error);
  }
}
