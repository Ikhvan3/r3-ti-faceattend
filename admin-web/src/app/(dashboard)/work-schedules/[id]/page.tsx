import { notFound, redirect } from "next/navigation";

import { DetailItem } from "@/app/(dashboard)/employees/_components/detail-item";
import {
  PageHeader,
  SecondaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { ConfirmScheduleStatusDialog } from "@/app/(dashboard)/work-schedules/_components/confirm-schedule-status-dialog";
import { ScheduleBadge } from "@/app/(dashboard)/work-schedules/_components/schedule-badge";
import { SafeApiError } from "@/lib/auth/types";
import { getWorkScheduleByID } from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";
import {
  formatDateTime,
  formatScheduleTime,
  scheduleStatusFromActive,
} from "@/lib/schedule/utils";

export default async function WorkScheduleDetailPage({
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
    <section className="mx-auto max-w-5xl space-y-6">
      <PageHeader
        action={
          <div className="flex flex-wrap gap-2">
            <SecondaryLink href="/work-schedules">Kembali</SecondaryLink>
            <SecondaryLink href={`/work-schedules/${schedule.id}/edit`}>
              Edit
            </SecondaryLink>
            <ConfirmScheduleStatusDialog schedule={schedule} />
          </div>
        }
        description="Detail jadwal kerja aman tanpa data pegawai atau record absensi."
        title={schedule.name}
      />
      <div className="rounded-md border border-slate-200 bg-white p-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-sm font-medium text-slate-500">Jam kerja</p>
            <p className="mt-1 text-lg font-semibold text-slate-950">
              {formatScheduleTime(schedule.start_time)} -{" "}
              {formatScheduleTime(schedule.end_time)}
            </p>
          </div>
          <ScheduleBadge status={scheduleStatusFromActive(schedule.is_active)} />
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <DetailItem label="ID" value={schedule.id} />
        <DetailItem label="Nama" value={schedule.name} />
        <DetailItem label="Jam Mulai" value={formatScheduleTime(schedule.start_time)} />
        <DetailItem label="Jam Selesai" value={formatScheduleTime(schedule.end_time)} />
        <DetailItem label="Toleransi" value={`${schedule.grace_minutes} menit`} />
        <DetailItem label="Status" value={schedule.is_active ? "Aktif" : "Nonaktif"} />
        <DetailItem label="Dibuat" value={formatDateTime(schedule.created_at)} />
        <DetailItem label="Diperbarui" value={formatDateTime(schedule.updated_at)} />
      </div>
    </section>
  );
}
