import { notFound, redirect } from "next/navigation";

import { DetailItem } from "@/app/(dashboard)/employees/_components/detail-item";
import {
  PageHeader,
  SecondaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { EndAssignmentDialog } from "@/app/(dashboard)/schedule-assignments/_components/end-assignment-dialog";
import { ScheduleBadge } from "@/app/(dashboard)/work-schedules/_components/schedule-badge";
import { SafeApiError } from "@/lib/auth/types";
import { getScheduleAssignmentByID } from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";
import {
  formatDateOnly,
  formatDateTime,
  formatScheduleTime,
} from "@/lib/schedule/utils";

export default async function ScheduleAssignmentDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  await requireAdmin();
  const { id } = await params;

  let assignment;
  try {
    assignment = await getScheduleAssignmentByID(id);
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
            <SecondaryLink href="/schedule-assignments">Kembali</SecondaryLink>
            <EndAssignmentDialog assignment={assignment} />
          </div>
        }
        description="Detail penugasan jadwal pegawai. Tanggal mulai dan akhir bersifat inklusif."
        title={`${assignment.user.name} - ${assignment.schedule.name}`}
      />
      <div className="rounded-md border border-slate-200 bg-white p-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-sm font-medium text-slate-500">
              {assignment.user.employee_number}
            </p>
            <p className="mt-1 text-lg font-semibold text-slate-950">
              {formatDateOnly(assignment.effective_from)} -{" "}
              {formatDateOnly(assignment.effective_to)}
            </p>
          </div>
          <ScheduleBadge status={assignment.status} />
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <DetailItem label="ID Assignment" value={assignment.id} />
        <DetailItem label="Nomor Pegawai" value={assignment.user.employee_number} />
        <DetailItem label="Nama Pegawai" value={assignment.user.name} />
        <DetailItem label="Email Pegawai" value={assignment.user.email} />
        <DetailItem label="Jadwal" value={assignment.schedule.name} />
        <DetailItem
          label="Jam Kerja"
          value={`${formatScheduleTime(assignment.schedule.start_time)} - ${formatScheduleTime(assignment.schedule.end_time)}`}
        />
        <DetailItem label="Tanggal Mulai" value={formatDateOnly(assignment.effective_from)} />
        <DetailItem label="Tanggal Akhir" value={formatDateOnly(assignment.effective_to)} />
        <DetailItem label="Status" value={assignment.status} />
        <DetailItem label="Dibuat" value={formatDateTime(assignment.created_at)} />
        <DetailItem label="Diperbarui" value={formatDateTime(assignment.updated_at)} />
      </div>
    </section>
  );
}
