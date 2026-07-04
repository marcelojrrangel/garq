<script>
  let { x, y, targetIndex, onClose } = $props();

  const menuItems = [
    { label: 'Abrir', icon: '📂', shortcut: 'Enter', action: 'open', disabled: targetIndex < 0 },
    { type: 'separator' },
    { label: 'Copiar', icon: '📋', shortcut: 'Ctrl+C', action: 'copy' },
    { label: 'Recortar', icon: '✂️', shortcut: 'Ctrl+X', action: 'cut' },
    { type: 'separator' },
    { label: 'Excluir da seleção', icon: '🗑️', shortcut: 'Del', action: 'delete' },
  ];

  function handleAction(action) {
    console.log('Action:', action);
    onClose();
  }

  function handleClickOutside(e) {
    if (!e.target.closest('.context-menu')) {
      onClose();
    }
  }
</script>

<svelte:window on:click={handleClickOutside} />

<div class="context-menu" style="left: {x}px; top: {y}px;">
  {#each menuItems as item}
    {#if item.type === 'separator'}
      <div class="context-menu-sep"></div>
    {:else}
      <div
        class="context-menu-item"
        class:disabled={item.disabled}
        onclick={() => handleAction(item.action)}
      >
        <span>{item.icon}</span>
        <span>{item.label}</span>
        {#if item.shortcut}
          <span class="context-menu-shortcut">{item.shortcut}</span>
        {/if}
      </div>
    {/if}
  {/each}
</div>

<style>
  .context-menu {
    position: fixed;
    background: var(--bg-elevated);
    border: 1px solid var(--border-strong);
    border-radius: 6px;
    padding: 4px 0;
    min-width: 180px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    z-index: 1000;
  }

  .context-menu-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    font-size: 12px;
    cursor: pointer;
    color: var(--text-primary);
  }

  .context-menu-item:hover {
    background: var(--bg-hover);
  }

  .context-menu-item.disabled {
    color: var(--text-disabled);
    cursor: default;
  }

  .context-menu-item.disabled:hover {
    background: transparent;
  }

  .context-menu-sep {
    height: 1px;
    background: var(--border);
    margin: 4px 0;
  }

  .context-menu-shortcut {
    margin-left: auto;
    color: var(--text-secondary);
    font-size: 11px;
  }
</style>
