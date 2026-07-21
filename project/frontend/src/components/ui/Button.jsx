export default function Button({ children, isLoading, className = "", ...props }) {
  return (
    <button
      disabled={isLoading}
      className={`w-full rounded-lg bg-indigo-600 px-4 py-2 font-medium text-white transition hover:bg-indigo-700 disabled:opacity-50 ${className}`}
      {...props}
    >
      {isLoading ? "Yükleniyor..." : children}
    </button>
  );
}