import { NextResponse } from "next/server";

import {
  locationBffError,
  readUpdateOfficeLocationBody,
} from "@/lib/server/location-bff";
import { updateOfficeLocationWithSession } from "@/lib/server/go-api";

export async function PUT(
  request: Request,
  { params }: { params: Promise<{ id: string }> },
): Promise<NextResponse> {
  const body = await readUpdateOfficeLocationBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const { id } = await params;
    const location = await updateOfficeLocationWithSession(id, body.value);
    return NextResponse.json({
      status: "ok",
      message: "Lokasi kantor berhasil diperbarui.",
      data: location,
    });
  } catch (error) {
    return locationBffError(error);
  }
}
