import api from "./api";

// --------------------------------------------------
// BACKEND (VERİTABANI + OLLAMA) FONKSİYONLARI
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

// Backend'e (Go) konuşma geçmişini gönderir; backend bunu Ollama'ya iletir,
// cevabı bekler, otomatik olarak kaydeder ve skorlar. Artık modelin kendisi
// tarayıcıda değil, sunucu tarafında (Ollama) çalışıyor.
export async function generateChat(sessionId, messages) {
  const { data } = await api.post(`/llm/sessions/${sessionId}/generate`, {
    messages,
  });
  return data;
}

export async function analyzeText(sessionId, text) {
  const { data } = await api.post(`/llm/sessions/${sessionId}/analyze`, { text });
  return data;
}

export async function getSessionScores(sessionId) {
  const { data } = await api.get(`/llm/sessions/${sessionId}/score`);
  return data;
}

export async function getMessages(sessionId) {
  const { data } = await api.get(`/llm/sessions/${sessionId}/messages`);

  const rawMessages = data?.messages ?? [];
  return rawMessages.map((message) => ({
    role: message.role,
    content: message.content,
  }));
}