import { notFound, redirect } from "next/navigation";

import {
  PageHeader,
  SecondaryLink,
} from "@/app/(dashboard)/employees/_components/page-header";
import { OfficeLocationForm } from "@/app/(dashboard)/office-locations/_components/office-location-form";
import { SafeApiError } from "@/lib/auth/types";
import { getOfficeLocationByID } from "@/lib/server/go-api";
import { requireAdmin } from "@/lib/server/session";

export default async function EditOfficeLocationPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  await requireAdmin();
  const { id } = await params;

  let location;
  try {
    location = await getOfficeLocationByID(id);
  } catch (error) {
    if (
      error instanceof SafeApiError &&
      (error.code === "NOT_FOUND" || error.code === "BAD_REQUEST")
    ) {
      notFound();
    }
    if (error instanceof SafeApiError && error.code === "UNAUTHORIZED") {
      redirect("/login");
    }
    throw error;
  }

  return (
    <section className="mx-auto max-w-4xl">
      <PageHeader
        action={<SecondaryLink href={`/office-locations/${location.id}`}>Kembali</SecondaryLink>}
        description="Perbarui nama, alamat, koordinat, dan radius lokasi kantor."
        title="Edit Lokasi Kantor"
      />
      <OfficeLocationForm location={location} mode="edit" />
    </section>
  );
}
