import Link from "next/link";
import type { ReactNode } from "react";

import { logoutAction } from "@/lib/server/logout-action";
import { requireAdmin } from "@/lib/server/session";

export default async function DashboardLayout({
  children,
}: {
  children: ReactNode;
}) {
  const user = await requireAdmin();

  return (
    <div className="min-h-screen bg-[#f6f7f9] text-slate-950">
      <div className="flex min-h-screen">
        <aside className="hidden w-72 border-r border-slate-200 bg-white px-5 py-6 md:block">
          <div>
            <p className="text-xs font-semibold uppercase text-emerald-700">
              R3 TI FaceAttend
            </p>
            <h2 className="mt-2 text-lg font-semibold">Admin TI</h2>
          </div>
          <nav className="mt-8">
            <Link
              className="block rounded-md bg-emerald-50 px-3 py-2 text-sm font-medium text-emerald-800"
              href="/dashboard"
            >
              Dashboard
            </Link>
            <Link
              className="mt-2 block rounded-md px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
              href="/employees"
            >
              Pegawai TI
            </Link>
          </nav>
        </aside>

        <div className="flex min-w-0 flex-1 flex-col">
          <header className="flex items-center justify-between border-b border-slate-200 bg-white px-4 py-4 md:px-6">
            <div>
              <p className="text-xs font-medium text-slate-500">
                Admin Divisi Teknologi Informasi
              </p>
              <p className="mt-1 text-sm font-semibold text-slate-900">
                {user.name}
              </p>
            </div>
            <form action={logoutAction}>
              <button
                className="h-10 rounded-md border border-slate-300 px-4 text-sm font-medium text-slate-800 hover:bg-slate-50"
                type="submit"
              >
                Logout
              </button>
            </form>
          </header>
          <main className="flex-1 px-4 py-6 md:px-6">{children}</main>
        </div>
      </div>
    </div>
  );
}
