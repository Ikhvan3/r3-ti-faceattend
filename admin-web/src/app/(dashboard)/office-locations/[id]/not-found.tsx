import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";
import { SecondaryLink } from "@/app/(dashboard)/employees/_components/page-header";

export default function OfficeLocationNotFound() {
  return (
    <section className="mx-auto max-w-4xl">
      <EmptyState
        action={<SecondaryLink href="/office-locations">Kembali</SecondaryLink>}
        description="Lokasi kantor tidak ditemukan atau sudah tidak tersedia."
        title="Lokasi kantor tidak ditemukan"
      />
    </section>
  );
}
