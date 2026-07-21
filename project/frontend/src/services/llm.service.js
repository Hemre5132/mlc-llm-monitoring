import api from "./api";

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