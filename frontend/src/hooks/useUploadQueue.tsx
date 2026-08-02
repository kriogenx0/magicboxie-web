import { createContext, useCallback, useContext, useState, type ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { uploadFile } from "../api/upload";

export interface UploadItem {
  id: string;
  name: string;
  receivedBytes: number;
  totalBytes: number;
  status: "uploading" | "done" | "error";
  error?: string;
}

interface UploadQueueContextValue {
  uploads: UploadItem[];
  addFiles: (files: File[]) => void;
}

const UploadQueueContext = createContext<UploadQueueContextValue | null>(null);

// Holds the upload queue at the app level (rather than inside the upload
// panel component) so a drop anywhere on the page -- not just onto the
// panel's own box -- lands in the same visible queue.
export function UploadQueueProvider({ children }: { children: ReactNode }) {
  const [uploads, setUploads] = useState<UploadItem[]>([]);
  const queryClient = useQueryClient();

  const addFiles = useCallback(
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
            // The upload could be a movie or a music file -- the server
            // decides via matching, so just refresh every list rather than
            // guessing which query key applies.
            queryClient.invalidateQueries();
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
