import assert from "node:assert/strict";
import { test } from "node:test";

import {
  isLocationAssignment,
  isLocationAssignmentListResponse,
  isOfficeLocation,
  isOfficeLocationListResponse,
} from "../src/lib/location/utils.ts";

const office = {
  id: "00000000-0000-4000-8000-000000000030",
  name: "Kantor Regional 3",
  address: null,
  latitude: -6.1,
  longitude: 106.8,
  radius_meters: 100,
  is_active: true,
  created_at: "2026-08-06T00:00:00Z",
  updated_at: "2026-08-06T00:00:00Z",
};

const user = {
  id: "00000000-0000-4000-8000-000000000001",
  employee_number: "EMP-DUMMY-001",
  name: "Pegawai Dummy",
  email: "pegawai.dummy@example.test",
  phone: null,
  position: null,
  role: "USER",
  account_status: "ACTIVE",
  created_at: "2026-08-06T00:00:00Z",
  updated_at: "2026-08-06T00:00:00Z",
};

test("office location list accepts empty items and pagination", () => {
  assert.equal(
    isOfficeLocationListResponse({
      items: [],
      page: 1,
      page_size: 10,
      total_items: 0,
      total_pages: 0,
    }),
    true,
  );
});

test("office location accepts address null and numeric coordinates", () => {
  assert.equal(isOfficeLocation(office), true);
  assert.equal(typeof office.latitude, "number");
  assert.equal(typeof office.longitude, "number");
});

test("office location list rejects malformed items null", () => {
  assert.equal(
    isOfficeLocationListResponse({
      items: null,
      page: 1,
      page_size: 10,
      total_items: 0,
      total_pages: 0,
    }),
    false,
  );
});

test("office location rejects string coordinates", () => {
  assert.equal(
    isOfficeLocation({
      ...office,
      latitude: "-6.1",
      longitude: "106.8",
    }),
    false,
  );
});

test("location assignment list accepts empty items", () => {
  assert.equal(
    isLocationAssignmentListResponse({
      items: [],
      page: 1,
      page_size: 10,
      total_items: 0,
      total_pages: 0,
    }),
    true,
  );
});

test("location assignment accepts office_location nested object", () => {
  assert.equal(
    isLocationAssignment({
      id: "00000000-0000-4000-8000-000000000040",
      user,
      office_location: office,
      effective_from: "2026-08-06",
      effective_to: null,
      status: "CURRENT",
      created_at: "2026-08-06T00:00:00Z",
      updated_at: "2026-08-06T00:00:00Z",
    }),
    true,
  );
});

test("location assignment rejects office nested object", () => {
  assert.equal(
    isLocationAssignment({
      id: "00000000-0000-4000-8000-000000000040",
      user,
      office,
      effective_from: "2026-08-06",
      effective_to: null,
      status: "CURRENT",
      created_at: "2026-08-06T00:00:00Z",
      updated_at: "2026-08-06T00:00:00Z",
    }),
    false,
  );
});
