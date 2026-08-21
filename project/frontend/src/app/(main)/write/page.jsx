"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import {
  getDailyTopic,
  getRandomTopic,
  createCustomTopic,
} from "@/services/topics.service";
import { submitEssay, getEssay } from "@/services/essays.service";

const CRITERIA_META = {
  task_achievement: {
    label: "Task Achievement",
    desc: "Konuyu tam olarak ele alıyor mu, ilgili fikir ve örnekler var mı?",
    color: "bg-blue-500",
  },
  coherence_cohesion: {
    label: "Coherence & Cohesion",
    desc: "Mantıksal akış, net paragraflar ve bağlaç kullanımı nasıl?",
    color: "bg-indigo-500",
  },
  grammar_accuracy: {
    label: "Grammar Accuracy",
    desc: "Cümleler dilbilgisi açısından doğru mu (zaman, uyum, artikel, edat)?",
    color: "bg-violet-500",
  },
  vocabulary_range: {
    label: "Vocabulary Range",
    desc: "Kelime dağarcığı çeşitli, hassas ve konuya uygun mu?",
    color: "bg-purple-500",
  },
  spelling_mechanics: {
    label: "Spelling & Mechanics",
    desc: "Yazım, noktalama ve büyük harf kullanımı doğru mu?",
    color: "bg-emerald-500",
  },
  sentence_structure: {
    label: "Sentence Structure",
    desc: "Cümle çeşitliliği var mı, run-on ve fragment'lerden kaçınılmış mı?",
    color: "bg-cyan-500",
  },
};

const ERROR_CATEGORY_META = {
  grammar: { label: "Grammar", color: "bg-red-100 text-red-700" },
  vocabulary: { label: "Vocabulary", color: "bg-orange-100 text-orange-700" },
  spelling: { label: "Spelling", color: "bg-yellow-100 text-yellow-700" },
  punctuation: { label: "Punctuation", color: "bg-blue-100 text-blue-700" },
  sentence_structure: { label: "Sentence Structure", color: "bg-purple-100 text-purple-700" },
};

function ScoreBar({ value, color }) {
  const pct = Math.min(100, Math.max(0, value));
  return (
    <div className="h-2 w-full overflow-hidden rounded-full bg-gray-200">
      <div
        className={`h-full rounded-full transition-all duration-500 ${color}`}
        style={{ width: `${pct}%` }}
      />
    </div>
  );
}

function parseErrorList(errorList) {
  if (!errorList) return [];
  try {
    if (typeof errorList === "string") return JSON.parse(errorList);
    return errorList;
  } catch {
    return [];
  }
}

function parseStrengths(strengths) {
  if (!strengths) return [];
  try {
    if (typeof strengths === "string") return JSON.parse(strengths);
    return strengths;
  } catch {
    return [];
  }
}

