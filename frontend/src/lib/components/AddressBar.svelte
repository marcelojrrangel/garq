<script>
  import { appState } from '$lib/stores/state.js';
  import { esc } from '$lib/utils/helpers.js';

  let searchQuery = $state('');

  function updateBreadcrumb(path) {
    if (!path) return [];
    const sep = path.includes('\\') ? '\\' : '/';
    const parts = path.split(sep).filter(Boolean);
    const items = [];
    let acc = parts[0].includes(':') ? parts[0] : sep;
    items.push({ path: acc, label: parts[0] });
    for (let i = 1; i < parts.length; i++) {
      acc += sep + parts[i];
      items.push({ path: acc, label: parts[i] });
    }
    return items;
  }

  let breadcrumbItems = $derived(updateBreadcrumb($appState.currentPath));
</script>

<div class="address-bar">
  <div class="address-breadcrumb">
    {#if breadcrumbItems.length === 0}
      <span class="breadcrumb-item" style="color: var(--text-secondary);">Selecione uma pasta...</span>
    {:else}
      {#each breadcrumbItems as item, i}
        {#if i > 0}
          <span class="breadcrumb-sep">›</span>
        {/if}
        <span class="breadcrumb-item" onclick={() => {}}>
          {item.label}
        </span>
      {/each}
    {/if}
  </div>
  <div class="search-wrapper">
    <span class="search-icon">🔍</span>
    <input
      class="address-search"
      placeholder="Pesquisar"
      bind:value={searchQuery}
    />
  </div>
</div>

<style>
  .address-bar {
    display: flex;
    align-items: center;
    height: 32px;
    background: var(--bg-surface);
    padding: 0 8px;
    gap: 4px;
  }

  .address-breadcrumb {
    flex: 1;
    display: flex;
    align-items: center;
    height: 26px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-strong);
    border-radius: 4px;
    padding: 0 8px;
    font-size: 12px;
    gap: 2px;
    overflow: hidden;
  }

  .address-breadcrumb:focus-within {
    border-color: var(--accent);
  }

  .breadcrumb-item {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 2px 4px;
    border-radius: 3px;
    cursor: pointer;
    white-space: nowrap;
    color: var(--text-primary);
  }

  .breadcrumb-item:hover {
    background: var(--bg-hover);
  }

  .breadcrumb-sep {
    color: var(--text-secondary);
    font-size: 8px;
    margin: 0 2px;
  }

  .search-wrapper {
    position: relative;
    width: 220px;
  }

  .search-icon {
    position: absolute;
    left: 8px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--text-secondary);
    font-size: 12px;
    pointer-events: none;
  }

  .address-search {
    width: 220px;
    height: 26px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-strong);
    border-radius: 4px;
    padding: 0 8px 0 28px;
    color: var(--text-primary);
    font-size: 12px;
    font-family: inherit;
    outline: none;
  }

  .address-search:focus {
    border-color: var(--accent);
  }

  .address-search::placeholder {
    color: var(--text-disabled);
  }
</style>
