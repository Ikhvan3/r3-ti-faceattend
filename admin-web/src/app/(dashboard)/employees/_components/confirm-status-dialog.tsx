"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import type { AccountStatus } from "@/lib/auth/types";
import type { Employee } from "@/lib/employee/types";
import {
  EMPLOYEE_STATUSES,
  safeEmployeeApiMessage,
  statusDescription,
  statusLabel,
} from "@/lib/employee/utils";
import { StatusBadge } from "@/app/(dashboard)/employees/_components/status-badge";

export function ConfirmStatusDialog({ employee }: { employee: Employee }) {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);
  const [selectedStatus, setSelectedStatus] = useState<AccountStatus>(
    employee.account_status,
  );
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function updateStatus() {
    setError("");
    setIsSubmitting(true);

    try {
      const response = await fetch(
        `/api/admin/employees/${employee.id}/status`,
        {
          method: "PATCH",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ account_status: selectedStatus }),
        },
      );
      const payload = await readStatusResponse(response);

      if (!response.ok || !payload.ok) {
        const message = payload.ok ? undefined : payload.message;
        setError(
          safeEmployeeApiMessage(
            response.status,
            message ?? "Status pegawai gagal diperbarui.",
          ),
        );
        setIsSubmitting(false);
        return;
      }

      setIsSubmitting(false);
      setIsOpen(false);
      router.refresh();
    } catch {
      setError("Layanan pegawai belum tersedia. Coba lagi nanti.");
      setIsSubmitting(false);
    }
  }

  return (
    <>
      <button
        className="inline-flex h-9 items-center justify-center rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-slate-800 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100"
        onClick={() => {
          setSelectedStatus(employee.account_status);
          setError("");
          setIsOpen(true);
        }}
        type="button"
      >
        Ubah Status
      </button>

      {isOpen ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 px-4">
          <div
            aria-labelledby="status-dialog-title"
            aria-modal="true"
            className="w-full max-w-md rounded-md bg-white p-5 shadow-xl"
            role="dialog"
          >
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2
                  className="text-lg font-semibold text-slate-950"
                  id="status-dialog-title"
                >
                  Ubah status pegawai
                </h2>
                <p className="mt-1 text-sm text-slate-600">{employee.name}</p>
              </div>
              <StatusBadge status={employee.account_status} />
            </div>

            <div className="mt-5 space-y-3">
              {EMPLOYEE_STATUSES.map((status) => (
                <label
                  className="flex cursor-pointer gap-3 rounded-md border border-slate-200 p-3 transition hover:bg-slate-50"
                  key={status}
                >
                  <input
                    checked={selectedStatus === status}
                    className="mt-1"
                    disabled={isSubmitting}
                    name="account_status"
                    onChange={() => setSelectedStatus(status)}
                    type="radio"
                    value={status}
                  />
                  <span>
                    <span className="block text-sm font-semibold text-slate-950">
                      {statusLabel(status)}
                    </span>
                    <span className="mt-1 block text-sm leading-5 text-slate-600">
                      {statusDescription(status)}
                    </span>
                  </span>
                </label>
              ))}
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
                disabled={isSubmitting}
                onClick={updateStatus}
                type="button"
              >
                {isSubmitting ? "Menyimpan..." : "Simpan Status"}
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}

type StatusPayload =
  | {
      ok: true;
    }
  | {
      ok: false;
      message: string;
    };

async function readStatusResponse(response: Response): Promise<StatusPayload> {
  try {
    const payload = (await response.json()) as unknown;
    if (isRecord(payload) && payload.status === "ok") {
      return { ok: true };
    }
    if (
      isRecord(payload) &&
      payload.status === "error" &&
      typeof payload.message === "string"
    ) {
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
