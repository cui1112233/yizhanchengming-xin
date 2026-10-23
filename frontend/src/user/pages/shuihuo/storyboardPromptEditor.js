export function readStoryboardPromptEditableText(root) {
  return String(root?.innerText || '').replace(/\r\n?/g, '\n').replace(/\u00a0/g, ' ');
}

export function storyboardPromptTextForSave(root, fallback = '') {
  const current = readStoryboardPromptEditableText(root).trim();
  return current || String(fallback || '');
}
