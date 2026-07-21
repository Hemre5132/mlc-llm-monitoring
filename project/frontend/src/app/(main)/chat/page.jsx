"use client";

import { useState, useEffect } from "react";
// llm.service.js dosyasını içe aktarıyoruz. (Yolun projedeki yerine göre gerekirse @/services/... yapabilirsin)
import { initLLMEngine, generateResponse } from "@/services/llm.service";

export default function ChatPage() {
  const [isWebGPUSupported, setIsWebGPUSupported] = useState(null);
  
  // Model yükleme durumları
  const [isEngineReady, setIsEngineReady] = useState(false);
  const [loadingText, setLoadingText] = useState("Model başlatılıyor...");
  const [isGenerating, setIsGenerating] = useState(false); // Yanıt akarken butonu kilitlemek için

  const [messages, setMessages] = useState([
    { role: "assistant", content: "Merhaba! Ben cihazınızda çalışan yerel yapay zekayım. Size nasıl yardımcı olabilirim?" },
  ]);
  const [input, setInput] = useState("");

  // 1. Adım: Sayfa açıldığında WebGPU kontrolü yap
  useEffect(() => {
    if ("gpu" in navigator) {
      setIsWebGPUSupported(true);
    } else {
      setIsWebGPUSupported(false);
    }
  }, []);

  // 2. Adım: WebGPU varsa modeli indirmeye/yüklemeye başla
  useEffect(() => {
    if (isWebGPUSupported) {
      const loadModel = async () => {
        try {
          await initLLMEngine((progress) => {
            // progress.text içinde "Fetching... %45" gibi MLC'nin kendi logları vardır
            setLoadingText(progress.text);
          });
          setIsEngineReady(true);
        } catch (error) {
          console.error(error);
          setLoadingText("Model yüklenirken bir hata oluştu.");
        }
      };
      loadModel();
    }
  }, [isWebGPUSupported]);

  // 3. Adım: Mesaj Gönderme ve Yanıtı Ekrana Yazdırma (Streaming)
  async function handleSend() {
    // Boş mesaj atılmasını, modelin hazır olmamasını veya halihazırda cevap yazıyor olmasını engelle
    if (!input.trim() || !isEngineReady || isGenerating) return;

    const userText = input;
    setInput("");
    setIsGenerating(true);

    // Ekrana kullanıcının mesajını ve asistanın içi boş "yükleniyor" mesajını ekle
    const newMessages = [
      ...messages,
      { role: "user", content: userText },
      { role: "assistant", content: "..." }, // Gelecek olan yanıt burayı dolduracak
    ];
    setMessages(newMessages);

    try {
      // Modele göndereceğimiz mesaj geçmişi (son eklediğimiz boş asistan mesajını hariç tutuyoruz)
      const messagesForModel = newMessages.slice(0, -1);

      // Modeli çağırıyoruz. streamCallback fonksiyonu her yeni kelimede tetiklenir
      await generateResponse(messagesForModel, (currentText) => {
        setMessages((prev) => {
          const updatedMessages = [...prev];
          // Son sıradaki asistan mesajının içeriğini güncelliyoruz (daktilo efekti)
          updatedMessages[updatedMessages.length - 1].content = currentText;
          return updatedMessages;
        });
      });

      // TODO: İleride logMessage(sessionId, {...}) backend çağrısını buraya ekleyeceğiz

    } catch (error) {
      console.error("Yanıt üretilirken hata:", error);
      // Hata olursa ekrandaki "..." kısmını hata mesajıyla değiştir
      setMessages((prev) => {
        const updated = [...prev];
        updated[updated.length - 1].content = "Yanıt oluşturulurken bir hata meydana geldi.";
        return updated;
      });
    } finally {
      setIsGenerating(false);
    }
  }

  // --- ARAYÜZ (RENDER) BÖLÜMÜ ---

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
      
      {/* Model Yükleme / Başarı Bildirimi */}
      <div className={`p-3 text-center text-xs font-medium border-b ${
        isEngineReady ? "bg-green-50 text-green-700 border-green-100" : "bg-blue-50 text-blue-700 border-blue-100 animate-pulse"
      }`}>
        {isEngineReady ? "✓ Yerel AI Motoru Hazır" : loadingText}
      </div>

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
          placeholder={isEngineReady ? "Mesajınızı yazın..." : "Model yükleniyor, lütfen bekleyin..."}
          disabled={!isEngineReady || isGenerating}
        />
        <button
          onClick={handleSend}
          disabled={!isEngineReady || isGenerating}
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed"
        >
          {isGenerating ? "Yazıyor..." : "Gönder"}
        </button>
      </div>
    </div>
  );
}