import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import type { Movie } from "../api/types";
import { movieImageUrl } from "../api/client";

const ROTATE_INTERVAL_MS = 8000;

export function Hero({ movies }: { movies: Movie[] }) {
  const [index, setIndex] = useState(0);

  useEffect(() => {
    if (movies.length <= 1) return;
    const id = setInterval(() => setIndex((i) => (i + 1) % movies.length), ROTATE_INTERVAL_MS);
    return () => clearInterval(id);
  }, [movies.length]);

  if (movies.length === 0) return null;
  const movie = movies[index % movies.length];

  return (
    <div className="relative h-[46vh] min-h-[320px] w-full overflow-hidden sm:h-[60vh]">
      {movie.hasBackdrop ? (
        <img
          key={movie.id}
          src={movieImageUrl(movie.id, "backdrop")}
          alt=""
          className="absolute inset-0 h-full w-full object-cover transition-opacity duration-700"
        />
      ) : (
        <div className="absolute inset-0 bg-gradient-to-br from-neutral-800 to-black" />
      )}
      <div className="absolute inset-0 bg-gradient-to-t from-black via-black/40 to-transparent" />
      <div className="absolute inset-0 bg-gradient-to-r from-black/80 via-black/10 to-transparent" />

      <div className="absolute bottom-0 left-0 max-w-xl p-4 sm:p-10">
        <h1 className="mb-2 text-3xl font-bold drop-shadow-lg sm:text-5xl">{movie.title}</h1>
        {movie.overview && (
          <p className="mb-4 line-clamp-3 text-sm text-neutral-200 drop-shadow sm:text-base">
            {movie.overview}
          </p>
        )}
        <Link
          to={`/movies/${movie.id}`}
          className="inline-block rounded bg-white px-5 py-2 font-semibold text-black transition hover:bg-neutral-200"
        >
          View details
        </Link>
      </div>
    </div>
  );
}
