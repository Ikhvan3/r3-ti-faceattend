import { NextResponse } from "next/server";

import { clearAuthCookies, readRefreshToken } from "@/lib/server/auth-cookies";
import { logoutToGoApi } from "@/lib/server/go-api";

export async function POST(): Promise<NextResponse> {
  const refreshToken = await readRefreshToken();

  if (refreshToken) {
    try {
      await logoutToGoApi(refreshToken);
    } catch {
      // Logout lokal tetap dilakukan agar aman dan idempotent.
    }
  }

  await clearAuthCookies();

  return NextResponse.json({
    status: "ok",
    message: "Logout berhasil.",
  });
}

export function GET(request: Request): NextResponse {
  return NextResponse.redirect(new URL("/login", request.url));
}
