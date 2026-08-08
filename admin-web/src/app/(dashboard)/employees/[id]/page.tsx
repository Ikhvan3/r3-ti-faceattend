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

  const face = employee.face_enrollment;

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
        description="Profil aman pegawai USER. Password hash, session, embedding, dan foto wajah tidak ditampilkan."
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

      <section className="rounded-md border border-slate-200 bg-white p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold text-slate-950">
              Enrollment Wajah
            </h2>
            <p className="mt-1 text-sm text-slate-600">
              Admin hanya melihat metadata enrollment. Template biometrik, embedding,
              similarity, threshold, dan gambar wajah tidak pernah ditampilkan.
            </p>
          </div>
          <span
            className={
              face?.enrolled
                ? "inline-flex rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-sm font-semibold text-emerald-700"
                : "inline-flex rounded-full border border-slate-200 bg-slate-100 px-3 py-1 text-sm font-semibold text-slate-600"
            }
          >
            {face?.enrolled ? "Terdaftar" : "Belum terdaftar"}
          </span>
        </div>

        <div className="mt-5 grid gap-4 md:grid-cols-2">
          <DetailItem
            label="Status Wajah"
            value={face?.enrolled ? "ENROLLED" : "NOT_ENROLLED"}
          />
          <DetailItem
            label="Tanggal Enrollment"
            value={face?.enrolled_at ? formatDateTime(face.enrolled_at) : "-"}
          />
          <DetailItem label="Model" value={face?.embedding_model ?? "-"} />
          <DetailItem label="Versi Model" value={face?.embedding_version ?? "-"} />
        </div>
      </section>
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
