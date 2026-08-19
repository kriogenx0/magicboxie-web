import { createSHA256 } from "hash-wasm";
import { apiFetch, getToken } from "./client";

const HASH_CHUNK_SIZE = 4 * 1024 * 1024;

async function hashFile(file: File, onProgress: (done: number, total: number) => void) {
  const hasher = await createSHA256();
  hasher.init();
  for (let offset = 0; offset < file.size; offset += HASH_CHUNK_SIZE) {
    const end = Math.min(offset + HASH_CHUNK_SIZE, file.size);
    hasher.update(new Uint8Array(await file.slice(offset, end).arrayBuffer()));
    onProgress(end, file.size);
  }
  return hasher.digest("hex");
}

/**
 * Streams a file in one request. Native XHR upload progress works reliably in
 * Safari and avoids a WebKit issue where a JavaScript chunk loop stopped after
 * its first successful Blob request.
 */
export function uploadFile(
  file: File,
  onProgress: (receivedBytes: number, totalBytes: number, phase: "hashing" | "uploading") => void,
): Promise<{ kind: string; status: string }> {
  return hashFile(file, (done, total) => onProgress(done, total, "hashing")).then(async (checksum) => {
    const result = await apiFetch<{ exists: boolean }>(`/api/uploads/checksum/${checksum}`);
    if (result.exists) throw new Error("this file has already been uploaded");

    return new Promise((resolve, reject) => {
    const request = new XMLHttpRequest();
    request.open("POST", `/api/uploads/direct?filename=${encodeURIComponent(file.name)}`);
    request.setRequestHeader("Content-Type", "application/octet-stream");
    const token = getToken();
    if (token) request.setRequestHeader("Authorization", `Bearer ${token}`);

    request.upload.onprogress = (event) => {
      onProgress(event.loaded, event.lengthComputable ? event.total : file.size, "uploading");
    };
    request.onload = () => {
      let body: { kind?: string; status?: string; error?: string } = {};
      try {
        body = JSON.parse(request.responseText) as typeof body;
      } catch {
        reject(new Error(`upload returned invalid JSON (${request.status})`));
        return;
      }
      if (request.status < 200 || request.status >= 300) {
        reject(new Error(body.error ?? `upload failed: ${request.status}`));
        return;
      }
      onProgress(file.size, file.size, "uploading");
      resolve({ kind: body.kind ?? "media", status: body.status ?? "importing" });
    };
    request.onerror = () => reject(new Error("upload network error"));
    request.onabort = () => reject(new Error("upload aborted"));
    request.send(file);
    });
  });
}
