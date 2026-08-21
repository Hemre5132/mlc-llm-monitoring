"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { listEssays } from "@/services/essays.service";
import { getDashboardStats, getMyStats } from "@/services/stats.service";

const CRITERIA_LABELS = {
  task_achievement: "Task Achievement",
  coherence_cohesion: "Coherence & Cohesion",
  grammar_accuracy: "Grammar Accuracy",
  vocabulary_range: "Vocabulary Range",
  spelling_mechanics: "Spelling & Mechanics",
  sentence_structure: "Sentence Structure",
};

const CRITERIA_COLORS = {
  task_achievement: "bg-blue-500",
  coherence_cohesion: "bg-indigo-500",
  grammar_accuracy: "bg-violet-500",
  vocabulary_range: "bg-purple-500",
  spelling_mechanics: "bg-emerald-500",
  sentence_structure: "bg-cyan-500",
};

const WEAKEST_ADVICE = {
  task_achievement: "Konuyu tam ele almak için ana fikri netleştir ve örneklerle destekle.",
  coherence_cohesion: "Paragraflar arası geçişler ve bağlaç kullanımına odaklan.",
  grammar_accuracy: "Gramer kurallarına odaklan: zamanlar, özne-yüklem uyumu, artikel ve edatlar.",
  vocabulary_range: "Kelime dağarcığını genişlet; eş anlamlılar ve konuya özel terimler kullan.",
  spelling_mechanics: "Yazım, noktalama ve büyük harf kullanımına dikkat et.",
  sentence_structure: "Cümle çeşitliliği ekle; basit, bileşik ve karmaşık cümleleri karıştır.",
};

