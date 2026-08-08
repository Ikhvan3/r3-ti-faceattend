import Link from "next/link";
import { redirect } from "next/navigation";

import { PageHeader } from "@/app/(dashboard)/employees/_components/page-header";
import type { AuditAction, AuditEntityType, AuditLogQuery } from "@/lib/audit/types";
import {
  auditActionLabel,
  auditEntityLabel,
  auditSnapshotSummary,
  formatAuditDateTime,
} from "@/lib/audit/utils";
import { SafeApiError } from "@/lib/auth/types";
import { getAdminAuditLogs } from "@/lib/server/admin-audit-bff";
import { requireAdmin } from "@/lib/server/session";

const ACTIONS: AuditAction[] = [
  "ATTENDANCE_CORRECTED",
  "FACE_ENROLLMENT_RESET",
];
const ENTITIES: AuditEntityType[] = ["ATTENDANCE_RECORD", "FACE_PROFILE"];

export default async function AuditLogsPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) {
  await requireAdmin();
  const params = await searchParams;
  const query = parseQuery(params);

  let logs;
  try {
    logs = await getAdminAuditLogs(query);
  } catch (error) {
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      redirect("/login");
    }
    throw error;
  }

  return (
    <section className="space-y-6">
      <PageHeader
        description="Jejak perubahan administratif yang sensitif. Audit log bersifat read-only dan menyimpan actor, target, alasan, serta ringkasan sebelum/sesudah perubahan."
        title="Audit Log"
      />

      {query.entity_id ? (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-900">
          <span>Menampilkan audit untuk satu record yang dipilih dari halaman detail.</span>
          <Link className="font-semibold hover:text-emerald-950" href="/audit-logs">
            Tampilkan semua audit
          </Link>
        </div>
      ) : null}

      <form className="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 md:grid-cols-5" method="GET">
        {query.entity_id ? (
          <input name="entity_id" type="hidden" value={query.entity_id} />
        ) : null}
        <label className="text-sm font-medium text-slate-700">
          Aksi
          <select
            className="mt-1 h-10 w-full rounded-md border border-slate-300 px-3 text-sm"
            defaultValue={query.action ?? ""}
            name="action"
          >
            <option value="">Semua aksi</option>
            {ACTIONS.map((action) => (
              <option key={action} value={action}>
                {auditActionLabel(action)}
              </option>
            ))}
          </select>
        </label>
        <label className="text-sm font-medium text-slate-700">
          Entitas
          <select
            className="mt-1 h-10 w-full rounded-md border border-slate-300 px-3 text-sm"
            defaultValue={query.entity_type ?? ""}
            name="entity_type"
          >
            <option value="">Semua entitas</option>
            {ENTITIES.map((entity) => (
              <option key={entity} value={entity}>
                {auditEntityLabel(entity)}
              </option>
            ))}
          </select>
        </label>
        <label className="text-sm font-medium text-slate-700">
          Dari tanggal
          <input
            className="mt-1 h-10 w-full rounded-md border border-slate-300 px-3 text-sm"
            defaultValue={query.date_from ?? ""}
            name="date_from"
            type="date"
          />
        </label>
        <label className="text-sm font-medium text-slate-700">
          Sampai tanggal
          <input
            className="mt-1 h-10 w-full rounded-md border border-slate-300 px-3 text-sm"
            defaultValue={query.date_to ?? ""}
            name="date_to"
            type="date"
          />
        </label>
        <div className="flex items-end gap-2">
          <button
            className="h-10 rounded-md bg-emerald-700 px-4 text-sm font-semibold text-white hover:bg-emerald-800"
            type="submit"
          >
            Terapkan
          </button>
          <Link
            className="inline-flex h-10 items-center rounded-md border border-slate-300 px-3 text-sm font-medium text-slate-700 hover:bg-slate-50"
            href="/audit-logs"
          >
            Reset
          </Link>
        </div>
      </form>

      <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
        <div className="border-b border-slate-200 px-5 py-4">
          <p className="text-sm text-slate-600">
            {logs.total_items} perubahan tercatat
          </p>
        </div>
        {logs.items.length === 0 ? (
          <p className="p-6 text-sm text-slate-500">Belum ada audit log untuk filter ini.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-slate-200 text-sm">
              <thead className="bg-slate-50 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">
                <tr>
                  <th className="px-4 py-3">Waktu</th>
                  <th className="px-4 py-3">Aksi</th>
                  <th className="px-4 py-3">Admin</th>
                  <th className="px-4 py-3">Target</th>
                  <th className="px-4 py-3">Perubahan</th>
                  <th className="px-4 py-3">Alasan</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {logs.items.map((item) => {
                  const snapshot = auditSnapshotSummary(item);
                  return (
                    <tr className="align-top" key={item.id}>
                      <td className="whitespace-nowrap px-4 py-4 text-slate-600">
                        {formatAuditDateTime(item.created_at)}
                      </td>
                      <td className="px-4 py-4">
                        <div className="font-semibold text-slate-900">
                          {auditActionLabel(item.action)}
                        </div>
                        <div className="mt-1 text-xs text-slate-500">
                          {auditEntityLabel(item.entity_type)}
                        </div>
                      </td>
                      <td className="px-4 py-4 text-slate-700">
                        {item.actor_email}
                      </td>
                      <td className="px-4 py-4 text-slate-700">
                        {item.target_label ?? item.target_user_id ?? "-"}
                      </td>
                      <td className="min-w-56 px-4 py-4 text-xs leading-5 text-slate-600">
                        <div><span className="font-semibold">Sebelum:</span> {snapshot.before}</div>
                        <div className="mt-1"><span className="font-semibold">Sesudah:</span> {snapshot.after}</div>
                      </td>
                      <td className="min-w-64 px-4 py-4 leading-6 text-slate-700">
                        {item.reason}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {logs.total_pages > 1 ? (
        <div className="flex items-center justify-between gap-3 text-sm text-slate-600">
          <span>
            Halaman {logs.page} dari {logs.total_pages}
          </span>
          <div className="flex gap-2">
            {logs.page > 1 ? (
              <Link className="rounded-md border border-slate-300 px-3 py-2 hover:bg-white" href={pageHref(query, logs.page - 1)}>
                Sebelumnya
              </Link>
            ) : null}
            {logs.page < logs.total_pages ? (
              <Link className="rounded-md border border-slate-300 px-3 py-2 hover:bg-white" href={pageHref(query, logs.page + 1)}>
                Berikutnya
              </Link>
            ) : null}
          </div>
        </div>
      ) : null}
    </section>
  );
}

function parseQuery(
  params: Record<string, string | string[] | undefined>,
): AuditLogQuery {
  const action = first(params.action);
  const entity = first(params.entity_type);
  const entityID = first(params.entity_id);
  return {
    action: ACTIONS.includes(action as AuditAction)
      ? (action as AuditAction)
      : undefined,
    entity_type: ENTITIES.includes(entity as AuditEntityType)
      ? (entity as AuditEntityType)
      : undefined,
    entity_id: isUuid(entityID) ? entityID : undefined,
    date_from: first(params.date_from) || undefined,
    date_to: first(params.date_to) || undefined,
    page: positiveInt(first(params.page), 1),
    page_size: 20,
  };
}

function pageHref(query: AuditLogQuery, page: number): string {
  const params = new URLSearchParams();
  if (query.action) params.set("action", query.action);
  if (query.entity_type) params.set("entity_type", query.entity_type);
  if (query.entity_id) params.set("entity_id", query.entity_id);
  if (query.date_from) params.set("date_from", query.date_from);
  if (query.date_to) params.set("date_to", query.date_to);
  params.set("page", String(page));
  return `/audit-logs?${params.toString()}`;
}

function first(value: string | string[] | undefined): string {
  return Array.isArray(value) ? value[0] ?? "" : value ?? "";
}

function positiveInt(value: string, fallback: number): number {
  const parsed = Number.parseInt(value, 10);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}

function isUuid(value: string): boolean {
  return /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value);
}
