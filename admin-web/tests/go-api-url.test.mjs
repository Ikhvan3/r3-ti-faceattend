import assert from "node:assert/strict";
import { test } from "node:test";

import { buildGoApiUrl } from "../src/lib/server/go-api-url.ts";

test("buildGoApiUrl preserves /api/v1 from GO_API_BASE_URL", () => {
  assert.equal(
    buildGoApiUrl("http://127.0.0.1:8080/api/v1", "/admin/employees"),
    "http://127.0.0.1:8080/api/v1/admin/employees",
  );
});

test("buildGoApiUrl builds employee pagination query safely", () => {
  assert.equal(
    buildGoApiUrl(
      "http://127.0.0.1:8080/api/v1/",
      "/admin/employees?page=1&page_size=10",
    ),
    "http://127.0.0.1:8080/api/v1/admin/employees?page=1&page_size=10",
  );
});
