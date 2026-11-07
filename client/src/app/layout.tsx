import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "LEFTN’T — Find your way",
  description: "A little maze. A few wrong turns. Find your way to the goal.",
};

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
