import { useParams, useNavigate } from "react-router-dom";
import { useArtist, useAlbumsByArtist } from "../hooks/useMusic";
import { MusicTile } from "../components/MusicTile";
import { albumImageUrl } from "../api/client";

export function ArtistPage() {
  const { id } = useParams<{ id: string }>();
  const artistId = Number(id);
  const { data: artist } = useArtist(artistId);
  const { data: albums, isLoading, error } = useAlbumsByArtist(artistId);
  const navigate = useNavigate();

  return (
    <div className="px-4 py-6 sm:px-8">
      <button
        onClick={() => navigate(-1)}
        className="mb-4 rounded-sm bg-neutral-900 px-3 py-1.5 text-sm text-neutral-200 hover:bg-neutral-800"
      >
        ← Back
      </button>
      <h1 className="mb-4 text-2xl font-extrabold tracking-tight text-neutral-100 sm:text-3xl">
        {artist?.name ?? "…"}
      </h1>

      {isLoading && <p className="text-neutral-400">Loading albums…</p>}
      {error && <p className="text-red-500">Failed to load albums.</p>}

      <div className="flex flex-wrap gap-4 sm:gap-5">
        {(albums ?? []).map((album) => (
          <MusicTile
            key={album.id}
            to={`/music/albums/${album.id}`}
            title={album.title}
            subtitle={album.year > 0 ? String(album.year) : undefined}
            imageUrl={album.hasCover ? albumImageUrl(album.id) : undefined}
          />
        ))}
      </div>
    </div>
  );
}
