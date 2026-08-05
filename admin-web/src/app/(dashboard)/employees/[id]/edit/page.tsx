import { notFound, redirect } from "next/navigation";

import { EmployeeForm } from "@/app/(dashboard)/employees/_components/employee-form";
import { PageHeader, SecondaryLink } from "@/app/(dashboard)/employees/_components/page-header";
import { SafeApiError } from "@/lib/auth/types";
import { getEmployeeByID } from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";

export default async function EditEmployeePage({
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
    <section className="mx-auto max-w-4xl">
      <PageHeader
        action={<SecondaryLink href={`/employees/${employee.id}`}>Kembali</SecondaryLink>}
        description="Ubah data profil pegawai. Password, role, status akun, dan timestamp tidak dapat diubah dari form ini."
        title="Edit Pegawai"
      />
      <EmployeeForm employee={employee} mode="edit" />
    </section>
  );
}
