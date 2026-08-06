"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import type { WorkSchedule } from "@/lib/schedule/types";
import { safeScheduleApiMessage } from "@/lib/schedule/utils";

type WorkScheduleFormMode = "create" | "edit";

type FormState = {
  name: string;
  start_time: string;
  end_time: string;
  grace_minutes: string;
  error: string;
  isSubmitting: boolean;
};

export function WorkScheduleForm({
  mode,
  schedule,
}: {
  mode: WorkScheduleFormMode;
  schedule?: WorkSchedule;
}) {
  const router = useRouter();
  const [state, setState] = useState<FormState>({
    name: schedule?.name ?? "",
    start_time: schedule?.start_time.slice(0, 5) ?? "",
    end_time: schedule?.end_time.slice(0, 5) ?? "",
    grace_minutes: String(schedule?.grace_minutes ?? 0),
    error: "",
    isSubmitting: false,
  });

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const name = state.name.trim();
    const graceMinutes = Number(state.grace_minutes);

    if (!name || !state.start_time || !state.end_time) {
      setState((current) => ({ ...current, error: "Nama dan jam kerja wajib diisi." }));
      return;
    }
    if (state.end_time <= state.start_time) {
      setState((current) => ({ ...current, error: "Jam selesai harus setelah jam mulai." }));
      return;
    }
    if (!Number.isInteger(graceMinutes) || graceMinutes < 0) {
      setState((current) => ({ ...current, error: "Toleransi wajib angka bulat dan tidak negatif." }));
      return;
    }

    setState((current) => ({ ...current, error: "", isSubmitting: true }));
    try {
      const response = await fetch(
        mode === "create"
          ? "/api/admin/work-schedules"
          : `/api/admin/work-schedules/${schedule?.id ?? ""}`,
        {
          method: mode === "create" ? "POST" : "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            name,
            start_time: state.start_time,
            end_time: state.end_time,
            grace_minutes: graceMinutes,
          }),
        },
      );
      const payload = await readScheduleFormResponse(response);

      if (!response.ok || !payload.ok) {
        setState((current) => ({
          ...current,
          error: safeScheduleApiMessage(
            response.status,
            payload.ok ? "Jadwal kerja gagal disimpan." : payload.message,
          ),
          isSubmitting: false,
        }));
        return;
      }

      setState((current) => ({ ...current, isSubmitting: false }));
      router.replace(`/work-schedules/${payload.schedule.id}`);
      router.refresh();
    } catch {
      setState((current) => ({
        ...current,
        error: "Layanan jadwal belum tersedia. Coba lagi nanti.",
        isSubmitting: false,
      }));
    }
  }

  return (
    <form className="mt-6 space-y-5 rounded-md border border-slate-200 bg-white p-5" onSubmit={handleSubmit}>
      <div className="grid gap-5 md:grid-cols-2">
        <Field htmlFor="name" label="Nama jadwal" required>
          <input className={inputClassName} disabled={state.isSubmitting} id="name" onChange={(event) => setState((current) => ({ ...current, name: event.target.value }))} value={state.name} />
        </Field>
        <Field htmlFor="grace_minutes" label="Toleransi terlambat" required>
          <input className={inputClassName} disabled={state.isSubmitting} id="grace_minutes" min="0" onChange={(event) => setState((current) => ({ ...current, grace_minutes: event.target.value }))} type="number" value={state.grace_minutes} />
        </Field>
        <Field htmlFor="start_time" label="Jam mulai" required>
          <input className={inputClassName} disabled={state.isSubmitting} id="start_time" onChange={(event) => setState((current) => ({ ...current, start_time: event.target.value }))} type="time" value={state.start_time} />
        </Field>
        <Field htmlFor="end_time" label="Jam selesai" required>
          <input className={inputClassName} disabled={state.isSubmitting} id="end_time" onChange={(event) => setState((current) => ({ ...current, end_time: event.target.value }))} type="time" value={state.end_time} />
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

type FormPayload =
  | { ok: true; schedule: WorkSchedule }
  | { ok: false; message: string };

async function readScheduleFormResponse(response: Response): Promise<FormPayload> {
  try {
    const payload = (await response.json()) as unknown;
    if (isRecord(payload) && payload.status === "ok" && isScheduleLike(payload.data)) {
      return { ok: true, schedule: payload.data };
    }
    if (isRecord(payload) && payload.status === "error" && typeof payload.message === "string") {
      return { ok: false, message: payload.message };
    }
  } catch {
    return { ok: false, message: "Respons tidak valid." };
  }
  return { ok: false, message: "Respons tidak valid." };
}

function isScheduleLike(value: unknown): value is WorkSchedule {
  return isRecord(value) && typeof value.id === "string" && typeof value.name === "string";
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

const inputClassName =
  "h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-emerald-700 focus:ring-2 focus:ring-emerald-100 disabled:cursor-not-allowed disabled:bg-slate-100";
const outlineButtonClassName =
  "h-10 rounded-md border border-slate-300 px-4 text-sm font-medium text-slate-800 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100 disabled:cursor-not-allowed disabled:text-slate-400";
const primaryButtonClassName =
  "h-10 rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white transition hover:bg-emerald-800 focus:outline-none focus:ring-2 focus:ring-emerald-200 disabled:cursor-not-allowed disabled:bg-slate-400";
const errorClassName =
  "rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700";
