import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "R3 TI FaceAttend Admin",
  description: "Website admin R3 TI FaceAttend",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="id">
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
