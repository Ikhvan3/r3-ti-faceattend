import Link from "next/link";

import { ConfirmStatusDialog } from "@/app/(dashboard)/employees/_components/confirm-status-dialog";
import { StatusBadge } from "@/app/(dashboard)/employees/_components/status-badge";
import type { EmployeeListItem } from "@/lib/employee/types";

export function EmployeeTable({ employees }: { employees: EmployeeListItem[] }) {
  return (
    <div className="overflow-hidden rounded-md border border-slate-200 bg-white">
      <div className="overflow-x-auto">
        <table className="min-w-[900px] w-full border-collapse text-left text-sm">
          <thead className="bg-slate-50 text-xs font-semibold uppercase text-slate-500">
            <tr>
              <th className="px-4 py-3">Nomor Pegawai</th>
              <th className="px-4 py-3">Nama</th>
              <th className="px-4 py-3">Email</th>
              <th className="px-4 py-3">Jabatan</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Wajah</th>
              <th className="px-4 py-3">Aksi</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-200">
            {employees.map((employee) => (
              <tr className="align-top" key={employee.id}>
                <td className="px-4 py-3 font-semibold text-slate-950">
                  {employee.employee_number}
                </td>
                <td className="px-4 py-3 text-slate-900">{employee.name}</td>
                <td className="px-4 py-3 text-slate-700">{employee.email}</td>
                <td className="px-4 py-3 text-slate-700">
                  {employee.position ?? "-"}
                </td>
                <td className="px-4 py-3">
                  <StatusBadge status={employee.account_status} />
                </td>
                <td className="px-4 py-3">
                  <FaceEnrollmentBadge enrolled={employee.face_enrollment?.enrolled === true} />
                </td>
                <td className="px-4 py-3">
                  <div className="flex flex-wrap gap-2">
                    <ActionLink href={`/employees/${employee.id}`}>
                      Detail
                    </ActionLink>
                    <ActionLink href={`/employees/${employee.id}/edit`}>
                      Edit
                    </ActionLink>
                    <ConfirmStatusDialog employee={employee} />
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function FaceEnrollmentBadge({ enrolled }: { enrolled: boolean }) {
  return (
    <span
      className={
        enrolled
          ? "inline-flex rounded-full border border-emerald-200 bg-emerald-50 px-2.5 py-1 text-xs font-semibold text-emerald-700"
          : "inline-flex rounded-full border border-slate-200 bg-slate-100 px-2.5 py-1 text-xs font-semibold text-slate-600"
      }
    >
      {enrolled ? "Terdaftar" : "Belum terdaftar"}
    </span>
  );
}

function ActionLink({
  href,
  children,
}: {
  href: string;
  children: string;
}) {
  return (
    <Link
      className="inline-flex h-9 items-center rounded-md border border-slate-300 bg-white px-3 text-sm font-medium text-slate-800 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-emerald-100"
      href={href}
    >
      {children}
    </Link>
  );
}
