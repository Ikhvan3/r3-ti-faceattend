import { redirect } from "next/navigation";

import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";
import { PageHeader } from "@/app/(dashboard)/employees/_components/page-header";
import { AttendanceFilterForm } from "@/app/(dashboard)/attendance/_components/attendance-filter-form";
import { AttendancePagination } from "@/app/(dashboard)/attendance/_components/attendance-pagination";
import { AttendanceTable } from "@/app/(dashboard)/attendance/_components/attendance-table";
import { SafeApiError } from "@/lib/auth/types";
import { parseAttendanceSearchParams } from "@/lib/attendance/utils";
import { getAdminAttendance } from "@/lib/server/admin-attendance-bff";
import { requireAdmin } from "@/lib/server/session";

export default async function AttendancePage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  await requireAdmin();
  const query = parseAttendanceSearchParams(await searchParams);

  let result;
  let errorMessage = "";
  try {
    result = await getAdminAttendance({
      date_from: query.date_from || undefined,
      date_to: query.date_to || undefined,
      search: query.search || undefined,
      attendance_state: query.attendance_state || undefined,
      is_late:
        query.is_late === "" ? undefined : query.is_late === "true",
      page: query.page,
      page_size: query.page_size,
    });
  } catch (error) {
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      redirect("/login");
    }
    if (error instanceof SafeApiError) {
      errorMessage = error.message;
    } else {
      throw error;
    }
  }

  const hasFilter =
    query.date_from !== "" ||
    query.date_to !== "" ||
    query.search !== "" ||
    query.attendance_state !== "" ||
    query.is_late !== "";

  return (
    <section className="mx-auto max-w-7xl space-y-6">
      <PageHeader
        description="Pantau presensi pegawai secara read-only berdasarkan tanggal bisnis, status, dan keterlambatan."
        title="Presensi Pegawai"
      />
      <AttendanceFilterForm query={query} />

      {errorMessage ? (
        <div className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
          {errorMessage}
        </div>
      ) : result && result.items.length > 0 ? (
        <>
          <AttendanceTable items={result.items} />
          <AttendancePagination
            query={query}
            totalItems={result.total_items}
            totalPages={result.total_pages}
          />
        </>
      ) : hasFilter ? (
        <EmptyState
          description="Tidak ada presensi yang sesuai dengan filter yang dipilih."
          title="Hasil filter kosong"
        />
      ) : (
        <EmptyState
          description="Belum ada data presensi untuk tanggal bisnis hari ini."
          title="Belum ada data presensi"
        />
      )}
    </section>
  );
}
