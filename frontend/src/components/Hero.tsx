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
    <section className="relative h-[38vh] w-full overflow-hidden sm:h-[44vh]">
      {movie.hasBackdrop ? (
        <img
          key={movie.id}
          src={movieImageUrl(movie.id, "backdrop")}
          alt=""
          className="absolute inset-0 h-full w-full object-cover object-[62%_center] transition-opacity duration-700"
        />
      ) : (
        <div className="absolute inset-0 bg-gradient-to-br from-neutral-800 to-black" />
      )}
      <div className="absolute inset-0 bg-gradient-to-t from-[#141414] via-[#141414]/30 to-transparent" />
      <div className="absolute inset-0 bg-gradient-to-r from-black via-black/45 to-transparent" />

      <div className="absolute bottom-2 left-0 max-w-2xl px-4 sm:bottom-4 sm:px-10 lg:px-14">
        <p className="mb-3 text-xs font-bold uppercase tracking-[0.24em] text-neutral-200/90">MagicBoxie presents</p>
        <h1 className="mb-4 text-4xl font-extrabold tracking-tight drop-shadow-lg sm:text-6xl lg:text-7xl">{movie.title}</h1>
        {movie.overview && (
          <p className="mb-6 line-clamp-3 max-w-xl text-sm leading-relaxed text-neutral-100 drop-shadow sm:text-base lg:text-lg">
            {movie.overview}
          </p>
        )}
        <div className="flex gap-3">
          <Link to={`/movies/${movie.id}`} className="inline-flex items-center gap-2 rounded-sm bg-white px-5 py-2.5 text-sm font-bold text-black transition hover:bg-white/80">
            <span className="text-base leading-none">▶</span> Details
          </Link>
          <Link to={`/movies/${movie.id}`} className="inline-flex items-center gap-2 rounded-sm bg-neutral-500/70 px-5 py-2.5 text-sm font-bold text-white transition hover:bg-neutral-500/50">
            <span className="text-base leading-none">ⓘ</span> More Info
          </Link>
        </div>
      </div>
    </section>
  );
}
