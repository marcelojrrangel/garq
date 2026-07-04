<script>
  import { appState } from '$lib/stores/state.js';
  import { esc, fileIcon, ext } from '$lib/utils/helpers.js';

  let { openDirectory, onContextMenu } = $props();

  let listElement;

  function applySort() {
    const entries = [...$appState.filteredEntries];
    if ($appState.sortDirection === 'default') {
      entries.sort((a, b) => {
        if (a.is_dir !== b.is_dir) return a.is_dir ? -1 : 1;
        return a.name.localeCompare(b.name, 'pt-BR');
      });
    } else {
      const dir = $appState.sortDirection === 'asc' ? 1 : -1;
      entries.sort((a, b) => {
        if (a.is_dir !== b.is_dir) return a.is_dir ? -1 : 1;
        let cmp = 0;
        switch ($appState.sortColumn) {
          case 'name': cmp = a.name.localeCompare(b.name, 'pt-BR'); break;
          case 'type': cmp = (a.is_dir ? '' : ext(a.name)).localeCompare(b.is_dir ? '' : ext(b.name)); break;
          case 'size': cmp = (a.size || 0) - (b.size || 0); break;
          case 'date': cmp = new Date(a.mod_time) - new Date(b.mod_time); break;
        }
        return cmp * dir;
      });
    }
    state.setDisplayEntries(entries);
  }

  function cycleSort(col) {
    if ($appState.sortColumn === col) {
      if ($appState.sortDirection === 'default') state.setSortDirection('asc');
      else if ($appState.sortDirection === 'asc') state.setSortDirection('desc');
      else {
        state.setSortDirection('default');
        state.setSortColumn('name');
      }
    } else {
      state.setSortColumn(col);
      state.setSortDirection('asc');
    }
    applySort();
  }

  function selectOnly(idx) {
    const newSelected = new Set([idx]);
    state.setSelectedIndices(newSelected);
    state.setFocusedIndex(idx);
    state.setAnchorIndex(idx);
  }

  function toggleSelect(idx) {
    const newSelected = new Set($appState.selectedIndices);
    if (newSelected.has(idx)) newSelected.delete(idx);
    else newSelected.add(idx);
    state.setSelectedIndices(newSelected);
    state.setFocusedIndex(idx);
    state.setAnchorIndex(idx);
  }

  function rangeSelect(from, to) {
    const start = Math.min(from, to);
    const end = Math.max(from, to);
    const newSelected = new Set($appState.selectedIndices);
    for (let i = start; i <= end; i++) newSelected.add(i);
    state.setSelectedIndices(newSelected);
    state.setFocusedIndex(to);
  }

  function handleClick(e, idx) {
    if (e.ctrlKey || e.metaKey) {
      toggleSelect(idx);
    } else if (e.shiftKey) {
      if ($appState.anchorIndex === -1) selectOnly(idx);
      else rangeSelect($appState.anchorIndex, idx);
    } else {
      selectOnly(idx);
    }
  }

  function handleDblClick(idx) {
    const entry = $appState.displayEntries[idx];
    if (entry?.is_dir) openDirectory(entry.path);
  }

  function handleContextMenu(e, idx) {
    e.preventDefault();
    if (idx >= 0 && !$appState.selectedIndices.has(idx)) selectOnly(idx);
    onContextMenu(e);
  }

  $effect(() => {
    applySort();
  });
</script>

