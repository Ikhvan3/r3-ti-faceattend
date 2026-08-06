import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";
import { SecondaryLink } from "@/app/(dashboard)/employees/_components/page-header";

export default function WorkScheduleNotFound() {
  return (
    <section className="mx-auto max-w-4xl">
      <EmptyState
        action={<SecondaryLink href="/work-schedules">Kembali ke Jadwal Kerja</SecondaryLink>}
        description="Jadwal kerja tidak ditemukan atau ID tidak valid."
        title="Jadwal kerja tidak ditemukan"
      />
    </section>
  );
}
