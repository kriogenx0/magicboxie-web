import { useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useMovie } from "../hooks/useMovies";
import { movieImageUrl, movieStreamUrl, movieDownloadUrl } from "../api/client";
import { useTranscodeProgress } from "../hooks/useTranscodeProgress";

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
  const navigate = useNavigate();
  const progress = useTranscodeProgress(movieId);

  if (isLoading) {
    return <div className="p-8 text-neutral-400">Loading…</div>;
  }
  if (error || !movie) {
    return <div className="p-8 text-red-500">Movie not found.</div>;
  }

  const ready = movie.status === "ready";

  return (
    <div>
      <div className="relative h-[40vh] min-h-[280px] w-full overflow-hidden">
        {movie.hasBackdrop ? (
          <img
            src={movieImageUrl(movie.id, "backdrop")}
            alt=""
            className="absolute inset-0 h-full w-full object-cover"
          />
        ) : (
          <div className="absolute inset-0 bg-gradient-to-br from-neutral-800 to-black" />
        )}
        <div className="absolute inset-0 bg-gradient-to-t from-black via-black/50 to-transparent" />
        <button
          onClick={() => navigate(-1)}
          className="absolute left-4 top-4 rounded bg-black/60 px-3 py-1.5 text-sm text-white transition hover:bg-black/80 sm:left-8 sm:top-6"
        >
          ← Back
        </button>
      </div>

      <div className="relative z-10 mx-auto -mt-24 max-w-4xl px-4 pb-16 sm:px-8">
        <div className="flex flex-col gap-6 sm:flex-row">
          <div className="w-40 shrink-0 overflow-hidden rounded-md bg-neutral-800 shadow-xl sm:w-56">
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
            <h1 className="text-3xl font-bold sm:text-4xl">{movie.title}</h1>
            <div className="mt-2 flex flex-wrap items-center gap-3 text-sm text-neutral-400">
              {movie.year > 0 && <span>{movie.year}</span>}
              {movie.runtimeSeconds > 0 && <span>{formatRuntime(movie.runtimeSeconds)}</span>}
              {movie.genres.length > 0 && <span>{movie.genres.join(", ")}</span>}
            </div>

            {!ready && (
              <div className="mt-4 rounded bg-neutral-900 px-4 py-3 text-sm text-neutral-300">
                {movie.status === "transcoding"
                  ? `Transcoding… ${(progress ?? 0).toFixed(0)}%`
                  : movie.status === "needs_transcode"
                    ? "Queued for transcode"
                    : movie.status === "error"
                      ? "This file failed to process."
                      : "Matching metadata…"}
              </div>
            )}

            {movie.overview && <p className="mt-4 max-w-2xl text-neutral-200">{movie.overview}</p>}

            {ready && (
              <div className="mt-6 flex gap-3">
                <button
                  onClick={() => setShowPlayer(true)}
                  className="rounded bg-red-600 px-6 py-2 font-semibold text-white transition hover:bg-red-700"
                >
                  ▶ Play
                </button>
                <a
                  href={movieDownloadUrl(movie.id)}
                  className="rounded border border-neutral-600 px-6 py-2 font-semibold text-neutral-200 transition hover:border-neutral-400"
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
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/95 p-4">
          <button
            onClick={() => setShowPlayer(false)}
            className="absolute right-4 top-4 rounded bg-white/10 px-3 py-1.5 text-white transition hover:bg-white/20"
          >
            ✕ Close
          </button>
          {/* eslint-disable-next-line jsx-a11y/media-has-caption */}
          <video key={movie.id} src={movieStreamUrl(movie.id)} controls autoPlay className="max-h-full max-w-full" />
        </div>
      )}
    </div>
  );
}
