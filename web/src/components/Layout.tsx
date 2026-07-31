import { useState, type ReactNode } from "react";
import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { UploadDropzone } from "./UploadDropzone";

export function Layout({ children }: { children: ReactNode }) {
  const [showUpload, setShowUpload] = useState(false);
  const { logout } = useAuth();
  const location = useLocation();
  const onMusic = location.pathname.startsWith("/music");

  return (
    <div className="min-h-screen bg-black pb-24">
      <header className="sticky top-0 z-20 flex items-center justify-between bg-gradient-to-b from-black to-transparent px-4 py-4 sm:px-8">
        <div className="flex items-center gap-6">
          <Link to="/" className="text-xl font-bold tracking-wide text-red-600">
            MagicBox
          </Link>
          <nav className="flex items-center gap-4 text-sm">
            <Link
              to="/"
              className={!onMusic ? "text-white" : "text-neutral-400 hover:text-white"}
            >
              Movies
            </Link>
            <Link
              to="/music"
              className={onMusic ? "text-white" : "text-neutral-400 hover:text-white"}
            >
              Music
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => setShowUpload((v) => !v)}
            className="rounded border border-neutral-700 px-3 py-1.5 text-sm text-neutral-200 transition hover:border-red-600 hover:text-white"
          >
            {showUpload ? "Close" : "Add Media"}
          </button>
          <button
            onClick={logout}
            className="rounded border border-neutral-700 px-3 py-1.5 text-sm text-neutral-400 transition hover:border-neutral-500 hover:text-white"
          >
            Log out
          </button>
        </div>
      </header>

      {showUpload && (
        <div className="mb-4">
          <UploadDropzone />
        </div>
      )}

      <main>{children}</main>
    </div>
  );
}
