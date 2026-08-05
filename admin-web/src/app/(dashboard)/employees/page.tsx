import { redirect } from "next/navigation";

import { EmployeeSearchForm } from "@/app/(dashboard)/employees/_components/employee-search-form";
import { EmployeeTable } from "@/app/(dashboard)/employees/_components/employee-table";
import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";
import { PageHeader, PrimaryLink } from "@/app/(dashboard)/employees/_components/page-header";
import { Pagination } from "@/app/(dashboard)/employees/_components/pagination";
import { SafeApiError } from "@/lib/auth/types";
import { parseEmployeeSearchParams } from "@/lib/employee/utils";
import { getEmployees } from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";

export default async function EmployeesPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  await requireAdmin();
  const params = await searchParams;
  const query = parseEmployeeSearchParams(params);

  let result;
  try {
    result = await getEmployees({
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
        action={<PrimaryLink href="/employees/new">Tambah Pegawai</PrimaryLink>}
        description="Kelola akun pegawai USER untuk Divisi Teknologi Informasi. Gunakan hanya data dummy pada development."
        title="Pegawai TI"
      />

      <EmployeeSearchForm query={query} />

      {result.items.length > 0 ? (
        <>
          <EmployeeTable employees={result.items} />
          <Pagination
            query={query}
            totalItems={result.total_items}
            totalPages={result.total_pages}
          />
        </>
      ) : hasFilter ? (
        <EmptyState
          description="Tidak ada pegawai yang cocok dengan kata kunci atau status yang dipilih."
          title="Hasil pencarian kosong"
        />
      ) : (
        <EmptyState
          action={<PrimaryLink href="/employees/new">Tambah Pegawai</PrimaryLink>}
          description="Belum ada pegawai dummy yang tersedia untuk dikelola pada modul ini."
          title="Belum ada pegawai"
        />
      )}
    </section>
  );
}
