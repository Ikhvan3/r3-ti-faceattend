import assert from "node:assert/strict";
import { test } from "node:test";

import {
  buildEmployeeQueryParams,
  safeEmployeeApiMessage,
} from "../src/lib/employee/utils.ts";

test("buildEmployeeQueryParams builds pagination query", () => {
  assert.equal(
    buildEmployeeQueryParams({ page: 1, page_size: 10 }),
    "page=1&page_size=10",
  );
});

test("buildEmployeeQueryParams includes search and status when present", () => {
  assert.equal(
    buildEmployeeQueryParams({
      page: 2,
      page_size: 25,
      search: "dummy ti",
      status: "ACTIVE",
    }),
    "page=2&page_size=25&search=dummy+ti&status=ACTIVE",
  );
});

test("safeEmployeeApiMessage keeps detail not found as 404 message", () => {
  assert.equal(safeEmployeeApiMessage(404, "fallback"), "Pegawai tidak ditemukan.");
});
