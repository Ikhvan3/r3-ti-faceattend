import Link from "next/link";

import type { AdminAttendanceListItem } from "@/lib/attendance/types";
import {
  attendanceStateLabel,
  formatBusinessDate,
  formatBusinessTime,
} from "@/lib/attendance/utils";

export function AttendanceTable({
  items,
}: {
  items: AdminAttendanceListItem[];
}) {
  return (
    <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-slate-200 text-sm">
          <thead className="bg-slate-50 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
            <tr>
              <th className="px-4 py-3">Pegawai</th>
              <th className="px-4 py-3">Tanggal</th>
              <th className="px-4 py-3">Jadwal</th>
              <th className="px-4 py-3">Masuk</th>
              <th className="px-4 py-3">Pulang</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Keterlambatan</th>
              <th className="px-4 py-3">Lokasi</th>
              <th className="px-4 py-3 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 text-slate-700">
            {items.map((item) => (
              <tr key={`${item.employee.id}-${item.attendance_date}`}>
                <td className="whitespace-nowrap px-4 py-3">
                  <p className="font-semibold text-slate-950">
                    {item.employee.name}
                  </p>
                  <p className="mt-1 text-xs text-slate-500">
                    {item.employee.employee_number}
                  </p>
                </td>
                <td className="whitespace-nowrap px-4 py-3">
                  {formatBusinessDate(item.attendance_date)}
                </td>
                <td className="whitespace-nowrap px-4 py-3">
                  <p>{item.schedule.name}</p>
                  <p className="mt-1 text-xs text-slate-500">
                    {item.schedule.start_time}–{item.schedule.end_time}
                  </p>
                </td>
                <td className="whitespace-nowrap px-4 py-3">
                  {formatBusinessTime(item.check_in_at)}
                </td>
                <td className="whitespace-nowrap px-4 py-3">
                  {formatBusinessTime(item.check_out_at)}
                </td>
                <td className="whitespace-nowrap px-4 py-3">
                  <StatusBadge state={item.attendance_state} />
                </td>
                <td className="whitespace-nowrap px-4 py-3">
                  {item.is_late ? (
                    <span className="rounded-full bg-amber-50 px-2.5 py-1 text-xs font-semibold text-amber-800">
                      Terlambat
                    </span>
                  ) : (
                    <span className="text-slate-500">Tidak</span>
                  )}
                </td>
                <td className="whitespace-nowrap px-4 py-3">
                  {item.office_location?.name ?? "-"}
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-right">
                  {item.id ? (
                    <Link
                      className="font-semibold text-emerald-700 hover:text-emerald-900"
                      href={`/attendance/${item.id}`}
                    >
                      Detail
                    </Link>
                  ) : (
                    <span className="text-xs text-slate-400">Belum tersedia</span>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export function StatusBadge({
  state,
}: {
  state: AdminAttendanceListItem["attendance_state"];
}) {
  const style =
    state === "CHECKED_OUT"
      ? "bg-emerald-50 text-emerald-800"
      : state === "CHECKED_IN"
        ? "bg-sky-50 text-sky-800"
        : "bg-slate-100 text-slate-700";
  return (
    <span className={`rounded-full px-2.5 py-1 text-xs font-semibold ${style}`}>
      {attendanceStateLabel(state)}
    </span>
  );
}
