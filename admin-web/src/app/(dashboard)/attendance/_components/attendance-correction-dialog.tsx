"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export function AttendanceCorrectionDialog({
  attendanceID,
  employeeName,
  currentCheckIn,
  currentCheckOut,
}: {
  attendanceID: string;
  employeeName: string;
  currentCheckIn: string;
  currentCheckOut: string | null;
}) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [checkIn, setCheckIn] = useState(currentCheckIn);
  const [checkOut, setCheckOut] = useState(currentCheckOut ?? "");
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");

  async function submitCorrection() {
    const normalizedReason = reason.trim();
    if (!checkIn || normalizedReason.length < 5) {
      setError("Jam masuk dan alasan minimal 5 karakter wajib diisi.");
      return;
    }

    setError("");
    setIsSubmitting(true);
    try {
      const response = await fetch(
        `/api/admin/attendance/${attendanceID}/correction`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            check_in_time: checkIn,
            check_out_time: checkOut || null,
            reason: normalizedReason,
          }),
        },
      );
      const payload = (await response.json()) as unknown;
      if (!response.ok) {
        setError(readMessage(payload) ?? "Koreksi presensi gagal diproses.");
        setIsSubmitting(false);
        return;
      }

      setIsSubmitting(false);
      setIsOpen(false);
      setReason("");
      router.refresh();
    } catch {
      setError("Layanan koreksi presensi belum tersedia. Coba lagi nanti.");
      setIsSubmitting(false);
    }
  }

  return (
    <>
      <button
        className="inline-flex h-10 items-center justify-center rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white transition hover:bg-emerald-800 focus:outline-none focus:ring-2 focus:ring-emerald-200"
        onClick={() => {
          setCheckIn(currentCheckIn);
          setCheckOut(currentCheckOut ?? "");
          setReason("");
          setError("");
          setIsOpen(true);
        }}
        type="button"
      >
        Koreksi Presensi
      </button>

      {isOpen ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
          <div
            aria-labelledby="attendance-correction-title"
            aria-modal="true"
            className="w-full max-w-lg rounded-xl bg-white p-5 shadow-xl"
            role="dialog"
          >
            <h2
              className="text-lg font-semibold text-slate-950"
              id="attendance-correction-title"
            >
              Koreksi presensi {employeeName}
            </h2>
            <p className="mt-2 text-sm leading-6 text-slate-600">
              Koreksi hanya mengubah waktu presensi. Bukti GPS yang sudah ada tidak
              dimanipulasi, dan setiap perubahan disimpan pada audit log.
            </p>

            <div className="mt-5 grid gap-4 sm:grid-cols-2">
              <label className="text-sm font-semibold text-slate-800">
                Jam masuk
                <input
                  className="mt-2 h-10 w-full rounded-lg border border-slate-300 px-3 text-sm font-normal text-slate-900 outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100"
                  disabled={isSubmitting}
                  onChange={(event) => setCheckIn(event.target.value)}
                  type="time"
                  value={checkIn}
                />
              </label>
              <label className="text-sm font-semibold text-slate-800">
                Jam pulang
                <input
                  className="mt-2 h-10 w-full rounded-lg border border-slate-300 px-3 text-sm font-normal text-slate-900 outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100"
                  disabled={isSubmitting}
                  onChange={(event) => setCheckOut(event.target.value)}
                  type="time"
                  value={checkOut}
                />
                <span className="mt-1 block text-xs font-normal text-slate-500">
                  Kosongkan jika pegawai memang belum check-out.
                </span>
              </label>
            </div>

            <label className="mt-4 block text-sm font-semibold text-slate-800">
              Alasan koreksi
              <textarea
                className="mt-2 min-h-24 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm font-normal text-slate-900 outline-none focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100"
                disabled={isSubmitting}
                maxLength={1000}
                onChange={(event) => setReason(event.target.value)}
                placeholder="Contoh: Pegawai lupa check-out dan jam pulang telah dikonfirmasi oleh atasan."
                value={reason}
              />
            </label>

            <div className="mt-4 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-900">
              Jika jam pulang ditambahkan secara manual, bukti lokasi check-out tetap
              kosong. Sistem tidak membuat data GPS palsu.
            </div>

            {error ? (
              <p className="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                {error}
              </p>
            ) : null}

            <div className="mt-5 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
              <button
                className="h-10 rounded-md border border-slate-300 px-4 text-sm font-medium text-slate-800 hover:bg-slate-50 disabled:cursor-not-allowed disabled:text-slate-400"
                disabled={isSubmitting}
                onClick={() => setIsOpen(false)}
                type="button"
              >
                Batal
              </button>
              <button
                className="h-10 rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white hover:bg-emerald-800 disabled:cursor-not-allowed disabled:bg-slate-400"
                disabled={isSubmitting || !checkIn || reason.trim().length < 5}
                onClick={submitCorrection}
                type="button"
              >
                {isSubmitting ? "Menyimpan..." : "Simpan Koreksi"}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}

function readMessage(value: unknown): string | null {
  if (typeof value !== "object" || value === null) return null;
  const message = (value as Record<string, unknown>).message;
  return typeof message === "string" ? message : null;
}
