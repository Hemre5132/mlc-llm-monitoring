"use client";

import { useEffect, useState } from "react";
import { listSessions } from "@/services/llm.service";
import { getDashboardStats, getMyStats } from "@/services/stats.service";

export default function DashboardPage() {
  const [stats, setStats] = useState(null);
  const [myStats, setMyStats] = useState(null);

  useEffect(() => {
    const loadStats = async () => {
      try {
        const [dashboardData, personalData, sessionsData] = await Promise.all([
          getDashboardStats(),
          getMyStats(),
          listSessions(),
        ]);

        const sessionList = Array.isArray(sessionsData?.sessions)
          ? sessionsData.sessions
          : [];

        const dashboardSessions = Array.isArray(dashboardData?.sessions) && dashboardData.sessions.length > 0
          ? dashboardData.sessions
          : sessionList.map((session) => ({
              id: session.id,
              title: session.title,
              model_name: session.model_name,
              average_score: 0,
            }));

        setStats({ ...dashboardData, sessions: dashboardSessions });
        setMyStats(personalData);
      } catch (error) {
        console.error("Dashboard verileri yüklenemedi", error);
      }
    };

    loadStats();
  }, []);

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
          <p className="text-2xl font-bold">{Number(stats.average_score ?? 0).toFixed(1)}/100</p>
        </div>
      </div>

      <div className="mb-6 rounded-xl border bg-white p-4">
        <p className="text-sm text-gray-500">Kişisel Ortalama</p>
        <p className="text-2xl font-bold">{Number(myStats?.average_score ?? 0).toFixed(1)}/100</p>
      </div>

      <div className="space-y-3">
        {sessionRows.length === 0 ? (
          <div className="rounded-lg border bg-white p-4 text-sm text-gray-500">
            Henüz skorlanmış sohbet bulunmuyor.
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
                {Number(session.average_score ?? session.AverageScore ?? 0).toFixed(1)}/100
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}