import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";

export default function OfficeLocationsLoading() {
  return (
    <section className="mx-auto max-w-4xl">
      <EmptyState
        description="Mengambil data lokasi kantor dari backend."
        title="Memuat lokasi kantor"
      />
    </section>
  );
}
