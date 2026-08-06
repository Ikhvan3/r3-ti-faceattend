import type { AssignmentStatus, ScheduleStatus } from "@/lib/schedule/types";
import {
  assignmentStatusLabel,
  badgeClass,
  scheduleStatusLabel,
} from "@/lib/schedule/utils";

export function ScheduleBadge({
  status,
}: {
  status: ScheduleStatus | AssignmentStatus;
}) {
  const label =
    status === "ACTIVE" || status === "INACTIVE"
      ? scheduleStatusLabel(status)
      : assignmentStatusLabel(status);

  return (
    <span
      className={`inline-flex h-7 items-center rounded-full border px-2.5 text-xs font-semibold ${badgeClass(status)}`}
    >
      {label}
    </span>
  );
}
