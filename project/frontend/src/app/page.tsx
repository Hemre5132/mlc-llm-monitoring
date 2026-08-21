import { redirect } from "next/navigation";

export default function HomePage() {
  // Şimdilik projeyi başlattığımızda doğrudan giriş sayfasına yönlendiriyoruz.
  // İlerleyen adımlarda Auth Context (JWT yönetimi) eklendiğinde, 
  // burada kullanıcının token'ı varsa '/write' veya '/dashboard' sayfasına, 
  // yoksa '/login' sayfasına yönlendirme mantığını kuracağız.
  
  redirect("/login");
}