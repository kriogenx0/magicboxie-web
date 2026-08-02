import type { Movie } from "../api/types";
import { PosterCard } from "./PosterCard";

export function PosterRow({ title, movies }: { title: string; movies: Movie[] }) {
  if (movies.length === 0) return null;

  return (
    <section className="mb-8">
      <h2 className="mb-3 px-4 text-lg font-semibold text-neutral-100 sm:px-8">{title}</h2>
      <div className="scroll-row flex gap-3 overflow-x-auto px-4 pb-4 sm:px-8">
        {movies.map((movie) => (
          <PosterCard key={movie.id} movie={movie} />
        ))}
      </div>
    </section>
  );
}
