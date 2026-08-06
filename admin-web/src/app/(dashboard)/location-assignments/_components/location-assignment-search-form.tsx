import Link from "next/link";

import type { Employee } from "@/lib/employee/types";
import type { OfficeLocation } from "@/lib/location/types";
import type { ParsedLocationAssignmentSearchParams } from "@/lib/location/utils";
import {
  LOCATION_ASSIGNMENT_STATUSES,
  locationStatusLabel,
} from "@/lib/location/utils";

export function LocationAssignmentSearchForm({
  query,
  employees,
  locations,
}: {
  query: ParsedLocationAssignmentSearchParams;
  employees: Employee[];
  locations: OfficeLocation[];
}) {
  return (
    <form
      action="/location-assignments"
      className="grid gap-3 rounded-md border border-slate-200 bg-white p-4 lg:grid-cols-[1fr_220px_220px_180px_auto_auto]"
    >
      <input type="hidden" name="page" value="1" />
      <input type="hidden" name="page_size" value={query.page_size} />
      <Field htmlFor="search" label="Cari pegawai">
        <input
          className={inputClassName}
          defaultValue={query.search}
          id="search"
          name="search"
          placeholder="Nomor, nama, atau email"
          type="search"
        />
      </Field>
      <Field htmlFor="user_id" label="Pegawai">
        <select className={inputClassName} defaultValue={query.user_id} id="user_id" name="user_id">
          <option value="">Semua pegawai</option>
          {employees.map((employee) => (
            <option key={employee.id} value={employee.id}>
              {employee.employee_number} - {employee.name}
            </option>
          ))}
        </select>
      </Field>
      <Field htmlFor="office_location_id" label="Lokasi">
        <select className={inputClassName} defaultValue={query.office_location_id} id="office_location_id" name="office_location_id">
          <option value="">Semua lokasi</option>
          {locations.map((location) => (
            <option key={location.id} value={location.id}>
              {location.name}
            </option>
          ))}
        </select>
      </Field>
      <Field htmlFor="status" label="Status">
        <select className={inputClassName} defaultValue={query.status} id="status" name="status">
          <option value="">Semua status</option>
          {LOCATION_ASSIGNMENT_STATUSES.map((status) => (
            <option key={status} value={status}>
              {locationStatusLabel(status)}
            </option>
          ))}
        </select>
      </Field>
      <div className="flex items-end">
        <button className={primaryButtonClassName} type="submit">
          Cari
        </button>
      </div>
      <div className="flex items-end">
        <Link className={secondaryLinkClassName} href="/location-assignments">
          Reset
        </Link>
      </div>
    </form>
  );
}

function Field({
  label,
  htmlFor,
  children,
}: {
  label: string;
  htmlFor: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label className="text-sm font-medium text-slate-800" htmlFor={htmlFor}>
        {label}
      </label>
      {children}
    </div>
  );
}

const inputClassName =
  "mt-2 h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none transition focus:border-emerald-700 focus:ring-2 focus:ring-emerald-100";
const primaryButtonClassName =
  "h-10 w-full rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white transition hover:bg-emerald-800 focus:outline-none focus:ring-2 focus:ring-emerald-200";
const secondaryLinkClassName =
  "inline-flex h-10 w-full items-center justify-center rounded-md border border-slate-300 px-4 text-sm font-medium text-slate-800 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100";
