"use client";

import { EmptyState } from "@/app/(dashboard)/employees/_components/empty-state";

export default function ScheduleAssignmentsError() {
  return (
    <section className="mx-auto max-w-4xl">
      <EmptyState
        description="Layanan penugasan jadwal belum tersedia atau response backend tidak sesuai."
        title="Penugasan jadwal belum dapat dibaca"
      />
    </section>
  );
}
