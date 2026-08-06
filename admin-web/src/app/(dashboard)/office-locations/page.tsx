import { redirect } from "next/navigation";

import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";
import {
  PageHeader,
  PrimaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { LocationPagination } from "@/app/(dashboard)/office-locations/_components/location-pagination";
import { OfficeLocationSearchForm } from "@/app/(dashboard)/office-locations/_components/office-location-search-form";
import { OfficeLocationTable } from "@/app/(dashboard)/office-locations/_components/office-location-table";
import { SafeApiError } from "@/lib/auth/types";
import { parseOfficeLocationSearchParams } from "@/lib/location/utils";
import { getOfficeLocations } from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";

export default async function OfficeLocationsPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  await requireAdmin();
  const query = parseOfficeLocationSearchParams(await searchParams);

  let result;
  try {
    result = await getOfficeLocations({
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
        action={<PrimaryLink href="/office-locations/new">Tambah Lokasi</PrimaryLink>}
        description="Kelola lokasi kantor yang dapat ditugaskan kepada pegawai USER Divisi Teknologi Informasi."
        title="Lokasi Kantor"
      />
      <OfficeLocationSearchForm query={query} />

      {result.items.length > 0 ? (
        <>
          <OfficeLocationTable locations={result.items} />
          <LocationPagination
            kind="office"
            query={query}
            totalItems={result.total_items}
            totalPages={result.total_pages}
          />
        </>
      ) : hasFilter ? (
        <EmptyState
          description="Tidak ada lokasi kantor yang cocok dengan kata kunci atau status yang dipilih."
          title="Hasil pencarian kosong"
        />
      ) : (
        <EmptyState
          action={<PrimaryLink href="/office-locations/new">Tambah Lokasi</PrimaryLink>}
          description="Belum ada lokasi kantor yang tersedia untuk penugasan pegawai."
          title="Belum ada lokasi kantor"
        />
      )}
    </section>
  );
}