function WriteContent() {
  const searchParams = useSearchParams();
  const [topic, setTopic] = useState(null);
  const [text, setText] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState(null);
  const [result, setResult] = useState(null);
  const [initialLoading, setInitialLoading] = useState(true);
  const [showCustomTopic, setShowCustomTopic] = useState(false);
  const [customTopicText, setCustomTopicText] = useState("");
  const [readOnlyEssay, setReadOnlyEssay] = useState(null);

  // Sayfa yüklendiğinde: essayId varsa o essay'i salt-okunur göster,
  // yoksa günün konusunu yükle
  useEffect(() => {
    const init = async () => {
      try {
        const essayId = searchParams?.get("essayId");

        if (essayId) {
          // Salt-okunur essay görünümü
          const essay = await getEssay(essayId);
          setReadOnlyEssay(essay);
          setTopic(essay.topic || null);
          setText(essay.content || "");
          if (essay.score) {
            setResult({
              overall_score: essay.score.overall_score,
              cefr_estimate: essay.score.cefr_estimate,
              task_achievement: essay.score.task_achievement,
              coherence_cohesion: essay.score.coherence_cohesion,
              grammar_accuracy: essay.score.grammar_accuracy,
              vocabulary_range: essay.score.vocabulary_range,
              spelling_mechanics: essay.score.spelling_mechanics,
              sentence_structure: essay.score.sentence_structure,
              error_list: parseErrorList(essay.score.error_list),
              strengths: parseStrengths(essay.score.strengths),
              reasoning: essay.score.reasoning,
            });
          }
          setInitialLoading(false);
          return;
        }

        // Günün konusunu yükle
        const dailyTopic = await getDailyTopic();
        setTopic(dailyTopic);
      } catch (err) {
        console.error("Konu yüklenemedi:", err);
        setError("Günün konusu alınamadı. Backend bağlantısını kontrol edin.");
      } finally {
        setInitialLoading(false);
      }
    };

    init();
  }, [searchParams]);

  async function handleNewTopic() {
    try {
      const newTopic = await getRandomTopic(topic?.id);
      setTopic(newTopic);
      setError(null);
    } catch (err) {
      console.error("Yeni konu alınamadı:", err);
      setError("Yeni konu alınamadı.");
    }
  }

  async function handleCreateCustomTopic() {
    const trimmed = customTopicText.trim();
    if (!trimmed) return;

    try {
      const newTopic = await createCustomTopic(trimmed);
      setTopic(newTopic);
      setCustomTopicText("");
      setShowCustomTopic(false);
      setError(null);
    } catch (err) {
      console.error("Özel konu oluşturulamadı:", err);
      setError("Özel konu oluşturulamadı.");
    }
  }

  async function handleSubmit() {
    const trimmed = text.trim();
    if (!trimmed || isSubmitting || !topic) return;

    setIsSubmitting(true);
    setError(null);
    setResult(null);

    try {
      const data = await submitEssay(topic.id, trimmed);
      const score = data?.score;

      if (!score) {
        setError("Skor sonucu alınamadı.");
        return;
      }

      setResult({
        overall_score: score.overall_score,
        cefr_estimate: score.cefr_estimate,
        task_achievement: score.task_achievement,
        coherence_cohesion: score.coherence_cohesion,
        grammar_accuracy: score.grammar_accuracy,
        vocabulary_range: score.vocabulary_range,
        spelling_mechanics: score.spelling_mechanics,
        sentence_structure: score.sentence_structure,
        error_list: parseErrorList(score.error_list),
        strengths: parseStrengths(score.strengths),
        reasoning: score.reasoning,
      });
    } catch (err) {
      console.error("Essay gönderim hatası:", err);
      setError(
        err?.response?.data?.error || "Essay gönderilirken bir hata oluştu."
      );
    } finally {
      setIsSubmitting(false);
    }
  }

  const criteriaKeys = Object.keys(CRITERIA_META);
  const overall = result?.overall_score ?? 0;
  const cefr = result?.cefr_estimate ?? "";
  const errorList = result?.error_list ?? [];
  const strengths = result?.strengths ?? [];
  const wordCount = text.trim().split(/\s+/).filter(Boolean).length;

  if (initialLoading) {
    return (
      <div className="flex h-[85vh] items-center justify-center">
        <p className="text-gray-500">Yükleniyor...</p>
      </div>
    );
  }

  return (
    <div className="mx-auto flex h-[85vh] max-w-6xl flex-col gap-4 p-4">
      {/* Başlık */}
      <div className="text-center">
        <h1 className="text-xl font-bold text-gray-800">
          Daily English Writing Coach
        </h1>
        <p className="text-sm text-gray-500">
          Günün konusu hakkında bir essay yazın, AI her kriter için puanlasın
        </p>
      </div>

      {/* Hata mesajı */}
      {error && (
        <div className="rounded-lg border border-orange-300 bg-orange-50 p-3 text-center text-sm text-orange-800">
          {error}
        </div>
      )}

      {/* Konu kartı */}
      {topic && (
        <div className="rounded-xl border border-indigo-200 bg-indigo-50 p-4">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0 flex-1">
              <div className="mb-1 flex items-center gap-2">
                <span className="rounded-full bg-indigo-100 px-2.5 py-0.5 text-xs font-semibold text-indigo-700">
                  {topic.category || "topic"}
                </span>
                {topic.difficulty && (
                  <span className="rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-600">
                    {topic.difficulty}
                  </span>
                )}
              </div>
              <p className="text-sm font-medium text-gray-800">{topic.text}</p>
            </div>
            {!readOnlyEssay && (
              <div className="flex flex-col gap-2">
                <button
                  onClick={handleNewTopic}
                  className="rounded-md bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-indigo-700"
                >
                  Başka konu göster
                </button>
                <button
                  onClick={() => setShowCustomTopic(!showCustomTopic)}
                  className="rounded-md border border-indigo-300 px-3 py-1.5 text-xs font-medium text-indigo-700 transition-colors hover:bg-indigo-100"
                >
                  Kendi konumu yaz
                </button>
              </div>
            )}
          </div>

          {showCustomTopic && !readOnlyEssay && (
            <div className="mt-3 flex gap-2">
              <input
                type="text"
                value={customTopicText}
                onChange={(e) => setCustomTopicText(e.target.value)}
                placeholder="Kendi konunuzu yazın..."
                className="flex-1 rounded-md border border-indigo-300 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-300"
              />
              <button
                onClick={handleCreateCustomTopic}
                disabled={!customTopicText.trim()}
                className="rounded-md bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-indigo-700 disabled:bg-indigo-400"
              >
                Kaydet
              </button>
            </div>
          )}
        </div>
      )}

      {/* Ana içerik: Sol panel + Sağ panel */}
      <div className="flex flex-1 gap-4 overflow-hidden">
        {/* === SOL PANEL: Essay Girişi === */}
        <div className="flex w-1/2 flex-col rounded-xl border bg-white shadow-sm">
          <div className="flex items-center justify-between border-b px-4 py-3">
            <span className="text-sm font-semibold text-gray-700">
              Your Essay
            </span>
            {isSubmitting && (
              <span className="text-xs font-medium text-indigo-600">
                Puanlanıyor...
              </span>
            )}
          </div>

          <textarea
            className="flex-1 resize-none px-4 py-3 text-sm text-gray-800 placeholder-gray-400 focus:outline-none"
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="Konu hakkında İngilizce essay'inizi buraya yazın..."
            disabled={isSubmitting || !!readOnlyEssay}
          />

          <div className="flex items-center justify-between border-t px-4 py-3">
            <div className="flex items-center gap-2">
              <span className="text-xs text-gray-400">{wordCount} kelime</span>
              {!readOnlyEssay && (
                <span className="text-xs text-gray-400">
                  • En az 100 kelime öneriyoruz
                </span>
              )}
            </div>
            {!readOnlyEssay && (
              <button
                onClick={handleSubmit}
                disabled={isSubmitting || !text.trim() || !topic}
                className="rounded-lg bg-indigo-600 px-6 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:cursor-not-allowed disabled:bg-indigo-400"
              >
                {isSubmitting ? "Puanlanıyor..." : "Submit"}
              </button>
            )}
          </div>
        </div>

        {/* === SAĞ PANEL: Skor Sonucu === */}
        <div className="flex w-1/2 flex-col rounded-xl border bg-white shadow-sm">
          <div className="flex items-center justify-between border-b px-4 py-3">
            <span className="text-sm font-semibold text-gray-700">
              Scoring Results
            </span>
            {result && (
              <div className="flex items-center gap-2">
                {cefr && (
                  <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-700">
                    {cefr}
                  </span>
                )}
                <span className="rounded-full bg-indigo-100 px-3 py-1 text-xs font-bold text-indigo-700">
                  {overall.toFixed(1)} / 100
                </span>
              </div>
            )}
          </div>

          <div className="flex-1 overflow-y-auto p-4">
            {!result ? (
              <div className="flex h-full items-center justify-center">
                <p className="text-center text-sm text-gray-400">
                  Henüz puanlama yapılmadı.
                  <br />
                  Soldaki essay'i yazıp Submit butonuna tıklayın.
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                {/* Kriter kartları */}
                {criteriaKeys.map((key) => {
                  const meta = CRITERIA_META[key];
                  const val = result[key];
                  if (val === undefined || val === null) return null;
                  return (
                    <div
                      key={key}
                      className="rounded-lg border border-gray-100 bg-gray-50 p-3"
                    >
                      <div className="mb-1 flex items-center justify-between">
                        <span className="text-xs font-semibold text-gray-700">
                          {meta.label}
                        </span>
                        <span className="text-xs font-bold text-gray-800">
                          {Number(val).toFixed(1)}
                        </span>
                      </div>
                      <ScoreBar value={val} color={meta.color} />
                      <p className="mt-1 text-[10px] leading-tight text-gray-500">
                        {meta.desc}
                      </p>
                    </div>
                  );
                })}

                {/* Hata Listesi */}
                <div className="rounded-lg border border-gray-200 bg-gray-50 p-3">
                  <h3 className="mb-2 text-xs font-bold text-gray-700">
                    Hata Listesi ({errorList.length})
                  </h3>
                  {errorList.length === 0 ? (
                    <p className="text-xs text-emerald-600">
                      Harika, belirgin bir hata bulunamadı!
                    </p>
                  ) : (
                    <div className="space-y-2">
                      {errorList.map((err, idx) => {
                        const meta = ERROR_CATEGORY_META[err.category] || {
                          label: err.category,
                          color: "bg-gray-100 text-gray-700",
                        };
                        return (
                          <div
                            key={idx}
                            className="rounded-md border border-gray-200 bg-white p-2"
                          >
                            <div className="mb-1 flex items-center gap-2">
                              <span
                                className={`rounded-full px-2 py-0.5 text-[10px] font-semibold ${meta.color}`}
                              >
                                {meta.label}
                              </span>
                            </div>
                            <p className="text-xs text-gray-500 line-through">
                              {err.original}
                            </p>
                            <p className="text-xs font-medium text-emerald-600">
                              → {err.correction}
                            </p>
                            <p className="mt-1 text-[10px] leading-tight text-gray-500">
                              {err.explanation}
                            </p>
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>

                {/* Strengths */}
                {strengths.length > 0 && (
                  <div className="rounded-lg border border-emerald-200 bg-emerald-50 p-3">
                    <h3 className="mb-2 text-xs font-bold text-emerald-800">
                      Güçlü Yönler
                    </h3>
                    <ul className="list-inside list-disc space-y-1">
                      {strengths.map((s, idx) => (
                        <li key={idx} className="text-xs text-emerald-700">
                          {s}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {/* Reasoning */}
                {result.reasoning && (
                  <div className="rounded-lg border border-gray-200 bg-gray-50 p-3">
                    <p className="mb-1 text-[10px] font-semibold text-gray-500 uppercase">
                      Genel Yorum
                    </p>
                    <p className="text-xs leading-relaxed text-gray-700">
                      {result.reasoning}
                    </p>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default function WritePage() {
  return (
    <Suspense
      fallback={
        <div className="flex h-[85vh] items-center justify-center">
          <p className="text-gray-500">Yükleniyor...</p>
        </div>
      }
    >
      <WriteContent />
    </Suspense>
  );
}