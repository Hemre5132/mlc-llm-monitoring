"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import Input from "@/components/ui/Input";
import Button from "@/components/ui/Button";
import { registerUser } from "@/services/auth.service";

export default function RegisterForm() {
  const router = useRouter();
  const [form, setForm] = useState({ name: "", email: "", password: "" });
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
      await registerUser(form);
      router.push("/login");
    } catch (err) {
      setError(err.response?.data?.message || "Kayıt başarısız. Tekrar deneyin.");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <h1 className="mb-6 text-xl font-semibold">Kayıt Ol</h1>

      <Input
        label="Ad Soyad"
        name="name"
        value={form.name}
        onChange={handleChange}
        required
      />
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
        Kayıt Ol
      </Button>

      <p className="mt-4 text-center text-sm text-gray-500">
        Zaten hesabın var mı?{" "}
        <Link href="/login" className="text-indigo-600 hover:underline">
          Giriş Yap
        </Link>
      </p>
    </form>
  );
}