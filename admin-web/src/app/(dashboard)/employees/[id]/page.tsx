import { notFound, redirect } from "next/navigation";

import { ConfirmStatusDialog } from "@/app/(dashboard)/employees/_components/confirm-status-dialog";
import { DetailItem } from "@/app/(dashboard)/employees/_components/detail-item";
import { PageHeader, SecondaryLink } from "@/app/(dashboard)/employees/_components/page-header";
import { StatusBadge } from "@/app/(dashboard)/employees/_components/status-badge";
import { SafeApiError } from "@/lib/auth/types";
import { getEmployeeByID } from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";

export default async function EmployeeDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  await requireAdmin();
  const { id } = await params;

  let employee;
  try {
    employee = await getEmployeeByID(id);
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
            <SecondaryLink href="/employees">Kembali</SecondaryLink>
            <SecondaryLink href={`/employees/${employee.id}/edit`}>
              Edit
            </SecondaryLink>
            <ConfirmStatusDialog employee={employee} />
          </div>
        }
        description="Profil aman pegawai USER. Password hash dan data session tidak ditampilkan."
        title={employee.name}
      />

      <div className="rounded-md border border-slate-200 bg-white p-4">
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-sm font-medium text-slate-500">
              {employee.employee_number}
            </p>
            <p className="mt-1 text-lg font-semibold text-slate-950">
              {employee.email}
            </p>
          </div>
          <StatusBadge status={employee.account_status} />
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <DetailItem label="Nomor Pegawai" value={employee.employee_number} />
        <DetailItem label="Nama" value={employee.name} />
        <DetailItem label="Email" value={employee.email} />
        <DetailItem label="Telepon" value={employee.phone ?? "-"} />
        <DetailItem label="Jabatan" value={employee.position ?? "-"} />
        <DetailItem label="Role" value={employee.role} />
        <DetailItem label="Status Akun" value={employee.account_status} />
        <DetailItem label="Dibuat" value={formatDateTime(employee.created_at)} />
        <DetailItem
          label="Diperbarui"
          value={formatDateTime(employee.updated_at)}
        />
      </div>
    </section>
  );
}

function formatDateTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }

  return new Intl.DateTimeFormat("id-ID", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}
