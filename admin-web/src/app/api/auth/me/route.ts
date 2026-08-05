import { NextResponse } from "next/server";

import { SafeApiError } from "@/lib/auth/types";
import {
  clearAuthCookies,
  readAccessToken,
  readRefreshToken,
  setAuthCookies,
} from "@/lib/server/auth-cookies";
import { getMeFromGoApi, refreshToGoApi } from "@/lib/server/go-api";

export async function GET(): Promise<NextResponse> {
  const user = await readAdminWithRefresh();

  if (!user) {
    return NextResponse.json(
      { status: "error", message: "Session tidak valid." },
      { status: 401 },
    );
  }

  return NextResponse.json({
    status: "ok",
    message: "Profil berhasil dibaca.",
    data: {
      user,
    },
  });
}

async function readAdminWithRefresh() {
  const accessToken = await readAccessToken();
  if (!accessToken) {
    return null;
  }

  try {
    const user = await getMeFromGoApi(accessToken);
    return user.role === "ADMIN" ? user : null;
  } catch (error) {
    if (!(error instanceof SafeApiError) || error.code !== "UNAUTHORIZED") {
      return null;
    }
  }

  const refreshToken = await readRefreshToken();
  if (!refreshToken) {
    await clearAuthCookies();
    return null;
  }

  try {
    const refreshed = await refreshToGoApi(refreshToken);
    if (refreshed.user.role !== "ADMIN") {
      await clearAuthCookies();
      return null;
    }

    await setAuthCookies({
      accessToken: refreshed.access_token,
      refreshToken: refreshed.refresh_token,
      accessMaxAge: refreshed.expires_in,
    });

    const user = await getMeFromGoApi(refreshed.access_token);
    return user.role === "ADMIN" ? user : null;
  } catch {
    await clearAuthCookies();
    return null;
  }
}
