import "server-only";

import { cookies } from "next/headers";
import type { ResponseCookie } from "next/dist/compiled/@edge-runtime/cookies";

import { isProduction } from "@/lib/server/env";

export const ACCESS_TOKEN_COOKIE = "r3_access_token";
export const REFRESH_TOKEN_COOKIE = "r3_refresh_token";

const DEFAULT_REFRESH_MAX_AGE_SECONDS = 7 * 24 * 60 * 60;

export async function readAccessToken(): Promise<string | undefined> {
  return (await cookies()).get(ACCESS_TOKEN_COOKIE)?.value;
}

export async function readRefreshToken(): Promise<string | undefined> {
  return (await cookies()).get(REFRESH_TOKEN_COOKIE)?.value;
}

export async function setAuthCookies(input: {
  accessToken: string;
  refreshToken: string;
  accessMaxAge: number;
}): Promise<void> {
  const store = await cookies();

  store.set(ACCESS_TOKEN_COOKIE, input.accessToken, {
    ...baseCookieOptions(),
    maxAge: Math.max(1, input.accessMaxAge),
  });

  store.set(REFRESH_TOKEN_COOKIE, input.refreshToken, {
    ...baseCookieOptions(),
    maxAge: DEFAULT_REFRESH_MAX_AGE_SECONDS,
  });
}

export async function clearAuthCookies(): Promise<void> {
  const store = await cookies();
  store.set(ACCESS_TOKEN_COOKIE, "", {
    ...baseCookieOptions(),
    maxAge: 0,
  });
  store.set(REFRESH_TOKEN_COOKIE, "", {
    ...baseCookieOptions(),
    maxAge: 0,
  });
}

function baseCookieOptions(): Pick<
  ResponseCookie,
  "httpOnly" | "secure" | "sameSite" | "path"
> {
  return {
    httpOnly: true,
    secure: isProduction(),
    sameSite: "lax",
    path: "/",
  };
}
