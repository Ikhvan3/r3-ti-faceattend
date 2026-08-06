import { redirect } from "next/navigation";

import { AssignmentSearchForm } from "@/app/(dashboard)/schedule-assignments/_components/assignment-search-form";
import { AssignmentTable } from "@/app/(dashboard)/schedule-assignments/_components/assignment-table";
import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";
import {
  PageHeader,
  PrimaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { SchedulePagination } from "@/app/(dashboard)/work-schedules/_components/schedule-pagination";
import { SafeApiError } from "@/lib/auth/types";
import {
  getActiveWorkSchedules,
  getEmployeeOptions,
  getScheduleAssignments,
} from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";
import { parseAssignmentSearchParams } from "@/lib/schedule/utils";

export default async function ScheduleAssignmentsPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  await requireAdmin();
  const query = parseAssignmentSearchParams(await searchParams);

  let result;
  let employees;
  let schedules;
  try {
    [result, employees, schedules] = await Promise.all([
      getScheduleAssignments({
        page: query.page,
        page_size: query.page_size,
        search: query.search,
        user_id: query.user_id || undefined,
        schedule_id: query.schedule_id || undefined,
        status: query.status || undefined,
      }),
      getEmployeeOptions(),
      getActiveWorkSchedules(),
    ]);
  } catch (error) {
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      redirect("/login");
    }
    throw error;
  }

  const hasFilter =
    query.search !== "" ||
    query.user_id !== "" ||
    query.schedule_id !== "" ||
    query.status !== "";

  return (
    <section className="mx-auto max-w-6xl space-y-6">
      <PageHeader
        action={
          <PrimaryLink href="/schedule-assignments/new">
            Tambah Penugasan
          </PrimaryLink>
        }
        description="Kelola periode penugasan jadwal pegawai USER. Periode tanggal bersifat inklusif."
        title="Penugasan Jadwal"
      />
      <AssignmentSearchForm
        employees={employees}
        query={query}
        schedules={schedules}
      />
      {result.items.length > 0 ? (
        <>
          <AssignmentTable assignments={result.items} />
          <SchedulePagination
            kind="assignment"
            query={query}
            totalItems={result.total_items}
            totalPages={result.total_pages}
          />
        </>
      ) : hasFilter ? (
        <EmptyState
          description="Tidak ada penugasan yang cocok dengan filter yang dipilih."
          title="Hasil pencarian kosong"
        />
      ) : (
        <EmptyState
          action={
            <PrimaryLink href="/schedule-assignments/new">
              Tambah Penugasan
            </PrimaryLink>
          }
          description="Belum ada penugasan jadwal pegawai yang tersedia."
          title="Belum ada penugasan jadwal"
        />
      )}
    </section>
  );
}
