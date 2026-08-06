import Link from "next/link";

import type { ParsedWorkScheduleSearchParams } from "@/lib/schedule/utils";
import { SCHEDULE_STATUSES, scheduleStatusLabel } from "@/lib/schedule/utils";

export function WorkScheduleSearchForm({
  query,
}: {
  query: ParsedWorkScheduleSearchParams;
}) {
  return (
    <form
      action="/work-schedules"
      className="grid gap-3 rounded-md border border-slate-200 bg-white p-4 md:grid-cols-[1fr_220px_auto_auto]"
    >
      <input type="hidden" name="page" value="1" />
      <input type="hidden" name="page_size" value={query.page_size} />
      <div>
        <label className="text-sm font-medium text-slate-800" htmlFor="search">
          Cari jadwal
        </label>
        <input
          className={inputClassName}
          defaultValue={query.search}
          id="search"
          name="search"
          placeholder="Nama jadwal kerja"
          type="search"
        />
      </div>
      <div>
        <label className="text-sm font-medium text-slate-800" htmlFor="status">
          Status
        </label>
        <select
          className={inputClassName}
          defaultValue={query.status}
          id="status"
          name="status"
        >
          <option value="">Semua status</option>
          {SCHEDULE_STATUSES.map((status) => (
            <option key={status} value={status}>
              {scheduleStatusLabel(status)}
            </option>
          ))}
        </select>
      </div>
      <div className="flex items-end">
        <button className={primaryButtonClassName} type="submit">
          Cari
        </button>
      </div>
      <div className="flex items-end">
        <Link className={secondaryLinkClassName} href="/work-schedules">
          Reset
        </Link>
      </div>
    </form>
  );
}

const inputClassName =
  "mt-2 h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none transition focus:border-emerald-700 focus:ring-2 focus:ring-emerald-100";
const primaryButtonClassName =
  "h-10 w-full rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white transition hover:bg-emerald-800 focus:outline-none focus:ring-2 focus:ring-emerald-200";
const secondaryLinkClassName =
  "inline-flex h-10 w-full items-center justify-center rounded-md border border-slate-300 px-4 text-sm font-medium text-slate-800 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100";
