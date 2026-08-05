import { redirect } from "next/navigation";

import { getCurrentAdmin } from "@/lib/server/session";
import { LoginForm } from "./login-form";

export default async function LoginPage() {
  const user = await getCurrentAdmin();

  if (user) {
    redirect("/dashboard");
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-[#f6f7f9] px-4 py-10 text-slate-950">
      <section className="w-full max-w-md rounded-lg border border-slate-200 bg-white p-6 shadow-sm">
        <div>
          <p className="text-sm font-medium text-emerald-700">
            Admin Divisi Teknologi Informasi
          </p>
          <h1 className="mt-2 text-2xl font-semibold">R3 TI FaceAttend</h1>
          <p className="mt-2 text-sm text-slate-600">
            Masuk menggunakan akun admin lokal yang sudah dibuat melalui
            backend.
          </p>
        </div>
        <LoginForm />
      </section>
    </main>
  );
}
