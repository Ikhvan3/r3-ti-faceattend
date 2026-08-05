import { EmployeeForm } from "@/app/(dashboard)/employees/_components/employee-form";
import { PageHeader, SecondaryLink } from "@/app/(dashboard)/employees/_components/page-header";
import { requireAdmin } from "@/lib/server/session";

export default async function NewEmployeePage() {
  await requireAdmin();

  return (
    <section className="mx-auto max-w-4xl">
      <PageHeader
        action={<SecondaryLink href="/employees">Kembali</SecondaryLink>}
        description="Buat akun pegawai USER baru. Role dan status awal ditentukan oleh backend, bukan dari form."
        title="Tambah Pegawai"
      />
      <EmployeeForm mode="create" />
    </section>
  );
}
