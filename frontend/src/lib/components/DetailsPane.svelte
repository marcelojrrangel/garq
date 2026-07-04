<script>
  import { appState } from '$lib/stores/state.js';
  import { jobs } from '$lib/stores/jobs.js';
  import { addCopyJob, addCompressJob, addExtractJob } from '$lib/api.js';
  import { settings } from '$lib/stores/settings.js';
  import { statusClass, translateStatus } from '$lib/utils/helpers.js';

  let copyDest = $state('');
  let compressDest = $state('');
  let extractArchive = $state('');
  let extractDest = $state('');

  async function createCopyJob() {
    const sources = [...$appState.selectedSources];
    if (sources.length === 0 || !copyDest) return;
    try {
      await addCopyJob(sources, copyDest, $settings.copy);
      jobs.refresh();
    } catch (e) {
      console.error('Failed to create copy job:', e);
    }
  }

  async function createCompressJob() {
    const sources = [...$appState.selectedSources];
    if (sources.length === 0 || !compressDest) return;
    try {
      await addCompressJob(sources, compressDest, $settings.compress);
      jobs.refresh();
    } catch (e) {
      console.error('Failed to create compress job:', e);
    }
  }

  async function createExtractJob() {
    if (!extractArchive || !extractDest) return;
    try {
      await addExtractJob(extractArchive, extractDest, $settings.extract);
      jobs.refresh();
    } catch (e) {
      console.error('Failed to create extract job:', e);
    }
  }
</script>

<div class="details-pane">
  <div class="details-section">
    <div class="details-section-title">Selecionados</div>
    <div class="details-selected-list">
      {#if $appState.selectedSources.size === 0}
        <div style="color: var(--text-disabled); font-size: 11px; text-align: center; padding: 8px;">
          Nenhum arquivo selecionado
        </div>
      {:else}
        {#each [...$appState.selectedSources] as path}
          <div class="details-selected-item">
            <span class="name" title={path}>{path.split(/[\\/]/).pop()}</span>
          </div>
        {/each}
      {/if}
    </div>
  </div>

  <div class="details-section">
    <div class="details-group">
      <div class="details-group-title">📋 Cópia</div>
      <label class="details-label">Destino</label>
      <input class="details-input" placeholder="E:\destino" bind:value={copyDest} />
      <button class="details-btn primary" onclick={createCopyJob}>Criar job de cópia</button>
    </div>
  </div>

  <div class="details-section">
    <div class="details-group">
      <div class="details-group-title">📦 Compactação (7z)</div>
      <label class="details-label">Arquivo destino</label>
      <input class="details-input" placeholder="arquivo.7z" bind:value={compressDest} />
      <button class="details-btn primary" onclick={createCompressJob}>Criar job de compactação</button>
    </div>
  </div>

  <div class="details-section">
    <div class="details-group">
      <div class="details-group-title">📂 Extração (7z)</div>
      <label class="details-label">Arquivo .7z</label>
      <input class="details-input" placeholder="arquivo.7z" bind:value={extractArchive} />
      <label class="details-label">Pasta destino</label>
      <input class="details-input" placeholder="pasta destino" bind:value={extractDest} />
      <button class="details-btn primary" onclick={createExtractJob}>Criar job de extração</button>
    </div>
  </div>

  <div class="details-section">
    <div class="details-section-title">Jobs</div>
    <div class="jobs-list">
      {#if $jobs.length === 0}
        <div style="color: var(--text-disabled); font-size: 11px; text-align: center; padding: 8px;">
          Nenhum job
        </div>
      {:else}
        {#each $jobs as job}
          <div class="job-item">
            <span class="job-id">#{job.id}</span>
            <span class="job-type">{job.type}</span>
            <span class="status-tag {statusClass(job.status)}">{translateStatus(job.status)}</span>
            <span class="job-progress">{Math.round((job.progress || 0) * 100)}%</span>
          </div>
        {/each}
      {/if}
    </div>
  </div>
</div>

<style>
  .details-pane {
    width: 280px;
    min-width: 200px;
    max-width: 400px;
    background: var(--bg-surface);
    border-left: 1px solid var(--border);
    overflow-y: auto;
    padding: 8px;
  }

  .details-section {
    margin-bottom: 12px;
  }

  .details-section-title {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-secondary);
    margin-bottom: 6px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .details-group {
    margin-bottom: 8px;
  }

  .details-group-title {
    font-size: 12px;
    font-weight: 500;
    margin-bottom: 4px;
  }

  .details-label {
    display: block;
    font-size: 11px;
    color: var(--text-secondary);
    margin-bottom: 2px;
  }

  .details-input {
    width: 100%;
    height: 24px;
    background: var(--bg-elevated);
    border: 1px solid var(--border-strong);
    border-radius: 3px;
    padding: 0 6px;
    color: var(--text-primary);
    font-size: 11px;
    font-family: inherit;
    margin-bottom: 4px;
  }

  .details-input:focus {
    border-color: var(--accent);
    outline: none;
  }

  .details-btn {
    width: 100%;
    height: 28px;
    border: none;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    margin-top: 4px;
  }

  .details-btn.primary {
    background: var(--accent);
    color: #000;
  }

  .details-btn.primary:hover {
    background: var(--accent-hover);
  }

  .details-selected-list {
    max-height: 120px;
    overflow-y: auto;
  }

  .details-selected-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 2px 4px;
    font-size: 11px;
    border-radius: 3px;
  }

  .details-selected-item:hover {
    background: var(--bg-hover);
  }

  .details-selected-item .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  .jobs-list {
    max-height: 200px;
    overflow-y: auto;
  }

  .job-item {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 2px 4px;
    font-size: 11px;
    border-radius: 3px;
  }

  .job-item:hover {
    background: var(--bg-hover);
  }

  .job-id {
    color: var(--text-secondary);
    min-width: 30px;
  }

  .job-type {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .status-tag {
    padding: 1px 4px;
    border-radius: 3px;
    font-size: 10px;
    text-transform: uppercase;
  }

  .status-tag.pending {
    background: #444;
    color: #aaa;
  }

  .status-tag.running {
    background: #004275;
    color: var(--accent);
  }

  .status-tag.done {
    background: #1a4d1a;
    color: var(--success);
  }

  .status-tag.failed {
    background: #4d1a1a;
    color: var(--danger);
  }

  .job-progress {
    min-width: 30px;
    text-align: right;
  }
</style>
