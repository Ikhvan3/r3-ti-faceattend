import { redirect } from "next/navigation";

import { AssignmentForm } from "@/app/(dashboard)/schedule-assignments/_components/assignment-form";
import {
  PageHeader,
  SecondaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { SafeApiError } from "@/lib/auth/types";
import {
  getActiveWorkSchedules,
  getEmployeeOptions,
} from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";

export default async function NewScheduleAssignmentPage() {
  await requireAdmin();

  let employees;
  let schedules;
  try {
    [employees, schedules] = await Promise.all([
      getEmployeeOptions(),
      getActiveWorkSchedules(),
    ]);
  } catch (error) {
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      redirect("/login");
    }
    throw error;
  }

  return (
    <section className="mx-auto max-w-4xl">
      <PageHeader
        action={
          <SecondaryLink href="/schedule-assignments">Kembali</SecondaryLink>
        }
        description="Gunakan ID pegawai dan jadwal dari pilihan aman. Jadwal nonaktif tidak ditampilkan."
        title="Tambah Penugasan Jadwal"
      />
      <AssignmentForm employees={employees} schedules={schedules} />
    </section>
  );
}
