import { Link } from "react-router-dom";
import type { Movie } from "../api/types";
import { movieImageUrl } from "../api/client";
import { useTranscodeProgress } from "../hooks/useTranscodeProgress";

function StatusBadge({ movie }: { movie: Movie }) {
  const progress = useTranscodeProgress(movie.id);

  if (movie.status === "pending" || movie.status === "probing") {
    return (
      <div className="absolute inset-x-0 bottom-0 bg-black/80 px-2 py-1 text-xs text-neutral-200">
        Matching metadata…
      </div>
    );
  }
  if (movie.status === "needs_transcode") {
    return (
      <div className="absolute inset-x-0 bottom-0 bg-black/80 px-2 py-1 text-xs text-neutral-200">
        Queued for transcode
      </div>
    );
  }
  if (movie.status === "transcoding") {
    const pct = progress ?? 0;
    return (
      <div className="absolute inset-x-0 bottom-0 bg-black/80 px-2 py-1">
        <div className="mb-1 text-xs text-neutral-200">Transcoding {pct.toFixed(0)}%</div>
        <div className="h-1 w-full overflow-hidden rounded-full bg-neutral-700">
          <div
            className="h-full bg-red-600 transition-all duration-500"
            style={{ width: `${Math.min(100, Math.max(0, pct))}%` }}
          />
        </div>
      </div>
    );
  }
  if (movie.status === "error") {
    return (
      <div className="absolute inset-x-0 bottom-0 bg-red-900/90 px-2 py-1 text-xs text-red-100">
        Failed
      </div>
    );
  }
  return null;
}

export function PosterCard({ movie }: { movie: Movie }) {
  return (
    <Link
      to={`/movies/${movie.id}`}
      className="group relative block w-40 shrink-0 overflow-hidden rounded-md bg-neutral-900 shadow-lg transition-transform duration-200 ease-out hover:z-10 hover:scale-105 focus:z-10 focus:scale-105 focus:outline-none focus:ring-2 focus:ring-red-600 sm:w-48"
    >
      <div className="aspect-[2/3] w-full bg-neutral-800">
        {movie.hasPoster ? (
          <img
            src={movieImageUrl(movie.id, "primary")}
            alt={movie.title}
            loading="lazy"
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center px-2 text-center text-sm text-neutral-500">
            {movie.title}
          </div>
        )}
      </div>
      <StatusBadge movie={movie} />
    </Link>
  );
}
