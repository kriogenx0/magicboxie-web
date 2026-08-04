import { useState, type FormEvent } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { ApiError } from "../api/client";

export function LoginPage() {
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      await login(password);
      const from = (location.state as { from?: string } | null)?.from ?? "/";
      navigate(from, { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Login failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-[radial-gradient(ellipse_at_top,_#333,_#141414_55%)] px-4">
      <form onSubmit={handleSubmit} className="w-full max-w-md rounded-sm bg-black/75 p-8 shadow-2xl backdrop-blur-sm sm:p-12">
        <h1 className="netflix-logo mb-8 text-3xl font-black uppercase text-[#e50914]">MagicBox</h1>
        <label className="mb-1 block text-sm text-neutral-300" htmlFor="password">
          Password
        </label>
        <input
          id="password"
          type="password"
          autoFocus
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="mb-4 w-full rounded-sm border border-neutral-700 bg-neutral-800 px-3 py-3 text-white focus:border-red-600 focus:outline-none"
        />
        {error && <p className="mb-4 text-sm text-red-500">{error}</p>}
        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-sm bg-[#e50914] py-3 font-bold text-white transition hover:bg-[#f6121d] disabled:opacity-50"
        >
          {submitting ? "Signing in…" : "Sign in"}
        </button>
      </form>
    </div>
  );
}
