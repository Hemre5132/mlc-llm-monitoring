"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { createSession, analyzeText, getMessages, getSessionScores } from "@/services/llm.service";

const SESSION_STORAGE_KEY = "currentSessionId";

const CRITERIA_META = {
  opening_hook: {
    label: "Opening Hook",
    desc: "İlk cümle dikkat çekici mi, ilgi kuruyor mu?",
    color: "bg-blue-500",
  },
  discovery: {
    label: "Discovery",
    desc: "İhtiyaç ve bağlamı net şekilde ortaya çıkarıyor mu?",
    color: "bg-indigo-500",
  },
  value_proposition: {
    label: "Value Proposition",
    desc: "Değer teklifi somut ve ikna edici mi?",
    color: "bg-violet-500",
  },
  objection_handling: {
    label: "Objection Handling",
    desc: "İtirazlara cevap veriyor ve savunma yaratıyor mu?",
    color: "bg-purple-500",
  },
  closing_power: {
    label: "Closing Power",
    desc: "Kapanış net, çağrı ve sonraki adım içeriyor mu?",
    color: "bg-pink-500",
  },
  persuasiveness_tone: {
    label: "Persuasiveness & Tone",
    desc: "Konuşma tonu güvenilir, ikna edici ve doğal mı?",
    color: "bg-rose-500",
  },
  compliance_safety: {
    label: "Compliance & Safety",
    desc: "Yüzeysel vaat, sahte aciliyet, manipülatif dil içermiyor mu?",
    color: "bg-emerald-500",
  },
  personalization: {
    label: "Personalization",
    desc: "Dinleyicinin bağlamına ve ihtiyacına uygun mu?",
    color: "bg-teal-500",
  },
  structure_flow: {
    label: "Structure & Flow",
    desc: "Akış mantıklı, bölüm geçişleri güçlü ve okunaklı mı?",
    color: "bg-cyan-500",
  },
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

function parseScoreCriteria(criteriaStr) {
  if (!criteriaStr) return null;
  try {
    if (typeof criteriaStr === "string") return JSON.parse(criteriaStr);
    return criteriaStr;
  } catch {
    return null;
  }
}

function ChatContent() {
  const searchParams = useSearchParams();
  const [sessionId, setSessionId] = useState(null);
  const [text, setText] = useState("");
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [error, setError] = useState(null);
  const [result, setResult] = useState(null);
  const [initialLoading, setInitialLoading] = useState(true);

  // Sayfa yüklendiğinde: sessionId varsa o session'ı yükle, yoksa yeni session oluştur
  useEffect(() => {
    const init = async () => {
      try {
        const urlSessionId = searchParams?.get("sessionId");

        if (urlSessionId) {
          setSessionId(urlSessionId);
          // Var olan session'ın mesajlarını ve skorlarını yükle
          try {
            const [messages, scoresData] = await Promise.all([
              getMessages(urlSessionId),
              getSessionScores(urlSessionId),
            ]);

            // Son user mesajını bul
            const userMsgs = (messages || []).filter((m) => m.role === "user");
            if (userMsgs.length > 0) {
              setText(userMsgs[userMsgs.length - 1].content);
            }

            // Skorları parse et
            const scores = scoresData?.scores || [];
            if (scores.length > 0) {
              const latestScore = scores[scores.length - 1];
              const criteria = parseScoreCriteria(latestScore.criteria);
              const catScores = parseScoreCriteria(latestScore.category_scores);

              if (criteria) {
                setResult({
                  ...criteria,
                  overall: latestScore.score,
                  effectiveness: catScores?.effectiveness ?? 0,
                  structure: catScores?.structure ?? 0,
                  compliance_safety: criteria.compliance_safety,
                  requires_revision: latestScore.requires_revision,
                  reasoning: "",
                });
              }
            }
          } catch (err) {
            console.error("Session verileri yüklenemedi:", err);
          }
          setInitialLoading(false);
          return;
        }

        // Yeni session oluştur
        const stored =
          typeof window !== "undefined"
            ? window.localStorage.getItem(SESSION_STORAGE_KEY)
            : null;

        if (stored) {
          setSessionId(stored);
          setInitialLoading(false);
          return;
        }

        const session = await createSession({ model_name: "gemma2:2b" });
        const sId = session.id || session.ID;

        if (typeof window !== "undefined") {
          window.localStorage.setItem(SESSION_STORAGE_KEY, sId);
        }

        setSessionId(sId);
      } catch (err) {
        console.error("Oturum başlatılamadı:", err);
        if (typeof window !== "undefined") {
          window.localStorage.removeItem(SESSION_STORAGE_KEY);
        }
      } finally {
        setInitialLoading(false);
      }
    };

    init();
  }, [searchParams]);

  async function handleAnalyze() {
    const trimmed = text.trim();
    if (!trimmed || isAnalyzing || !sessionId) return;

    setIsAnalyzing(true);
    setError(null);
    setResult(null);

    try {
      const data = await analyzeText(sessionId, trimmed);
      const r = data?.result;

      if (!r) {
        setError("Analiz sonucu alınamadı.");
        return;
      }

      setResult(r);
    } catch (err) {
      console.error("Analiz hatası:", err);
      setError(
        err?.response?.data?.error || "Analiz sırasında bir hata oluştu."
      );
    } finally {
      setIsAnalyzing(false);
    }
  }

  const criteriaKeys = Object.keys(CRITERIA_META);
  const overall = result?.overall ?? 0;
  const effectiveness = result?.effectiveness ?? 0;
  const structure = result?.structure ?? 0;
  const complianceSafety = result?.compliance_safety ?? 0;
  const requiresRevision = result?.requires_revision ?? false;

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
          Sales Script Analyzer
        </h1>
        <p className="text-sm text-gray-500">
          Bir satış metni yazın, AI her kriter için puanlasın
        </p>
      </div>

      {/* Hata mesajı */}
      {error && (
        <div className="rounded-lg border border-orange-300 bg-orange-50 p-3 text-center text-sm text-orange-800">
          {error}
        </div>
      )}

      {/* Ana içerik: Sol panel + Sağ panel */}
      <div className="flex flex-1 gap-4 overflow-hidden">
        {/* === SOL PANEL: Metin Girişi === */}
        <div className="flex w-1/2 flex-col rounded-xl border bg-white shadow-sm">
          <div className="flex items-center justify-between border-b px-4 py-3">
            <span className="text-sm font-semibold text-gray-700">
              Sales Script
            </span>
            {isAnalyzing && (
              <span className="text-xs font-medium text-indigo-600">
                Analiz ediliyor...
              </span>
            )}
          </div>

          <textarea
            className="flex-1 resize-none px-4 py-3 text-sm text-gray-800 placeholder-gray-400 focus:outline-none"
            value={text}
            onChange={(e) => setText(e.target.value)}
            placeholder="Satış scriptinizi buraya yapıştırın veya yazın..."
            disabled={isAnalyzing}
          />

          <div className="flex items-center justify-between border-t px-4 py-3">
            <span className="text-xs text-gray-400">
              {text.trim().split(/\s+/).filter(Boolean).length} kelime
            </span>
            <button
              onClick={handleAnalyze}
              disabled={isAnalyzing || !text.trim() || !sessionId}
              className="rounded-lg bg-indigo-600 px-6 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:cursor-not-allowed disabled:bg-indigo-400"
            >
              {isAnalyzing ? "Analiz ediliyor..." : "Analyze"}
            </button>
          </div>
        </div>

        {/* === SAĞ PANEL: Skor Tablosu === */}
        <div className="flex w-1/2 flex-col rounded-xl border bg-white shadow-sm">
          <div className="flex items-center justify-between border-b px-4 py-3">
            <span className="text-sm font-semibold text-gray-700">
              Scoring Results
            </span>
            {result && (
              <span className="rounded-full bg-indigo-100 px-3 py-1 text-xs font-bold text-indigo-700">
                {overall.toFixed(1)} / 100
              </span>
            )}
          </div>

          <div className="flex-1 overflow-y-auto p-4">
            {!result ? (
              <div className="flex h-full items-center justify-center">
                <p className="text-center text-sm text-gray-400">
                  Henüz analiz yapılmadı.
                  <br />
                  Soldaki metni yazıp Analyze butonuna tıklayın.
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

                {/* Özet skor kartı */}
                <div className="rounded-lg border border-indigo-200 bg-indigo-50 p-3">
                  <h3 className="mb-2 text-xs font-bold text-indigo-800">
                    Summary Scores
                  </h3>
                  <div className="grid grid-cols-3 gap-2">
                    <div className="rounded-md bg-white p-2 text-center">
                      <p className="text-[10px] font-medium text-gray-500">
                        Effectiveness
                      </p>
                      <p className="text-sm font-bold text-indigo-700">
                        {effectiveness.toFixed(1)}
                      </p>
                    </div>
                    <div className="rounded-md bg-white p-2 text-center">
                      <p className="text-[10px] font-medium text-gray-500">
                        Structure
                      </p>
                      <p className="text-sm font-bold text-indigo-700">
                        {structure.toFixed(1)}
                      </p>
                    </div>
                    <div className="rounded-md bg-white p-2 text-center">
                      <p className="text-[10px] font-medium text-gray-500">
                        Safety
                      </p>
                      <p className="text-sm font-bold text-indigo-700">
                        {complianceSafety.toFixed(1)}
                      </p>
                    </div>
                  </div>

                  {requiresRevision && (
                    <p className="mt-2 text-center text-xs font-semibold text-red-600">
                      ⚠ Revizyon gerekli — compliance/safety skoru düşük
                    </p>
                  )}
                </div>

                {/* Reasoning */}
                {result.reasoning && (
                  <div className="rounded-lg border border-gray-200 bg-gray-50 p-3">
                    <p className="mb-1 text-[10px] font-semibold text-gray-500 uppercase">
                      Reasoning
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

export default function SalesScriptAnalyzerPage() {
  return (
    <Suspense fallback={
      <div className="flex h-[85vh] items-center justify-center">
        <p className="text-gray-500">Yükleniyor...</p>
      </div>
    }>
      <ChatContent />
    </Suspense>
  );
}