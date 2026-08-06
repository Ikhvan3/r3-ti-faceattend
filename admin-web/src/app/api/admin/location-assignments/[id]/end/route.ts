import { NextResponse } from "next/server";

import {
  locationBffError,
  readEndLocationAssignmentBody,
} from "@/lib/server/location-bff";
import { endLocationAssignmentWithSession } from "@/lib/server/go-api";

export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string }> },
): Promise<NextResponse> {
  const body = await readEndLocationAssignmentBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const { id } = await params;
    const assignment = await endLocationAssignmentWithSession(id, body.value);
    return NextResponse.json({
      status: "ok",
      message: "Penugasan lokasi berhasil diakhiri.",
      data: assignment,
    });
  } catch (error) {
    return locationBffError(error);
  }
}
