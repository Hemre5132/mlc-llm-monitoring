import api from "./api";

export async function getDailyTopic() {
  const { data } = await api.get("/topics/daily");
  return data;
}

export async function getRandomTopic(excludeId) {
  const { data } = await api.get("/topics/random", {
    params: excludeId ? { exclude: excludeId } : {},
  });
  return data;
}

export async function createCustomTopic(text) {
  const { data } = await api.post("/topics/custom", { text });
  return data;
}