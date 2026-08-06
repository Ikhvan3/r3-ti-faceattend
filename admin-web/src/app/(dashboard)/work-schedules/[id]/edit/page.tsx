import { notFound, redirect } from "next/navigation";

import {
  PageHeader,
  SecondaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { WorkScheduleForm } from "@/app/(dashboard)/work-schedules/_components/work-schedule-form";
import { SafeApiError } from "@/lib/auth/types";
import { getWorkScheduleByID } from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";

export default async function EditWorkSchedulePage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  await requireAdmin();
  const { id } = await params;

  let schedule;
  try {
    schedule = await getWorkScheduleByID(id);
  } catch (error) {
    if (
      error instanceof SafeApiError &&
      (error.code === "NOT_FOUND" || error.code === "BAD_REQUEST")
    ) {
      notFound();
    }
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      redirect("/login");
    }
    throw error;
  }

  return (
    <section className="mx-auto max-w-4xl">
      <PageHeader
        action={<SecondaryLink href={`/work-schedules/${schedule.id}`}>Kembali</SecondaryLink>}
        description="Perbarui nama, jam kerja, dan toleransi keterlambatan."
        title="Edit Jadwal Kerja"
      />
      <WorkScheduleForm mode="edit" schedule={schedule} />
    </section>
  );
}
