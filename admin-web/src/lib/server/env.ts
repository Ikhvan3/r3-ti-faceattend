import "server-only";

export function getGoApiBaseUrl(): string {
  const value = process.env.GO_API_BASE_URL?.trim();

  if (!value) {
    throw new Error("GO_API_BASE_URL belum dikonfigurasi untuk admin-web.");
  }

  return value.replace(/\/+$/, "");
}

export function isProduction(): boolean {
  return process.env.APP_ENV === "production";
}
