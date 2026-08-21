import api from "./api";

export async function submitEssay(topicId, content) {
  const { data } = await api.post("/essays", { topic_id: topicId, content });
  return data;
}

export async function listEssays(limit = 20, offset = 0) {
  const { data } = await api.get("/essays", { params: { limit, offset } });
  return data;
}

export async function getEssay(id) {
  const { data } = await api.get(`/essays/${id}`);
  return data;
}