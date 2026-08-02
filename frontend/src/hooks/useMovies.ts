import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../api/client";
import {
  mapItemToMovie,
  type ItemsResponse,
  type JellyfinItem,
  type Movie,
  type TMDBSearchResult,
} from "../api/types";

// "1" is the fixed single-user id -- MagicBox has one shared login, not
// per-user accounts, so this is never actually looked up server-side.
const USER_ID = "1";

export function useMovies() {
  return useQuery({
    queryKey: ["movies"],
    queryFn: async () => {
      const res = await apiFetch<ItemsResponse>(
        `/Users/${USER_ID}/Items?IncludeItemTypes=Movie&Recursive=true`,
      );
      return res.Items.map(mapItemToMovie);
    },
  });
}

export function useMovie(id: number) {
  return useQuery<Movie>({
    queryKey: ["movies", id],
    queryFn: async () => {
      const item = await apiFetch<JellyfinItem>(`/Users/${USER_ID}/Items/movie-${id}`);
      return mapItemToMovie(item);
    },
    enabled: Number.isFinite(id),
  });
}

// Looks up TMDB candidates by title (+ optional year) so the user can
// manually correct a movie whose automatic match was missing or wrong.
// Not a useQuery -- this only runs when the user explicitly searches, not on
// mount, so it's a plain async function the picker UI calls itself.
export async function searchTMDB(query: string, year?: number): Promise<TMDBSearchResult[]> {
  const params = new URLSearchParams({ query });
  if (year) params.set("year", String(year));
  const res = await apiFetch<{ results: TMDBSearchResult[] }>(`/magicbox/items/search?${params}`);
  return res.results;
}

export function useApplyMatch(movieId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (tmdbId: number) =>
      apiFetch(`/magicbox/items/movie-${movieId}/match`, {
        method: "POST",
        body: JSON.stringify({ tmdb_id: tmdbId }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["movies"] });
    },
  });
}
