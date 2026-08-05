import type { AccountStatus } from "@/lib/auth/types";
import { statusBadgeClass, statusLabel } from "@/lib/employee/utils";

export function StatusBadge({ status }: { status: AccountStatus }) {
  return (
    <span
      className={`inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-semibold ${statusBadgeClass(status)}`}
    >
      {statusLabel(status)}
    </span>
  );
}
