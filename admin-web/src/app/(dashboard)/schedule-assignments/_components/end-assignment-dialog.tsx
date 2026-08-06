"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { readAssignmentResponse } from "@/app/(dashboard)/schedule-assignments/_components/assignment-form";
import type { ScheduleAssignment } from "@/lib/schedule/types";
import { safeScheduleApiMessage } from "@/lib/schedule/utils";

export function EndAssignmentDialog({
  assignment,
}: {
  assignment: ScheduleAssignment;
}) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [effectiveTo, setEffectiveTo] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  if (assignment.status === "ENDED") {
    return null;
  }

  async function endAssignment() {
    if (!effectiveTo) {
      setError("Tanggal akhir wajib diisi.");
      return;
    }
    if (effectiveTo < assignment.effective_from) {
      setError("Tanggal akhir tidak boleh sebelum tanggal mulai.");
      return;
    }

    setError("");
    setIsSubmitting(true);
    try {
      const response = await fetch(
        `/api/admin/schedule-assignments/${assignment.id}/end`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ effective_to: effectiveTo }),
        },
      );
      const payload = await readAssignmentResponse(response);

      if (!response.ok || !payload.ok) {
        setError(
          safeScheduleApiMessage(
            response.status,
            payload.ok ? "Penugasan jadwal gagal diakhiri." : payload.message,
          ),
        );
        setIsSubmitting(false);
        return;
      }

      setIsSubmitting(false);
      setIsOpen(false);
      router.refresh();
    } catch {
      setError("Layanan jadwal belum tersedia. Coba lagi nanti.");
      setIsSubmitting(false);
    }
  }

  return (
    <>
      <button
        className="inline-flex h-10 items-center justify-center rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white transition hover:bg-emerald-800 focus:outline-none focus:ring-2 focus:ring-emerald-200"
        onClick={() => {
          setEffectiveTo(assignment.effective_to ?? "");
          setError("");
          setIsOpen(true);
        }}
        type="button"
      >
        Akhiri Penugasan
      </button>
      {isOpen ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
          <div
            aria-labelledby="end-assignment-dialog-title"
            aria-modal="true"
            className="w-full max-w-md rounded-md bg-white p-5 shadow-xl"
            role="dialog"
          >
            <h2
              className="text-lg font-semibold text-slate-950"
              id="end-assignment-dialog-title"
            >
              Akhiri penugasan jadwal
            </h2>
            <p className="mt-2 text-sm leading-6 text-slate-600">
              Tanggal akhir bersifat inklusif. Pegawai masih memakai jadwal ini
              sampai akhir tanggal yang dipilih.
            </p>
            <div className="mt-5 space-y-2">
              <label className="text-sm font-medium text-slate-800" htmlFor="effective_to">
                Tanggal akhir
              </label>
              <input
                className="h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-emerald-700 focus:ring-2 focus:ring-emerald-100 disabled:cursor-not-allowed disabled:bg-slate-100"
                disabled={isSubmitting}
                id="effective_to"
                min={assignment.effective_from}
                onChange={(event) => setEffectiveTo(event.target.value)}
                type="date"
                value={effectiveTo}
              />
            </div>
            {error ? <p className={errorClassName}>{error}</p> : null}
            <div className="mt-5 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
              <button
                className={outlineButtonClassName}
                disabled={isSubmitting}
                onClick={() => setIsOpen(false)}
                type="button"
              >
                Batal
              </button>
              <button
                className={primaryButtonClassName}
                disabled={isSubmitting}
                onClick={endAssignment}
                type="button"
              >
                {isSubmitting ? "Menyimpan..." : "Simpan"}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}

const outlineButtonClassName =
  "h-10 rounded-md border border-slate-300 px-4 text-sm font-medium text-slate-800 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100 disabled:cursor-not-allowed disabled:text-slate-400";
const primaryButtonClassName =
  "h-10 rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white hover:bg-emerald-800 disabled:cursor-not-allowed disabled:bg-slate-400";
const errorClassName =
  "mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700";
