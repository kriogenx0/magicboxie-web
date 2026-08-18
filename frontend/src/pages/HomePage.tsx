import { useMemo } from "react";
import { useMovies } from "../hooks/useMovies";
import { Hero } from "../components/Hero";
import { PosterRow } from "../components/PosterRow";
import type { Movie } from "../api/types";

export function HomePage() {
  const { data: movies, isLoading, error } = useMovies();

  const { heroMovies, recentlyAdded, inProgress, byGenre } = useMemo(() => {
    const all = movies ?? [];
    const ready = all.filter((m) => m.status === "ready");
    const inProgress = all.filter((m) =>
      ["pending", "probing", "needs_transcode", "transcoding"].includes(m.status),
    );
    const recentlyAdded = [...all].sort(
      (a, b) => new Date(b.addedAt).getTime() - new Date(a.addedAt).getTime(),
    );

    const genreMap = new Map<string, Movie[]>();
    for (const m of ready) {
      for (const genre of m.genres.length > 0 ? m.genres : ["Unsorted"]) {
        const bucket = genreMap.get(genre) ?? [];
        bucket.push(m);
        genreMap.set(genre, bucket);
      }
    }

    return {
      heroMovies: ready.slice(0, 5),
      recentlyAdded,
      inProgress,
      byGenre: Array.from(genreMap.entries()),
    };
  }, [movies]);

  if (isLoading) {
    return <div className="p-8 text-neutral-400">Loading your library…</div>;
  }
  if (error) {
    return <div className="p-8 text-red-500">Failed to load movies.</div>;
  }
  if ((movies ?? []).length === 0) {
    return (
      <div className="p-8 text-center text-neutral-400">
        No movies yet. Use "Add Media" above to upload your first one.
      </div>
    );
  }

  return (
    <div>
      <Hero movies={heroMovies} />
      <div className="relative z-10 -mt-2 sm:-mt-6">
        <PosterRow title="Continue Processing" movies={inProgress} />
        <PosterRow title="Recently Added" movies={recentlyAdded} />
        {byGenre.map(([label, list]) => (
          <PosterRow key={label} title={label} movies={list} />
        ))}
      </div>
    </div>
  );
}