const ERROR_CATEGORY_LABELS = {
  grammar: "Grammar",
  vocabulary: "Vocabulary",
  spelling: "Spelling",
  punctuation: "Punctuation",
  sentence_structure: "Sentence Structure",
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
  const [streak, setStreak] = useState(null);
  const [essays, setEssays] = useState([]);
  const [error, setError] = useState(null);
  const [expandedEssay, setExpandedEssay] = useState(null);

  useEffect(() => {
    const loadStats = async () => {
      try {
        const [dashboardData, personalData, essaysData, streakData] =
          await Promise.all([
            getDashboardStats(),
            getMyStats(),
            listEssays(50, 0),
            fetch("/api/stats/streak").then((r) => r.json()).catch(() => null),
          ]);

        setStats(dashboardData);
        setMyStats(personalData);
        setStreak(streakData);
        setEssays(essaysData.essays || []);
      } catch (error) {
        console.error("Dashboard verileri yüklenemedi", error);
        setError("Dashboard verileri alınamadı. Backend bağlantısını ve oturumunuzu kontrol edin.");
      }
    };

    loadStats();
  }, []);

  if (error) return <p className="p-8 text-red-600">{error}</p>;
  if (!stats) return <p className="p-8">Yükleniyor...</p>;

  const criteriaAverages = stats.criteria_averages || {};
  const weakestCriterion = stats.weakest_criterion || "task_achievement";
  const weakestValue = criteriaAverages[weakestCriterion] ?? 0;
  const errorCounts = stats.error_category_counts || {};

  // Hata kategorilerini en çoktan aza sırala
  const sortedErrorCategories = Object.entries(errorCounts)
    .filter(([, count]) => count > 0)
    .sort((a, b) => b[1] - a[1]);

  const maxErrorCount = Math.max(1, ...Object.values(errorCounts));

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="mb-6 text-xl font-semibold">Dashboard</h1>

      {/* Streak rozeti */}
      {streak && streak.current_streak > 0 && (
        <div className="mb-4 inline-block rounded-full bg-orange-100 px-4 py-1.5 text-sm font-semibold text-orange-700">
          🔥 {streak.current_streak} gün
        </div>
      )}

      {/* Üst özet kartları */}
      <div className="mb-6 grid gap-4 md:grid-cols-2">
        <div className="rounded-xl border bg-white p-4">
          <p className="text-sm text-gray-500">Toplam Essay</p>
          <p className="text-2xl font-bold">{stats.total_essays ?? 0}</p>
        </div>
        <div className="rounded-xl border bg-white p-4">
          <p className="text-sm text-gray-500">Ortalama Skor</p>
          <p className="text-2xl font-bold">
            {stats.total_essays > 0
              ? `${Number(stats.average_score).toFixed(1)}/100`
              : "Henüz skor yok"}
          </p>
        </div>
      </div>

      {/* En zayıf alan kartı */}
      {stats.total_essays > 0 && (
        <div className="mb-6 rounded-xl border border-red-200 bg-red-50 p-4">
          <p className="text-sm text-gray-500">En Zayıf Alanın</p>
          <p className="text-2xl font-bold text-red-700">
            {CRITERIA_LABELS[weakestCriterion] || weakestCriterion}
          </p>
          <p className="text-sm font-semibold text-red-600">
            Son essay'lerde ortalama {Number(weakestValue).toFixed(1)}/100
          </p>
          <p className="mt-1 text-xs text-gray-600">
            {WEAKEST_ADVICE[weakestCriterion] || "Bu alana odaklanarak geliştir."}
          </p>
        </div>
      )}

      {/* Hata kategorisi dağılımı */}
      {sortedErrorCategories.length > 0 && (
        <div className="mb-6 rounded-xl border bg-white p-4">
          <p className="mb-3 text-sm font-semibold text-gray-700">
            Hata Kategorileri
          </p>
          <div className="space-y-2">
            {sortedErrorCategories.map(([category, count]) => (
              <div key={category} className="flex items-center gap-2">
                <span className="w-32 text-xs text-gray-600">
                  {ERROR_CATEGORY_LABELS[category] || category}
                </span>
                <div className="h-2 flex-1 overflow-hidden rounded-full bg-gray-200">
                  <div
                    className="h-full rounded-full bg-indigo-500"
                    style={{ width: `${(count / maxErrorCount) * 100}%` }}
                  />
                </div>
                <span className="w-8 text-right text-xs font-semibold text-gray-700">
                  {count}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Kriter ortalamaları */}
      {stats.total_essays > 0 && (
        <div className="mb-6 rounded-xl border bg-white p-4">
          <p className="mb-3 text-sm font-semibold text-gray-700">
            Kriter Ortalamaları
          </p>
          <div className="space-y-2">
            {Object.entries(CRITERIA_LABELS).map(([key, label]) => {
              const val = criteriaAverages[key];
              if (val === undefined) return null;
              return (
                <div key={key} className="flex items-center gap-2">
                  <span className="w-40 text-xs text-gray-600">{label}</span>
                  <div className="flex-1">
                    <ScoreBar value={val} color={CRITERIA_COLORS[key]} />
                  </div>
                  <span className="w-10 text-right text-xs font-semibold text-gray-700">
                    {Number(val).toFixed(1)}
                  </span>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Essay listesi */}
      <div className="space-y-3">
        <p className="text-sm font-semibold text-gray-700">Essay Geçmişi</p>
        {essays.length === 0 ? (
          <div className="rounded-lg border bg-white p-4 text-sm text-gray-500">
            Henüz essay bulunmuyor. Write sayfasına gidip bir essay yazın.
          </div>
        ) : (
          essays.map((essay) => {
            const isExpanded = expandedEssay === essay.id;
            const score = essay.score;
            const topic = essay.topic;

            return (
              <div key={essay.id}>
                {/* Kart */}
                <div
                  onClick={() => setExpandedEssay(isExpanded ? null : essay.id)}
                  className={`cursor-pointer rounded-lg border bg-white p-3 transition-colors hover:bg-gray-50 ${
                    isExpanded ? "border-indigo-300 ring-1 ring-indigo-200" : ""
                  }`}
                >
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-medium text-gray-800">
                        {topic?.text || "Essay"}
                      </p>
                      <p className="text-xs text-gray-400">
                        {new Date(essay.created_at).toLocaleString("tr-TR")}
                      </p>
                    </div>
                    <div className="flex flex-col items-end gap-1">
                      {score?.cefr_estimate && (
                        <span className="rounded-full bg-emerald-100 px-2.5 py-0.5 text-xs font-semibold text-emerald-700">
                          {score.cefr_estimate}
                        </span>
                      )}
                      <span className="font-semibold text-indigo-600">
                        {score
                          ? `${Number(score.overall_score).toFixed(1)}/100`
                          : "Henüz skorlanmadı"}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Detay paneli */}
                {isExpanded && (
                  <div className="rounded-b-lg border border-t-0 border-indigo-200 bg-indigo-50 p-4">
                    <div className="space-y-3">
                      {/* Essay metni */}
                      <div>
                        <p className="mb-1 text-xs font-semibold text-gray-500 uppercase">
                          Essay
                        </p>
                        <p className="max-h-24 overflow-y-auto rounded-md bg-white p-2 text-xs leading-relaxed text-gray-700">
                          {essay.content}
                        </p>
                      </div>

                      {/* Kriter skorları */}
                      {score && (
                        <div>
                          <p className="mb-1 text-xs font-semibold text-gray-500 uppercase">
                            Kriter Skorları
                          </p>
                          <div className="space-y-1.5">
                            {Object.entries(CRITERIA_LABELS).map(
                              ([key, label]) => {
                                const val = score[key];
                                if (val === undefined) return null;
                                return (
                                  <div
                                    key={key}
                                    className="flex items-center gap-2"
                                  >
                                    <span className="w-40 text-xs text-gray-600">
                                      {label}
                                    </span>
                                    <div className="flex-1">
                                      <ScoreBar
                                        value={val}
                                        color={CRITERIA_COLORS[key]}
                                      />
                                    </div>
                                    <span className="w-10 text-right text-xs font-semibold text-gray-700">
                                      {Number(val).toFixed(1)}
                                    </span>
                                  </div>
                                );
                              }
                            )}
                          </div>
                        </div>
                      )}

                      {/* Write sayfasında aç butonu */}
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          router.push(`/write?essayId=${essay.id}`);
                        }}
                        className="w-full rounded-md bg-indigo-600 px-3 py-2 text-xs font-medium text-white transition-colors hover:bg-indigo-700"
                      >
                        Write sayfasında aç
                      </button>
                    </div>
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