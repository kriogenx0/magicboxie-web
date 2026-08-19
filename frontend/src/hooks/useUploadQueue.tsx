import { createContext, useCallback, useContext, useRef, useState, type ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { uploadFile } from "../api/upload";

export interface UploadItem {
  id: string;
  name: string;
  receivedBytes: number;
  totalBytes: number;
  status: "queued" | "hashing" | "uploading" | "done" | "error";
  error?: string;
}

interface UploadQueueContextValue {
  uploads: UploadItem[];
  addFiles: (files: File[]) => void;
}

const UploadQueueContext = createContext<UploadQueueContextValue | null>(null);
let nextUploadId = 0;

function localUploadId() {
  nextUploadId += 1;
  return `upload-${Date.now()}-${nextUploadId}`;
}

// Holds the upload queue at the app level (rather than inside the upload
// panel component) so a drop anywhere on the page -- not just onto the
// panel's own box -- lands in the same visible queue.
export function UploadQueueProvider({ children }: { children: ReactNode }) {
  const [uploads, setUploads] = useState<UploadItem[]>([]);
  const queryClient = useQueryClient();
  const queue = useRef(Promise.resolve());

  const addFiles = useCallback(
    (files: File[]) => {
      for (const file of files) {
        // randomUUID is unavailable in older Safari releases and on some
        // non-secure HTTP origins. This ID is only for matching UI state;
        // the server creates its own UUID for the actual upload session.
        const id = localUploadId();
        setUploads((prev) => [
          ...prev,
          { id, name: file.name, receivedBytes: 0, totalBytes: file.size, status: "queued" },
        ]);

        // Browsers limit simultaneous HTTP/1.1 requests per host. Starting a
        // large batch in parallel can leave every transfer stuck after its
        // first chunk, particularly while the SSE connection is open. Keep a
        // single active upload; the server starts import/transcode as soon as
        // each file completes.
        queue.current = queue.current.then(async () => {
          setUploads((prev) =>
            prev.map((u) => (u.id === id ? { ...u, status: "uploading" } : u)),
          );
          try {
            await uploadFile(file, (receivedBytes, totalBytes, phase) => {
              setUploads((prev) =>
                prev.map((u) =>
                  u.id === id ? { ...u, receivedBytes, totalBytes, status: phase } : u,
                ),
              );
            });
            setUploads((prev) =>
              prev.map((u) => (u.id === id ? { ...u, status: "done" } : u)),
            );
            queryClient.invalidateQueries();
          } catch (err) {
            const message = err instanceof Error ? err.message : "upload failed";
            setUploads((prev) =>
              prev.map((u) => (u.id === id ? { ...u, status: "error", error: message } : u)),
            );
          }
        });
      }
    },
    [queryClient],
  );

  return (
    <UploadQueueContext.Provider value={{ uploads, addFiles }}>
      {children}
    </UploadQueueContext.Provider>
  );
}

export function useUploadQueue() {
  const ctx = useContext(UploadQueueContext);
  if (!ctx) throw new Error("useUploadQueue must be used within UploadQueueProvider");
  return ctx;
}
