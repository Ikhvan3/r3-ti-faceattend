import { NextRequest, NextResponse } from "next/server";

const ACCESS_TOKEN_COOKIE = "r3_access_token";
const REFRESH_TOKEN_COOKIE = "r3_refresh_token";

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const hasAccessToken = request.cookies.has(ACCESS_TOKEN_COOKIE);
  const hasRefreshToken = request.cookies.has(REFRESH_TOKEN_COOKIE);

  if (pathname.startsWith("/dashboard") && !hasAccessToken && !hasRefreshToken) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*"],
};
