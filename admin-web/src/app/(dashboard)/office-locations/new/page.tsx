import {
  PageHeader,
  SecondaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { OfficeLocationForm } from "@/app/(dashboard)/office-locations/_components/office-location-form";
import { requireAdmin } from "@/lib/server/session";

export default async function NewOfficeLocationPage() {
  await requireAdmin();

  return (
    <section className="mx-auto max-w-4xl">
      <PageHeader
        action={<SecondaryLink href="/office-locations">Kembali</SecondaryLink>}
        description="Tambahkan lokasi kantor aktif untuk penugasan pegawai."
        title="Tambah Lokasi Kantor"
      />
      <OfficeLocationForm mode="create" />
    </section>
  );
}
