import { notFound, redirect } from "next/navigation";

import { DetailItem } from "@/app/(dashboard)/employees/_components/detail-item";
import {
  PageHeader,
  SecondaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { SafeApiError } from "@/lib/auth/types";
import type { AdminAttendanceLocationEvidence } from "@/lib/attendance/types";
import {
  attendanceStateLabel,
  formatBusinessDate,
  formatBusinessTime,
} from "@/lib/attendance/utils";
import { getAdminAttendanceDetail } from "@/lib/server/admin-attendance-bff";
import { requireAdmin } from "@/lib/server/session";

export default async function AttendanceDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  await requireAdmin();
  const { id } = await params;

  let attendance;
  try {
    attendance = await getAdminAttendanceDetail(id);
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
        action={<SecondaryLink href="/attendance">Kembali</SecondaryLink>}
        description="Detail presensi read-only. Data biometrik, token, dan verification grant tidak ditampilkan."
        title={`Presensi ${attendance.employee.name}`}
      />

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <DetailItem label="Nomor Pegawai" value={attendance.employee.employee_number} />
        <DetailItem label="Jabatan" value={attendance.employee.position ?? "-"} />
        <DetailItem label="Tanggal" value={formatBusinessDate(attendance.attendance_date)} />
        <DetailItem label="Jadwal" value={`${attendance.schedule.name} (${attendance.schedule.start_time}–${attendance.schedule.end_time})`} />
        <DetailItem label="Jam Masuk" value={formatBusinessTime(attendance.check_in_at)} />
        <DetailItem label="Jam Pulang" value={formatBusinessTime(attendance.check_out_at)} />
        <DetailItem label="Status" value={attendanceStateLabel(attendance.attendance_state)} />
        <DetailItem label="Keterlambatan" value={attendance.is_late ? "Terlambat" : "Tidak terlambat"} />
        <DetailItem label="Toleransi Jadwal" value={`${attendance.schedule.grace_minutes} menit`} />
      </div>

      <EvidenceCard title="Lokasi Check-in" evidence={attendance.check_in_location} />
      <EvidenceCard title="Lokasi Check-out" evidence={attendance.check_out_location} />
    </section>
  );
}

function EvidenceCard({
  title,
  evidence,
}: {
  title: string;
  evidence: AdminAttendanceLocationEvidence | null;
}) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5">
      <h2 className="text-base font-semibold text-slate-950">{title}</h2>
      {evidence ? (
        <dl className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <EvidenceItem label="Lokasi kantor" value={evidence.office_location_name} />
          <EvidenceItem label="Radius" value={`${evidence.radius_meters} m`} />
          <EvidenceItem label="Akurasi GPS" value={`${evidence.accuracy_meters.toFixed(1)} m`} />
          <EvidenceItem label="Jarak dari kantor" value={`${evidence.distance_meters.toFixed(1)} m`} />
          <EvidenceItem label="Latitude" value={evidence.latitude.toFixed(6)} />
          <EvidenceItem label="Longitude" value={evidence.longitude.toFixed(6)} />
        </dl>
      ) : (
        <p className="mt-3 text-sm text-slate-500">Bukti lokasi belum tersedia.</p>
      )}
    </div>
  );
}

function EvidenceItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">
        {label}
      </dt>
      <dd className="mt-1 break-words text-sm font-medium text-slate-900">
        {value}
      </dd>
    </div>
  );
}
