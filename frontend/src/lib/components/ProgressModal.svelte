<script>
  import { jobs } from '$lib/stores/jobs.js';
  import { statusClass, translateStatus } from '$lib/utils/helpers.js';

  let { jobId, type, onClose } = $props();

  let job = $state(null);
  let progress = $state(0);
  let status = $state('pending');

  function getJobIcon(t) {
    if (t === 'copy') return '📋';
    if (t === 'compress') return '📦';
    if (t === 'extract') return '📂';
    if (t === 'move') return '✂️';
    if (t === 'delete') return '🗑️';
    return '⏳';
  }

  function getJobTitle(t) {
    if (t === 'copy') return 'Copiando';
    if (t === 'compress') return 'Compactando';
    if (t === 'extract') return 'Extraindo';
    if (t === 'move') return 'Movendo';
    if (t === 'delete') return 'Excluindo';
    return 'Processando';
  }

  $effect(() => {
    const interval = setInterval(async () => {
      try {
        const jobsList = await jobs.refresh();
        job = $jobs.find(j => j.id === jobId);
        if (job) {
          progress = Math.round((job.progress || 0) * 100);
          status = job.status;
          if (job.status === 'done' || job.status === 'failed') {
            clearInterval(interval);
          }
        }
      } catch (e) {}
    }, 300);
    return () => clearInterval(interval);
  });
</script>

<div class="modal-overlay">
  <div class="modal-card">
    <div class="modal-title">
      <span class="icon">{getJobIcon(type)}</span>
      <span>{getJobTitle(type)} (Job #{jobId})</span>
    </div>
    <div class="progress-track">
      <div class="progress-fill" style="width: {progress}%"></div>
    </div>
    <div class="progress-info">
      <span>{progress}%</span>
      <span class="status-tag {statusClass(status)}">{translateStatus(status)}</span>
    </div>
    {#if status === 'failed'}
      <div class="modal-error">{job?.error || 'Erro desconhecido'}</div>
    {/if}
    <div class="modal-footer">
      <button class="modal-btn secondary" onclick={onClose}>Fechar</button>
    </div>
  </div>
</div>

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
  }

  .modal-card {
    background: var(--bg-elevated);
    border: 1px solid var(--border-strong);
    border-radius: 8px;
    padding: 20px;
    min-width: 300px;
    max-width: 400px;
  }

  .modal-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 500;
    margin-bottom: 16px;
  }

  .modal-title .icon {
    font-size: 20px;
  }

  .progress-track {
    height: 8px;
    background: var(--bg-hover);
    border-radius: 4px;
    overflow: hidden;
    margin-bottom: 8px;
  }

  .progress-fill {
    height: 100%;
    background: var(--accent);
    border-radius: 4px;
    transition: width 0.3s ease;
  }

  .progress-info {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 12px;
  }

  .status-tag {
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 11px;
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

  .modal-error {
    background: #4d1a1a;
    color: var(--danger);
    padding: 8px 12px;
    border-radius: 4px;
    font-size: 12px;
    margin-bottom: 12px;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  .modal-btn {
    padding: 6px 16px;
    border: none;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
  }

  .modal-btn.secondary {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .modal-btn.secondary:hover {
    background: var(--bg-selected);
  }
</style>
