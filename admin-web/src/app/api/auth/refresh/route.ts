import { NextRequest, NextResponse } from "next/server";

import { clearAuthCookies, readRefreshToken, setAuthCookies } from "@/lib/server/auth-cookies";
import { refreshToGoApi } from "@/lib/server/go-api";

export async function POST(): Promise<NextResponse> {
  const result = await refreshSession();

  if (!result.ok) {
    return NextResponse.json(
      { status: "error", message: "Session tidak valid." },
      { status: 401 },
    );
  }

  return NextResponse.json({
    status: "ok",
    message: "Session diperbarui.",
    data: {
      user: result.user,
    },
  });
}

export async function GET(request: NextRequest): Promise<NextResponse> {
  const next = sanitizeNextPath(request.nextUrl.searchParams.get("next"));
  const result = await refreshSession();

  if (!result.ok) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.redirect(new URL(next, request.url));
}

type RefreshResult =
  | {
      ok: true;
      user: unknown;
    }
  | {
      ok: false;
    };

async function refreshSession(): Promise<RefreshResult> {
  const refreshToken = await readRefreshToken();
  if (!refreshToken) {
    await clearAuthCookies();
    return { ok: false };
  }

  try {
    const result = await refreshToGoApi(refreshToken);
    if (result.user.role !== "ADMIN") {
      await clearAuthCookies();
      return { ok: false };
    }

    await setAuthCookies({
      accessToken: result.access_token,
      refreshToken: result.refresh_token,
      accessMaxAge: result.expires_in,
    });

    return { ok: true, user: result.user };
  } catch {
    await clearAuthCookies();
    return { ok: false };
  }
}

function sanitizeNextPath(value: string | null): string {
  if (!value || !value.startsWith("/") || value.startsWith("//")) {
    return "/dashboard";
  }

  return value;
}
