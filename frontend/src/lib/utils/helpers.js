export function esc(s) {
  return String(s).replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;');
}

export function fileIcon(e) {
  if (e.is_dir) return '📁';
  const n = e.name.toLowerCase();
  if (n.endsWith('.exe') || n.endsWith('.msi')) return '⚙️';
  if (n.endsWith('.zip') || n.endsWith('.7z') || n.endsWith('.rar')) return '📦';
  if (n.endsWith('.txt') || n.endsWith('.md') || n.endsWith('.log')) return '📄';
  if (n.endsWith('.go') || n.endsWith('.js') || n.endsWith('.ts') || n.endsWith('.py')) return '📝';
  if (n.endsWith('.png') || n.endsWith('.jpg') || n.endsWith('.gif') || n.endsWith('.svg')) return '🖼️';
  if (n.endsWith('.mp3') || n.endsWith('.wav')) return '🎵';
  if (n.endsWith('.mp4') || n.endsWith('.avi')) return '🎬';
  return '📄';
}

export function ext(name) {
  const i = name.lastIndexOf('.');
  return i > 0 ? name.substring(i + 1).toLowerCase() : '';
}

export function statusClass(s) {
  if (s === 'done') return 'done';
  if (s === 'failed') return 'failed';
  if (s === 'running') return 'running';
  return 'pending';
}

export function translateStatus(s) {
  if (s === 'pending') return 'pendente';
  if (s === 'running') return 'executando';
  if (s === 'done') return 'concluído';
  if (s === 'failed') return 'falhou';
  return s;
}

export function isInputFocused() {
  const el = document.activeElement;
  return el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable);
}
