import Link from "next/link";

import { ConfirmOfficeLocationStatusDialog } from "@/app/(dashboard)/office-locations/_components/confirm-office-location-status-dialog";
import { LocationBadge } from "@/app/(dashboard)/office-locations/_components/location-badge";
import type { OfficeLocation } from "@/lib/location/types";
import {
  formatCoordinate,
  formatRadius,
  officeLocationStatusFromActive,
} from "@/lib/location/utils";

export function OfficeLocationTable({
  locations,
}: {
  locations: OfficeLocation[];
}) {
  return (
    <div className="overflow-hidden rounded-md border border-slate-200 bg-white">
      <div className="overflow-x-auto">
        <table className="min-w-[920px] w-full border-collapse text-left text-sm">
          <thead className="bg-slate-50 text-xs font-semibold uppercase text-slate-500">
            <tr>
              <th className="px-4 py-3">Nama</th>
              <th className="px-4 py-3">Alamat</th>
              <th className="px-4 py-3">Koordinat</th>
              <th className="px-4 py-3">Radius</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Aksi</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200">
            {locations.map((location) => (
              <tr className="align-top" key={location.id}>
                <td className="px-4 py-3 font-semibold text-slate-950">
                  {location.name}
                </td>
                <td className="max-w-xs px-4 py-3 text-slate-700">
                  {location.address || "-"}
                </td>
                <td className="px-4 py-3 text-slate-700">
                  {formatCoordinate(location.latitude)},{" "}
                  {formatCoordinate(location.longitude)}
                </td>
                <td className="px-4 py-3 text-slate-700">
                  {formatRadius(location.radius_meters)}
                </td>
                <td className="px-4 py-3">
                  <LocationBadge
                    status={officeLocationStatusFromActive(location.is_active)}
                  />
                </td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap gap-2">
                    <ActionLink href={`/office-locations/${location.id}`}>
                      Detail
                    </ActionLink>
                    <ActionLink href={`/office-locations/${location.id}/edit`}>
                      Edit
                    </ActionLink>
                    <ConfirmOfficeLocationStatusDialog location={location} />
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function ActionLink({
  href,
  children,
}: {
  href: string;
  children: string;
}) {
  return (
    <Link
      className="inline-flex h-9 items-center rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-slate-800 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100"
      href={href}
    >
      {children}
    </Link>
  );
}
