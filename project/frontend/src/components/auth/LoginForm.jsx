"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import Input from "@/components/ui/Input";
import Button from "@/components/ui/Button";
import { loginUser, getCurrentUser } from "@/services/auth.service";
import { useAuth } from "@/hooks/useAuth";

export default function LoginForm() {
  const router = useRouter();
  const { setUser } = useAuth();
  const [form, setForm] = useState({ email: "", password: "" });
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  function handleChange(e) {
    setForm({ ...form, [e.target.name]: e.target.value });
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setError("");
    setIsLoading(true);
    try {
      await loginUser(form);
      const me = await getCurrentUser();
      setUser(me);
      router.push("/dashboard");
    } catch (err) {
      setError(err.response?.data?.message || "Giriş başarısız. Bilgilerinizi kontrol edin.");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <h1 className="mb-6 text-xl font-semibold">Giriş Yap</h1>

      <Input
        label="E-posta"
        type="email"
        name="email"
        value={form.email}
        onChange={handleChange}
        required
      />
      <Input
        label="Şifre"
        type="password"
        name="password"
        value={form.password}
        onChange={handleChange}
        required
      />

      {error && <p className="mb-4 text-sm text-red-500">{error}</p>}

      <Button type="submit" isLoading={isLoading}>
        Giriş Yap
      </Button>

      <p className="mt-4 text-center text-sm text-gray-500">
        Hesabın yok mu?{" "}
        <Link href="/register" className="text-indigo-600 hover:underline">
          Kayıt Ol
        </Link>
      </p>
    </form>
  );
}