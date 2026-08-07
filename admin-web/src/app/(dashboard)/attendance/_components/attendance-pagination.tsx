import Link from "next/link";

import type { AttendanceSearchQuery } from "@/lib/attendance/utils";

export function AttendancePagination({
  query,
  totalItems,
  totalPages,
}: {
  query: AttendanceSearchQuery;
  totalItems: number;
  totalPages: number;
}) {
  if (totalPages <= 1) {
    return (
      <p className="text-sm text-slate-500">
        Total {totalItems} data.
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <p className="text-sm text-slate-500">
        Halaman {query.page} dari {totalPages} · {totalItems} data
      </p>
      <div className="flex gap-2">
        <PageLink
          disabled={query.page <= 1}
          href={buildPageHref(query, query.page - 1)}
          label="Sebelumnya"
        />
        <PageLink
          disabled={query.page >= totalPages}
          href={buildPageHref(query, query.page + 1)}
          label="Berikutnya"
        />
      </div>
    </div>
  );
}

function PageLink({
  href,
  label,
  disabled,
}: {
  href: string;
  label: string;
  disabled: boolean;
}) {
  if (disabled) {
    return (
      <span className="inline-flex h-9 cursor-not-allowed items-center rounded-md border border-slate-200 bg-slate-50 px-3 text-sm text-slate-400">
        {label}
      </span>
    );
  }
  return (
    <Link
      className="inline-flex h-9 items-center rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 hover:bg-slate-50"
      href={href}
    >
      {label}
    </Link>
  );
}

function buildPageHref(query: AttendanceSearchQuery, page: number): string {
  const params = new URLSearchParams();
  if (query.date_from) params.set("date_from", query.date_from);
  if (query.date_to) params.set("date_to", query.date_to);
  if (query.search) params.set("search", query.search);
  if (query.attendance_state)
    params.set("attendance_state", query.attendance_state);
  if (query.is_late) params.set("is_late", query.is_late);
  params.set("page", String(Math.max(page, 1)));
  params.set("page_size", String(query.page_size));
  return `/attendance?${params.toString()}`;
}
