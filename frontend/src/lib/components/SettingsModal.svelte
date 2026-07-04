<script>
  import { settings } from '$lib/stores/settings.js';

  let { onClose } = $props();

  let copySetting = $state($settings.copy);
  let moveSetting = $state($settings.move);
  let compressSetting = $state($settings.compress);
  let extractSetting = $state($settings.extract);

  function save() {
    settings.save({
      copy: copySetting,
      move: moveSetting,
      compress: compressSetting,
      extract: extractSetting
    });
    onClose();
  }
</script>

<div class="modal-overlay" onclick={(e) => { if (e.target === e.currentTarget) onClose(); }}>
  <div class="modal-card">
    <div class="modal-title">
      <span class="icon">⚙️</span>
      <span>Configurações de Conflito</span>
    </div>

    <div class="settings-group">
      <div class="details-group">
        <div class="details-group-title">📋 Cópia</div>
        <label class="details-label">Quando o arquivo destino já existir:</label>
        <select class="details-input" bind:value={copySetting}>
          <option value="replace">Substituir (sobrescrever)</option>
          <option value="skip">Pular (manter existente)</option>
          <option value="rename">Renomear (adicionar número)</option>
        </select>
      </div>
    </div>

    <div class="settings-group">
      <div class="details-group">
        <div class="details-group-title">✂️ Movimentação</div>
        <label class="details-label">Quando o arquivo destino já existir:</label>
        <select class="details-input" bind:value={moveSetting}>
          <option value="replace">Substituir (sobrescrever)</option>
          <option value="skip">Pular (manter existente)</option>
          <option value="rename">Renomear (adicionar número)</option>
        </select>
      </div>
    </div>

    <div class="settings-group">
      <div class="details-group">
        <div class="details-group-title">📦 Compactação (7z)</div>
        <label class="details-label">Quando o archive destino já existir:</label>
        <select class="details-input" bind:value={compressSetting}>
          <option value="replace">Substituir (sobrescrever)</option>
          <option value="skip">Pular (não compactar)</option>
          <option value="rename">Renomear (adicionar número)</option>
        </select>
      </div>
    </div>

    <div class="settings-group">
      <div class="details-group">
        <div class="details-group-title">📂 Extração (7z)</div>
        <label class="details-label">Quando o arquivo extraído já existir:</label>
        <select class="details-input" bind:value={extractSetting}>
          <option value="replace">Substituir (sobrescrever)</option>
          <option value="skip">Pular (manter existente)</option>
        </select>
      </div>
    </div>

    <div class="modal-footer">
      <button class="modal-btn secondary" onclick={onClose}>Cancelar</button>
      <button class="modal-btn primary" onclick={save}>Salvar</button>
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
    width: 380px;
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

  .settings-group {
    margin-bottom: 16px;
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
    height: 28px;
    background: var(--bg-hover);
    border: 1px solid var(--border-strong);
    border-radius: 4px;
    padding: 0 8px;
    color: var(--text-primary);
    font-size: 12px;
    font-family: inherit;
  }

  .details-input:focus {
    border-color: var(--accent);
    outline: none;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 16px;
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

  .modal-btn.primary {
    background: var(--accent);
    color: #000;
  }

  .modal-btn.primary:hover {
    background: var(--accent-hover);
  }
</style>
