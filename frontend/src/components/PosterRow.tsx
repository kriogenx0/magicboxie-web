import type { Movie } from "../api/types";
import { PosterCard } from "./PosterCard";

export function PosterRow({ title, movies }: { title: string; movies: Movie[] }) {
  if (movies.length === 0) return null;

  return (
    <section className="mb-10 sm:mb-12">
      <h2 className="mb-2 px-4 text-lg font-bold tracking-tight text-neutral-100 sm:px-10 sm:text-xl">{title}</h2>
      <div className="grid grid-cols-3 gap-x-2 gap-y-5 px-4 pt-1 sm:grid-cols-4 sm:gap-x-2.5 sm:gap-y-6 sm:px-10 md:grid-cols-5 lg:grid-cols-6 xl:grid-cols-7 2xl:grid-cols-8">
        {movies.map((movie) => (
          <PosterCard key={movie.id} movie={movie} />
        ))}
      </div>
    </section>
  );
}
