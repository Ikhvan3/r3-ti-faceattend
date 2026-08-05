import "server-only";

import { redirect } from "next/navigation";

import type { SafeUserProfile } from "@/lib/auth/types";
import { SafeApiError } from "@/lib/auth/types";
import {
  readAccessToken,
  readRefreshToken,
} from "@/lib/server/auth-cookies";
import { getMeFromGoApi } from "@/lib/server/go-api";

export async function getCurrentAdmin(): Promise<SafeUserProfile | null> {
  const accessToken = await readAccessToken();
  if (!accessToken) {
    return null;
  }

  try {
    const user = await getMeFromGoApi(accessToken);
    return user.role === "ADMIN" ? user : null;
  } catch (error) {
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      const refreshToken = await readRefreshToken();
      if (refreshToken) {
        redirect("/api/auth/refresh?next=/dashboard");
      }
    }

    return null;
  }
}

export async function requireAdmin(): Promise<SafeUserProfile> {
  const user = await getCurrentAdmin();

  if (!user) {
    redirect("/login");
  }

  return user;
}
