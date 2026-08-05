export function buildGoApiUrl(baseUrl: string, path: string): string {
  const normalizedBase = baseUrl.trim().replace(/\/+$/, "");
  const normalizedPath = path.trim().replace(/^\/+/, "");

  if (!normalizedBase) {
    throw new Error("GO_API_BASE_URL belum dikonfigurasi untuk admin-web.");
  }
  if (!normalizedPath) {
    return normalizedBase;
  }

  return `${normalizedBase}/${normalizedPath}`;
}
