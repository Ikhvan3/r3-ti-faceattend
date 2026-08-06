import { NextResponse } from "next/server";

import {
  locationBffError,
  readCreateOfficeLocationBody,
} from "@/lib/server/location-bff";
import { createOfficeLocationWithSession } from "@/lib/server/go-api";

export async function POST(request: Request): Promise<NextResponse> {
  const body = await readCreateOfficeLocationBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const location = await createOfficeLocationWithSession(body.value);
    return NextResponse.json(
      {
        status: "ok",
        message: "Lokasi kantor berhasil dibuat.",
        data: location,
      },
      { status: 201 },
    );
  } catch (error) {
    return locationBffError(error);
  }
}
