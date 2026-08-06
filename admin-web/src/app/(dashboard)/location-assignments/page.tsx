import { redirect } from "next/navigation";

import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";
import {
  PageHeader,
  PrimaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { LocationAssignmentSearchForm } from "@/app/(dashboard)/location-assignments/_components/location-assignment-search-form";
import { LocationAssignmentTable } from "@/app/(dashboard)/location-assignments/_components/location-assignment-table";
import { LocationPagination } from "@/app/(dashboard)/office-locations/_components/location-pagination";
import { SafeApiError } from "@/lib/auth/types";
import { parseLocationAssignmentSearchParams } from "@/lib/location/utils";
import {
  getActiveOfficeLocations,
  getEmployeeOptions,
  getLocationAssignments,
} from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";

export default async function LocationAssignmentsPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  await requireAdmin();
  const query = parseLocationAssignmentSearchParams(await searchParams);

  let result;
  let employees;
  let locations;
  try {
    [result, employees, locations] = await Promise.all([
      getLocationAssignments({
        page: query.page,
        page_size: query.page_size,
        search: query.search,
        user_id: query.user_id || undefined,
        office_location_id: query.office_location_id || undefined,
        status: query.status || undefined,
      }),
      getEmployeeOptions(),
      getActiveOfficeLocations(),
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
    query.office_location_id !== "" ||
    query.status !== "";

  return (
    <section className="mx-auto max-w-6xl space-y-6">
      <PageHeader
        action={
          <PrimaryLink href="/location-assignments/new">
            Tambah Penugasan
          </PrimaryLink>
        }
        description="Kelola periode penugasan lokasi kantor pegawai USER. Periode tanggal bersifat inklusif."
        title="Penugasan Lokasi"
      />
      <LocationAssignmentSearchForm
        employees={employees}
        locations={locations}
        query={query}
      />
      {result.items.length > 0 ? (
        <>
          <LocationAssignmentTable assignments={result.items} />
          <LocationPagination
            kind="assignment"
            query={query}
            totalItems={result.total_items}
            totalPages={result.total_pages}
          />
        </>
      ) : hasFilter ? (
        <EmptyState
          description="Tidak ada penugasan lokasi yang cocok dengan filter yang dipilih."
          title="Hasil pencarian kosong"
        />
      ) : (
        <EmptyState
          action={
            <PrimaryLink href="/location-assignments/new">
              Tambah Penugasan
            </PrimaryLink>
          }
          description="Belum ada penugasan lokasi pegawai yang tersedia."
          title="Belum ada penugasan lokasi"
        />
      )}
    </section>
  );
}
