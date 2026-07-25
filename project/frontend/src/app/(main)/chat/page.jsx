"use client";

import { useEffect, useState } from "react";
import {
  createSession,
  generateChat,
  getMessages,
  logMessage,
} from "@/services/llm.service";

const SESSION_STORAGE_KEY = "currentSessionId";
const DEFAULT_ASSISTANT_MESSAGE = {
  role: "assistant",
  content: "Hi! I am your AI assistant. How can I help you today?",
};

const SYSTEM_MESSAGE = {
  role: "system",
  content:
    "You are a highly intelligent and helpful AI assistant. You MUST always respond strictly in English, regardless of the language the user uses to ask the question. Be concise and accurate.",
};

function estimateTokenCount(text) {
  if (!text) return 0;
  return Math.max(1, Math.ceil(text.trim().split(/\s+/).filter(Boolean).length));
}

export default function ChatPage() {
  const [isGenerating, setIsGenerating] = useState(false);
  const [sessionId, setSessionId] = useState(null);
  const [messages, setMessages] = useState([DEFAULT_ASSISTANT_MESSAGE]);
  const [input, setInput] = useState("");
  const [error, setError] = useState(null);

  const startNewChat = async () => {
    try {
      const session = await createSession({ model_name: "gemma2:2b" });
      const nextSessionId = session.id || session.ID;

      if (typeof window !== "undefined") {
        window.localStorage.setItem(SESSION_STORAGE_KEY, nextSessionId);
      }

      setSessionId(nextSessionId);
      setMessages([DEFAULT_ASSISTANT_MESSAGE]);
      setInput("");
      setError(null);
    } catch (err) {
      console.error("Yeni sohbet başlatılamadı:", err);
    }
  };

  useEffect(() => {
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

        const session = await createSession({ model_name: "gemma2:2b" });
        const sId = session.id || session.ID;

        if (typeof window !== "undefined") {
          window.localStorage.setItem(SESSION_STORAGE_KEY, sId);
        }

        setSessionId(sId);
      } catch (err) {
        console.error("Oturum başlatılamadı:", err);
        // Eski session silinmiş olabilir, localStorage'ı temizle
        if (typeof window !== "undefined") {
          window.localStorage.removeItem(SESSION_STORAGE_KEY);
        }
      }
    };

    initChat();
  }, []);

  async function handleSend() {
    if (!input.trim() || isGenerating || !sessionId) return;

    const userText = input;
    setInput("");
    setIsGenerating(true);
    setError(null);

    const newMessages = [
      ...messages,
      { role: "user", content: userText },
      { role: "assistant", content: "..." },
    ];
    setMessages(newMessages);

    try {
      await logMessage(sessionId, {
        role: "user",
        content: userText,
        latency_ms: 0,
        token_count: estimateTokenCount(userText),
      });
    } catch (err) {
      console.error("Kullanıcı mesajı kaydedilemedi", err);
    }

    try {
      const conversation = [
        SYSTEM_MESSAGE,
        ...newMessages.slice(0, -1).map((m) => ({ role: m.role, content: m.content })),
      ];

      const result = await generateChat(sessionId, conversation);
      const reply = result?.message?.content || "Yanıt alınamadı.";

      setMessages((prev) => {
        const updated = [...prev];
        updated[updated.length - 1].content = reply;
        return updated;
      });
    } catch (err) {
      console.error("Yanıt üretilirken hata:", err);
      const backendMsg = err?.response?.data?.error || "Ollama bağlantı hatası";
      setError(backendMsg);
      setMessages((prev) => {
        const updated = [...prev];
        updated[updated.length - 1].content = "Yanıt oluşturulurken bir hata meydana geldi.";
        return updated;
      });
    } finally {
      setIsGenerating(false);
    }
  }

  return (
    <div className="mx-auto flex h-[80vh] max-w-2xl flex-col rounded-xl border bg-white shadow-sm">
      <div className="flex items-center justify-between border-b p-3">
        <div className="flex-1 text-center text-xs font-medium text-green-700">
          {isGenerating ? "Yanıt üretiliyor..." : "✓ Bağlı"}
        </div>
        <button
          onClick={startNewChat}
          className="ml-2 rounded-md bg-gray-100 px-3 py-1 text-xs font-semibold text-gray-600 hover:bg-gray-200 transition-colors"
        >
          + New Chat
        </button>
      </div>

      {error && (
        <div className="mx-4 mt-2 rounded-lg border border-orange-300 bg-orange-50 p-3 text-center text-sm text-orange-800">
          {error}
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
          placeholder="Type your message..."
          disabled={isGenerating}
        />
        <button
          onClick={handleSend}
          disabled={isGenerating}
          className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed"
        >
          {isGenerating ? "Writing..." : "Send"}
        </button>
      </div>
    </div>
  );
}