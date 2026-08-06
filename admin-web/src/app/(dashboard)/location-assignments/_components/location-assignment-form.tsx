"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import type { Employee } from "@/lib/employee/types";
import type { LocationAssignment, OfficeLocation } from "@/lib/location/types";
import { isDateOnly, safeLocationApiMessage } from "@/lib/location/utils";

type FormState = {
  user_id: string;
  office_location_id: string;
  effective_from: string;
  effective_to: string;
  error: string;
  isSubmitting: boolean;
};

export function LocationAssignmentForm({
  employees,
  locations,
}: {
  employees: Employee[];
  locations: OfficeLocation[];
}) {
  const router = useRouter();
  const [state, setState] = useState<FormState>({
    user_id: "",
    office_location_id: "",
    effective_from: "",
    effective_to: "",
    error: "",
    isSubmitting: false,
  });

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!state.user_id || !state.office_location_id || !state.effective_from) {
      setState((current) => ({ ...current, error: "Pegawai, lokasi, dan tanggal mulai wajib diisi." }));
      return;
    }
    if (!isDateOnly(state.effective_from) || (state.effective_to && !isDateOnly(state.effective_to))) {
      setState((current) => ({ ...current, error: "Format tanggal tidak valid." }));
      return;
    }
    if (state.effective_to && state.effective_to < state.effective_from) {
      setState((current) => ({ ...current, error: "Tanggal akhir tidak boleh sebelum tanggal mulai." }));
      return;
    }

    setState((current) => ({ ...current, error: "", isSubmitting: true }));
    try {
      const response = await fetch("/api/admin/location-assignments", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          user_id: state.user_id,
          office_location_id: state.office_location_id,
          effective_from: state.effective_from,
          effective_to: state.effective_to || null,
        }),
      });
      const payload = await readLocationAssignmentResponse(response);
      if (!response.ok || !payload.ok) {
        setState((current) => ({
          ...current,
          error: safeLocationApiMessage(
            response.status,
            payload.ok ? "Penugasan lokasi gagal dibuat." : payload.message,
          ),
          isSubmitting: false,
        }));
        return;
      }

      setState((current) => ({ ...current, isSubmitting: false }));
      router.replace(`/location-assignments/${payload.assignment.id}`);
      router.refresh();
    } catch {
      setState((current) => ({
        ...current,
        error: "Layanan lokasi belum tersedia. Coba lagi nanti.",
        isSubmitting: false,
      }));
    }
  }

  return (
    <form className="mt-6 space-y-5 rounded-md border border-slate-200 bg-white p-5" onSubmit={handleSubmit}>
      <div className="grid gap-5 md:grid-cols-2">
        <Field htmlFor="user_id" label="Pegawai" required>
          <select className={inputClassName} disabled={state.isSubmitting} id="user_id" onChange={(event) => setState((current) => ({ ...current, user_id: event.target.value }))} value={state.user_id}>
            <option value="">Pilih pegawai</option>
            {employees.map((employee) => (
              <option key={employee.id} value={employee.id}>
                {employee.employee_number} - {employee.name}
              </option>
            ))}
          </select>
        </Field>
        <Field htmlFor="office_location_id" label="Lokasi aktif" required>
          <select className={inputClassName} disabled={state.isSubmitting} id="office_location_id" onChange={(event) => setState((current) => ({ ...current, office_location_id: event.target.value }))} value={state.office_location_id}>
            <option value="">Pilih lokasi</option>
            {locations.map((location) => (
              <option key={location.id} value={location.id}>
                {location.name}
              </option>
            ))}
          </select>
        </Field>
        <Field htmlFor="effective_from" label="Tanggal mulai" required>
          <input className={inputClassName} disabled={state.isSubmitting} id="effective_from" onChange={(event) => setState((current) => ({ ...current, effective_from: event.target.value }))} type="date" value={state.effective_from} />
        </Field>
        <Field htmlFor="effective_to" label="Tanggal akhir">
          <input className={inputClassName} disabled={state.isSubmitting} id="effective_to" onChange={(event) => setState((current) => ({ ...current, effective_to: event.target.value }))} type="date" value={state.effective_to} />
        </Field>
      </div>
      {state.error ? <p className={errorClassName}>{state.error}</p> : null}
      <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
        <button className={outlineButtonClassName} disabled={state.isSubmitting} onClick={() => router.back()} type="button">
          Batal
        </button>
        <button className={primaryButtonClassName} disabled={state.isSubmitting} type="submit">
          {state.isSubmitting ? "Menyimpan..." : "Simpan"}
        </button>
      </div>
    </form>
  );
}

type AssignmentPayload =
  | { ok: true; assignment: LocationAssignment }
  | { ok: false; message: string };

export async function readLocationAssignmentResponse(
  response: Response,
): Promise<AssignmentPayload> {
  try {
    const payload = (await response.json()) as unknown;
    if (isRecord(payload) && payload.status === "ok" && isRecord(payload.data) && typeof payload.data.id === "string") {
      return { ok: true, assignment: payload.data as LocationAssignment };
    }
    if (isRecord(payload) && payload.status === "error" && typeof payload.message === "string") {
      return { ok: false, message: payload.message };
    }
  } catch {
    return { ok: false, message: "Respons tidak valid." };
  }
  return { ok: false, message: "Respons tidak valid." };
}

function Field({
  label,
  htmlFor,
  required = false,
  children,
}: {
  label: string;
  htmlFor: string;
  required?: boolean;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <label className="text-sm font-medium text-slate-800" htmlFor={htmlFor}>
        {label}
        {required ? <span className="text-red-600"> *</span> : null}
      </label>
      {children}
    </div>
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

const inputClassName =
  "h-10 w-full rounded-md border border-slate-300 bg-white px-3 text-sm outline-none transition focus:border-emerald-700 focus:ring-2 focus:ring-emerald-100 disabled:cursor-not-allowed disabled:bg-slate-100";
const outlineButtonClassName =
  "h-10 rounded-md border border-slate-300 px-4 text-sm font-medium text-slate-800 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100 disabled:cursor-not-allowed disabled:text-slate-400";
const primaryButtonClassName =
  "h-10 rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white transition hover:bg-emerald-800 focus:outline-none focus:ring-2 focus:ring-emerald-200 disabled:cursor-not-allowed disabled:bg-slate-400";
const errorClassName =
  "rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700";
