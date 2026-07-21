"use client";

import { useEffect, useState } from "react";
// import { getDashboardStats } from "@/services/stats.service"; // ileride aç

const MOCK_STATS = {
  totalSessions: 12,
  averageScore: 78,
  scores: [
    { session: "Session #1", score: 82 },
    { session: "Session #2", score: 65 },
    { session: "Session #3", score: 91 },
  ],
};

export default function DashboardPage() {
  const [stats, setStats] = useState(null);

  useEffect(() => {
    // TODO: gerçek backend hazır olunca değiştir:
    // getDashboardStats().then(setStats);
    setStats(MOCK_STATS);
  }, []);

  if (!stats) return <p className="p-8">Yükleniyor...</p>;

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="mb-6 text-xl font-semibold">Deci.Scoring Dashboard</h1>

      <div className="mb-6 grid grid-cols-2 gap-4">
        <div className="rounded-xl border bg-white p-4">
          <p className="text-sm text-gray-500">Toplam Oturum</p>
          <p className="text-2xl font-bold">{stats.totalSessions}</p>
        </div>
        <div className="rounded-xl border bg-white p-4">
          <p className="text-sm text-gray-500">Ortalama Skor</p>
          <p className="text-2xl font-bold">{stats.averageScore}/100</p>
        </div>
      </div>

      <div className="space-y-3">
        {stats.scores.map((s) => (
          <div
            key={s.session}
            className="flex items-center justify-between rounded-lg border bg-white p-3"
          >
            <span className="text-sm">{s.session}</span>
            <span className="font-semibold text-indigo-600">{s.score}/100</span>
          </div>
        ))}
      </div>
    </div>
  );
}