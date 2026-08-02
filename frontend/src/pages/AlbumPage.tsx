import { useParams, useNavigate } from "react-router-dom";
import { useAlbum, useTracksByAlbum } from "../hooks/useMusic";
import { usePlayerStore } from "../stores/playerStore";
import { albumImageUrl } from "../api/client";

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m}:${s.toString().padStart(2, "0")}`;
}

export function AlbumPage() {
  const { id } = useParams<{ id: string }>();
  const albumId = Number(id);
  const { data: album } = useAlbum(albumId);
  const { data: tracks, isLoading, error } = useTracksByAlbum(albumId);
  const navigate = useNavigate();
  const playQueue = usePlayerStore((s) => s.playQueue);
  const currentTrackId = usePlayerStore((s) => s.currentTrack()?.id);
  const isPlaying = usePlayerStore((s) => s.isPlaying);

  return (
    <div className="px-4 py-6 sm:px-8">
      <button
        onClick={() => navigate(-1)}
        className="mb-4 rounded bg-neutral-900 px-3 py-1.5 text-sm text-neutral-200 hover:bg-neutral-800"
      >
        ← Back
      </button>

      <div className="mb-6 flex items-end gap-5">
        <div className="h-36 w-36 shrink-0 overflow-hidden rounded-md bg-neutral-800 shadow-lg sm:h-44 sm:w-44">
          {album?.hasCover ? (
            <img src={albumImageUrl(albumId)} alt={album.title} className="h-full w-full object-cover" />
          ) : (
            <div className="flex h-full w-full items-center justify-center px-2 text-center text-sm text-neutral-500">
              {album?.title}
            </div>
          )}
        </div>
        <div>
          <h1 className="text-2xl font-bold text-neutral-100">{album?.title ?? "…"}</h1>
          <p className="text-neutral-400">
            {album?.artistName}
            {album?.year ? ` · ${album.year}` : ""}
          </p>
          {tracks && tracks.length > 0 && (
            <button
              onClick={() => playQueue(tracks, 0)}
              className="mt-3 rounded bg-red-600 px-5 py-2 font-semibold text-white transition hover:bg-red-700"
            >
              ▶ Play Album
            </button>
          )}
        </div>
      </div>

      {isLoading && <p className="text-neutral-400">Loading tracks…</p>}
      {error && <p className="text-red-500">Failed to load tracks.</p>}

      <ul className="max-w-2xl divide-y divide-neutral-800">
        {(tracks ?? []).map((track, i) => {
          const active = track.id === currentTrackId;
          return (
            <li key={track.id}>
              <button
                onClick={() => playQueue(tracks!, i)}
                className={`flex w-full items-center gap-3 px-2 py-2.5 text-left hover:bg-neutral-900 ${active ? "text-red-500" : "text-neutral-200"}`}
              >
                <span className="w-6 shrink-0 text-sm text-neutral-500">
                  {active && isPlaying ? "♪" : track.trackNumber || i + 1}
                </span>
                <span className="min-w-0 flex-1 truncate">{track.title}</span>
                <span className="shrink-0 text-sm text-neutral-500">{formatDuration(track.durationSeconds)}</span>
              </button>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
