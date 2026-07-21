"use client";

import Link from "next/link";
import { useAuth } from "@/hooks/useAuth";

export default function MainLayout({ children }) {
  const { logout } = useAuth();

  return (
    <div>
      <nav className="flex items-center justify-between border-b bg-white px-6 py-3">
        <div className="flex gap-4 text-sm font-medium">
          <Link href="/chat">Chat</Link>
          <Link href="/dashboard">Dashboard</Link>
        </div>
        <button onClick={logout} className="text-sm text-red-500">
          Çıkış Yap
        </button>
      </nav>
      <main>{children}</main>
    </div>
  );
}