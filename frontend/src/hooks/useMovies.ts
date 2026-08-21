import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiFetch } from "../api/client";
import {
  mapItemToMovie,
  type ItemsResponse,
  type JellyfinItem,
  type Movie,
  type TMDBSearchResult,
  type ThumbnailCandidate,
  type BackgroundJob,
  type Device,
} from "../api/types";

// "1" is the fixed single-user id -- MagicBoxie has one shared login, not
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

export function useBackgroundJobs() {
  return useQuery({
    queryKey: ["backgroundJobs"],
    queryFn: async () => (await apiFetch<{ jobs: BackgroundJob[] }>("/api/jobs")).jobs,
    refetchInterval: 3000,
  });
}

// Every magicboxie-device Pi that has ever checked in, and when it last did
// - see AdminDevicesPage.
export function useDevices() {
  return useQuery({
    queryKey: ["devices"],
    queryFn: async () => (await apiFetch<{ devices: Device[] }>("/api/devices")).devices,
    refetchInterval: 15000,
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

export function useRenameMovie(movieId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (title: string) =>
      apiFetch<JellyfinItem>(`/api/items/movie-${movieId}`, {
        method: "PATCH",
        body: JSON.stringify({ title }),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["movies"] }),
  });
}

export function useDeleteMovie() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (movieId: number) =>
      apiFetch<void>(`/api/items/movie-${movieId}`, { method: "DELETE" }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["movies"] }),
  });
}

// Looks up TMDB candidates by title (+ optional year) so the user can
// manually correct a movie whose automatic match was missing or wrong.
// Not a useQuery -- this only runs when the user explicitly searches, not on
// mount, so it's a plain async function the picker UI calls itself.
export async function searchTMDB(query: string, year?: number): Promise<TMDBSearchResult[]> {
  const params = new URLSearchParams({ query });
  if (year) params.set("year", String(year));
  const res = await apiFetch<{ results: TMDBSearchResult[] }>(`/api/items/search?${params}`);
  return res.results;
}

export function useApplyMatch(movieId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (tmdbId: number) =>
      apiFetch(`/api/items/movie-${movieId}/match`, {
        method: "POST",
        body: JSON.stringify({ tmdb_id: tmdbId }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["movies"] });
    },
  });
}

export async function getThumbnailCandidates(movieId: number): Promise<ThumbnailCandidate[]> {
  const res = await apiFetch<{ candidates: ThumbnailCandidate[] }>(
    `/api/items/movie-${movieId}/thumbnails`,
  );
  return res.candidates;
}

export function useSelectThumbnail(movieId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (index: number) =>
      apiFetch(`/api/items/movie-${movieId}/thumbnails/select`, {
        method: "POST",
        body: JSON.stringify({ index }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["movies"] });
    },
  });
}

// Marks/unmarks a movie as available to magicboxie-device Pis, which only
// download items with syncEnabled set (see POST /devices/register).
export function useSetDeviceSync(movieId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (enabled: boolean) =>
      apiFetch(`/api/items/movie-${movieId}/sync`, {
        method: "POST",
        body: JSON.stringify({ enabled }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["movies"] });
    },
  });
}
