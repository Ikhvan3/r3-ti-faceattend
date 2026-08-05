import { NextResponse } from "next/server";

import { SafeApiError } from "@/lib/auth/types";
import { clearAuthCookies, setAuthCookies } from "@/lib/server/auth-cookies";
import { loginToGoApi } from "@/lib/server/go-api";

const MAX_BODY_BYTES = 10_000;

export async function POST(request: Request): Promise<NextResponse> {
  const body = await readLoginBody(request);

  if (!body.ok) {
    return jsonError(body.message, 400);
  }

  try {
    const result = await loginToGoApi({
      email: body.email,
      password: body.password,
    });

    if (result.user.role !== "ADMIN") {
      await clearAuthCookies();
      return jsonError("Akun ini tidak memiliki akses admin.", 403);
    }

    await setAuthCookies({
      accessToken: result.access_token,
      refreshToken: result.refresh_token,
      accessMaxAge: result.expires_in,
    });

    return NextResponse.json({
      status: "ok",
      message: "Login berhasil.",
      data: {
        user: result.user,
      },
    });
  } catch (error) {
    await clearAuthCookies();

    if (error instanceof SafeApiError) {
      if (error.code === "INVALID_CREDENTIALS") {
        return jsonError("Email atau password tidak valid.", 401);
      }

      return jsonError(error.message, error.status);
    }

    return jsonError("Login gagal. Coba lagi nanti.", 500);
  }
}

export function GET(): NextResponse {
  return methodNotAllowed();
}

type LoginBodyResult =
  | {
      ok: true;
      email: string;
      password: string;
    }
  | {
      ok: false;
      message: string;
    };

async function readLoginBody(request: Request): Promise<LoginBodyResult> {
  const contentLength = Number(request.headers.get("content-length") ?? "0");
  if (contentLength > MAX_BODY_BYTES) {
    return { ok: false, message: "Request terlalu besar." };
  }

  let payload: unknown;
  try {
    payload = await request.json();
  } catch {
    return { ok: false, message: "Request tidak valid." };
  }

  if (!isLoginPayload(payload)) {
    return { ok: false, message: "Email dan password wajib diisi." };
  }

  return {
    ok: true,
    email: payload.email.trim(),
    password: payload.password,
  };
}

function isLoginPayload(
  value: unknown,
): value is { email: string; password: string } {
  if (typeof value !== "object" || value === null) {
    return false;
  }

  const record = value as Record<string, unknown>;
  const keys = Object.keys(record);

  return (
    keys.every((key) => key === "email" || key === "password") &&
    typeof record.email === "string" &&
    record.email.trim() !== "" &&
    typeof record.password === "string" &&
    record.password !== ""
  );
}

function jsonError(message: string, status: number): NextResponse {
  return NextResponse.json({ status: "error", message }, { status });
}

function methodNotAllowed(): NextResponse {
  return NextResponse.json(
    { status: "error", message: "Method tidak diizinkan." },
    {
      status: 405,
      headers: {
        Allow: "POST",
      },
    },
  );
}
