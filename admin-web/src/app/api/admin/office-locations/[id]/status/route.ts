import { NextResponse } from "next/server";

import {
  locationBffError,
  readUpdateOfficeLocationStatusBody,
} from "@/lib/server/location-bff";
import { updateOfficeLocationStatusWithSession } from "@/lib/server/go-api";

export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string }> },
): Promise<NextResponse> {
  const body = await readUpdateOfficeLocationStatusBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const { id } = await params;
    const location = await updateOfficeLocationStatusWithSession(id, body.value);
    return NextResponse.json({
      status: "ok",
      message: "Status lokasi kantor berhasil diperbarui.",
      data: location,
    });
  } catch (error) {
    return locationBffError(error);
  }
}
