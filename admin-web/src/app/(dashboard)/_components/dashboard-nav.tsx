"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const NAV_ITEMS = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/employees", label: "Pegawai TI" },
  { href: "/attendance", label: "Presensi" },
  { href: "/work-schedules", label: "Jadwal Kerja" },
  { href: "/schedule-assignments", label: "Penugasan Jadwal" },
  { href: "/office-locations", label: "Lokasi Kantor" },
  { href: "/location-assignments", label: "Penugasan Lokasi" },
];

export function DashboardNav() {
  const pathname = usePathname();

  return (
    <nav className="mt-8 space-y-2">
      {NAV_ITEMS.map((item) => {
        const isActive =
          pathname === item.href || pathname.startsWith(`${item.href}/`);
        return (
          <Link
            className={
              isActive
                ? "block rounded-md bg-emerald-50 px-3 py-2 text-sm font-medium text-emerald-800"
                : "block rounded-md px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
            }
            href={item.href}
            key={item.href}
          >
            {item.label}
          </Link>
        );
      })}
    </nav>
  );
}
