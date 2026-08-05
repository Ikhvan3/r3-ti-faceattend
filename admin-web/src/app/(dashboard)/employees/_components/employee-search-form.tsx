import Link from "next/link";

import { EMPLOYEE_STATUSES, statusLabel } from "@/lib/employee/utils";
import type { ParsedEmployeeSearchParams } from "@/lib/employee/utils";

export function EmployeeSearchForm({
  query,
}: {
  query: ParsedEmployeeSearchParams;
}) {
  return (
    <form
      action="/employees"
      className="grid gap-3 rounded-md border border-slate-200 bg-white p-4 md:grid-cols-[1fr_220px_auto_auto]"
    >
      <input type="hidden" name="page" value="1" />
      <input type="hidden" name="page_size" value={query.page_size} />
      <div>
        <label className="text-sm font-medium text-slate-800" htmlFor="search">
          Cari pegawai
        </label>
        <input
          className="mt-2 h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-emerald-700 focus:ring-2 focus:ring-emerald-100"
          defaultValue={query.search}
          id="search"
          name="search"
          placeholder="Nomor, nama, email, atau jabatan"
          type="search"
        />
      </div>
      <div>
        <label className="text-sm font-medium text-slate-800" htmlFor="status">
          Status
        </label>
        <select
          className="mt-2 h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none transition focus:border-emerald-700 focus:ring-2 focus:ring-emerald-100"
          defaultValue={query.status}
          id="status"
          name="status"
        >
          <option value="">Semua status</option>
          {EMPLOYEE_STATUSES.map((status) => (
            <option key={status} value={status}>
              {statusLabel(status)}
            </option>
          ))}
        </select>
      </div>
      <div className="flex items-end">
        <button
          className="h-10 w-full rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white transition hover:bg-emerald-800 focus:outline-none focus:ring-2 focus:ring-emerald-200"
          type="submit"
        >
          Cari
        </button>
      </div>
      <div className="flex items-end">
        <Link
          className="inline-flex h-10 w-full items-center justify-center rounded-md border border-slate-300 px-4 text-sm font-medium text-slate-800 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100"
          href="/employees"
        >
          Reset
        </Link>
      </div>
    </form>
  );
}
