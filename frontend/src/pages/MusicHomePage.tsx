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
    <div className="px-4 py-8 sm:px-10 sm:py-10">
      <p className="mb-1 text-xs font-bold uppercase tracking-[0.2em] text-neutral-500">Your collection</p>
      <h1 className="mb-6 text-2xl font-bold tracking-tight text-neutral-100">Music</h1>
      <div className="flex flex-wrap gap-4 sm:gap-5">
        {artists!.map((artist) => (
          <MusicTile key={artist.id} to={`/music/artists/${artist.id}`} title={artist.name} round />
        ))}
      </div>
    </div>
  );
}
