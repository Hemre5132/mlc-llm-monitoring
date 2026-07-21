"use client";

import { useEffect, useState } from "react";
import { backfillScores, listSessions } from "@/services/llm.service";
import { getDashboardStats, getMyStats } from "@/services/stats.service";

export default function DashboardPage() {
  const [stats, setStats] = useState(null);
  const [myStats, setMyStats] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    const loadStats = async () => {
      try {
        await backfillScores();

        const [dashboardData, personalData, sessionsData] = await Promise.all([
          getDashboardStats(),
          getMyStats(),
          listSessions(),
        ]);

        const scoreBySessionId = new Map(
          (dashboardData.sessions || []).map((session) => [String(session.id), session])
        );
        const sessions = (sessionsData.sessions || []).map((session) => ({
          ...session,
          ...(scoreBySessionId.get(String(session.id)) || {}),
        }));

        setStats({ ...dashboardData, sessions });
        setMyStats(personalData);
      } catch (error) {
        console.error("Dashboard verileri yüklenemedi", error);
        setError("Dashboard verileri alınamadı. Backend bağlantısını ve oturumunuzu kontrol edin.");
      }
    };

    loadStats();
  }, []);

  if (error) return <p className="p-8 text-red-600">{error}</p>;
  if (!stats) return <p className="p-8">Yükleniyor...</p>;

  const sessionRows = Array.isArray(stats.sessions)
    ? stats.sessions
    : [];

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="mb-6 text-xl font-semibold">Deci.Scoring Dashboard</h1>

      <div className="mb-6 grid gap-4 md:grid-cols-2">
        <div className="rounded-xl border bg-white p-4">
          <p className="text-sm text-gray-500">Toplam Oturum</p>
          <p className="text-2xl font-bold">{stats.total_sessions ?? 0}</p>
        </div>
        <div className="rounded-xl border bg-white p-4">
          <p className="text-sm text-gray-500">Ortalama Skor</p>
          <p className="text-2xl font-bold">
            {stats.scored_messages > 0 ? `${Number(stats.average_score).toFixed(1)}/100` : "Henüz skor yok"}
          </p>
        </div>
      </div>

      <div className="mb-6 rounded-xl border bg-white p-4">
        <p className="text-sm text-gray-500">Kişisel Ortalama</p>
        <p className="text-2xl font-bold">
          {myStats?.scored_messages > 0
            ? `${Number(myStats.average_score).toFixed(1)}/100`
            : "Henüz skor yok"}
        </p>
      </div>

      <div className="space-y-3">
        {sessionRows.length === 0 ? (
          <div className="rounded-lg border bg-white p-4 text-sm text-gray-500">
            Henüz sohbet bulunmuyor.
          </div>
        ) : (
          sessionRows.map((session) => (
            <div
              key={session.id}
              className="flex items-center justify-between rounded-lg border bg-white p-3"
            >
              <div>
                <p className="text-sm font-medium">{session.title || session.id}</p>
                <p className="text-xs text-gray-500">{session.model_name || "Model"}</p>
              </div>
              <span className="font-semibold text-indigo-600">
                {session.scored_messages > 0
                  ? `${Number(session.average_score ?? session.AverageScore).toFixed(1)}/100`
                  : "Henüz skorlanmadı"}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}