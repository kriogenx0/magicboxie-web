const TOKEN_KEY = "magicbox_token";

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string | null) {
  if (token) localStorage.setItem(TOKEN_KEY, token);
  else localStorage.removeItem(TOKEN_KEY);
}

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

// apiFetch hits bare Jellyfin-style paths (e.g. "/Users/1/Items") or
// MagicBox extension paths (e.g. "/api/library/scan") directly -- there
// is no shared "/api/v1" prefix in the Jellyfin-compatible contract.
export async function apiFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers = new Headers(options.headers);
  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (options.body && typeof options.body === "string" && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const res = await fetch(path, { ...options, headers });

  if (res.status === 401) {
    setToken(null);
    if (!window.location.pathname.startsWith("/login")) {
      window.location.href = "/login";
    }
    throw new ApiError(401, "unauthorized");
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}) as { error?: string });
    throw new ApiError(res.status, body.error ?? res.statusText);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

// Item ids are "<kind>-<n>" server-side (e.g. "movie-5", "track-12") since
// movies/artists/albums/tracks each have their own primary key and would
// otherwise collide in Jellyfin's flat item-id space (see
// internal/controllers/items_controller.go). These helpers take the plain
// numeric id the rest of the app works with and add the right prefix.

function authQuery(): string {
  return `api_key=${encodeURIComponent(getToken() ?? "")}`;
}

/** Builds a stream URL for a <video> tag, carrying the token as ?api_key=
 * (Jellyfin's convention for URLs that can't attach headers). Unlike the
 * old bespoke API, this reuses the long-lived session token directly --
 * real Jellyfin does the same, no separate short-lived token needed. */
export function movieStreamUrl(movieId: number): string {
  return `/Videos/movie-${movieId}/stream?static=true&${authQuery()}`;
}

export function movieDownloadUrl(movieId: number): string {
  return `${movieStreamUrl(movieId)}&download=1`;
}

export function audioStreamUrl(trackId: number): string {
  return `/Audio/track-${trackId}/stream?static=true&${authQuery()}`;
}

/** Poster/backdrop URLs are unauthenticated per Jellyfin convention (see
 * backend routes.go), so no token is needed here. */
export function movieImageUrl(movieId: number, kind: "primary" | "backdrop"): string {
  return kind === "primary"
    ? `/Items/movie-${movieId}/Images/Primary`
    : `/Items/movie-${movieId}/Images/Backdrop/0`;
}

export function albumImageUrl(albumId: number): string {
  return `/Items/album-${albumId}/Images/Primary`;
}
