import { writable } from 'svelte/store';
import { getJobs } from '../api.js';

function createJobs() {
  const { subscribe, set, update } = writable([]);

  const refresh = async () => {
    try {
      const jobs = await getJobs();
      set(jobs);
    } catch (e) {
      console.error('Failed to refresh jobs:', e);
    }
  };

  return {
    subscribe,
    set,
    update,
    refresh
  };
}

export const jobs = createJobs();
