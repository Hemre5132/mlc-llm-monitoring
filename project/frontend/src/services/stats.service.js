import api from "./api";

export async function getDashboardStats() {
  const { data } = await api.get("/stats/dashboard");
  return data;
}

export async function getMyStats() {
  const { data } = await api.get("/stats/me");
  return data;
}