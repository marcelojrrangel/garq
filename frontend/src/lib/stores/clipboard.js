import { writable } from 'svelte/store';

function createClipboard() {
  const { subscribe, set, update } = writable({
    mode: null,
    paths: []
  });

  return {
    subscribe,
    set,
    update,
    copy: (paths) => set({ mode: 'copy', paths }),
    cut: (paths) => set({ mode: 'cut', paths }),
    clear: () => set({ mode: null, paths: [] })
  };
}

export const clipboard = createClipboard();
