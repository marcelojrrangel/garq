import { writable } from 'svelte/store';

function createAppState() {
  const { subscribe, set, update } = writable({
    currentPath: '',
    currentEntries: [],
    filteredEntries: [],
    displayEntries: [],
    sortColumn: 'name',
    sortDirection: 'default',
    selectedIndices: new Set(),
    focusedIndex: -1,
    anchorIndex: -1,
    searchQuery: '',
    expanded: new Set(),
    navHistory: [],
    historyIndex: -1,
    selectedSources: new Set()
  });

  return {
    subscribe,
    set,
    update,
    setCurrentPath: (path) => update(s => ({ ...s, currentPath: path })),
    setCurrentEntries: (entries) => update(s => ({ ...s, currentEntries: entries })),
    setFilteredEntries: (entries) => update(s => ({ ...s, filteredEntries: entries })),
    setDisplayEntries: (entries) => update(s => ({ ...s, displayEntries: entries })),
    setSortColumn: (col) => update(s => ({ ...s, sortColumn: col })),
    setSortDirection: (dir) => update(s => ({ ...s, sortDirection: dir })),
    setSelectedIndices: (indices) => update(s => ({ ...s, selectedIndices: indices })),
    setFocusedIndex: (idx) => update(s => ({ ...s, focusedIndex: idx })),
    setAnchorIndex: (idx) => update(s => ({ ...s, anchorIndex: idx })),
    setSearchQuery: (query) => update(s => ({ ...s, searchQuery: query })),
    setExpanded: (expanded) => update(s => ({ ...s, expanded: expanded })),
    setNavHistory: (history) => update(s => ({ ...s, navHistory: history })),
    setHistoryIndex: (idx) => update(s => ({ ...s, historyIndex: idx })),
    setSelectedSources: (sources) => update(s => ({ ...s, selectedSources: sources }))
  };
}

export const appState = createAppState();
