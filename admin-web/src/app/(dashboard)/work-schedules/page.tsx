import { redirect } from "next/navigation";

import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";
import {
  PageHeader,
  PrimaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { SchedulePagination } from "@/app/(dashboard)/work-schedules/_components/schedule-pagination";
import { WorkScheduleSearchForm } from "@/app/(dashboard)/work-schedules/_components/work-schedule-search-form";
import { WorkScheduleTable } from "@/app/(dashboard)/work-schedules/_components/work-schedule-table";
import { SafeApiError } from "@/lib/auth/types";
import { getWorkSchedules } from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";
import { parseWorkScheduleSearchParams } from "@/lib/schedule/utils";

export default async function WorkSchedulesPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  await requireAdmin();
  const query = parseWorkScheduleSearchParams(await searchParams);

  let result;
  try {
    result = await getWorkSchedules({
      page: query.page,
      page_size: query.page_size,
      search: query.search,
      status: query.status || undefined,
    });
  } catch (error) {
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      redirect("/login");
    }
    throw error;
  }

  const hasFilter = query.search !== "" || query.status !== "";

  return (
    <section className="mx-auto max-w-6xl space-y-6">
      <PageHeader
        action={<PrimaryLink href="/work-schedules/new">Tambah Jadwal</PrimaryLink>}
        description="Kelola jadwal kerja dasar untuk pegawai USER Divisi Teknologi Informasi."
        title="Jadwal Kerja"
      />
      <WorkScheduleSearchForm query={query} />

      {result.items.length > 0 ? (
        <>
          <WorkScheduleTable schedules={result.items} />
          <SchedulePagination
            kind="schedule"
            query={query}
            totalItems={result.total_items}
            totalPages={result.total_pages}
          />
        </>
      ) : hasFilter ? (
        <EmptyState
          description="Tidak ada jadwal kerja yang cocok dengan kata kunci atau status yang dipilih."
          title="Hasil pencarian kosong"
        />
      ) : (
        <EmptyState
          action={<PrimaryLink href="/work-schedules/new">Tambah Jadwal</PrimaryLink>}
          description="Belum ada jadwal kerja yang tersedia untuk penugasan pegawai."
          title="Belum ada jadwal kerja"
        />
      )}
    </section>
  );
}
