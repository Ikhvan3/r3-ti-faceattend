import { PageHeader, SecondaryLink } from "@/app/(dashboard)/employees/_components/page-header";
import { WorkScheduleForm } from "@/app/(dashboard)/work-schedules/_components/work-schedule-form";
import { requireAdmin } from "@/lib/server/session";

export default async function NewWorkSchedulePage() {
  await requireAdmin();

  return (
    <section className="mx-auto max-w-4xl">
      <PageHeader
        action={<SecondaryLink href="/work-schedules">Kembali</SecondaryLink>}
        description="Tambahkan jadwal kerja dalam satu hari. Jadwal baru otomatis aktif."
        title="Tambah Jadwal Kerja"
      />
      <WorkScheduleForm mode="create" />
    </section>
  );
}
