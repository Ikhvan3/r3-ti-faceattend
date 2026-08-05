import Link from "next/link";

import type { ParsedEmployeeSearchParams } from "@/lib/employee/utils";
import { employeeListHref } from "@/lib/employee/utils";

export function Pagination({
  query,
  totalPages,
  totalItems,
}: {
  query: ParsedEmployeeSearchParams;
  totalPages: number;
  totalItems: number;
}) {
  const currentPage = Math.min(Math.max(1, query.page), Math.max(1, totalPages));
  const previousDisabled = currentPage <= 1;
  const nextDisabled = totalPages === 0 || currentPage >= totalPages;

  return (
    <div className="flex flex-col gap-3 border-t border-slate-200 px-4 py-4 text-sm text-slate-600 sm:flex-row sm:items-center sm:justify-between">
      <p>
        Total <span className="font-semibold text-slate-950">{totalItems}</span>{" "}
        pegawai
      </p>
      <div className="flex items-center gap-2">
        <PageLink
          disabled={previousDisabled}
          href={employeeListHref(query, currentPage - 1)}
        >
          Sebelumnya
        </PageLink>
        <span className="min-w-24 text-center">
          {totalPages === 0 ? "0 / 0" : `${currentPage} / ${totalPages}`}
        </span>
        <PageLink
          disabled={nextDisabled}
          href={employeeListHref(query, currentPage + 1)}
        >
          Berikutnya
        </PageLink>
      </div>
    </div>
  );
}

function PageLink({
  href,
  disabled,
  children,
}: {
  href: string;
  disabled: boolean;
  children: string;
}) {
  if (disabled) {
    return (
      <span className="inline-flex h-9 cursor-not-allowed items-center rounded-md border border-slate-200 px-3 text-slate-400">
        {children}
      </span>
    );
  }

  return (
    <Link
      className="inline-flex h-9 items-center rounded-md border border-slate-300 bg-white px-3 text-slate-800 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100"
      href={href}
    >
      {children}
    </Link>
  );
}
