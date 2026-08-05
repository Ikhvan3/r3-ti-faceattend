"use client";

export default function EmployeesError({
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <section className="mx-auto max-w-3xl rounded-md border border-red-200 bg-red-50 p-5">
      <h1 className="text-lg font-semibold text-red-800">
        Data pegawai belum dapat ditampilkan
      </h1>
      <p className="mt-2 text-sm leading-6 text-red-700">
        Layanan pegawai mungkin belum aktif atau session tidak valid. Coba lagi
        beberapa saat.
      </p>
      <button
        className="mt-4 h-10 rounded-md bg-red-700 px-4 text-sm font-semibold text-white hover:bg-red-800"
        onClick={reset}
        type="button"
      >
        Coba Lagi
      </button>
    </section>
  );
}
