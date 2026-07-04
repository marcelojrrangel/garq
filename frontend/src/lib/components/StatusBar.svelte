<script>
  import { appState } from '$lib/stores/state.js';
  import { jobs } from '$lib/stores/jobs.js';
</script>

<div class="statusbar">
  <div class="statusbar-left">
    <span>{$appState.displayEntries.length} itens</span>
    <span>{$appState.selectedIndices.size} selecionados</span>
  </div>
  <div class="statusbar-right">
    <span>{$jobs.length} jobs</span>
  </div>
</div>

<style>
  .statusbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 24px;
    background: var(--bg-surface);
    border-top: 1px solid var(--border);
    padding: 0 12px;
    font-size: 11px;
    color: var(--text-secondary);
  }

  .statusbar-left,
  .statusbar-right {
    display: flex;
    gap: 12px;
  }
</style>
