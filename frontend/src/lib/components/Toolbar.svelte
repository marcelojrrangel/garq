<script>
  import { appState } from '$lib/stores/state.js';

  let { onBack, onForward, onUp, onRefresh } = $props();

  let backDisabled = $derived($appState.historyIndex <= 0);
  let forwardDisabled = $derived($appState.historyIndex >= $appState.navHistory.length - 1);
  let upDisabled = $derived(!$appState.currentPath);
</script>

<div class="toolbar">
  <button class="toolbar-btn" onclick={onBack} disabled={backDisabled} title="Voltar (Alt+←)">←</button>
  <button class="toolbar-btn" onclick={onForward} disabled={forwardDisabled} title="Avançar (Alt+→)">→</button>
  <button class="toolbar-btn" onclick={onUp} disabled={upDisabled} title="Pasta pai (Alt+↑)">↑</button>
  <div class="toolbar-sep"></div>
  <button class="toolbar-btn" onclick={onRefresh} title="Atualizar (F5)">⟳</button>
  <button class="toolbar-btn" title="Usar pasta atual como destino">📋</button>
  <div class="toolbar-sep"></div>
  <button class="toolbar-btn" title="Configurações">⚙️</button>
</div>

<style>
  .toolbar {
    display: flex;
    align-items: center;
    height: 40px;
    background: var(--bg-surface);
    padding: 0 4px;
    gap: 2px;
  }

  .toolbar-btn {
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    border-radius: 4px;
    font-size: 14px;
  }

  .toolbar-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .toolbar-btn:disabled {
    color: var(--text-disabled);
    cursor: default;
  }

  .toolbar-btn:disabled:hover {
    background: transparent;
  }

  .toolbar-sep {
    width: 1px;
    height: 20px;
    background: var(--border-strong);
    margin: 0 4px;
  }
</style>
