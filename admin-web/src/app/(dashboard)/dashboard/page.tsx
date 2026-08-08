import Link from "next/link";
import { redirect } from "next/navigation";

import { SafeApiError } from "@/lib/auth/types";
import {
  attendanceStateLabel,
  formatBusinessDate,
  formatBusinessTime,
} from "@/lib/attendance/utils";
import {
  getAdminAttendance,
  getAdminAttendanceSummary,
} from "@/lib/server/admin-attendance-bff";
import { requireAdmin } from "@/lib/server/session";

export default async function DashboardPage() {
  const user = await requireAdmin();

  let summary;
  let attendance;
  let errorMessage = "";
  try {
    summary = await getAdminAttendanceSummary();
    attendance = await getAdminAttendance({ page: 1, page_size: 8 });
  } catch (error) {
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      redirect("/login");
    }
    if (error instanceof SafeApiError) {
      errorMessage = error.message;
    } else {
      throw error;
    }
  }

  return (
    <section className="mx-auto max-w-7xl space-y-6">
      <div>
        <p className="text-sm font-medium text-emerald-700">Dashboard Admin TI</p>
        <h1 className="mt-1 text-2xl font-semibold text-slate-950">
          Selamat datang, {user.name}
        </h1>
        <p className="mt-2 text-sm text-slate-600">
          Ringkasan presensi pegawai berdasarkan tanggal bisnis Asia/Jakarta.
        </p>
      </div>

      {errorMessage ? (
        <div className="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-800">
          {errorMessage}
        </div>
      ) : summary ? (
        <>
          <div>
            <div className="flex items-end justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold text-slate-950">
                  Ringkasan Hari Ini
                </h2>
                <p className="mt-1 text-sm text-slate-500">
                  {formatBusinessDate(summary.date)}
                </p>
              </div>
            </div>
            <div className="mt-4 grid gap-4 sm:grid-cols-2 xl:grid-cols-5">
              <MetricCard label="Pegawai dijadwalkan" value={summary.active_employees} />
              <MetricCard label="Sudah check-in" value={summary.checked_in} />
              <MetricCard label="Belum check-in" value={summary.not_checked_in} />
              <MetricCard label="Sudah check-out" value={summary.checked_out} />
              <MetricCard label="Terlambat" value={summary.late} warning />
            </div>
          </div>

          <div className="rounded-xl border border-slate-200 bg-white">
            <div className="flex items-center justify-between border-b border-slate-200 px-5 py-4">
              <div>
                <h2 className="font-semibold text-slate-950">Presensi Hari Ini</h2>
                <p className="mt-1 text-sm text-slate-500">
                  Status terbaru pegawai yang memiliki jadwal aktif hari ini.
                </p>
              </div>
              <Link
                className="text-sm font-semibold text-emerald-700 hover:text-emerald-900"
                href="/attendance"
              >
                Lihat semua
              </Link>
            </div>
            {attendance && attendance.items.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-slate-200 text-sm">
                  <thead className="bg-slate-50 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
                    <tr>
                      <th className="px-5 py-3">Pegawai</th>
                      <th className="px-5 py-3">Masuk</th>
                      <th className="px-5 py-3">Pulang</th>
                      <th className="px-5 py-3">Status</th>
                      <th className="px-5 py-3">Keterangan</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {attendance.items.map((item) => (
                      <tr key={`${item.employee.id}-${item.attendance_date}`}>
                        <td className="px-5 py-3">
                          <p className="font-semibold text-slate-950">{item.employee.name}</p>
                          <p className="mt-1 text-xs text-slate-500">{item.employee.employee_number}</p>
                        </td>
                        <td className="px-5 py-3 text-slate-700">{formatBusinessTime(item.check_in_at)}</td>
                        <td className="px-5 py-3 text-slate-700">{formatBusinessTime(item.check_out_at)}</td>
                        <td className="px-5 py-3 text-slate-700">{attendanceStateLabel(item.attendance_state)}</td>
                        <td className="px-5 py-3">
                          {item.is_late ? (
                            <span className="rounded-full bg-amber-50 px-2.5 py-1 text-xs font-semibold text-amber-800">
                              Terlambat
                            </span>
                          ) : (
                            <span className="text-slate-500">-</span>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="px-5 py-8 text-sm text-slate-500">
                Belum ada pegawai dengan jadwal aktif pada tanggal bisnis hari ini.
              </p>
            )}
          </div>
        </>
      ) : null}
    </section>
  );
}

function MetricCard({
  label,
  value,
  warning = false,
}: {
  label: string;
  value: number;
  warning?: boolean;
}) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5">
      <p className="text-sm font-medium text-slate-500">{label}</p>
      <p className={`mt-2 text-3xl font-semibold ${warning ? "text-amber-700" : "text-slate-950"}`}>
        {value}
      </p>
    </div>
  );
}
