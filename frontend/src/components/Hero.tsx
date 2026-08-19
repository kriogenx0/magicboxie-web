import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import type { Movie } from "../api/types";
import { movieImageUrl, moviePreviewUrl } from "../api/client";

const ROTATE_INTERVAL_MS = 8000;

export function Hero({ movies, selectedMovie }: { movies: Movie[]; selectedMovie?: Movie | null }) {
  const [index, setIndex] = useState(0);
  const [previewReady, setPreviewReady] = useState(false);

  useEffect(() => {
    if (movies.length <= 1) return;
    const id = setInterval(() => setIndex((i) => (i + 1) % movies.length), ROTATE_INTERVAL_MS);
    return () => clearInterval(id);
  }, [movies.length]);

  const movie = selectedMovie ?? (movies.length > 0 ? movies[index % movies.length] : null);

  useEffect(() => setPreviewReady(false), [movie?.id]);

  if (!movie) return null;

  return (
    <section className="relative h-[58vh] min-h-[440px] w-full overflow-hidden sm:h-[72vh]">
      {movie.hasBackdrop ? (
        <img
          key={movie.id}
          src={movieImageUrl(movie.id, "backdrop", movie.backdropTag)}
          alt=""
          className="absolute inset-0 h-full w-full object-cover object-[62%_center] transition-opacity duration-700"
        />
      ) : (
        <div className="absolute inset-0 bg-gradient-to-br from-neutral-800 to-black" />
      )}
      {movie.status === "ready" && (
        <video
          key={movie.id}
          src={moviePreviewUrl(movie.id)}
          muted
          autoPlay
          loop
          playsInline
          onCanPlay={() => setPreviewReady(true)}
          className={`absolute inset-0 h-full w-full object-cover object-[62%_center] transition-opacity duration-700 ${previewReady ? "opacity-100" : "opacity-0"}`}
        />
      )}
      <div className="absolute inset-0 bg-gradient-to-t from-[#141414] via-[#141414]/10 to-transparent" />
      <div className="absolute inset-0 bg-gradient-to-r from-black via-black/55 to-transparent" />

      <div className="absolute bottom-16 left-0 max-w-2xl px-4 sm:bottom-24 sm:px-10 lg:px-14">
        <h1 className="mb-4 text-4xl font-extrabold tracking-tight drop-shadow-lg sm:text-6xl lg:text-7xl">{movie.title}</h1>
        {movie.overview && (
          <p className="mb-6 line-clamp-3 max-w-xl text-sm leading-relaxed text-neutral-100 drop-shadow sm:text-base lg:text-lg">
            {movie.overview}
          </p>
        )}
        <div className="flex gap-3">
          <Link to={`/movies/${movie.id}`} className="inline-flex items-center gap-2 rounded bg-white px-6 py-3 text-sm font-bold text-black transition hover:bg-white/80">
            <span className="text-base leading-none">▶</span> Play
          </Link>
          <Link to={`/movies/${movie.id}`} className="inline-flex items-center gap-2 rounded bg-neutral-500/70 px-6 py-3 text-sm font-bold text-white backdrop-blur-sm transition hover:bg-neutral-500/50">
            <span className="text-base leading-none">ⓘ</span> More Info
          </Link>
        </div>
      </div>
    </section>
  );
}
