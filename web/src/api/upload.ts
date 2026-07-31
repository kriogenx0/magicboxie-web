import { apiFetch, getToken } from "./client";

const MAX_RETRIES = 5;

interface CreateUploadResponse {
  upload_id: string;
  chunk_size_bytes: number;
}

interface UploadStatusResponse {
  received_bytes: number;
  total_size_bytes: number;
  status: string;
}

async function putChunk(uploadId: string, offset: number, chunk: Blob): Promise<number> {
  const token = getToken();
  const res = await fetch(`/magicbox/uploads/${uploadId}/chunk?offset=${offset}`, {
    method: "PUT",
    headers: {
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      "Content-Type": "application/octet-stream",
    },
    body: chunk,
  });
  if (res.status === 409) {
    // Offset mismatch -- caller should resync via GET and retry.
    const body = (await res.json()) as { received_bytes: number };
    throw new OffsetMismatchError(body.received_bytes);
  }
  if (!res.ok) {
    throw new Error(`chunk upload failed: ${res.status}`);
  }
  const body = (await res.json()) as { received_bytes: number };
  return body.received_bytes;
}

class OffsetMismatchError extends Error {
  receivedBytes: number;
  constructor(receivedBytes: number) {
    super("offset mismatch");
    this.receivedBytes = receivedBytes;
  }
}

function sleep(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Uploads a file in sequential chunks (per the backend's simple
 * offset-based protocol), retrying with backoff on network failure and
 * resyncing via GET on an offset mismatch. Calls onProgress after each
 * successfully-written chunk. Resolves with the completion response.
 */
export async function uploadFile(
  file: File,
  onProgress: (receivedBytes: number, totalBytes: number) => void,
): Promise<{ kind: string; status: string }> {
  const created = await apiFetch<CreateUploadResponse>("/magicbox/uploads", {
    method: "POST",
    body: JSON.stringify({ filename: file.name, size_bytes: file.size }),
  });

  const uploadId = created.upload_id;
  const chunkSize = created.chunk_size_bytes;

  let offset = 0;
  while (offset < file.size) {
    const chunk = file.slice(offset, offset + chunkSize);

    let attempt = 0;
    for (;;) {
      try {
        const received = await putChunk(uploadId, offset, chunk);
        offset = received;
        onProgress(offset, file.size);
        break;
      } catch (err) {
        if (err instanceof OffsetMismatchError) {
          offset = err.receivedBytes;
          break; // retry the loop from the corrected offset
        }
        attempt += 1;
        if (attempt > MAX_RETRIES) throw err;
        await sleep(Math.min(1000 * 2 ** attempt, 15000));

        // Resync in case some bytes did land before the failure.
        const status = await apiFetch<UploadStatusResponse>(`/magicbox/uploads/${uploadId}`);
        offset = status.received_bytes;
      }
    }
  }

  return apiFetch(`/magicbox/uploads/${uploadId}/complete`, { method: "POST" });
}
