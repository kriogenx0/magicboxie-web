import { useArtists } from "../hooks/useMusic";
import { MusicTile } from "../components/MusicTile";

export function MusicHomePage() {
  const { data: artists, isLoading, error } = useArtists();

  if (isLoading) {
    return <div className="p-8 text-neutral-400">Loading your music…</div>;
  }
  if (error) {
    return <div className="p-8 text-red-500">Failed to load artists.</div>;
  }
  if ((artists ?? []).length === 0) {
    return (
      <div className="p-8 text-center text-neutral-400">
        No music yet. Use "Add Media" above to upload some tracks.
      </div>
    );
  }

  return (
    <div className="px-4 py-6 sm:px-8">
      <h1 className="mb-4 text-lg font-semibold text-neutral-100">Artists</h1>
      <div className="flex flex-wrap gap-5">
        {artists!.map((artist) => (
          <MusicTile key={artist.id} to={`/music/artists/${artist.id}`} title={artist.name} round />
        ))}
      </div>
    </div>
  );
}
