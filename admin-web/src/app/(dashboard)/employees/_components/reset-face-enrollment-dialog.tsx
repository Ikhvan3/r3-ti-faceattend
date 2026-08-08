"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export function ResetFaceEnrollmentDialog({
  employeeID,
  employeeName,
}: {
  employeeID: string;
  employeeName: string;
}) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");

  async function resetEnrollment() {
    const normalizedReason = reason.trim();
    if (normalizedReason.length < 5) {
      setError("Alasan reset wajib diisi minimal 5 karakter.");
      return;
    }

    setError("");
    setIsSubmitting(true);

    try {
      const response = await fetch(
        `/api/admin/employees/${employeeID}/face-enrollment`,
        {
          method: "DELETE",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ reason: normalizedReason }),
        },
      );
      const payload = (await response.json()) as unknown;

      if (!response.ok) {
        const message = readMessage(payload);
        setError(message ?? "Reset enrollment wajah gagal diproses.");
        setIsSubmitting(false);
        return;
      }

      setIsSubmitting(false);
      setIsOpen(false);
      setReason("");
      router.refresh();
    } catch {
      setError("Layanan reset enrollment belum tersedia. Coba lagi nanti.");
      setIsSubmitting(false);
    }
  }

  return (
    <>
      <button
        className="inline-flex h-9 items-center justify-center rounded-md border border-red-200 bg-white px-3 text-sm font-semibold text-red-700 transition hover:bg-red-50 focus:outline-none focus:ring-2 focus:ring-red-100"
        onClick={() => {
          setError("");
          setReason("");
          setIsOpen(true);
        }}
        type="button"
      >
        Reset Enrollment
      </button>

      {isOpen ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
          <div
            aria-labelledby="reset-face-title"
            aria-modal="true"
            className="w-full max-w-md rounded-xl bg-white p-5 shadow-xl"
            role="dialog"
          >
            <h2
              className="text-lg font-semibold text-slate-950"
              id="reset-face-title"
            >
              Reset enrollment wajah?
            </h2>
            <p className="mt-2 text-sm leading-6 text-slate-600">
              Enrollment wajah <strong>{employeeName}</strong> akan dihapus. Pegawai
              harus melakukan enrollment ulang dari aplikasi mobile sebelum dapat
              menggunakan verifikasi wajah untuk absensi.
            </p>

            <label className="mt-4 block text-sm font-semibold text-slate-800">
              Alasan reset
              <textarea
                className="mt-2 min-h-24 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm font-normal text-slate-900 outline-none transition focus:border-emerald-500 focus:ring-2 focus:ring-emerald-100"
                disabled={isSubmitting}
                maxLength={1000}
                onChange={(event) => setReason(event.target.value)}
                placeholder="Contoh: Enrollment perlu diulang karena perubahan data biometrik telah diverifikasi admin."
                value={reason}
              />
            </label>
            <p className="mt-2 text-xs text-slate-500">
              Alasan ini akan disimpan permanen pada audit log.
            </p>

            <p className="mt-3 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
              Tindakan ini tidak menampilkan atau mengunduh template biometrik.
            </p>

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
                className="h-10 rounded-md bg-red-700 px-4 text-sm font-semibold text-white hover:bg-red-800 disabled:cursor-not-allowed disabled:bg-slate-400"
                disabled={isSubmitting || reason.trim().length < 5}
                onClick={resetEnrollment}
                type="button"
              >
                {isSubmitting ? "Mereset..." : "Ya, Reset Enrollment"}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}

function readMessage(value: unknown): string | null {
  if (typeof value !== "object" || value === null) {
    return null;
  }
  const message = (value as Record<string, unknown>).message;
  return typeof message === "string" ? message : null;
}
