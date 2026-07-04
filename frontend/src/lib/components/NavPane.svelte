<script>
  import { appState } from '$lib/stores/state.js';
  import { listRoots, listDirectory } from '$lib/api.js';

  let { openDirectory } = $props();

  let roots = $state([]);
  let expanded = $state(new Set());
  let activePath = $state('');

  async function loadRoots() {
    try {
      roots = await listRoots();
    } catch (e) {
      console.error('Failed to load roots:', e);
    }
  }

  async function toggleExpand(path) {
    if (expanded.has(path)) {
      expanded.delete(path);
    } else {
      expanded.add(path);
    }
    expanded = expanded;
  }

  async function loadChildren(path) {
    try {
      const entries = await listDirectory(path);
      return entries.filter(e => e.is_dir);
    } catch (e) {
      console.error('Failed to load children:', e);
      return [];
    }
  }

  function handleClick(path) {
    activePath = path;
    openDirectory(path);
  }

  $effect(() => {
    loadRoots();
  });
</script>

<div class="nav-pane">
  <div class="nav-section">
    <div class="nav-section-header">
      <span>Este Computador</span>
    </div>
    <ul class="nav-tree">
      {#each roots as root}
        <li>
          <div class="nav-tree-item" class:active={activePath === root} onclick={() => handleClick(root)}>
            <span class="nav-tree-toggle" onclick={(e) => { e.stopPropagation(); toggleExpand(root); }}>
              {expanded.has(root) ? '▾' : '▸'}
            </span>
            <span class="nav-tree-icon">💻</span>
            <span class="nav-tree-label">{root}</span>
          </div>
        </li>
      {/each}
    </ul>
  </div>
</div>

<style>
  .nav-pane {
    width: 240px;
    min-width: 180px;
    max-width: 400px;
    background: var(--bg-surface);
    border-right: 1px solid var(--border);
    overflow-y: auto;
    overflow-x: hidden;
    padding: 4px 0;
  }

  .nav-section {
    margin-bottom: 4px;
  }

  .nav-section-header {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px 4px 12px;
    font-size: 12px;
    color: var(--text-secondary);
    cursor: pointer;
  }

  .nav-section-header:hover {
    color: var(--text-primary);
  }

  .nav-tree {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .nav-tree-item {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 3px 8px;
    cursor: pointer;
    border-radius: 4px;
    margin: 0 4px;
    font-size: 12px;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .nav-tree-item:hover {
    background: var(--bg-hover);
  }

  .nav-tree-item.active {
    background: var(--bg-selected);
  }

  .nav-tree-toggle {
    width: 16px;
    height: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 8px;
    color: var(--text-secondary);
    flex-shrink: 0;
    cursor: pointer;
    border-radius: 3px;
  }

  .nav-tree-toggle:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .nav-tree-icon {
    font-size: 14px;
    flex-shrink: 0;
  }

  .nav-tree-label {
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>
