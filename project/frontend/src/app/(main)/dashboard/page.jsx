"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { listSessions, getMessages } from "@/services/llm.service";
import { getDashboardStats, getMyStats } from "@/services/stats.service";

const CRITERIA_LABELS = {
  opening_hook: "Opening Hook",
  discovery: "Discovery",
  value_proposition: "Value Proposition",
  objection_handling: "Objection Handling",
  closing_power: "Closing Power",
  persuasiveness_tone: "Persuasiveness & Tone",
  compliance_safety: "Compliance & Safety",
  personalization: "Personalization",
  structure_flow: "Structure & Flow",
};

function ScoreBar({ value, color }) {
  const pct = Math.min(100, Math.max(0, value));
  return (
    <div className="h-1.5 w-full overflow-hidden rounded-full bg-gray-200">
      <div
        className={`h-full rounded-full transition-all duration-500 ${color}`}
        style={{ width: `${pct}%` }}
      />
    </div>
  );
}

export default function DashboardPage() {
  const router = useRouter();
  const [stats, setStats] = useState(null);
  const [myStats, setMyStats] = useState(null);
  const [error, setError] = useState(null);
  const [expandedSession, setExpandedSession] = useState(null);
  const [sessionDetail, setSessionDetail] = useState(null);
  const [loadingDetail, setLoadingDetail] = useState(false);

  useEffect(() => {
    const loadStats = async () => {
      try {
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

  async function handleToggleDetail(session) {
    if (expandedSession === session.id) {
      setExpandedSession(null);
      setSessionDetail(null);
      return;
    }

    setExpandedSession(session.id);
    setLoadingDetail(true);
    setSessionDetail(null);

    try {
      const data = await getMessages(session.id);
      const msgs = data?.messages ?? [];

      // Son user mesajını bul (analiz edilen metin)
      const userMessages = msgs.filter((m) => m.role === "user");
      const lastUserMsg = userMessages.length > 0 ? userMessages[userMessages.length - 1] : null;

      // Skor bilgilerini session'dan al
      const categoryScores = session.category_scores || session.CategoryScores || null;
      const hasScores = session.scored_messages > 0;

      setSessionDetail({
        text: lastUserMsg?.content || null,
        categoryScores: categoryScores,
        hasScores: hasScores,
        averageScore: session.average_score ?? session.AverageScore ?? 0,
        requiresRevision: Boolean(session.requires_revision || session.RequiresRevision),
      });
    } catch (err) {
      console.error("Detay yüklenemedi:", err);
      setSessionDetail({ error: "Detaylar yüklenemedi." });
    } finally {
      setLoadingDetail(false);
    }
  }

  if (error) return <p className="p-8 text-red-600">{error}</p>;
  if (!stats) return <p className="p-8">Yükleniyor...</p>;

  const sessionRows = Array.isArray(stats.sessions) ? stats.sessions : [];

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="mb-6 text-xl font-semibold">Analysis History</h1>

      {/* Üst özet kartları */}
      <div className="mb-6 grid gap-4 md:grid-cols-2">
        <div className="rounded-xl border bg-white p-4">
          <p className="text-sm text-gray-500">Toplam Analiz</p>
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

      {/* Session listesi */}
      <div className="space-y-3">
        {sessionRows.length === 0 ? (
          <div className="rounded-lg border bg-white p-4 text-sm text-gray-500">
            Henüz analiz bulunmuyor. Chat sayfasına gidip bir metin analiz edin.
          </div>
        ) : (
          sessionRows.map((session) => {
            const hasScores = session.scored_messages > 0;
            const revisionRequired = Boolean(session.requires_revision || session.RequiresRevision);
            const isExpanded = expandedSession === session.id;

            return (
              <div key={session.id}>
                {/* Kart */}
                <div
                  onClick={() => handleToggleDetail(session)}
                  className={`cursor-pointer rounded-lg border bg-white p-3 transition-colors hover:bg-gray-50 ${
                    isExpanded ? "border-indigo-300 ring-1 ring-indigo-200" : ""
                  }`}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium text-gray-800">
                        {session.title || `Analysis ${new Date(session.created_at).toLocaleString("tr-TR")}`}
                      </p>
                      <p className="text-xs text-gray-400">
                        {new Date(session.created_at).toLocaleString("tr-TR")}
                      </p>
                    </div>
                    <div className="flex flex-col items-end gap-1">
                      {revisionRequired && (
                        <span className="rounded-full bg-red-100 px-2.5 py-0.5 text-xs font-semibold text-red-700">
                          Revizyon gerekli
                        </span>
                      )}
                      <span className="font-semibold text-indigo-600">
                        {hasScores
                          ? `${Number(session.average_score ?? session.AverageScore).toFixed(1)}/100`
                          : "Henüz skorlanmadı"}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Detay paneli */}
                {isExpanded && (
                  <div className="rounded-b-lg border border-t-0 border-indigo-200 bg-indigo-50 p-4">
                    {loadingDetail ? (
                      <p className="text-sm text-gray-500">Yükleniyor...</p>
                    ) : sessionDetail?.error ? (
                      <p className="text-sm text-red-600">{sessionDetail.error}</p>
                    ) : (
                      <div className="space-y-3">
                        {/* Analiz edilen metin */}
                        {sessionDetail?.text && (
                          <div>
                            <p className="mb-1 text-xs font-semibold text-gray-500 uppercase">
                              Analiz Edilen Metin
                            </p>
                            <p className="max-h-24 overflow-y-auto rounded-md bg-white p-2 text-xs leading-relaxed text-gray-700">
                              {sessionDetail.text}
                            </p>
                          </div>
                        )}

                        {/* Kategori skorları */}
                        {sessionDetail?.hasScores && sessionDetail?.categoryScores && (
                          <div>
                            <p className="mb-1 text-xs font-semibold text-gray-500 uppercase">
                              Kategori Skorları
                            </p>
                            <div className="grid grid-cols-3 gap-2">
                              <div className="rounded-md bg-white p-2 text-center">
                                <p className="text-[10px] font-medium text-gray-500">Effectiveness</p>
                                <p className="text-sm font-bold text-indigo-700">
                                  {Number(sessionDetail.categoryScores.effectiveness ?? 0).toFixed(1)}
                                </p>
                              </div>
                              <div className="rounded-md bg-white p-2 text-center">
                                <p className="text-[10px] font-medium text-gray-500">Structure</p>
                                <p className="text-sm font-bold text-indigo-700">
                                  {Number(sessionDetail.categoryScores.structure ?? 0).toFixed(1)}
                                </p>
                              </div>
                              <div className="rounded-md bg-white p-2 text-center">
                                <p className="text-[10px] font-medium text-gray-500">Safety</p>
                                <p className="text-sm font-bold text-indigo-700">
                                  {Number(sessionDetail.categoryScores.safety ?? 0).toFixed(1)}
                                </p>
                              </div>
                            </div>
                          </div>
                        )}

                        {/* Overall skor */}
                        {sessionDetail?.hasScores && (
                          <div className="flex items-center justify-between rounded-md bg-white p-2">
                            <span className="text-xs font-semibold text-gray-600">Overall Skor</span>
                            <span className="text-sm font-bold text-indigo-700">
                              {Number(sessionDetail.averageScore).toFixed(1)} / 100
                            </span>
                          </div>
                        )}

                        {/* Revizyon uyarısı */}
                        {sessionDetail?.requiresRevision && (
                          <p className="text-xs font-semibold text-red-600">
                            ⚠ Revizyon gerekli — compliance/safety skoru düşük
                          </p>
                        )}

                        {/* Chat sayfasında aç butonu */}
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            router.push(`/chat?sessionId=${session.id}`);
                          }}
                          className="w-full rounded-md bg-indigo-600 px-3 py-2 text-xs font-medium text-white transition-colors hover:bg-indigo-700"
                        >
                          Chat sayfasında aç
                        </button>
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}