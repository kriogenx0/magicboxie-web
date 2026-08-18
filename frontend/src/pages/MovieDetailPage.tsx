import { useEffect, useRef, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { getThumbnailCandidates, useMovie, useSelectThumbnail, useSetDeviceSync } from "../hooks/useMovies";
import { movieImageUrl, movieStreamUrl, movieDownloadUrl } from "../api/client";
import type { ThumbnailCandidate } from "../api/types";
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
  const [posterVersion, setPosterVersion] = useState(0);
  const [generatedBackdropUrl, setGeneratedBackdropUrl] = useState<string | null>(null);
  const [thumbnailCandidates, setThumbnailCandidates] = useState<ThumbnailCandidate[]>([]);
  const [thumbnailError, setThumbnailError] = useState<string | null>(null);
  const videoRef = useRef<HTMLVideoElement>(null);
  const navigate = useNavigate();
  const progress = useTranscodeProgress(movieId);
  const selectThumbnail = useSelectThumbnail(movieId);
  const setDeviceSync = useSetDeviceSync(movieId);

  useEffect(() => {
    let cancelled = false;
    if (!movie?.posterIsGenerated) {
      setGeneratedBackdropUrl(null);
      setThumbnailCandidates([]);
      return;
    }

    getThumbnailCandidates(movie.id)
      .then((candidates) => {
        if (cancelled) return;
        setThumbnailCandidates(candidates);
        setThumbnailError(null);
        // Prefer a later frame for the wide hero, but never reuse the frame
        // currently selected as the poster.
        const backdrop = [...candidates].reverse().find((candidate) => !candidate.selected);
        setGeneratedBackdropUrl(backdrop?.url ?? null);
      })
      .catch(() => {
        if (!cancelled) {
          setGeneratedBackdropUrl(null);
          setThumbnailError("Could not load thumbnail choices.");
        }
      });
    return () => {
      cancelled = true;
    };
  }, [movie?.id, movie?.posterIsGenerated, posterVersion]);

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
      <div className="relative h-[55vh] min-h-[350px] w-full overflow-hidden sm:h-[65vh]">
        {movie.hasBackdrop ? (
          <img
            src={movieImageUrl(movie.id, "backdrop")}
            alt=""
            className="absolute inset-0 h-full w-full object-cover"
          />
        ) : generatedBackdropUrl ? (
          <img
            src={`${generatedBackdropUrl}?v=${posterVersion}`}
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

      <div className="relative z-10 mx-auto -mt-44 max-w-6xl px-4 pb-16 sm:-mt-60 sm:px-10">
        <div className="flex flex-col gap-6 sm:flex-row">
          <div className="hidden w-40 shrink-0 overflow-hidden rounded-sm bg-neutral-800 shadow-2xl sm:block sm:w-56">
            <div className="aspect-[2/3] w-full">
              {movie.hasPoster ? (
                <img
                  src={`${movieImageUrl(movie.id, "primary")}?v=${posterVersion}`}
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

          <div className="flex-1">
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

            {movie.posterIsGenerated && (
              <section className="mt-6">
                <h2 className="text-sm font-semibold uppercase tracking-wide text-neutral-300">
                  Choose thumbnail
                </h2>
                <p className="mt-1 text-sm text-neutral-400">
                  Select a frame to use as the movie poster.
                </p>
                {thumbnailError && <p className="mt-3 text-sm text-red-400">{thumbnailError}</p>}
                {selectThumbnail.isError && (
                  <p className="mt-3 text-sm text-red-400">
                    {selectThumbnail.error instanceof Error
                      ? selectThumbnail.error.message
                      : "Could not select thumbnail."}
                  </p>
                )}
                <div className="mt-3 grid max-w-3xl grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
                  {thumbnailCandidates.map((candidate) => (
                    <button
                      key={candidate.index}
                      onClick={() =>
                        selectThumbnail.mutate(candidate.index, {
                          onSuccess: () => setPosterVersion((version) => version + 1),
                        })
                      }
                      disabled={selectThumbnail.isPending}
                      className={`group overflow-hidden rounded-sm border-2 bg-neutral-800 text-left transition hover:border-white disabled:opacity-50 ${
                        candidate.selected ? "border-[#e50914]" : "border-transparent"
                      }`}
                    >
                      <img
                        src={candidate.url}
                        alt={`Thumbnail option ${candidate.index + 1}`}
                        className="aspect-video w-full object-cover"
                      />
                      <span className="block px-2 py-1.5 text-center text-xs font-medium text-neutral-300 group-hover:text-white">
                        {candidate.selected ? "Current" : "Select"}
                      </span>
                    </button>
                  ))}
                </div>
              </section>
            )}

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
                <button
                  onClick={() => setDeviceSync.mutate(!movie.syncEnabled)}
                  disabled={setDeviceSync.isPending}
                  aria-pressed={movie.syncEnabled}
                  className={`rounded-sm border px-6 py-2.5 font-bold transition disabled:opacity-50 ${
                    movie.syncEnabled
                      ? "border-transparent bg-[#e50914] text-white hover:bg-[#f6121d]"
                      : "border-neutral-500 text-neutral-100 hover:border-white"
                  }`}
                >
                  {movie.syncEnabled ? "✓ Synced to device" : "Sync to device"}
                </button>
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
