<script>
  import Toolbar from '$lib/components/Toolbar.svelte';
  import AddressBar from '$lib/components/AddressBar.svelte';
  import NavPane from '$lib/components/NavPane.svelte';
  import ContentList from '$lib/components/ContentList.svelte';
  import DetailsPane from '$lib/components/DetailsPane.svelte';
  import StatusBar from '$lib/components/StatusBar.svelte';
  import ContextMenu from '$lib/components/ContextMenu.svelte';
  import ProgressModal from '$lib/components/ProgressModal.svelte';
  import SettingsModal from '$lib/components/SettingsModal.svelte';
  import { appState } from '$lib/stores/state.js';
  import { clipboard } from '$lib/stores/clipboard.js';
  import { settings } from '$lib/stores/settings.js';
  import { jobs } from '$lib/stores/jobs.js';
  import { listRoots, listDirectory, addCopyJob, addMoveJob, addDeleteJob, addCompressJob, addExtractJob } from '$lib/api.js';
  import { isInputFocused } from '$lib/utils/helpers.js';

  let contextMenuState = $state({ visible: false, x: 0, y: 0, targetIndex: -1 });
  let progressModalState = $state({ visible: false, jobId: null, type: null });
  let settingsModalState = $state(false);

  async function openDirectory(path, fromHistory = false) {
    appState.setCurrentPath(path);

    if (!fromHistory) {
      appState.update(s => {
        const newHistory = s.historyIndex < s.navHistory.length - 1
          ? s.navHistory.slice(0, s.historyIndex + 1)
          : [...s.navHistory];
        newHistory.push(path);
        return {
          ...s,
          navHistory: newHistory,
          historyIndex: newHistory.length - 1
        };
      });
    }

    try {
      const entries = await listDirectory(path);
      appState.setCurrentEntries(entries);
      appState.setFilteredEntries(entries);
      appState.setDisplayEntries(entries);
      appState.setSelectedIndices(new Set());
      appState.setFocusedIndex(-1);
      appState.setAnchorIndex(-1);
    } catch (e) {
      console.error('Failed to open directory:', e);
    }
  }

  function handleKeyDown(e) {
    if (isInputFocused()) return;
    const ctrl = e.ctrlKey || e.metaKey;
    const shift = e.shiftKey;

    switch (e.key) {
      case 'ArrowUp':
        e.preventDefault();
        // moveFocus(-1, shift);
        break;
      case 'ArrowDown':
        e.preventDefault();
        // moveFocus(1, shift);
        break;
      case 'Enter':
        // open focused directory
        break;
      case 'Backspace':
        e.preventDefault();
        navigateUp();
        break;
      case 'Delete':
        // delete selected
        break;
      case 'F5':
        e.preventDefault();
        navigateRefresh();
        break;
      case 'Escape':
        contextMenuState.visible = false;
        break;
    }

    if (ctrl) {
      switch (e.key.toLowerCase()) {
        case 'a':
          e.preventDefault();
          // selectAll();
          break;
        case 'c':
          e.preventDefault();
          // clipboardCopy();
          break;
        case 'x':
          e.preventDefault();
          // clipboardCut();
          break;
        case 'v':
          e.preventDefault();
          // clipboardPaste();
          break;
      }
    }
  }

  function navigateBack() {
    appState.update(s => {
      if (s.historyIndex > 0) {
        const newIndex = s.historyIndex - 1;
        openDirectory(s.navHistory[newIndex], true);
        return { ...s, historyIndex: newIndex };
      }
      return s;
    });
  }

  function navigateForward() {
    appState.update(s => {
      if (s.historyIndex < s.navHistory.length - 1) {
        const newIndex = s.historyIndex + 1;
        openDirectory(s.navHistory[newIndex], true);
        return { ...s, historyIndex: newIndex };
      }
      return s;
    });
  }

  function navigateUp() {
    appState.update(s => {
      if (!s.currentPath) return s;
      const sep = s.currentPath.includes('\\') ? '\\' : '/';
      const parts = s.currentPath.split(sep).filter(Boolean);
      if (parts.length > 1) {
        const parent = parts.slice(0, -1).join(sep);
        openDirectory(parent.includes(':') ? parent : sep + parent, false);
      }
      return s;
    });
  }

  function navigateRefresh() {
    if ($appState.currentPath) {
      openDirectory($appState.currentPath, true);
    }
  }

  $effect(() => {
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  });

  $effect(() => {
    const interval = setInterval(() => jobs.refresh(), 2000);
    return () => clearInterval(interval);
  });

  $effect(() => {
    listRoots().then(roots => {
      // Initialize nav pane with roots
    });
  });
</script>

<div class="app">
  <Toolbar
    onBack={navigateBack}
    onForward={navigateForward}
    onUp={navigateUp}
    onRefresh={navigateRefresh}
  />
  <AddressBar />
  <div class="main-layout">
    <NavPane {openDirectory} />
    <ContentList
      {openDirectory}
      onContextMenu={(e) => contextMenu = { visible: true, x: e.clientX, y: e.clientY, targetIndex: -1 }}
    />
    <DetailsPane />
  </div>
  <StatusBar />

  {#if contextMenuState.visible}
    <ContextMenu
      x={contextMenuState.x}
      y={contextMenuState.y}
      targetIndex={contextMenuState.targetIndex}
      onClose={() => contextMenuState.visible = false}
    />
  {/if}

  {#if progressModalState.visible}
    <ProgressModal
      jobId={progressModalState.jobId}
      type={progressModalState.type}
      onClose={() => progressModalState.visible = false}
    />
  {/if}

  {#if settingsModalState}
    <SettingsModal onClose={() => settingsModalState = false} />
  {/if}
</div>

<style>
  .app {
    display: flex;
    flex-direction: column;
    height: 100vh;
    background: var(--bg-app);
    color: var(--text-primary);
    font-family: 'Segoe UI', -apple-system, sans-serif;
    font-size: 12px;
    user-select: none;
  }

  .main-layout {
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  :root {
    color-scheme: dark;
    --bg-app: #191919;
    --bg-surface: #202020;
    --bg-elevated: #2d2d2d;
    --bg-hover: #383838;
    --bg-selected: #004275;
    --border: #2d2d2d;
    --border-strong: #3d3d3d;
    --text-primary: #ffffff;
    --text-secondary: #999999;
    --text-disabled: #666666;
    --accent: #4cc2ff;
    --accent-hover: #6cd0ff;
    --danger: #ff5d5d;
    --success: #53d769;
    --scrollbar-thumb: #555;
    --scrollbar-track: transparent;
  }

  :global(*) {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
  }

  :global(body) {
    font-family: 'Segoe UI', -apple-system, sans-serif;
    font-size: 12px;
    background: var(--bg-app);
    color: var(--text-primary);
    overflow: hidden;
    height: 100vh;
    user-select: none;
  }

  :global(::-webkit-scrollbar) {
    width: 8px;
    height: 8px;
  }

  :global(::-webkit-scrollbar-track) {
    background: var(--scrollbar-track);
  }

  :global(::-webkit-scrollbar-thumb) {
    background: var(--scrollbar-thumb);
    border-radius: 4px;
  }

  :global(::-webkit-scrollbar-thumb:hover) {
    background: #777;
  }
</style>
