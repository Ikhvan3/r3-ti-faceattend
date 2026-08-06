"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { LocationBadge } from "@/app/(dashboard)/office-locations/_components/location-badge";
import type { OfficeLocation } from "@/lib/location/types";
import {
  officeLocationStatusFromActive,
  safeLocationApiMessage,
} from "@/lib/location/utils";

export function ConfirmOfficeLocationStatusDialog({
  location,
}: {
  location: OfficeLocation;
}) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const nextActive = !location.is_active;

  async function updateStatus() {
    setError("");
    setIsSubmitting(true);

    try {
      const response = await fetch(
        `/api/admin/office-locations/${location.id}/status`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ is_active: nextActive }),
        },
      );
      const payload = await readDialogResponse(response);

      if (!response.ok || !payload.ok) {
        setError(
          safeLocationApiMessage(
            response.status,
            payload.ok ? "Status lokasi kantor gagal diperbarui." : payload.message,
          ),
        );
        setIsSubmitting(false);
        return;
      }

      setIsSubmitting(false);
      setIsOpen(false);
      router.refresh();
    } catch {
      setError("Layanan lokasi belum tersedia. Coba lagi nanti.");
      setIsSubmitting(false);
    }
  }

  return (
    <>
      <button className={outlineButtonClassName} onClick={() => setIsOpen(true)} type="button">
        {nextActive ? "Aktifkan" : "Nonaktifkan"}
      </button>
      {isOpen ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
          <div
            aria-labelledby="office-location-status-dialog-title"
            aria-modal="true"
            className="w-full max-w-md rounded-md bg-white p-5 shadow-xl"
            role="dialog"
          >
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold text-slate-950" id="office-location-status-dialog-title">
                  {nextActive ? "Aktifkan lokasi" : "Nonaktifkan lokasi"}
                </h2>
                <p className="mt-1 text-sm text-slate-600">{location.name}</p>
              </div>
              <LocationBadge status={officeLocationStatusFromActive(location.is_active)} />
            </div>

            <p className="mt-5 text-sm leading-6 text-slate-700">
              {nextActive
                ? "Lokasi kantor ini akan dapat dipilih lagi untuk penugasan pegawai."
                : "Lokasi kantor hanya dapat dinonaktifkan jika tidak sedang ditugaskan kepada pegawai."}
            </p>

            {error ? <p className={errorClassName}>{error}</p> : null}

            <div className="mt-5 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
              <button className={outlineButtonClassName} disabled={isSubmitting} onClick={() => setIsOpen(false)} type="button">
                Batal
              </button>
              <button className={primaryButtonClassName} disabled={isSubmitting} onClick={updateStatus} type="button">
                {isSubmitting ? "Menyimpan..." : "Konfirmasi"}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}

type DialogPayload = { ok: true } | { ok: false; message: string };

async function readDialogResponse(response: Response): Promise<DialogPayload> {
  try {
    const payload = (await response.json()) as unknown;
    if (isRecord(payload) && payload.status === "ok") {
      return { ok: true };
    }
    if (isRecord(payload) && payload.status === "error" && typeof payload.message === "string") {
      return { ok: false, message: payload.message };
    }
  } catch {
    return { ok: false, message: "Respons tidak valid." };
  }
  return { ok: false, message: "Respons tidak valid." };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

const outlineButtonClassName =
  "inline-flex h-9 items-center justify-center rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-slate-800 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100 disabled:cursor-not-allowed disabled:text-slate-400";
const primaryButtonClassName =
  "h-10 rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white hover:bg-emerald-800 disabled:cursor-not-allowed disabled:bg-slate-400";
const errorClassName =
  "mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700";
