"use client";

import { useState, useEffect } from "react";

export default function ChatPage() {
  // WebGPU destek durumunu tutacağımız state
  const [isWebGPUSupported, setIsWebGPUSupported] = useState(null);
  
  // Claude'un verdiği sohbet state'leri
  const [messages, setMessages] = useState([
    { role: "assistant", content: "Merhaba! WebGPU taramasını geçtin. Ben şu an mock (sahte) asistanım." },
  ]);
  const [input, setInput] = useState("");

  // Sayfa açıldığında WebGPU kontrolü yap
  useEffect(() => {
    if ("gpu" in navigator) {
      setIsWebGPUSupported(true);
    } else {
      setIsWebGPUSupported(false);
    }
  }, []);

  function handleSend() {
    if (!input.trim()) return;
    
    // TODO: llm.service.js -> logMessage() ile backend'e logla
    // TODO: @mlc-ai/web-llm entegrasyonu burada devreye girecek
    
    setMessages((prev) => [
      ...prev,
      { role: "user", content: input },
      { role: "assistant", content: "(mock yanıt - model henüz bağlı değil)" },
    ]);
    setInput("");
  }

  // 1. Aşama: Kontrol ediliyor (Ekran titremesini önler)
  if (isWebGPUSupported === null) {
    return (
      <div className="flex h-[80vh] items-center justify-center">
        <p className="text-gray-500 animate-pulse">Sistem gereksinimleri kontrol ediliyor...</p>
      </div>
    );
  }

  // 2. Aşama: Hata Durumu (WebGPU yoksa kırmızı uyarı)
  if (isWebGPUSupported === false) {
    return (
      <div className="mx-auto flex h-[80vh] max-w-2xl flex-col items-center justify-center p-6 text-center">
        <div className="rounded-xl border border-red-200 bg-red-50 p-6 text-red-700 shadow-sm">
          <svg className="mx-auto mb-4 h-12 w-12 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <h3 className="mb-2 text-xl font-bold">WebGPU Desteklenmiyor</h3>
          <p className="text-sm">
            Tarayıcınız yapay zeka modelini cihazınızda yerel olarak çalıştırmak için gereken WebGPU teknolojisini desteklemiyor. Lütfen güncel bir Google Chrome, Microsoft Edge veya Brave tarayıcısı kullanın.
          </p>
        </div>
      </div>
    );
  }

  // 3. Aşama: Başarılı Durum (Claude'un verdiği UI + Yeşil Bildirim)
  return (
    <div className="mx-auto flex h-[80vh] max-w-2xl flex-col rounded-xl border bg-white shadow-sm">
      <div className="flex-1 space-y-3 overflow-y-auto p-4">
        
        {/* Başarı Bildirimi */}
        <div className="mb-6 rounded bg-green-50 p-2 text-center text-xs font-medium text-green-700 border border-green-100">
          ✓ WebGPU aktif. Donanımınız modeli yerel çalıştırmak için uygun.
        </div>

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
          className="flex-1 rounded-lg border px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSend()}
          placeholder="Mesajınızı yazın..."
        />
        <button
          onClick={handleSend}
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700"
        >
          Gönder
        </button>
      </div>
    </div>
  );
}