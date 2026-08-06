"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import type { OfficeLocation } from "@/lib/location/types";
import {
  parseLatitude,
  parseLongitude,
  parseRadiusMeters,
  safeLocationApiMessage,
} from "@/lib/location/utils";

type OfficeLocationFormMode = "create" | "edit";

type FormState = {
  name: string;
  address: string;
  latitude: string;
  longitude: string;
  radius_meters: string;
  error: string;
  isSubmitting: boolean;
};

export function OfficeLocationForm({
  mode,
  location,
}: {
  mode: OfficeLocationFormMode;
  location?: OfficeLocation;
}) {
  const router = useRouter();
  const [state, setState] = useState<FormState>({
    name: location?.name ?? "",
    address: location?.address ?? "",
    latitude: location ? String(location.latitude) : "",
    longitude: location ? String(location.longitude) : "",
    radius_meters: location ? String(location.radius_meters) : "",
    error: "",
    isSubmitting: false,
  });

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const name = state.name.trim();
    const address = state.address.trim();
    const latitude = parseLatitude(state.latitude);
    const longitude = parseLongitude(state.longitude);
    const radiusMeters = parseRadiusMeters(state.radius_meters);

    if (!name) {
      setState((current) => ({ ...current, error: "Nama lokasi wajib diisi." }));
      return;
    }
    if (latitude === null) {
      setState((current) => ({ ...current, error: "Latitude wajib angka antara -90 dan 90." }));
      return;
    }
    if (longitude === null) {
      setState((current) => ({ ...current, error: "Longitude wajib angka antara -180 dan 180." }));
      return;
    }
    if (radiusMeters === null) {
      setState((current) => ({ ...current, error: "Radius wajib angka bulat antara 10 dan 2000 meter." }));
      return;
    }

    setState((current) => ({ ...current, error: "", isSubmitting: true }));
    try {
      const response = await fetch(
        mode === "create"
          ? "/api/admin/office-locations"
          : `/api/admin/office-locations/${location?.id ?? ""}`,
        {
          method: mode === "create" ? "POST" : "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            name,
            address: address || null,
            latitude,
            longitude,
            radius_meters: radiusMeters,
          }),
        },
      );
      const payload = await readOfficeLocationResponse(response);

      if (!response.ok || !payload.ok) {
        setState((current) => ({
          ...current,
          error: safeLocationApiMessage(
            response.status,
            payload.ok ? "Lokasi kantor gagal disimpan." : payload.message,
          ),
          isSubmitting: false,
        }));
        return;
      }

      setState((current) => ({ ...current, isSubmitting: false }));
      router.replace(`/office-locations/${payload.location.id}`);
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
        <Field htmlFor="name" label="Nama lokasi" required>
          <input className={inputClassName} disabled={state.isSubmitting} id="name" onChange={(event) => setState((current) => ({ ...current, name: event.target.value }))} value={state.name} />
        </Field>
        <Field htmlFor="radius_meters" label="Radius geofence" required>
          <input className={inputClassName} disabled={state.isSubmitting} id="radius_meters" min="10" max="2000" onChange={(event) => setState((current) => ({ ...current, radius_meters: event.target.value }))} type="number" value={state.radius_meters} />
        </Field>
        <Field htmlFor="latitude" label="Latitude" required>
          <input className={inputClassName} disabled={state.isSubmitting} id="latitude" inputMode="decimal" onChange={(event) => setState((current) => ({ ...current, latitude: event.target.value }))} placeholder="-6.984..." value={state.latitude} />
        </Field>
        <Field htmlFor="longitude" label="Longitude" required>
          <input className={inputClassName} disabled={state.isSubmitting} id="longitude" inputMode="decimal" onChange={(event) => setState((current) => ({ ...current, longitude: event.target.value }))} placeholder="110.409..." value={state.longitude} />
        </Field>
        <div className="md:col-span-2">
          <Field htmlFor="address" label="Alamat">
            <textarea className={`${inputClassName} min-h-24 py-2`} disabled={state.isSubmitting} id="address" onChange={(event) => setState((current) => ({ ...current, address: event.target.value }))} value={state.address} />
          </Field>
        </div>
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

type LocationPayload =
  | { ok: true; location: OfficeLocation }
  | { ok: false; message: string };

export async function readOfficeLocationResponse(
  response: Response,
): Promise<LocationPayload> {
  try {
    const payload = (await response.json()) as unknown;
    if (isRecord(payload) && payload.status === "ok" && isOfficeLocationLike(payload.data)) {
      return { ok: true, location: payload.data };
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

function isOfficeLocationLike(value: unknown): value is OfficeLocation {
  return isRecord(value) && typeof value.id === "string" && typeof value.name === "string";
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
