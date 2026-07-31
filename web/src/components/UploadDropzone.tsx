import { useCallback, useState } from "react";
import { useDropzone } from "react-dropzone";
import { useQueryClient } from "@tanstack/react-query";
import { uploadFile } from "../api/upload";

interface UploadItem {
  id: string;
  name: string;
  receivedBytes: number;
  totalBytes: number;
  status: "uploading" | "done" | "error";
  error?: string;
}

export function UploadDropzone() {
  const [uploads, setUploads] = useState<UploadItem[]>([]);
  const queryClient = useQueryClient();

  const onDrop = useCallback(
    (files: File[]) => {
      for (const file of files) {
        const id = `${file.name}-${file.size}-${Date.now()}`;
        setUploads((prev) => [
          ...prev,
          { id, name: file.name, receivedBytes: 0, totalBytes: file.size, status: "uploading" },
        ]);

        uploadFile(file, (receivedBytes, totalBytes) => {
          setUploads((prev) =>
            prev.map((u) => (u.id === id ? { ...u, receivedBytes, totalBytes } : u)),
          );
        })
          .then(() => {
            setUploads((prev) => prev.map((u) => (u.id === id ? { ...u, status: "done" } : u)));
            queryClient.invalidateQueries({ queryKey: ["movies"] });
          })
          .catch((err: Error) => {
            setUploads((prev) =>
              prev.map((u) => (u.id === id ? { ...u, status: "error", error: err.message } : u)),
            );
          });
      }
    },
    [queryClient],
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop });

  return (
    <div className="px-4 sm:px-8">
      <div
        {...getRootProps()}
        className={`mb-4 cursor-pointer rounded-lg border-2 border-dashed p-8 text-center transition-colors ${
          isDragActive
            ? "border-red-600 bg-red-950/20"
            : "border-neutral-700 bg-neutral-900/50 hover:border-neutral-500"
        }`}
      >
        <input {...getInputProps()} />
        <p className="text-neutral-300">
          {isDragActive ? "Drop to upload…" : "Drag movies or music here, or click to browse"}
        </p>
        <p className="mt-1 text-xs text-neutral-500">
          Large files are fine — uploads resume automatically if the connection drops.
        </p>
      </div>

      {uploads.length > 0 && (
        <ul className="mb-6 space-y-2">
          {uploads.map((u) => {
            const pct = u.totalBytes > 0 ? (u.receivedBytes / u.totalBytes) * 100 : 0;
            return (
              <li key={u.id} className="rounded bg-neutral-900 px-3 py-2 text-sm">
                <div className="mb-1 flex justify-between text-neutral-300">
                  <span className="truncate">{u.name}</span>
                  <span className="ml-2 shrink-0">
                    {u.status === "error" ? "Failed" : u.status === "done" ? "Done" : `${pct.toFixed(0)}%`}
                  </span>
                </div>
                <div className="h-1 w-full overflow-hidden rounded-full bg-neutral-700">
                  <div
                    className={`h-full transition-all duration-300 ${
                      u.status === "error" ? "bg-red-700" : "bg-red-600"
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
