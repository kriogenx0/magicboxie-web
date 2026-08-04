import { useRef, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useMovie } from "../hooks/useMovies";
import { movieImageUrl, movieStreamUrl, movieDownloadUrl } from "../api/client";
import { useTranscodeProgress } from "../hooks/useTranscodeProgress";
import { MatchPicker } from "../components/MatchPicker";

function formatRuntime(seconds: number): string {
  const mins = Math.round(seconds / 60);
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return h === 0 ? `${m}m` : `${h}h ${m}m`;
}

export function MovieDetailPage() {
  const { id } = useParams<{ id: string }>();
  const movieId = Number(id);
  const { data: movie, isLoading, error } = useMovie(movieId);
  const [showPlayer, setShowPlayer] = useState(false);
  const [showMatchPicker, setShowMatchPicker] = useState(false);
  const videoRef = useRef<HTMLVideoElement>(null);
  const navigate = useNavigate();
  const progress = useTranscodeProgress(movieId);

  if (isLoading) {
    return <div className="p-8 text-neutral-400">Loading…</div>;
  }
  if (error || !movie) {
    return <div className="p-8 text-red-500">Movie not found.</div>;
  }

  const ready = movie.status === "ready";

  const enterFullscreen = async () => {
    const video = videoRef.current;
    if (!video) return;
    await video.requestFullscreen?.().catch(() => {
      // Fullscreen can be unavailable in an embedded browser; native video
      // controls still expose the platform's fullscreen option where present.
    });
  };

  return (
    <div>
      <div className="relative -mt-16 h-[55vh] min-h-[350px] w-full overflow-hidden sm:-mt-[68px] sm:h-[65vh]">
        {movie.hasBackdrop ? (
          <img
            src={movieImageUrl(movie.id, "backdrop")}
            alt=""
            className="absolute inset-0 h-full w-full object-cover"
          />
        ) : (
          <div className="absolute inset-0 bg-gradient-to-br from-neutral-800 to-black" />
        )}
        <div className="absolute inset-0 bg-gradient-to-t from-[#141414] via-black/45 to-transparent" />
        <div className="absolute inset-0 bg-gradient-to-r from-black/55 to-transparent" />
        <button
          onClick={() => navigate(-1)}
          className="absolute left-4 top-20 rounded-sm bg-black/50 px-3 py-1.5 text-sm text-white transition hover:bg-black/80 sm:left-10 sm:top-24"
        >
          ← Back
        </button>
      </div>

      <div className="relative z-10 mx-auto -mt-40 max-w-6xl px-4 pb-16 sm:-mt-52 sm:px-10">
        <div className="flex flex-col gap-6 sm:flex-row">
          <div className="hidden w-40 shrink-0 overflow-hidden rounded-sm bg-neutral-800 shadow-2xl sm:block sm:w-56">
            <div className="aspect-[2/3] w-full">
              {movie.hasPoster ? (
                <img
                  src={movieImageUrl(movie.id, "primary")}
                  alt={movie.title}
                  className="h-full w-full object-cover"
                />
              ) : (
                <div className="flex h-full w-full items-center justify-center px-2 text-center text-sm text-neutral-500">
                  {movie.title}
                </div>
              )}
            </div>
          </div>

          <div className="flex-1 pt-2">
            <h1 className="text-4xl font-extrabold tracking-tight sm:text-5xl">{movie.title}</h1>
            <div className="mt-3 flex flex-wrap items-center gap-3 text-sm font-medium text-neutral-300">
              {movie.year > 0 && <span>{movie.year}</span>}
              {movie.runtimeSeconds > 0 && <span>{formatRuntime(movie.runtimeSeconds)}</span>}
              {movie.genres.length > 0 && <span>{movie.genres.join(", ")}</span>}
            </div>

            {!ready && (
              <div className="mt-5 rounded-sm bg-neutral-900/90 px-4 py-3 text-sm text-neutral-300">
                {movie.status === "transcoding"
                  ? `Transcoding… ${(progress ?? 0).toFixed(0)}%`
                  : movie.status === "needs_transcode"
                    ? "Queued for transcode"
                    : movie.status === "error"
                      ? "This file failed to process."
                      : "Matching metadata…"}
              </div>
            )}

            {(!movie.tmdbId || movie.needsReview) && (
              <div className="mt-5 flex items-center justify-between gap-4 rounded-sm bg-neutral-900/90 px-4 py-3 text-sm text-neutral-300">
                <span>
                  {movie.tmdbId
                    ? "This match might not be right."
                    : "No metadata match was found for this file."}
                </span>
                <button
                  onClick={() => setShowMatchPicker(true)}
                  className="shrink-0 rounded-sm border border-neutral-500 px-3 py-1.5 text-neutral-200 transition hover:border-white hover:text-white"
                >
                  Search TMDB
                </button>
              </div>
            )}

            {movie.overview && <p className="mt-5 max-w-2xl leading-relaxed text-neutral-200">{movie.overview}</p>}

            {ready && (
              <div className="mt-7 flex gap-3">
                <button
                  onClick={() => setShowPlayer(true)}
                  className="rounded-sm bg-white px-6 py-2.5 font-bold text-black transition hover:bg-white/80"
                >
                  ▶ Play
                </button>
                <a
                  href={movieDownloadUrl(movie.id)}
                  className="rounded-sm bg-neutral-500/70 px-6 py-2.5 font-bold text-white transition hover:bg-neutral-500/50"
                >
                  Download
                </a>
              </div>
            )}

            {movie.cast.length > 0 && (
              <div className="mt-8">
                <h2 className="mb-2 text-sm font-semibold uppercase tracking-wide text-neutral-400">
                  Cast
                </h2>
                <p className="text-neutral-300">{movie.cast.slice(0, 8).map((c) => c.name).join(", ")}</p>
              </div>
            )}
          </div>
        </div>
      </div>

      {showPlayer && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black p-4 sm:p-8">
          <div className="absolute inset-x-4 top-4 z-10 flex items-center justify-between sm:inset-x-8 sm:top-6">
            <p className="max-w-[65%] truncate text-sm font-medium text-white">{movie.title}</p>
            <div className="flex items-center gap-2">
              <button
                onClick={enterFullscreen}
                className="rounded-sm bg-white/10 px-3 py-1.5 text-sm text-white transition hover:bg-white/20"
              >
                ⛶ Full screen
              </button>
              <button
                onClick={() => setShowPlayer(false)}
                className="rounded-sm bg-white/10 px-3 py-1.5 text-sm text-white transition hover:bg-white/20"
              >
                ✕ Close
              </button>
            </div>
          </div>
          {/* eslint-disable-next-line jsx-a11y/media-has-caption */}
          <video
            ref={videoRef}
            key={movie.id}
            src={movieStreamUrl(movie.id)}
            controls
            autoPlay
            className="max-h-full max-w-full rounded-sm shadow-2xl"
          />
        </div>
      )}

      {showMatchPicker && (
        <MatchPicker
          movieId={movie.id}
          initialQuery={movie.title}
          initialYear={movie.year || undefined}
          onClose={() => setShowMatchPicker(false)}
        />
      )}
    </div>
  );
}
