import { writable } from 'svelte/store';

const defaultSettings = {
  copy: 'replace',
  move: 'replace',
  compress: 'replace',
  extract: 'replace'
};

function loadSettings() {
  try {
    return JSON.parse(localStorage.getItem('garq_conflict_settings')) || { ...defaultSettings };
  } catch (e) {
    return { ...defaultSettings };
  }
}

function createSettings() {
  const { subscribe, set, update } = writable(loadSettings());

  return {
    subscribe,
    set,
    update,
    save: (settings) => {
      localStorage.setItem('garq_conflict_settings', JSON.stringify(settings));
      set(settings);
    }
  };
}

export const settings = createSettings();
