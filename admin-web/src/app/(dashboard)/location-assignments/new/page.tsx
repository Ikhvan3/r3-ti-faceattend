import { redirect } from "next/navigation";

import {
  PageHeader,
  SecondaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { LocationAssignmentForm } from "@/app/(dashboard)/location-assignments/_components/location-assignment-form";
import { SafeApiError } from "@/lib/auth/types";
import {
  getActiveOfficeLocations,
  getEmployeeOptions,
} from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";

export default async function NewLocationAssignmentPage() {
  await requireAdmin();

  let employees;
  let locations;
  try {
    [employees, locations] = await Promise.all([
      getEmployeeOptions(),
      getActiveOfficeLocations(),
    ]);
  } catch (error) {
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      redirect("/login");
    }
    throw error;
  }

  return (
    <section className="mx-auto max-w-4xl">
      <PageHeader
        action={<SecondaryLink href="/location-assignments">Kembali</SecondaryLink>}
        description="Tambahkan penugasan lokasi kantor untuk pegawai USER."
        title="Tambah Penugasan Lokasi"
      />
      <LocationAssignmentForm employees={employees} locations={locations} />
    </section>
  );
}
