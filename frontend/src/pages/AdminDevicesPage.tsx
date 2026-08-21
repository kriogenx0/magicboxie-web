import { Link } from "react-router-dom";
import type { Movie } from "../api/types";
import { useDevices, useMovies, useSetDeviceSync } from "../hooks/useMovies";
import { PageLoader } from "../components/PageLoader";

function time(value?: string) {
  return value ? new Date(value).toLocaleString() : "—";
}

function DeviceSyncRow({ movie }: { movie: Movie }) {
  const setSync = useSetDeviceSync(movie.id);
  const onDevice = movie.status === "ready" || movie.status === "needs_transcode";

  return (
    <article className="flex flex-wrap items-center justify-between gap-3 rounded bg-neutral-900 p-4 ring-1 ring-white/10">
      <div>
        <Link to={`/movies/${movie.id}`} className="font-semibold text-white hover:underline">{movie.title}</Link>
        <p className="mt-1 text-xs text-neutral-500">{movie.originalFilename} · added {time(movie.addedAt)}</p>
      </div>
      <div className="flex items-center gap-3">
        <span className={`rounded-full px-3 py-1 text-xs font-semibold capitalize ${onDevice ? "bg-emerald-500/20 text-emerald-200" : "bg-amber-500/20 text-amber-200"}`}>
          {onDevice ? "ready to sync" : movie.status.replaceAll("_", " ")}
        </span>
        <button
          onClick={() => setSync.mutate(false)}
          disabled={setSync.isPending}
          className="rounded border border-neutral-600 px-3 py-1.5 text-xs text-neutral-200 hover:border-white disabled:opacity-50"
        >
          Remove from device
        </button>
      </div>
    </article>
  );
}

export function AdminDevicesPage() {
  const { data: devices = [], isLoading: devicesLoading, error: devicesError } = useDevices();
  const { data: movies = [], isLoading: moviesLoading } = useMovies();
  const syncedMovies = movies.filter((m) => m.syncEnabled);

  if (devicesLoading || moviesLoading) {
    return <PageLoader label="Loading devices" />;
  }

  return (
    <section className="mx-auto max-w-6xl px-4 py-10 sm:px-10">
      <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="mb-2 text-sm font-semibold uppercase tracking-[0.2em] text-[#e50914]">Admin</p>
          <h1 className="text-3xl font-bold text-white">Devices</h1>
          <p className="mt-2 text-sm text-neutral-400">Which Pis have checked in, and everything set to sync to one.</p>
        </div>
        <Link to="/admin/movies" className="rounded border border-neutral-600 px-4 py-2 text-sm text-neutral-200 hover:border-white">
          Manage movies
        </Link>
      </div>

      {devicesError && <p className="mb-6 rounded bg-red-950 p-4 text-red-200">Could not load devices.</p>}

      <h2 className="mb-3 text-lg font-semibold text-white">Check-ins</h2>
      {devices.length === 0 ? (
        <p className="mb-8 rounded bg-neutral-900 p-6 text-neutral-400">No device has checked in yet.</p>
      ) : (
        <div className="mb-8 space-y-3">
          {devices.map((device) => (
            <article key={device.id} className="flex flex-wrap items-center justify-between gap-3 rounded bg-neutral-900 p-4 ring-1 ring-white/10">
              <span className="font-mono text-sm text-white">{device.device_id}</span>
              <span className="text-xs text-neutral-400">Last checked in {time(device.last_seen_at)}</span>
            </article>
          ))}
        </div>
      )}

      <h2 className="mb-3 text-lg font-semibold text-white">Set to sync</h2>
      {syncedMovies.length === 0 ? (
        <p className="rounded bg-neutral-900 p-6 text-neutral-400">No movies are set to sync to a device yet - toggle "Sync to device" from a movie's admin row.</p>
      ) : (
        <div className="space-y-3">
          {syncedMovies.map((movie) => (
            <DeviceSyncRow key={movie.id} movie={movie} />
          ))}
        </div>
      )}
    </section>
  );
}
