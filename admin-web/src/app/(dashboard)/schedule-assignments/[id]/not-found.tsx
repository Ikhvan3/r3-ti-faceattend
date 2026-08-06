import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";
import { SecondaryLink } from "@/app/(dashboard)/employees/_components/page-header";

export default function ScheduleAssignmentNotFound() {
  return (
    <section className="mx-auto max-w-4xl">
      <EmptyState
        action={
          <SecondaryLink href="/schedule-assignments">
            Kembali ke Penugasan Jadwal
          </SecondaryLink>
        }
        description="Penugasan jadwal tidak ditemukan atau ID tidak valid."
        title="Penugasan jadwal tidak ditemukan"
      />
    </section>
  );
}
