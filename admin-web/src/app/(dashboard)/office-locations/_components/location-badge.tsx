import type { AssignmentStatus } from "@/lib/schedule/types";
import type { OfficeLocationStatus } from "@/lib/location/types";
import {
  locationBadgeClass,
  locationStatusLabel,
  officeLocationStatusLabel,
} from "@/lib/location/utils";

export function LocationBadge({
  status,
}: {
  status: OfficeLocationStatus | AssignmentStatus;
}) {
  const label =
    status === "ACTIVE" || status === "INACTIVE"
      ? officeLocationStatusLabel(status)
      : locationStatusLabel(status);

  return (
    <span
      className={`inline-flex h-7 items-center rounded-full border px-2.5 text-xs font-semibold ${locationBadgeClass(status)}`}
    >
      {label}
    </span>
  );
}
