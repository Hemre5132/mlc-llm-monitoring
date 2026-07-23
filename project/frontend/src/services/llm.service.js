import api from "./api";
import { CreateMLCEngine } from "@mlc-ai/web-llm";

// --- WEBLLM AYARLARI ---
let engine = null;
let isLoading = false;
let loadPromise = null;
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
  // payload: { role: "user" | "assistant", content, score? }
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

/**
 * Modül-seviyesinde engine singleton'ı.
 * isLoading ve loadPromise ile aynı anda birden fazla init çağrısını engeller.
 */
export function getEngine() {
  return engine;
}

export function isModelLoading() {
  return isLoading;
}

export function isModelReady() {
  return engine !== null;
}

export async function initLLMEngine(progressCallback) {
  // Eğer motor zaten çalışıyorsa, tekrar kurma
  if (engine) return engine;

  // Aynı anda birden fazla init çağrısını engelle
  if (isLoading) {
    // Yükleme devam ediyorsa, mevcut promise'ı döndür
    if (loadPromise) return loadPromise;
    throw new Error("Model is already being loaded but no promise tracked.");
  }

  try {
    isLoading = true;

    // Promise'ı kaydet ki concurrent çağrılar aynı sonucu beklesin
    loadPromise = CreateMLCEngine(
      SELECTED_MODEL,
      { initProgressCallback: progressCallback }
    );

    engine = await loadPromise;
    return engine;
  } catch (error) {
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

  // 1. DOKUNUŞ: SİSTEM KOMUTU (Sadece İngilizce)
  const cleanMessages = messages.filter(m => m.role !== 'system');

  const formattedMessages = [
    { 
      role: "system", 
      content: "You are a highly intelligent and helpful AI assistant. You MUST always respond strictly in English, regardless of the language the user uses to ask the question. Be concise and accurate." 
    },
    ...cleanMessages
  ];

  // Modele mesajları gönder ve stream (akan) yanıt iste
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
}

export async function getMessages(sessionId) {
  const { data } = await api.get(`/llm/sessions/${sessionId}/messages`);

  const rawMessages = data?.messages ?? [];
  return rawMessages.map((message) => ({
    role: message.role,
    content: message.content,
  }));
}