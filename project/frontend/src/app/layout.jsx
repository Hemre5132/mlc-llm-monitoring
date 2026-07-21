import { AuthProvider } from "@/context/AuthContext";
import "./globals.css";

export const metadata = { title: "MasterFabric Academy App" };

export default function RootLayout({ children }) {
  return (
    <html lang="tr">
      <body className="bg-gray-50 text-gray-900 antialiased">
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}