import api from "./api";
import { TOKEN_KEY } from "@/lib/constants";

export async function registerUser(payload) {
  // payload: { name, email, password }
  const { data } = await api.post("/auth/register", payload);
  return data;
}

function setTokenCookie(token) {
  // Middleware'in okuyabilmesi için cookie de set et
  document.cookie = `${TOKEN_KEY}=${token}; path=/; max-age=86400; SameSite=Lax`;
}

function removeTokenCookie() {
  document.cookie = `${TOKEN_KEY}=; path=/; max-age=0`;
}

export async function loginUser(payload) {
  // payload: { email, password }
  const { data } = await api.post("/auth/login", payload);
  if (data?.access_token) {
    localStorage.setItem(TOKEN_KEY, data.access_token);
    setTokenCookie(data.access_token);
  }
  return data;
}

export async function getCurrentUser() {
  const { data } = await api.get("/auth/me");
  return data;
}

export function logoutUser() {
  localStorage.removeItem(TOKEN_KEY);
  removeTokenCookie();
}
