import type { Metadata } from "next";
import { Inter, Space_Grotesk } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/components/auth/AuthProvider";
import { Navbar } from "@/components/layout/Navbar";
import { ToastProvider } from "@/components/ui/Toast";

const inter = Inter({ subsets: ["latin"], variable: "--font-body" });
const spaceGrotesk = Space_Grotesk({ subsets: ["latin"], variable: "--font-display" });

export const metadata: Metadata = {
  title: "GoTube Lite",
  description: "A mini YouTube-like video upload and streaming platform",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${inter.variable} ${spaceGrotesk.variable} font-[family:var(--font-body)]`}>
        <ToastProvider>
          <AuthProvider>
            <Navbar />
            <main className="mx-auto max-w-7xl px-4 pb-10 pt-6">{children}</main>
          </AuthProvider>
        </ToastProvider>
      </body>
    </html>
  );
}
