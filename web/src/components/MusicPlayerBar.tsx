import { useEffect, useRef, useState } from "react";
import { usePlayerStore } from "../stores/playerStore";
import { audioStreamUrl } from "../api/client";

function formatTime(seconds: number): string {
  if (!Number.isFinite(seconds)) return "0:00";
  const m = Math.floor(seconds / 60);
  const s = Math.floor(seconds % 60);
  return `${m}:${s.toString().padStart(2, "0")}`;
}

/**
 * Mounted once at the app root (see App.tsx) so audio keeps playing while
 * the user navigates elsewhere -- a single <audio> element controlled
 * imperatively from playerStore's intent (current track / play-pause),
 * mirroring how a real media player bar works.
 */
export function MusicPlayerBar() {
  const audioRef = useRef<HTMLAudioElement>(null);
  const { queue, currentIndex, isPlaying, togglePlay, next, prev } = usePlayerStore();
  const track = currentIndex >= 0 ? queue[currentIndex] : undefined;
  const [currentTime, setCurrentTime] = useState(0);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio || !track) return;
    audio.src = audioStreamUrl(track.id);
    audio.play().catch(() => {
      /* autoplay may be blocked until a user gesture; the play button covers that */
    });
  }, [track?.id]);

  useEffect(() => {
    const audio = audioRef.current;
    if (!audio || !track) return;
    if (isPlaying) audio.play().catch(() => {});
    else audio.pause();
  }, [isPlaying, track]);

  if (!track) return null;

  return (
    <div className="fixed inset-x-0 bottom-0 z-40 border-t border-neutral-800 bg-neutral-950/95 px-4 py-3 backdrop-blur sm:px-8">
      {/* eslint-disable-next-line jsx-a11y/media-has-caption */}
      <audio
        ref={audioRef}
        onTimeUpdate={(e) => setCurrentTime(e.currentTarget.currentTime)}
        onEnded={next}
      />
      <div className="mx-auto flex max-w-4xl items-center gap-4">
        <div className="min-w-0 flex-1">
          <p className="truncate text-sm font-medium text-white">{track.title}</p>
          <p className="truncate text-xs text-neutral-400">{track.artistName}</p>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={prev}
            disabled={currentIndex <= 0}
            className="text-neutral-300 hover:text-white disabled:opacity-30"
            aria-label="Previous"
          >
            ⏮
          </button>
          <button
            onClick={togglePlay}
            className="flex h-9 w-9 items-center justify-center rounded-full bg-white text-black hover:bg-neutral-200"
            aria-label={isPlaying ? "Pause" : "Play"}
          >
            {isPlaying ? "⏸" : "▶"}
          </button>
          <button
            onClick={next}
            disabled={currentIndex >= queue.length - 1}
            className="text-neutral-300 hover:text-white disabled:opacity-30"
            aria-label="Next"
          >
            ⏭
          </button>
        </div>

        <div className="hidden items-center gap-2 text-xs text-neutral-400 sm:flex">
          <span>{formatTime(currentTime)}</span>
          <span>/</span>
          <span>{formatTime(track.durationSeconds)}</span>
        </div>
      </div>
    </div>
  );
}
