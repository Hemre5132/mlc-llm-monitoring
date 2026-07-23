"use client";

import { useEffect, useRef, useState } from "react";
import {
  createSession,
  generateResponse,
  getMessages,
  initLLMEngine,
  isModelLoading,
  isModelReady,
  logMessage,
  resetEngine,
} from "@/services/llm.service";

const SESSION_STORAGE_KEY = "currentSessionId";
const DEFAULT_ASSISTANT_MESSAGE = {
  role: "assistant",
  content: "Hi! I am your local AI assistant running directly on your device. How can I help you today?",
};

function estimateTokenCount(text) {
  if (!text) return 0;
  return Math.max(1, Math.ceil(text.trim().split(/\s+/).filter(Boolean).length));
}

export default function ChatPage() {
  const [isWebGPUSupported, setIsWebGPUSupported] = useState(null);
  const [isEngineReady, setIsEngineReady] = useState(false);
  const [loadingText, setLoadingText] = useState("Model başlatılıyor...");
  const [isGenerating, setIsGenerating] = useState(false);
  const [sessionId, setSessionId] = useState(null);
  const [messages, setMessages] = useState([DEFAULT_ASSISTANT_MESSAGE]);
  const [input, setInput] = useState("");
  const [deviceError, setDeviceError] = useState(null); // device lost hatası UI'ı

  // Engine init'inin bir kez çalışmasını garanti altına almak için ref guard
  const engineInitStarted = useRef(false);

  // Device lost sonrası modeli yeniden yükle
  const reloadModel = async () => {
    setDeviceError(null);
    setLoadingText("Model yeniden başlatılıyor...");
    setIsEngineReady(false);
    engineInitStarted.current = false;
    resetEngine();

    // Kısa bir bekleme — GPU'nun toparlanmasına izin ver
    await new Promise((r) => setTimeout(r, 1500));

    try {
      await initLLMEngine((progress) => {
        setLoadingText(progress.text);
      });
      setIsEngineReady(true);
      setLoadingText("✓ Model ready");
    } catch (error) {
      console.error("Yeniden yükleme hatası:", error);
      setDeviceError("Model yeniden yüklenemedi. Sayfayı tazelemeyi deneyin.");
    }
  };

  const startNewChat = async () => {
    try {
      const session = await createSession({ model_name: "gemma-2b-it-q4f16_1-MLC" });
      const nextSessionId = session.id || session.ID;

      if (typeof window !== "undefined") {
        window.localStorage.setItem(SESSION_STORAGE_KEY, nextSessionId);
      }

      setSessionId(nextSessionId);
      setMessages([DEFAULT_ASSISTANT_MESSAGE]);
      setInput("");
    } catch (error) {
      console.error("Yeni sohbet başlatılamadı:", error);
    }
  };

  useEffect(() => {
    // --- WebGPU ve Engine kontrolü (tek effect, tek kaynak) ---
    if ("gpu" in navigator) {
      setIsWebGPUSupported(true);
    } else {
      setIsWebGPUSupported(false);
      return;
    }

    // --- Eğer engine singleton'ı zaten yüklüyse state'i hemen güncelle ---
    if (isModelReady()) {
      setIsEngineReady(true);
      setLoadingText("✓ Model ready");
      engineInitStarted.current = true;
    }

    // --- Eğer yükleme halihazırda devam ediyorsa bekleyen promise'ı yakala ---
    if (isModelLoading() && !engineInitStarted.current) {
      engineInitStarted.current = true;
    }

    // --- Session yönetimi ---
    const initChat = async () => {
      try {
        const storedSessionId =
          typeof window !== "undefined" ? window.localStorage.getItem(SESSION_STORAGE_KEY) : null;

        if (storedSessionId) {
          setSessionId(storedSessionId);

          const previousMessages = await getMessages(storedSessionId);
          if (Array.isArray(previousMessages) && previousMessages.length > 0) {
            setMessages(previousMessages);
          }
          return;
        }

        const session = await createSession({ model_name: "gemma-2b-it-q4f16_1-MLC" });
        const sId = session.id || session.ID;

        if (typeof window !== "undefined") {
          window.localStorage.setItem(SESSION_STORAGE_KEY, sId);
        }

        setSessionId(sId);
      } catch (error) {
        console.error("Oturum başlatılamadı:", error);
      }
    };

    initChat();

    // --- Engine başlatma (ref guard ile) ---
    if (!engineInitStarted.current) {
      engineInitStarted.current = true;

      const loadModel = async () => {
        try {
          await initLLMEngine((progress) => {
            setLoadingText(progress.text);
          });
          setIsEngineReady(true);
          setLoadingText("✓ Model ready");
        } catch (error) {
          console.error(error);
          setLoadingText("Model yüklenirken bir hata oluştu.");
        }
      };

      loadModel();
    }
  }, []); // ← boş dependency: sadece mount'ta bir kez çalışır

  async function handleSend() {
    if (!input.trim() || !isEngineReady || isGenerating) return;

    const userText = input;
    const startedAt = Date.now();
    setInput("");
    setIsGenerating(true);

    const newMessages = [
      ...messages,
      { role: "user", content: userText },
      { role: "assistant", content: "..." },
    ];
    setMessages(newMessages);

    if (sessionId) {
      try {
        await logMessage(sessionId, {
          role: "user",
          content: userText,
          latency_ms: 0,
          token_count: estimateTokenCount(userText),
        });
      } catch (error) {
        console.error("Kullanıcı mesajı kaydedilemedi", error);
      }
    }

    try {
      const messagesForModel = newMessages.slice(0, -1);
      const finalReply = await generateResponse(messagesForModel, (currentText) => {
        setMessages((prev) => {
          const updatedMessages = [...prev];
          updatedMessages[updatedMessages.length - 1].content = currentText;
          return updatedMessages;
        });
      });

      if (sessionId) {
        try {
          await logMessage(sessionId, {
            role: "assistant",
            content: finalReply,
            raw_output: finalReply,
            latency_ms: Date.now() - startedAt,
            token_count: estimateTokenCount(finalReply),
          });
        } catch (error) {
          console.error("Asistan mesajı kaydedilemedi", error);
        }
      }
    } catch (error) {
      console.error("Yanıt üretilirken hata:", error);

      // Device lost hatası mı kontrol et
      const isDeviceLost = 
        error?.message?.includes("Device was lost") || 
        error?.message?.includes("GPU") ||
        error?.toString?.()?.includes("Device was lost");

      if (isDeviceLost) {
        setIsEngineReady(false);
        setDeviceError("GPU belleği tükendi. Model sıfırlandı. Lütfen 'Modeli Yeniden Yükle' butonuna tıklayın.");
      }

      setMessages((prev) => {
        const updated = [...prev];
        updated[updated.length - 1].content = "Yanıt oluşturulurken bir hata meydana geldi.";
        return updated;
      });
    } finally {
      setIsGenerating(false);
    }
  }

  if (isWebGPUSupported === null) {
    return (
      <div className="flex h-[80vh] items-center justify-center">
        <p className="text-gray-500 animate-pulse">Sistem gereksinimleri kontrol ediliyor...</p>
      </div>
    );
  }

  if (isWebGPUSupported === false) {
    return (
      <div className="mx-auto flex h-[80vh] max-w-2xl flex-col items-center justify-center p-6 text-center">
        <div className="rounded-xl border border-red-200 bg-red-50 p-6 text-red-700 shadow-sm">
          <h3 className="mb-2 text-xl font-bold">WebGPU Desteklenmiyor</h3>
          <p className="text-sm">Tarayıcınız yerel AI modelini desteklemiyor.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto flex h-[80vh] max-w-2xl flex-col rounded-xl border bg-white shadow-sm">
      <div className="flex items-center justify-between border-b p-3">
        <div
          className={`flex-1 text-center text-xs font-medium ${
            isEngineReady ? "text-green-700" : "text-blue-700 animate-pulse"
          }`}
        >
          {isEngineReady ? "✓ Local AI Engine Ready" : loadingText}
        </div>
        <button
          onClick={startNewChat}
          className="ml-2 rounded-md bg-gray-100 px-3 py-1 text-xs font-semibold text-gray-600 hover:bg-gray-200 transition-colors"
        >
          + New Chat
        </button>
      </div>

      {/* Device lost hatası banner'ı */}
      {deviceError && (
        <div className="mx-4 mt-2 rounded-lg border border-orange-300 bg-orange-50 p-3 text-center text-sm text-orange-800">
          <p className="mb-2">{deviceError}</p>
          <button
            onClick={reloadModel}
            className="rounded-md bg-orange-500 px-4 py-1.5 text-xs font-semibold text-white hover:bg-orange-600 transition-colors"
          >
            🔄 Modeli Yeniden Yükle
          </button>
        </div>
      )}

      <div className="flex-1 space-y-3 overflow-y-auto p-4">
        {messages.map((m, i) => (
          <div
            key={i}
            className={`max-w-[75%] rounded-lg px-3 py-2 text-sm ${
              m.role === "user"
                ? "ml-auto bg-indigo-600 text-white"
                : "bg-gray-100 text-gray-800"
            }`}
          >
            {m.content}
          </div>
        ))}
      </div>

      <div className="flex gap-2 border-t p-3 bg-gray-50 rounded-b-xl">
        <input
          className="flex-1 rounded-lg border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50 disabled:bg-gray-200"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSend()}
          placeholder={isEngineReady ? "Type your message..." : "Loading model, please wait..."}
          disabled={!isEngineReady || isGenerating}
        />
        <button
          onClick={handleSend}
          disabled={!isEngineReady || isGenerating}
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed"
        >
          {isGenerating ? "Writing..." : "Send"}
        </button>
      </div>
    </div>
  );
}