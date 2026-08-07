import type { AttendanceSearchQuery } from "@/lib/attendance/utils";

export function AttendanceFilterForm({
  query,
}: {
  query: AttendanceSearchQuery;
}) {
  return (
    <form className="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 lg:grid-cols-6" method="get">
      <label className="space-y-1 text-sm text-slate-700">
        <span className="font-medium">Tanggal mulai</span>
        <input
          className="h-10 w-full rounded-md border border-slate-300 px-3 outline-none focus:border-emerald-600 focus:ring-2 focus:ring-emerald-100"
          defaultValue={query.date_from}
          name="date_from"
          type="date"
        />
      </label>
      <label className="space-y-1 text-sm text-slate-700">
        <span className="font-medium">Tanggal selesai</span>
        <input
          className="h-10 w-full rounded-md border border-slate-300 px-3 outline-none focus:border-emerald-600 focus:ring-2 focus:ring-emerald-100"
          defaultValue={query.date_to}
          name="date_to"
          type="date"
        />
      </label>
      <label className="space-y-1 text-sm text-slate-700 lg:col-span-2">
        <span className="font-medium">Cari pegawai</span>
        <input
          className="h-10 w-full rounded-md border border-slate-300 px-3 outline-none focus:border-emerald-600 focus:ring-2 focus:ring-emerald-100"
          defaultValue={query.search}
          name="search"
          placeholder="Nama, nomor pegawai, atau email"
          type="search"
        />
      </label>
      <label className="space-y-1 text-sm text-slate-700">
        <span className="font-medium">Status</span>
        <select
          className="h-10 w-full rounded-md border border-slate-300 bg-white px-3 outline-none focus:border-emerald-600 focus:ring-2 focus:ring-emerald-100"
          defaultValue={query.attendance_state}
          name="attendance_state"
        >
          <option value="">Semua</option>
          <option value="NOT_CHECKED_IN">Belum check-in</option>
          <option value="CHECKED_IN">Sudah check-in</option>
          <option value="CHECKED_OUT">Sudah check-out</option>
        </select>
      </label>
      <label className="space-y-1 text-sm text-slate-700">
        <span className="font-medium">Keterlambatan</span>
        <select
          className="h-10 w-full rounded-md border border-slate-300 bg-white px-3 outline-none focus:border-emerald-600 focus:ring-2 focus:ring-emerald-100"
          defaultValue={query.is_late}
          name="is_late"
        >
          <option value="">Semua</option>
          <option value="true">Terlambat</option>
          <option value="false">Tidak terlambat</option>
        </select>
      </label>
      <input name="page_size" type="hidden" value={query.page_size} />
      <div className="flex gap-2 lg:col-span-6">
        <button
          className="inline-flex h-10 items-center justify-center rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white transition hover:bg-emerald-800"
          type="submit"
        >
          Terapkan
        </button>
        <a
          className="inline-flex h-10 items-center justify-center rounded-md border border-slate-300 bg-white px-4 text-sm font-medium text-slate-700 transition hover:bg-slate-50"
          href="/attendance"
        >
          Reset
        </a>
      </div>
    </form>
  );
}
