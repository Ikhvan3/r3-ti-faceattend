import Link from "next/link";

import { ConfirmScheduleStatusDialog } from "@/app/(dashboard)/work-schedules/_components/confirm-schedule-status-dialog";
import { ScheduleBadge } from "@/app/(dashboard)/work-schedules/_components/schedule-badge";
import type { WorkSchedule } from "@/lib/schedule/types";
import {
  formatScheduleTime,
  scheduleStatusFromActive,
} from "@/lib/schedule/utils";

export function WorkScheduleTable({
  schedules,
}: {
  schedules: WorkSchedule[];
}) {
  return (
    <div className="overflow-hidden rounded-md border border-slate-200 bg-white">
      <div className="overflow-x-auto">
        <table className="min-w-[760px] w-full border-collapse text-left text-sm">
          <thead className="bg-slate-50 text-xs font-semibold uppercase text-slate-500">
            <tr>
              <th className="px-4 py-3">Nama</th>
              <th className="px-4 py-3">Jam Kerja</th>
              <th className="px-4 py-3">Toleransi</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Aksi</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200">
            {schedules.map((schedule) => (
              <tr className="align-top" key={schedule.id}>
                <td className="px-4 py-3 font-semibold text-slate-950">
                  {schedule.name}
                </td>
                <td className="px-4 py-3 text-slate-700">
                  {formatScheduleTime(schedule.start_time)} -{" "}
                  {formatScheduleTime(schedule.end_time)}
                </td>
                <td className="px-4 py-3 text-slate-700">
                  {schedule.grace_minutes} menit
                </td>
                <td className="px-4 py-3">
                  <ScheduleBadge
                    status={scheduleStatusFromActive(schedule.is_active)}
                  />
                </td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap gap-2">
                    <ActionLink href={`/work-schedules/${schedule.id}`}>
                      Detail
                    </ActionLink>
                    <ActionLink href={`/work-schedules/${schedule.id}/edit`}>
                      Edit
                    </ActionLink>
                    <ConfirmScheduleStatusDialog schedule={schedule} />
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function ActionLink({
  href,
  children,
}: {
  href: string;
  children: string;
}) {
  return (
    <Link
      className="inline-flex h-9 items-center rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-slate-800 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100"
      href={href}
    >
      {children}
    </Link>
  );
}
