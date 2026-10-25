import { useCallback, useLayoutEffect, useMemo, useRef } from 'react';
import { promptMentionRanges, removeAssetMention } from '../pages/agent-canvas-asset-input';

function plainTextFromNode(node) {
  if (!node) return '';
  if (node.nodeType === Node.TEXT_NODE) return node.nodeValue || '';
  if (node.nodeType === Node.ELEMENT_NODE && node.dataset?.assetMentionValue) return node.dataset.assetMentionValue;
  if (node.nodeType === Node.ELEMENT_NODE && node.tagName === 'BR') return '\n';
  return [...node.childNodes].map(plainTextFromNode).join('');
}

function plainTextFromEditor(root) {
  let text = '';
  for (const node of [...(root?.childNodes || [])]) {
    const block = node.nodeType === Node.ELEMENT_NODE && /^(DIV|P|LI)$/u.test(node.tagName);
    if (block && text && !text.endsWith('\n')) text += '\n';
    text += plainTextFromNode(node);
  }
  return text.replace(/\r/g, '');
}

function selectionOffset(root) {
  const selection = window.getSelection();
  if (!root || !selection?.rangeCount || !root.contains(selection.anchorNode)) return 0;
  const range = selection.getRangeAt(0);
  const before = range.cloneRange();
  before.selectNodeContents(root);
  before.setEnd(range.startContainer, range.startOffset);
  return plainTextFromEditor(before.cloneContents()).length;
}

function selectionRect(root) {
  const selection = window.getSelection();
  if (!selection?.rangeCount) return null;
  const range = selection.getRangeAt(0);
  const rect = [...range.getClientRects()].find(item => item.height) || range.getBoundingClientRect();
  return rect.width || rect.height ? rect : root?.getBoundingClientRect?.() || null;
}

function setSelectionOffset(root, wantedOffset) {
  if (!root || typeof document === 'undefined') return;
  root.focus({ preventScroll: true });
  let remaining = Math.max(0, wantedOffset);
  for (const node of [...root.childNodes]) {
    const mentionValue = node.nodeType === Node.ELEMENT_NODE ? node.dataset?.assetMentionValue : '';
    const size = mentionValue ? mentionValue.length : node.textContent.length;
    if (remaining <= size) {
      const range = document.createRange();
      if (mentionValue) {
        if (remaining === 0) range.setStartBefore(node);
        else range.setStartAfter(node);
      } else range.setStart(node.firstChild || node, Math.min(remaining, node.textContent.length));
      range.collapse(true);
      const selection = window.getSelection();
      selection.removeAllRanges();
      selection.addRange(range);
      return;
    }
    remaining -= size;
  }
  const range = document.createRange();
  range.selectNodeContents(root);
  range.collapse(false);
  const selection = window.getSelection();
  selection.removeAllRanges();
  selection.addRange(range);
}

function mentionAtBoundary(ranges, cursor, direction) {
  return ranges.find(range => (
    (direction === 'backspace' && cursor === range.end)
    || (direction === 'delete' && cursor === range.start)
  )) || null;
}

function appendAssetChip(fragment, range) {
  const { asset, label } = range;
  const chip = document.createElement('span');
  chip.className = 'agent-asset-mention-chip';
  chip.contentEditable = 'false';
  chip.dataset.assetMentionValue = label;
  chip.dataset.assetId = String(asset?.id || asset?.assetId || '');
  chip.dataset.kind = String(asset?.kind || 'text');

  const fallback = document.createElement('span');
  fallback.className = 'agent-asset-mention-chip__fallback';
  fallback.dataset.kind = String(asset?.kind || 'text');
  fallback.setAttribute('aria-hidden', 'true');

  if (asset?.kind === 'image' && asset?.url) {
    const image = document.createElement('img');
    image.src = asset.url;
    image.alt = '';
    image.loading = 'lazy';
    image.addEventListener('error', () => chip.classList.add('is-fallback'), { once: true });
    chip.append(image, fallback);
  } else chip.append(fallback);

  const name = document.createElement('span');
  name.className = 'agent-asset-mention-chip__name';
  name.textContent = label;
  chip.append(name);
  fragment.append(chip);
}

