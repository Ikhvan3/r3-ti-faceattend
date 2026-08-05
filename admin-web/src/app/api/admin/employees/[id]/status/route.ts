import { NextResponse } from "next/server";

import {
  employeeBffError,
  readUpdateStatusBody,
} from "@/lib/server/employee-bff";
import { updateEmployeeStatusWithSession } from "@/lib/server/go-api";

export async function PATCH(
  request: Request,
  ctx: { params: Promise<{ id: string }> },
): Promise<NextResponse> {
  const { id } = await ctx.params;
  const body = await readUpdateStatusBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const employee = await updateEmployeeStatusWithSession(id, body.value);
    return NextResponse.json({
      status: "ok",
      message: "Status pegawai berhasil diperbarui.",
      data: employee,
    });
  } catch (error) {
    return employeeBffError(error);
  }
}
