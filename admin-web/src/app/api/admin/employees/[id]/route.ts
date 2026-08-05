import { NextResponse } from "next/server";

import {
  employeeBffError,
  readUpdateEmployeeBody,
} from "@/lib/server/employee-bff";
import { updateEmployeeWithSession } from "@/lib/server/go-api";

export async function PUT(
  request: Request,
  ctx: { params: Promise<{ id: string }> },
): Promise<NextResponse> {
  const { id } = await ctx.params;
  const body = await readUpdateEmployeeBody(request);
  if (!body.ok) {
    return body.response;
  }

  try {
    const employee = await updateEmployeeWithSession(id, body.value);
    return NextResponse.json({
      status: "ok",
      message: "Pegawai berhasil diperbarui.",
      data: employee,
    });
  } catch (error) {
    return employeeBffError(error);
  }
}
