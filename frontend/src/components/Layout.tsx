import { useCallback, useState, type ReactNode } from "react";
import { useDropzone } from "react-dropzone";
import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../hooks/useAuth";
import { useUploadQueue } from "../hooks/useUploadQueue";
import { UploadDropzone } from "./UploadDropzone";

export function Layout({ children }: { children: ReactNode }) {
  const [showUpload, setShowUpload] = useState(false);
  const { logout } = useAuth();
  const location = useLocation();
  const onMusic = location.pathname.startsWith("/music");
  const hasHero = location.pathname === "/" || location.pathname.startsWith("/movies/");
  const { addFiles } = useUploadQueue();

  // The whole page is a drop target on every authenticated route, not just
  // the upload panel's own box -- open the panel on drop so progress is
  // visible even if the user dropped somewhere else entirely.
  const onDrop = useCallback(
    (files: File[]) => {
      if (files.length === 0) return;
      addFiles(files);
      setShowUpload(true);
    },
    [addFiles],
  );

  const { getRootProps, isDragActive } = useDropzone({ onDrop, noClick: true, noKeyboard: true });

  return (
    <div {...getRootProps()} className="app-surface min-h-screen pb-24">
      {isDragActive && (
        <div className="pointer-events-none fixed inset-0 z-50 flex items-center justify-center border-4 border-dashed border-red-600 bg-black/80">
          <p className="text-2xl font-semibold text-white">Drop to upload</p>
        </div>
      )}
      <header
        className={`${hasHero ? "absolute" : "sticky"} top-0 z-20 flex h-16 w-full items-center justify-between bg-gradient-to-b from-black via-black/85 to-transparent px-4 sm:h-[68px] sm:px-10`}
      >
        <div className="flex items-center gap-7">
          <Link to="/" className="inline-flex h-10 w-10 items-center justify-center rounded-sm bg-[#e50914] text-sm font-black tracking-[0.16em] text-white shadow-sm sm:h-11 sm:w-11 sm:text-base" aria-label="MagicBox">
            MB
          </Link>
          <nav className="flex items-center gap-5 text-sm font-medium">
            <Link
              to="/"
              className={!onMusic ? "text-white" : "text-neutral-400 transition hover:text-neutral-200"}
            >
              Movies
            </Link>
            <Link
              to="/music"
              className={onMusic ? "text-white" : "text-neutral-400 transition hover:text-neutral-200"}
            >
              Music
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-2 sm:gap-3">
          <button
            onClick={() => setShowUpload((v) => !v)}
            className="rounded-sm border border-neutral-500 px-2.5 py-1 text-xs font-medium text-neutral-100 transition hover:border-white hover:bg-white/10 sm:px-3 sm:text-sm"
          >
            {showUpload ? "Close" : "Add Media"}
          </button>
          <button
            onClick={logout}
            className="hidden rounded-sm px-2 py-1 text-xs text-neutral-400 transition hover:text-white sm:block"
          >
            Log out
          </button>
        </div>
      </header>

      {showUpload && (
        <div className={`mb-4 ${hasHero ? "pt-16 sm:pt-[68px]" : ""}`}>
          <UploadDropzone />
        </div>
      )}

      <main>{children}</main>
    </div>
  );
}
