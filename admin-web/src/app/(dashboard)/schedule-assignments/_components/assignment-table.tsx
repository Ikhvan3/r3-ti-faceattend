import Link from "next/link";

import { ScheduleBadge } from "@/app/(dashboard)/work-schedules/_components/schedule-badge";
import type { ScheduleAssignment } from "@/lib/schedule/types";
import { formatDateOnly } from "@/lib/schedule/utils";

export function AssignmentTable({
  assignments,
}: {
  assignments: ScheduleAssignment[];
}) {
  return (
    <div className="overflow-hidden rounded-md border border-slate-200 bg-white">
      <div className="overflow-x-auto">
        <table className="min-w-[900px] w-full border-collapse text-left text-sm">
          <thead className="bg-slate-50 text-xs font-semibold uppercase text-slate-500">
            <tr>
              <th className="px-4 py-3">Pegawai</th>
              <th className="px-4 py-3">Jadwal</th>
              <th className="px-4 py-3">Mulai</th>
              <th className="px-4 py-3">Selesai</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Aksi</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200">
            {assignments.map((assignment) => (
              <tr className="align-top" key={assignment.id}>
                <td className="px-4 py-3">
                  <p className="font-semibold text-slate-950">
                    {assignment.user.employee_number}
                  </p>
                  <p className="mt-1 text-slate-700">{assignment.user.name}</p>
                </td>
                <td className="px-4 py-3 text-slate-700">
                  {assignment.schedule.name}
                </td>
                <td className="px-4 py-3 text-slate-700">
                  {formatDateOnly(assignment.effective_from)}
                </td>
                <td className="px-4 py-3 text-slate-700">
                  {formatDateOnly(assignment.effective_to)}
                </td>
                <td className="px-4 py-3">
                  <ScheduleBadge status={assignment.status} />
                </td>
                <td className="px-4 py-3">
                  <Link
                    className="inline-flex h-9 items-center rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-slate-800 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100"
                    href={`/schedule-assignments/${assignment.id}`}
                  >
                    Detail
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
