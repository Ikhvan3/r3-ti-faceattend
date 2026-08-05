"use server";

import { redirect } from "next/navigation";

import { clearAuthCookies, readRefreshToken } from "@/lib/server/auth-cookies";
import { logoutToGoApi } from "@/lib/server/go-api";

export async function logoutAction(): Promise<void> {
  const refreshToken = await readRefreshToken();

  if (refreshToken) {
    try {
      await logoutToGoApi(refreshToken);
    } catch {
      // Logout lokal tetap dilanjutkan agar idempotent dan tidak membocorkan state backend.
    }
  }

  await clearAuthCookies();
  redirect("/login");
}
