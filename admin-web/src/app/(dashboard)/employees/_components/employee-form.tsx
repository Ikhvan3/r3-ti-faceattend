"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import type { Employee } from "@/lib/employee/types";
import { safeEmployeeApiMessage } from "@/lib/employee/utils";

type EmployeeFormMode = "create" | "edit";

type FormState = {
  employee_number: string;
  name: string;
  email: string;
  initial_password: string;
  confirm_password: string;
  phone: string;
  position: string;
  error: string;
  isSubmitting: boolean;
};

export function EmployeeForm({
  mode,
  employee,
}: {
  mode: EmployeeFormMode;
  employee?: Employee;
}) {
  const router = useRouter();
  const [state, setState] = useState<FormState>({
    employee_number: employee?.employee_number ?? "",
    name: employee?.name ?? "",
    email: employee?.email ?? "",
    initial_password: "",
    confirm_password: "",
    phone: employee?.phone ?? "",
    position: employee?.position ?? "",
    error: "",
    isSubmitting: false,
  });

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const employeeNumber = state.employee_number.trim();
    const name = state.name.trim();
    const email = state.email.trim();
    const phone = nullableInput(state.phone);
    const position = nullableInput(state.position);

    if (!employeeNumber || !name || !email) {
      setState((current) => ({
        ...current,
        error: "Nomor pegawai, nama, dan email wajib diisi.",
      }));
      return;
    }

    if (mode === "create") {
      if (!state.initial_password || state.initial_password.length < 8) {
        setState((current) => ({
          ...current,
          error: "Password awal minimal 8 karakter.",
        }));
        return;
      }
      if (state.initial_password !== state.confirm_password) {
        setState((current) => ({
          ...current,
          error: "Konfirmasi password tidak sama.",
        }));
        return;
      }
    }

    setState((current) => ({ ...current, error: "", isSubmitting: true }));

    try {
      const response = await fetch(
        mode === "create"
          ? "/api/admin/employees"
          : `/api/admin/employees/${employee?.id ?? ""}`,
        {
          method: mode === "create" ? "POST" : "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(
            mode === "create"
              ? {
                  employee_number: employeeNumber,
                  name,
                  email,
                  initial_password: state.initial_password,
                  phone,
                  position,
                }
              : {
                  employee_number: employeeNumber,
                  name,
                  email,
                  phone,
                  position,
                },
          ),
        },
      );

      const payload = await readEmployeeFormResponse(response);

      if (!response.ok || !payload.ok) {
        setState((current) => ({
          ...current,
          initial_password: "",
          confirm_password: "",
          error: safeEmployeeApiMessage(
            response.status,
            payload.message ?? "Data pegawai sudah digunakan.",
          ),
          isSubmitting: false,
        }));
        return;
      }

      setState((current) => ({
        ...current,
        initial_password: "",
        confirm_password: "",
        isSubmitting: false,
      }));
      router.replace(`/employees/${payload.employee.id}`);
      router.refresh();
    } catch {
      setState((current) => ({
        ...current,
        initial_password: "",
        confirm_password: "",
        error: "Layanan pegawai belum tersedia. Coba lagi nanti.",
        isSubmitting: false,
      }));
    }
  }

  return (
    <form
      className="mt-6 space-y-5 rounded-md border border-slate-200 bg-white p-5"
      onSubmit={handleSubmit}
    >
      <div className="grid gap-5 md:grid-cols-2">
        <Field label="Nomor pegawai" htmlFor="employee_number" required>
          <input
            autoComplete="off"
            className={inputClassName}
            disabled={state.isSubmitting}
            id="employee_number"
            name="employee_number"
            onChange={(event) =>
              setState((current) => ({
                ...current,
                employee_number: event.target.value,
              }))
            }
            value={state.employee_number}
          />
        </Field>
        <Field label="Nama" htmlFor="name" required>
          <input
            autoComplete="name"
            className={inputClassName}
            disabled={state.isSubmitting}
            id="name"
            name="name"
            onChange={(event) =>
              setState((current) => ({ ...current, name: event.target.value }))
            }
            value={state.name}
          />
        </Field>
        <Field label="Email" htmlFor="email" required>
          <input
            autoComplete="email"
            className={inputClassName}
            disabled={state.isSubmitting}
            id="email"
            name="email"
            onChange={(event) =>
              setState((current) => ({ ...current, email: event.target.value }))
            }
            type="email"
            value={state.email}
          />
        </Field>
        <Field label="Nomor telepon" htmlFor="phone">
          <input
            autoComplete="tel"
            className={inputClassName}
            disabled={state.isSubmitting}
            id="phone"
            name="phone"
            onChange={(event) =>
              setState((current) => ({ ...current, phone: event.target.value }))
            }
            value={state.phone}
          />
        </Field>
        <Field label="Jabatan" htmlFor="position">
          <input
            autoComplete="organization-title"
            className={inputClassName}
            disabled={state.isSubmitting}
            id="position"
            name="position"
            onChange={(event) =>
              setState((current) => ({
                ...current,
                position: event.target.value,
              }))
            }
            value={state.position}
          />
        </Field>
      </div>

      {mode === "create" ? (
        <div className="grid gap-5 md:grid-cols-2">
          <Field label="Password awal" htmlFor="initial_password" required>
            <input
              autoComplete="new-password"
              className={inputClassName}
              disabled={state.isSubmitting}
              id="initial_password"
              name="initial_password"
              onChange={(event) =>
                setState((current) => ({
                  ...current,
                  initial_password: event.target.value,
                }))
              }
              type="password"
              value={state.initial_password}
            />
          </Field>
          <Field label="Konfirmasi password" htmlFor="confirm_password" required>
            <input
              autoComplete="new-password"
              className={inputClassName}
              disabled={state.isSubmitting}
              id="confirm_password"
              name="confirm_password"
              onChange={(event) =>
                setState((current) => ({
                  ...current,
                  confirm_password: event.target.value,
                }))
              }
              type="password"
              value={state.confirm_password}
            />
          </Field>
        </div>
      ) : null}

      {state.error ? (
        <p className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
          {state.error}
        </p>
      ) : null}

      <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
        <button
          className="h-10 rounded-md border border-slate-300 px-4 text-sm font-medium text-slate-800 transition hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100 disabled:cursor-not-allowed disabled:text-slate-400"
          disabled={state.isSubmitting}
          onClick={() => router.back()}
          type="button"
        >
          Batal
        </button>
        <button
          className="h-10 rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white transition hover:bg-emerald-800 focus:outline-none focus:ring-2 focus:ring-emerald-200 disabled:cursor-not-allowed disabled:bg-slate-400"
          disabled={state.isSubmitting}
          type="submit"
        >
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

type EmployeeFormPayload =
  | {
      ok: true;
      employee: Employee;
      message: string;
    }
  | {
      ok: false;
      message: string;
    };

async function readEmployeeFormResponse(
  response: Response,
): Promise<EmployeeFormPayload> {
  try {
    const payload = (await response.json()) as unknown;
    if (isRecord(payload) && payload.status === "ok" && isEmployee(payload.data)) {
      return {
        ok: true,
        employee: payload.data,
        message:
          typeof payload.message === "string"
            ? payload.message
            : "Pegawai berhasil disimpan.",
      };
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

function nullableInput(value: string): string | null {
  const normalized = value.trim();
  return normalized === "" ? null : normalized;
}

function isEmployee(value: unknown): value is Employee {
  return (
    isRecord(value) &&
    typeof value.id === "string" &&
    typeof value.employee_number === "string" &&
    typeof value.name === "string" &&
    typeof value.email === "string"
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

const inputClassName =
  "h-10 w-full rounded-md border border-slate-300 px-3 text-sm outline-none transition focus:border-emerald-700 focus:ring-2 focus:ring-emerald-100 disabled:cursor-not-allowed disabled:bg-slate-100";
