import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";

export default function LocationAssignmentsLoading() {
  return (
    <section className="mx-auto max-w-4xl">
      <EmptyState
        description="Mengambil data penugasan lokasi dari backend."
        title="Memuat penugasan lokasi"
      />
    </section>
  );
}
