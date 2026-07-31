import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../api/client";
import { mapItemToMovie, type ItemsResponse, type JellyfinItem, type Movie } from "../api/types";

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
