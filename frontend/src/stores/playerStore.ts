import { create } from "zustand";
import type { Track } from "../api/types";

interface PlayerState {
  queue: Track[];
  currentIndex: number;
  isPlaying: boolean;
  currentTrack: () => Track | undefined;
  playQueue: (tracks: Track[], startIndex: number) => void;
  togglePlay: () => void;
  pause: () => void;
  next: () => void;
  prev: () => void;
}

/**
 * Playback intent only -- what should be playing, not the actual <audio>
 * element (that lives once at the app root in components/MusicPlayerBar.tsx,
 * which reacts to this store via useEffect). Kept separate from
 * react-query's server-state cache since this is pure client-side UI state.
 */
export const usePlayerStore = create<PlayerState>((set, get) => ({
  queue: [],
  currentIndex: -1,
  isPlaying: false,

  currentTrack: () => {
    const { queue, currentIndex } = get();
    return currentIndex >= 0 ? queue[currentIndex] : undefined;
  },

  playQueue: (tracks, startIndex) => {
    set({ queue: tracks, currentIndex: startIndex, isPlaying: true });
  },

  togglePlay: () => {
    if (get().currentIndex < 0) return;
    set((s) => ({ isPlaying: !s.isPlaying }));
  },

  pause: () => set({ isPlaying: false }),

  next: () => {
    const { queue, currentIndex } = get();
    if (currentIndex < queue.length - 1) {
      set({ currentIndex: currentIndex + 1, isPlaying: true });
    } else {
      set({ isPlaying: false });
    }
  },

  prev: () => {
    const { currentIndex } = get();
    if (currentIndex > 0) {
      set({ currentIndex: currentIndex - 1, isPlaying: true });
    }
  },
}));
