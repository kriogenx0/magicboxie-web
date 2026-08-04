import { useUploadQueue } from "../hooks/useUploadQueue";

export function UploadDropzone() {
  const { uploads, addFiles } = useUploadQueue();

  return (
    <div className="px-4 sm:px-8">
      <label className="mb-4 flex cursor-pointer flex-col items-center rounded-sm border-2 border-dashed border-neutral-700 bg-neutral-900/50 p-8 text-center transition-colors hover:border-neutral-400">
        <input
          type="file"
          multiple
          accept="video/*,audio/*"
          className="hidden"
          onChange={(e) => {
            if (e.target.files && e.target.files.length > 0) {
              addFiles(Array.from(e.target.files));
            }
            e.target.value = "";
          }}
        />
        <p className="text-neutral-300">Click to browse, or drop movies or music anywhere on the page</p>
        <p className="mt-1 text-xs text-neutral-500">
          Large files are fine — uploads resume automatically if the connection drops.
        </p>
      </label>

      {uploads.length > 0 && (
        <ul className="mb-6 space-y-2">
          {uploads.map((u) => {
            const pct = u.totalBytes > 0 ? (u.receivedBytes / u.totalBytes) * 100 : 0;
            return (
              <li key={u.id} className="rounded-sm bg-neutral-900 px-3 py-2 text-sm">
                <div className="mb-1 flex justify-between text-neutral-300">
                  <span className="truncate">{u.name}</span>
                  <span className="ml-2 shrink-0">
                    {u.status === "error" ? "Failed" : u.status === "done" ? "Done" : `${pct.toFixed(0)}%`}
                  </span>
                </div>
                <div className="h-1 w-full overflow-hidden rounded-full bg-neutral-700">
                  <div
                    className={`h-full transition-all duration-300 ${
                      u.status === "error" ? "bg-red-700" : "bg-white"
                    }`}
                    style={{ width: `${Math.min(100, pct)}%` }}
                  />
                </div>
                {u.error && <p className="mt-1 text-xs text-red-400">{u.error}</p>}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
