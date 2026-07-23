import api from "./api";
import { CreateMLCEngine } from "@mlc-ai/web-llm";

// --- WEBLLM AYARLARI ---
let engine = null;
let isLoading = false;
let loadPromise = null;

// Orijinal çalışan model — reload fix öncesi buydu ve sorunsuz çalışıyordu.
// Sadece reload fix'i (useRef + isModelLoading guard) koruyoruz.
const SELECTED_MODEL = "Qwen2.5-1.5B-Instruct-q4f16_1-MLC";

// --------------------------------------------------
// 1. BACKEND (VERİTABANI) FONKSİYONLARI
// --------------------------------------------------

export async function createSession(payload = {}) {
  const { data } = await api.post("/llm/sessions", payload);
  return data;
}

export async function listSessions() {
  const { data } = await api.get("/llm/sessions");
  return data;
}

export async function logMessage(sessionId, payload) {
  const { data } = await api.post(`/llm/sessions/${sessionId}/messages`, payload);
  return data;
}

export async function backfillScores() {
  const { data } = await api.post("/llm/scores/backfill");
  return data;
}

// --------------------------------------------------
// 2. YEREL YAPAY ZEKA (WEBLLM) FONKSİYONLARI
// --------------------------------------------------

export function getEngine() {
  return engine;
}

export function isModelLoading() {
  return isLoading;
}

export function isModelReady() {
  return engine !== null;
}

/** Device lost veya hata durumunda engine'i sıfırla */
export function resetEngine() {
  engine = null;
  isLoading = false;
  loadPromise = null;
}

export async function initLLMEngine(progressCallback) {
  // Eğer motor zaten çalışıyorsa, tekrar kurma
  if (engine) return engine;

  // Aynı anda birden fazla init çağrısını engelle
  if (isLoading) {
    if (loadPromise) return loadPromise;
    throw new Error("Model is already being loaded but no promise tracked.");
  }

  try {
    isLoading = true;

    loadPromise = CreateMLCEngine(
      SELECTED_MODEL,
      { initProgressCallback: progressCallback }
    );

    engine = await loadPromise;
    return engine;
  } catch (error) {
    engine = null;
    console.error("LLM Motoru başlatılamadı:", error);
    throw error;
  } finally {
    isLoading = false;
    loadPromise = null;
  }
}

export async function generateResponse(messages, streamCallback) {
  if (!engine) {
    console.warn("Motor henüz hazır değil, bekleniyor...");
    throw new Error("Engine is not initialized yet.");
  }

  const cleanMessages = messages.filter(m => m.role !== 'system');

  const formattedMessages = [
    { 
      role: "system", 
      content: "You are a highly intelligent and helpful AI assistant. You MUST always respond strictly in English, regardless of the language the user uses to ask the question. Be concise and accurate." 
    },
    ...cleanMessages
  ];

  try {
    const chunks = await engine.chat.completions.create({
      messages: formattedMessages,
      temperature: 0.4,
      stream: true,
    });

    let fullReply = "";
    for await (const chunk of chunks) {
      const text = chunk.choices[0]?.delta?.content || "";
      fullReply += text;
      
      if (streamCallback) {
        streamCallback(fullReply);
      }
    }

    return fullReply;
  } catch (error) {
    const errorStr = error?.toString?.() || error?.message || "";
    const isDeviceLost = 
      errorStr.includes("Device was lost") || 
      errorStr.includes("GPU") ||
      error?.message?.includes("Device was lost");

    if (isDeviceLost) {
      console.error("GPU device lost — engine sıfırlanıyor:", error);
      resetEngine();
    }

    throw error;
  }
}

export async function getMessages(sessionId) {
  const { data } = await api.get(`/llm/sessions/${sessionId}/messages`);

  const rawMessages = data?.messages ?? [];
  return rawMessages.map((message) => ({
    role: message.role,
    content: message.content,
  }));
}