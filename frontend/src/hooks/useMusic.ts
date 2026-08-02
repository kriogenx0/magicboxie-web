import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "../api/client";
import {
  mapItemToAlbum,
  mapItemToArtist,
  mapItemToTrack,
  type Album,
  type Artist,
  type ItemsResponse,
  type Track,
} from "../api/types";

const USER_ID = "1"; // MagicBox has one shared login, not per-user accounts

export function useArtists() {
  return useQuery({
    queryKey: ["artists"],
    queryFn: async () => {
      const res = await apiFetch<ItemsResponse>(`/Users/${USER_ID}/Items?IncludeItemTypes=MusicArtist`);
      return res.Items.map(mapItemToArtist);
    },
  });
}

export function useAlbumsByArtist(artistId: number) {
  return useQuery({
    queryKey: ["albums", "artist", artistId],
    queryFn: async () => {
      const res = await apiFetch<ItemsResponse>(
        `/Users/${USER_ID}/Items?IncludeItemTypes=MusicAlbum&ParentId=artist-${artistId}`,
      );
      return res.Items.map(mapItemToAlbum);
    },
    enabled: Number.isFinite(artistId),
  });
}

export function useArtist(artistId: number) {
  return useQuery<Artist>({
    queryKey: ["artists", artistId],
    queryFn: async () => {
      const item = await apiFetch(`/Users/${USER_ID}/Items/artist-${artistId}`);
      return mapItemToArtist(item as Parameters<typeof mapItemToArtist>[0]);
    },
    enabled: Number.isFinite(artistId),
  });
}

export function useAlbum(albumId: number) {
  return useQuery<Album>({
    queryKey: ["albums", albumId],
    queryFn: async () => {
      const item = await apiFetch(`/Users/${USER_ID}/Items/album-${albumId}`);
      return mapItemToAlbum(item as Parameters<typeof mapItemToAlbum>[0]);
    },
    enabled: Number.isFinite(albumId),
  });
}

export function useTracksByAlbum(albumId: number) {
  return useQuery({
    queryKey: ["tracks", "album", albumId],
    queryFn: async () => {
      const res = await apiFetch<ItemsResponse>(
        `/Users/${USER_ID}/Items?IncludeItemTypes=Audio&ParentId=album-${albumId}`,
      );
      return res.Items.map(mapItemToTrack);
    },
    enabled: Number.isFinite(albumId),
  });
}

export type { Artist, Album, Track };
