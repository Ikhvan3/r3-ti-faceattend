import { notFound, redirect } from "next/navigation";

import { DetailItem } from "@/app/(dashboard)/employees/_components/detail-item";
import {
  PageHeader,
  SecondaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { ConfirmOfficeLocationStatusDialog } from "@/app/(dashboard)/office-locations/_components/confirm-office-location-status-dialog";
import { LocationBadge } from "@/app/(dashboard)/office-locations/_components/location-badge";
import { SafeApiError } from "@/lib/auth/types";
import {
  formatCoordinate,
  formatRadius,
  officeLocationStatusFromActive,
} from "@/lib/location/utils";
import { getOfficeLocationByID } from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";
import { formatDateTime } from "@/lib/schedule/utils";

export default async function OfficeLocationDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  await requireAdmin();
  const { id } = await params;

  let location;
  try {
    location = await getOfficeLocationByID(id);
  } catch (error) {
    if (
      error instanceof SafeApiError &&
      (error.code === "NOT_FOUND" || error.code === "BAD_REQUEST")
    ) {
      notFound();
    }
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      redirect("/login");
    }
    throw error;
  }

  return (
    <section className="mx-auto max-w-5xl space-y-6">
      <PageHeader
        action={
          <div className="flex flex-wrap gap-2">
            <SecondaryLink href="/office-locations">Kembali</SecondaryLink>
            <SecondaryLink href={`/office-locations/${location.id}/edit`}>
              Edit
            </SecondaryLink>
            <ConfirmOfficeLocationStatusDialog location={location} />
          </div>
        }
        description="Detail lokasi kantor untuk penugasan pegawai. Validasi geofence absensi belum diaktifkan pada tahap ini."
        title={location.name}
      />
      <div className="rounded-md border border-slate-200 bg-white p-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-sm font-medium text-slate-500">Koordinat</p>
            <p className="mt-1 text-lg font-semibold text-slate-950">
              {formatCoordinate(location.latitude)},{" "}
              {formatCoordinate(location.longitude)}
            </p>
          </div>
          <LocationBadge status={officeLocationStatusFromActive(location.is_active)} />
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <DetailItem label="ID" value={location.id} />
        <DetailItem label="Nama" value={location.name} />
        <DetailItem label="Alamat" value={location.address || "-"} />
        <DetailItem label="Latitude" value={formatCoordinate(location.latitude)} />
        <DetailItem label="Longitude" value={formatCoordinate(location.longitude)} />
        <DetailItem label="Radius" value={formatRadius(location.radius_meters)} />
        <DetailItem label="Status" value={location.is_active ? "Aktif" : "Nonaktif"} />
        <DetailItem label="Dibuat" value={formatDateTime(location.created_at)} />
        <DetailItem label="Diperbarui" value={formatDateTime(location.updated_at)} />
      </div>
    </section>
  );
}