<div class="content-area">
  <div class="content-header">
    <div class="col-name" class:sort-active={$appState.sortColumn === 'name' && $appState.sortDirection !== 'default'}
         onclick={() => cycleSort('name')}>
      Nome
      {#if $appState.sortColumn === 'name' && $appState.sortDirection !== 'default'}
        <span class="sort-arrow">{$appState.sortDirection === 'asc' ? '▴' : '▾'}</span>
      {/if}
    </div>
    <div class="col-type" class:sort-active={$appState.sortColumn === 'type' && $appState.sortDirection !== 'default'}
         onclick={() => cycleSort('type')}>Tipo</div>
    <div class="col-size" class:sort-active={$appState.sortColumn === 'size' && $appState.sortDirection !== 'default'}
         onclick={() => cycleSort('size')}>Tamanho</div>
    <div class="col-mode">Permissões</div>
    <div class="col-date" class:sort-active={$appState.sortColumn === 'date' && $appState.sortDirection !== 'default'}
         onclick={() => cycleSort('date')}>Modificado em</div>
    <div class="col-action"></div>
  </div>

  <div class="content-list" bind:this={listElement}>
    {#if $appState.displayEntries.length === 0}
      <div class="empty-state">
        <div class="icon">📂</div>
        <div class="text">Pasta vazia</div>
      </div>
    {:else}
      {#each $appState.displayEntries as entry, idx}
        <div
          class="content-row"
          class:selected={$appState.selectedIndices.has(idx)}
          class:focused={idx === $appState.focusedIndex}
          onclick={(e) => handleClick(e, idx)}
          ondblclick={() => handleDblClick(idx)}
          oncontextmenu={(e) => handleContextMenu(e, idx)}
        >
          <div class="col-name">
            <span class="file-icon">{fileIcon(entry)}</span>
            <span class="file-name">{esc(entry.name)}</span>
          </div>
          <div class="col-type">{entry.is_dir ? 'Pasta' : 'Arquivo'}</div>
          <div class="col-size">{entry.is_dir ? '' : entry.size}</div>
          <div class="col-mode">{esc(entry.mode || '')}</div>
          <div class="col-date">{esc(entry.mod_time)}</div>
          <div class="col-action">
            {#if entry.is_dir}
              <button class="row-btn" onclick={(e) => { e.stopPropagation(); openDirectory(entry.path); }} title="Abrir">📂</button>
            {:else}
              <button class="row-btn" onclick={(e) => e.stopPropagation()} title="Adicionar">➕</button>
            {/if}
          </div>
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .content-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .content-header {
    display: flex;
    align-items: center;
    height: 28px;
    background: var(--bg-surface);
    border-bottom: 1px solid var(--border);
    padding: 0 12px;
    font-size: 12px;
    color: var(--text-secondary);
    flex-shrink: 0;
  }

  .col-name {
    flex: 3;
    min-width: 200px;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .col-name:hover {
    color: var(--text-primary);
  }

  .col-type {
    flex: 1;
    min-width: 80px;
    cursor: pointer;
  }

  .col-type:hover {
    color: var(--text-primary);
  }

  .col-size {
    flex: 1;
    min-width: 80px;
    text-align: right;
    cursor: pointer;
  }

  .col-size:hover {
    color: var(--text-primary);
  }

  .col-mode {
    flex: 1;
    min-width: 100px;
    cursor: pointer;
  }

  .col-mode:hover {
    color: var(--text-primary);
  }

  .col-date {
    flex: 1.5;
    min-width: 120px;
    cursor: pointer;
  }

  .col-date:hover {
    color: var(--text-primary);
  }

  .col-action {
    width: 80px;
    text-align: center;
  }

  .sort-active {
    color: var(--accent) !important;
  }

  .sort-arrow {
    font-size: 10px;
    margin-left: 2px;
  }

  .content-list {
    flex: 1;
    overflow-y: auto;
    padding: 2px 0;
    position: relative;
  }

  .content-row {
    display: flex;
    align-items: center;
    height: 24px;
    padding: 0 12px;
    cursor: pointer;
    font-size: 12px;
  }

  .content-row:hover {
    background: var(--bg-hover);
  }

  .content-row.selected {
    background: var(--bg-selected);
  }

  .content-row.focused {
    outline: 1px solid var(--accent);
    outline-offset: -1px;
  }

  .file-icon {
    margin-right: 4px;
  }

  .file-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .row-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 2px 4px;
    border-radius: 3px;
  }

  .row-btn:hover {
    background: var(--bg-hover);
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-disabled);
  }

  .empty-state .icon {
    font-size: 48px;
    margin-bottom: 8px;
  }
</style>
