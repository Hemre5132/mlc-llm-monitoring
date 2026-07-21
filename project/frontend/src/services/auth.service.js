import api from "./api";
import { TOKEN_KEY } from "@/lib/constants";

export async function registerUser(payload) {
  // payload: { name, email, password }
  const { data } = await api.post("/auth/register", payload);
  return data;
}

export async function loginUser(payload) {
  // payload: { email, password }
  const { data } = await api.post("/auth/login", payload);
  if (data?.access_token) {
    localStorage.setItem(TOKEN_KEY, data.access_token);
  }
  return data;
}

export async function getCurrentUser() {
  const { data } = await api.get("/auth/me");
  return data;
}

export function logoutUser() {
  localStorage.removeItem(TOKEN_KEY);
}