import { useState, type FormEvent } from "react";
import { searchTMDB, useApplyMatch } from "../hooks/useMovies";
import type { TMDBSearchResult } from "../api/types";

interface MatchPickerProps {
  movieId: number;
  initialQuery: string;
  initialYear?: number;
  onClose: () => void;
}

export function MatchPicker({ movieId, initialQuery, initialYear, onClose }: MatchPickerProps) {
  const [query, setQuery] = useState(initialQuery);
  const [year, setYear] = useState(initialYear ? String(initialYear) : "");
  const [results, setResults] = useState<TMDBSearchResult[] | null>(null);
  const [searching, setSearching] = useState(false);
  const [searchError, setSearchError] = useState<string | null>(null);
  const applyMatch = useApplyMatch(movieId);

  const runSearch = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!query.trim()) return;
    setSearching(true);
    setSearchError(null);
    try {
      const parsedYear = year.trim() ? Number(year) : undefined;
      setResults(await searchTMDB(query.trim(), parsedYear));
    } catch (err) {
      setSearchError(err instanceof Error ? err.message : "Search failed");
    } finally {
      setSearching(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/80 p-4 pt-16">
      <div className="w-full max-w-lg rounded-sm bg-neutral-900 p-6 shadow-2xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-white">Find the right match</h2>
          <button onClick={onClose} className="text-neutral-400 hover:text-white">
            ✕
          </button>
        </div>

        <form onSubmit={runSearch} className="mb-4 flex gap-2">
          <input
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Movie title"
            autoFocus
            className="min-w-0 flex-1 rounded-sm border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white placeholder-neutral-500 focus:border-red-600 focus:outline-none"
          />
          <input
            value={year}
            onChange={(e) => setYear(e.target.value)}
            placeholder="Year"
            className="w-20 rounded-sm border border-neutral-700 bg-neutral-800 px-3 py-2 text-sm text-white placeholder-neutral-500 focus:border-red-600 focus:outline-none"
          />
          <button
            type="submit"
            disabled={searching || !query.trim()}
            className="shrink-0 rounded-sm bg-[#e50914] px-4 py-2 text-sm font-semibold text-white transition hover:bg-[#f6121d] disabled:opacity-50"
          >
            {searching ? "Searching…" : "Search"}
          </button>
        </form>

        {searchError && <p className="mb-4 text-sm text-red-400">{searchError}</p>}
        {applyMatch.isError && (
          <p className="mb-4 text-sm text-red-400">
            {applyMatch.error instanceof Error ? applyMatch.error.message : "Failed to apply match"}
          </p>
        )}

        {results !== null && results.length === 0 && !searchError && (
          <p className="text-sm text-neutral-400">No results.</p>
        )}

        {results && results.length > 0 && (
          <ul className="max-h-96 space-y-2 overflow-y-auto">
            {results.map((r) => (
              <li key={r.tmdb_id} className="flex gap-3 rounded-sm bg-neutral-800 p-2">
                <div className="h-24 w-16 shrink-0 overflow-hidden rounded-sm bg-neutral-700">
                  {r.poster_url && <img src={r.poster_url} alt="" className="h-full w-full object-cover" />}
                </div>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-white">
                    {r.title} {r.year ? <span className="text-neutral-400">({r.year})</span> : null}
                  </p>
                  {r.overview && <p className="mt-1 line-clamp-2 text-xs text-neutral-400">{r.overview}</p>}
                  <button
                    onClick={() => applyMatch.mutate(r.tmdb_id, { onSuccess: onClose })}
                    disabled={applyMatch.isPending}
                    className="mt-2 rounded-sm border border-neutral-500 px-3 py-1 text-xs text-neutral-200 transition hover:border-white hover:text-white disabled:opacity-50"
                  >
                    Use this match
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
