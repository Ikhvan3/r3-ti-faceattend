import { requireAdmin } from "@/lib/server/session";

export default async function DashboardPage() {
  const user = await requireAdmin();

  return (
    <section className="mx-auto max-w-5xl">
      <div>
        <h1 className="text-2xl font-semibold">Selamat datang, {user.name}</h1>
        <p className="mt-2 text-sm text-slate-600">
          Modul dashboard statistik akan dibuat pada tahap berikutnya.
        </p>
      </div>

      <div className="mt-6 grid gap-4 md:grid-cols-2">
        <ProfileItem label="Nomor Pegawai" value={user.employee_number} />
        <ProfileItem label="Email" value={user.email} />
        <ProfileItem label="Jabatan" value={user.position ?? "-"} />
        <ProfileItem label="Role" value={user.role} />
        <ProfileItem label="Status Akun" value={user.account_status} />
      </div>
    </section>
  );
}

function ProfileItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border border-slate-200 bg-white p-4">
      <p className="text-xs font-medium uppercase text-slate-500">{label}</p>
      <p className="mt-2 break-words text-sm font-semibold text-slate-950">
        {value}
      </p>
    </div>
  );
}
