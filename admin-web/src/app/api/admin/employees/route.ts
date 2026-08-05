import { NextResponse } from "next/server";

import {
  employeeBffError,
  readCreateEmployeeBody,
} from "@/lib/server/employee-bff";
import { createEmployeeWithSession } from "@/lib/server/go-api";

export async function POST(request: Request): Promise<NextResponse> {
  const body = await readCreateEmployeeBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const employee = await createEmployeeWithSession(body.value);
    return NextResponse.json(
      {
        status: "ok",
        message: "Pegawai berhasil dibuat.",
        data: employee,
      },
      { status: 201 },
    );
  } catch (error) {
    return employeeBffError(error);
  }
}
