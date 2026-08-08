import Link from "next/link";
import { notFound, redirect } from "next/navigation";

import { AttendanceCorrectionDialog } from "@/app/(dashboard)/attendance/_components/attendance-correction-dialog";
import { DetailItem } from "@/app/(dashboard)/employees/_components/detail-item";
import {
  PageHeader,
  SecondaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import type { AuditLogItem } from "@/lib/audit/types";
import {
  auditSnapshotSummary,
  formatAuditDateTime,
} from "@/lib/audit/utils";
import { SafeApiError } from "@/lib/auth/types";
import type { AdminAttendanceLocationEvidence } from "@/lib/attendance/types";
import {
  attendanceStateLabel,
  formatBusinessDate,
  formatBusinessTime,
  formatBusinessTimeInput,
} from "@/lib/attendance/utils";
import { getAdminAttendanceDetail } from "@/lib/server/admin-attendance-bff";
import { getAdminAuditLogs } from "@/lib/server/admin-audit-bff";
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

  let correctionHistory: AuditLogItem[] = [];
  try {
    const audit = await getAdminAuditLogs({
      action: "ATTENDANCE_CORRECTED",
      entity_type: "ATTENDANCE_RECORD",
      entity_id: attendance.id,
      page: 1,
      page_size: 10,
    });
    correctionHistory = audit.items;
  } catch (error) {
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      redirect("/login");
    }
    throw error;
  }

  const checkInInput = formatBusinessTimeInput(attendance.check_in_at) ?? "";
  const checkOutInput = formatBusinessTimeInput(attendance.check_out_at);

  return (
    <section className="mx-auto max-w-5xl space-y-6">
      <PageHeader
        action={
          <div className="flex flex-wrap gap-2">
            <AttendanceCorrectionDialog
              attendanceID={attendance.id}
              currentCheckIn={checkInInput}
              currentCheckOut={checkOutInput}
              employeeName={attendance.employee.name}
            />
            <SecondaryLink href="/attendance">Kembali</SecondaryLink>
          </div>
        }
        description="Detail presensi dan bukti lokasi. Koreksi administratif selalu membutuhkan alasan dan dicatat pada audit log."
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

      <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm leading-6 text-amber-900">
        Koreksi Admin tidak membuat bukti GPS baru. Jika jam check-out ditambahkan
        secara manual, bukti lokasi check-out tetap kosong.
      </div>

      <CorrectionHistory items={correctionHistory} />
      <EvidenceCard title="Lokasi Check-in" evidence={attendance.check_in_location} />
      <EvidenceCard title="Lokasi Check-out" evidence={attendance.check_out_location} />
    </section>
  );
}

function CorrectionHistory({ items }: { items: AuditLogItem[] }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-slate-950">Riwayat Koreksi</h2>
          <p className="mt-1 text-sm text-slate-500">
            Maksimal 10 koreksi terbaru untuk record presensi ini.
          </p>
        </div>
        <Link
          className="text-sm font-semibold text-emerald-700 hover:text-emerald-800"
          href="/audit-logs?action=ATTENDANCE_CORRECTED&entity_type=ATTENDANCE_RECORD"
        >
          Buka Audit Log
        </Link>
      </div>

      {items.length === 0 ? (
        <p className="mt-4 rounded-lg bg-slate-50 px-4 py-3 text-sm text-slate-500">
          Presensi ini belum pernah dikoreksi Admin.
        </p>
      ) : (
        <div className="mt-4 space-y-3">
          {items.map((item) => {
            const snapshot = auditSnapshotSummary(item);
            return (
              <div className="rounded-lg border border-slate-200 p-4" key={item.id}>
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <span className="text-sm font-semibold text-slate-900">
                    {item.actor_email}
                  </span>
                  <span className="text-xs text-slate-500">
                    {formatAuditDateTime(item.created_at)}
                  </span>
                </div>
                <div className="mt-3 grid gap-2 text-sm text-slate-700 sm:grid-cols-2">
                  <p><span className="font-semibold">Sebelum:</span> {snapshot.before}</p>
                  <p><span className="font-semibold">Sesudah:</span> {snapshot.after}</p>
                </div>
                <p className="mt-3 text-sm leading-6 text-slate-600">
                  <span className="font-semibold text-slate-800">Alasan:</span> {item.reason}
                </p>
              </div>
            );
          })}
        </div>
      )}
    </div>
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
