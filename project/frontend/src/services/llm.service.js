import api from "./api";
import { CreateMLCEngine } from "@mlc-ai/web-llm";

// --- WEBLLM AYARLARI ---
let engine = null;
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

// --------------------------------------------------
// 2. YEREL YAPAY ZEKA (WEBLLM) FONKSİYONLARI
// --------------------------------------------------

export async function initLLMEngine(progressCallback) {
  // Eğer motor zaten çalışıyorsa, tekrar kurma
  if (engine) return engine;

  try {
    // Motoru oluştur, modeli indir ve ilerlemeyi (yüzdeyi) arayüze bildir
    engine = await CreateMLCEngine(
      SELECTED_MODEL,
      { initProgressCallback: progressCallback }
    );
    return engine;
  } catch (error) {
    console.error("LLM Motoru başlatılamadı:", error);
    throw error;
  }
}

export async function generateResponse(messages, streamCallback) {
  if (!engine) {
    console.warn("Motor henüz hazır değil, bekleniyor...");
    // Küçük bir bekleme süresi veya doğrudan hata fırlatma
    throw new Error("Engine is not initialized yet.");
  }

  // 1. DOKUNUŞ: SİSTEM KOMUTU (Sadece İngilizce)
  // Kullanıcının mesajlarının en başına gizli bir "system" mesajı ekleyerek modeli yönlendiriyoruz.
  const cleanMessages = messages.filter(m => m.role !== 'system');

  const formattedMessages = [
    { 
      role: "system", 
      content: "You are a highly intelligent and helpful AI assistant. You MUST always respond strictly in English, regardless of the language the user uses to ask the question. Be concise and accurate." 
    },
    ...cleanMessages // Senin arayüzden gönderdiğin mesaj geçmişi bunun altına ekleniyor
  ];

  // Modele mesajları gönder ve stream (akan) yanıt iste
  const chunks = await engine.chat.completions.create({
    messages: formattedMessages, // Artık ham messages yerine formatlanmış olanı gönderiyoruz
    temperature: 0.4,
    stream: true,
  });

  let fullReply = "";
  for await (const chunk of chunks) {
    const text = chunk.choices[0]?.delta?.content || "";
    fullReply += text;
    
    // Her yeni kelime geldiğinde arayüzü güncelle (Daktilo efekti)
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