function renderEditor(root, value, ranges) {
  const fragment = document.createDocumentFragment();
  let cursor = 0;
  for (const range of ranges) {
    if (range.start > cursor) fragment.append(document.createTextNode(value.slice(cursor, range.start)));
    appendAssetChip(fragment, range);
    cursor = range.end;
  }
  if (cursor < value.length) fragment.append(document.createTextNode(value.slice(cursor)));
  root.replaceChildren(fragment);
}

export default function AgentAssetMentionEditor({
  value = '',
  references = [],
  disabled = false,
  placeholder = '',
  maxLength = 4000,
  ariaLabel = '资产提示词编辑器',
  mentionOpen = false,
  selectionOffset: initialSelectionOffset,
  onChange,
  onQueryChange,
  onBlur,
  onEditorKeyDown,
  onRemoveReference
}) {
  const editorRef = useRef(null);
  const composingRef = useRef(false);
  const restoreOffsetRef = useRef(null);
  const ranges = useMemo(() => promptMentionRanges(value, references), [value, references]);

  useLayoutEffect(() => {
    const root = editorRef.current;
    if (!root) return;
    const activeOffset = root.contains(window.getSelection()?.anchorNode) ? selectionOffset(root) : null;
    if (Number.isInteger(initialSelectionOffset)) restoreOffsetRef.current = initialSelectionOffset;
    else if (restoreOffsetRef.current === null && activeOffset !== null) restoreOffsetRef.current = activeOffset;
    renderEditor(root, value, ranges);
    if (restoreOffsetRef.current === null) return;
    setSelectionOffset(root, restoreOffsetRef.current);
    restoreOffsetRef.current = null;
  }, [initialSelectionOffset, ranges, value]);

  const emitQuery = useCallback(() => {
    const root = editorRef.current;
    if (!root) return;
    onQueryChange?.({ text: plainTextFromEditor(root), cursor: selectionOffset(root), rect: selectionRect(root) });
  }, [onQueryChange]);

  const emitTextChange = useCallback(() => {
    const root = editorRef.current;
    if (!root) return;
    const rawText = plainTextFromEditor(root);
    const text = rawText.slice(0, maxLength);
    const cursor = Math.min(selectionOffset(root), text.length);
    restoreOffsetRef.current = cursor;
    onChange?.(text, cursor);
    onQueryChange?.({ text, cursor, rect: selectionRect(root) });
  }, [maxLength, onChange, onQueryChange]);

  const handlePaste = useCallback(event => {
    event.preventDefault();
    const pasted = event.clipboardData.getData('text/plain');
    const selection = window.getSelection();
    if (!selection?.rangeCount) return;
    const range = selection.getRangeAt(0);
    range.deleteContents();
    const textNode = document.createTextNode(pasted);
    range.insertNode(textNode);
    range.setStartAfter(textNode);
    range.collapse(true);
    selection.removeAllRanges();
    selection.addRange(range);
    emitTextChange();
  }, [emitTextChange]);

  const handleKeyDown = useCallback(event => {
    onEditorKeyDown?.(event);
    if (event.defaultPrevented || disabled || !['Backspace', 'Delete'].includes(event.key)) return;
    const cursor = selectionOffset(editorRef.current);
    const boundary = mentionAtBoundary(ranges, cursor, event.key === 'Backspace' ? 'backspace' : 'delete');
    if (!boundary) return;
    event.preventDefault();
    const nextValue = removeAssetMention(value, boundary.asset);
    restoreOffsetRef.current = Math.min(boundary.start, nextValue.length);
    if (onRemoveReference) onRemoveReference?.(boundary.asset, nextValue);
    else onChange?.(nextValue, restoreOffsetRef.current);
  }, [disabled, onChange, onEditorKeyDown, onRemoveReference, ranges, value]);

  return <div
    ref={editorRef}
    className="agent-asset-mention-editor"
    contentEditable={!disabled}
    suppressContentEditableWarning
    role="textbox"
    aria-multiline="true"
    aria-label={ariaLabel}
    aria-expanded={mentionOpen}
    aria-disabled={disabled}
    data-placeholder={placeholder}
    onInput={() => { if (!composingRef.current) emitTextChange(); }}
    onKeyDown={handleKeyDown}
    onKeyUp={() => { if (!composingRef.current) emitQuery(); }}
    onClick={emitQuery}
    onPaste={handlePaste}
    onBlur={onBlur}
    onCompositionStart={() => { composingRef.current = true; }}
    onCompositionEnd={() => { composingRef.current = false; emitTextChange(); }}
  />;
}
