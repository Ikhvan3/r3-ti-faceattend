"use client";

import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";

export default function WorkSchedulesError() {
  return (
    <section className="mx-auto max-w-4xl">
      <EmptyState
        description="Layanan jadwal belum tersedia atau response backend tidak sesuai."
        title="Jadwal kerja belum dapat dibaca"
      />
    </section>
  );
}
