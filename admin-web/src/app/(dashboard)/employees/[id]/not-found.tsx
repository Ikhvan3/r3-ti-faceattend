import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";
import { SecondaryLink } from "@/app/(dashboard)/employees/_components/page-header";

export default function EmployeeNotFound() {
  return (
    <section className="mx-auto max-w-3xl">
      <EmptyState
        action={<SecondaryLink href="/employees">Kembali ke Pegawai</SecondaryLink>}
        description="Pegawai tidak ditemukan atau bukan akun USER yang dapat dikelola melalui modul ini."
        title="Pegawai tidak ditemukan"
      />
    </section>
  );
}
