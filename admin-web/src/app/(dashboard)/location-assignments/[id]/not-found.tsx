import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";
import { SecondaryLink } from "@/app/(dashboard)/employees/_components/page-header";

export default function LocationAssignmentNotFound() {
  return (
    <section className="mx-auto max-w-4xl">
      <EmptyState
        action={<SecondaryLink href="/location-assignments">Kembali</SecondaryLink>}
        description="Penugasan lokasi tidak ditemukan atau sudah tidak tersedia."
        title="Penugasan lokasi tidak ditemukan"
      />
    </section>
  );
}
