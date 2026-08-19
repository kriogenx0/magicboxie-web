import { Link } from "react-router-dom";
import type { BackgroundJob } from "../api/types";
import { useBackgroundJobs } from "../hooks/useMovies";

function time(value?: string) {
  return value ? new Date(value).toLocaleString() : "—";
}

function statusStyle(status: BackgroundJob["status"]) {
  if (status === "running") return "bg-blue-500/20 text-blue-200";
  if (status === "queued") return "bg-amber-500/20 text-amber-200";
  if (status === "failed") return "bg-red-500/20 text-red-200";
  return "bg-emerald-500/20 text-emerald-200";
}

export function AdminActivityPage() {
  const { data: jobs = [], isLoading, error } = useBackgroundJobs();
  const active = jobs.filter((job) => job.status === "running" || job.status === "queued");

  return (
    <section className="mx-auto max-w-6xl px-4 py-10 sm:px-10">
      <div className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="mb-2 text-sm font-semibold uppercase tracking-[0.2em] text-[#e50914]">Admin</p>
          <h1 className="text-3xl font-bold text-white">Background activity</h1>
          <p className="mt-2 text-sm text-neutral-400">Updates automatically every few seconds.</p>
        </div>
        <Link to="/admin/movies" className="rounded border border-neutral-600 px-4 py-2 text-sm text-neutral-200 hover:border-white">
          Manage movies
        </Link>
      </div>

      <div className="mb-6 grid gap-3 sm:grid-cols-3">
        <div className="rounded bg-neutral-900 p-4 ring-1 ring-white/10"><div className="text-2xl font-bold">{active.length}</div><div className="text-sm text-neutral-400">Active or queued</div></div>
        <div className="rounded bg-neutral-900 p-4 ring-1 ring-white/10"><div className="text-2xl font-bold">{jobs.filter((j) => j.status === "completed").length}</div><div className="text-sm text-neutral-400">Recently completed</div></div>
        <div className="rounded bg-neutral-900 p-4 ring-1 ring-white/10"><div className="text-2xl font-bold">{jobs.filter((j) => j.status === "failed").length}</div><div className="text-sm text-neutral-400">Recent failures</div></div>
      </div>

      {isLoading && <p className="text-neutral-400">Loading activity…</p>}
      {error && <p className="rounded bg-red-950 p-4 text-red-200">Could not load background activity.</p>}
      {!isLoading && !error && jobs.length === 0 && <p className="rounded bg-neutral-900 p-6 text-neutral-400">No background jobs have run yet.</p>}

      <div className="space-y-3">
        {jobs.map((job) => (
          <article key={job.id} className="rounded bg-neutral-900 p-4 ring-1 ring-white/10">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <Link to={`/movies/${job.movie_id}`} className="font-semibold text-white hover:underline">{job.movie_title}</Link>
                <p className="mt-1 text-xs text-neutral-500">{job.type.replaceAll("_", " ")} · Job #{job.id} · queued {time(job.created_at)}</p>
              </div>
              <span className={`rounded-full px-3 py-1 text-xs font-semibold capitalize ${statusStyle(job.status)}`}>{job.status}</span>
            </div>
            {(job.status === "running" || job.progress_percent > 0) && (
              <div className="mt-4">
                <div className="mb-1 flex justify-between text-xs text-neutral-400"><span>Progress</span><span>{job.progress_percent.toFixed(0)}%</span></div>
                <div className="h-2 overflow-hidden rounded-full bg-neutral-700"><div className="h-full bg-[#e50914] transition-all" style={{ width: `${Math.max(0, Math.min(100, job.progress_percent))}%` }} /></div>
              </div>
            )}
            <div className="mt-3 flex flex-wrap gap-x-6 gap-y-1 text-xs text-neutral-500">
              <span>Started: {time(job.started_at)}</span><span>Finished: {time(job.finished_at)}</span>
            </div>
            {job.status === "failed" && job.log_tail && <pre className="mt-3 max-h-36 overflow-auto whitespace-pre-wrap rounded bg-black p-3 text-xs text-red-200">{job.log_tail}</pre>}
          </article>
        ))}
      </div>
    </section>
  );
}
