"use client";

import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";

export default function LocationAssignmentsError() {
  return (
    <section className="mx-auto max-w-4xl">
      <EmptyState
        description="Layanan lokasi belum tersedia atau respons backend tidak sesuai."
        title="Penugasan lokasi belum dapat dibaca"
      />
    </section>
  );
}
