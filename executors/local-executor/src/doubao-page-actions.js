const DOUBAO_VIDEO_WORKSPACE_URL = 'https://www.doubao.com/chat/create-image';

class DoubaoControlError extends Error {
  constructor(code, message) {
    super(message);
    this.name = 'DoubaoControlError';
    this.code = code;
  }
}

class DoubaoPageActions {
  async setPrompt(webContents, prompt) {
    const text = String(prompt || '').trim();
    if (!text) throw new DoubaoControlError('PROMPT_REQUIRED', 'prompt is required');
    const result = await webContents.executeJavaScript(buildSetPromptScript(text), true);
    if (!result?.ok) {
      throw new DoubaoControlError('PROMPT_INPUT_NOT_FOUND', 'no usable Doubao prompt editor is visible');
    }
    return result;
  }

  async clickExactControl(webContents, text) {
    const wanted = String(text || '').trim();
    if (!wanted) throw new DoubaoControlError('CONTROL_TEXT_REQUIRED', 'control text is required');
    const result = await webContents.executeJavaScript(buildExactClickScript(wanted), true);
    if (result?.count === 0) {
      throw new DoubaoControlError('CONTROL_NOT_FOUND', `Doubao control not found: ${wanted}`);
    }
    if (result?.count !== 1 || !result?.clicked) {
      throw new DoubaoControlError('CONTROL_AMBIGUOUS', `Doubao control is ambiguous: ${wanted}`);
    }
    return true;
  }

  async openVideoWorkspace(webContents) {
    if (!webContents || typeof webContents.loadURL !== 'function') {
      throw new DoubaoControlError('VIDEO_WORKSPACE_NAVIGATION_UNAVAILABLE', 'Doubao account browser cannot navigate to the video workspace');
    }
    await webContents.loadURL(DOUBAO_VIDEO_WORKSPACE_URL);
    return true;
  }

  async openCreationWorkspace(webContents) {
    return this.clickOptionalPreferred(webContents, ['AI创作', 'AI 创作'], 'CREATION_ENTRY_AMBIGUOUS');
  }

  async openVideoMode(webContents) {
    return this.clickOptionalPreferred(webContents, ['视频', '视频生成', 'Seedance 2.0', 'Seedance 2.0 Fast'], 'VIDEO_ENTRY_AMBIGUOUS');
  }

  async clickOptionalPreferred(webContents, labels, ambiguousCode) {
    const result = await webContents.executeJavaScript(buildPreferredClickScript(labels), true);
    if (!result || result.count === 0) return false;
    if (result.count !== 1 || !result.clicked) {
      const detail = result?.rawCount > result?.count ? ` (${result.count}/${result.rawCount} actionable)` : '';
      throw new DoubaoControlError(ambiguousCode, `Doubao control is ambiguous: ${result.label || labels[0]}${detail}`);
    }
    return true;
  }

  async setReferenceImages(webContents, filePaths) {
    const files = (Array.isArray(filePaths) ? filePaths : []).map(value => String(value || '').trim()).filter(Boolean);
    if (files.length === 0) return false;
    const debug = webContents?.debugger;
    if (!debug) throw new DoubaoControlError('CDP_UNAVAILABLE', 'Chromium debugger is unavailable for reference image upload');
    if (!debug.isAttached?.()) debug.attach('1.3');
    await debug.sendCommand('DOM.enable');
    await debug.sendCommand('Runtime.enable');
    const evaluated = await debug.sendCommand('Runtime.evaluate', {
      expression: buildImageInputExpression(),
      returnByValue: false,
      awaitPromise: false,
      userGesture: true
    });
    const objectId = evaluated?.result?.objectId;
    if (!objectId) {
      throw new DoubaoControlError('REFERENCE_FILE_INPUT_NOT_FOUND', 'no compatible Doubao reference-image file input is available');
    }
    await debug.sendCommand('DOM.setFileInputFiles', { files, objectId });
    return true;
  }

  async submit(webContents) {
    const result = await webContents.executeJavaScript(buildPreferredClickScript([
      '生成视频',
      '开始制作',
      '立即生成',
      '开始生成',
      '生成',
      '发送'
    ]), true);
    if (result?.count === 0) throw new DoubaoControlError('SUBMIT_CONTROL_NOT_FOUND', 'no Doubao video submit control is visible');
    if (result?.count !== 1 || !result?.clicked) throw new DoubaoControlError('SUBMIT_CONTROL_AMBIGUOUS', 'Doubao video submit control is ambiguous');
    return true;
  }

  async confirmNormal(webContents) {
    const result = await webContents.executeJavaScript(buildPreferredClickScript([
      '确认生成',
      '继续生成',
      '确认并生成'
    ]), true);
    if (!result || result.count === 0) return false;
    if (result.count !== 1 || !result.clicked) {
      throw new DoubaoControlError('CONFIRM_CONTROL_AMBIGUOUS', 'Doubao generation confirmation control is ambiguous');
    }
    return true;
  }
}

function buildSetPromptScript(prompt) {
  return `(() => {
    const wanted = ${JSON.stringify(prompt)};
    const isVisible = element => {
      if (!element) return false;
      const style = getComputedStyle(element);
      const rect = element.getBoundingClientRect();
      return style.display !== 'none' && style.visibility !== 'hidden' && rect.width > 0 && rect.height > 0;
    };
    const candidates = Array.from(document.querySelectorAll('textarea,[contenteditable="true"]')).filter(isVisible);
    const score = element => {
      const hint = String(element.getAttribute('placeholder') || element.getAttribute('data-placeholder') || element.getAttribute('aria-label') || '');
      let value = 0;
      if (/提示词|描述|告诉豆包|输入/i.test(hint)) value += 10;
      if (element.tagName.toLowerCase() === 'textarea') value += 2;
      return value;
    };
    candidates.sort((a, b) => score(b) - score(a));
    const editor = candidates[0];
    if (!editor) return { ok: false, reason: 'not_found' };
    editor.focus();
    const kind = editor.tagName.toLowerCase() === 'textarea' ? 'textarea' : 'contenteditable';
    if (kind === 'textarea') {
      const descriptor = Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype, 'value');
      descriptor?.set?.call(editor, wanted);
      editor.dispatchEvent(new InputEvent('input', { bubbles: true, inputType: 'insertText', data: wanted }));
      editor.dispatchEvent(new Event('change', { bubbles: true }));
    } else {
      const selection = getSelection();
      const range = document.createRange();
      range.selectNodeContents(editor);
      selection.removeAllRanges();
      selection.addRange(range);
      let inserted = false;
      try { inserted = document.execCommand('insertText', false, wanted); } catch {}
      if (!inserted || String(editor.innerText || editor.textContent || '').trim() !== wanted) {
        editor.textContent = wanted;
        editor.dispatchEvent(new InputEvent('input', { bubbles: true, inputType: 'insertText', data: wanted }));
      }
    }
    return { ok: true, kind };
  })()`;
}

function buildExactClickScript(text) {
  return `(() => {
    const wanted = ${JSON.stringify(text)}.replace(/\\s+/g, ' ').trim();
    const isVisible = element => {
      const style = getComputedStyle(element);
      const rect = element.getBoundingClientRect();
      return style.display !== 'none' && style.visibility !== 'hidden' && rect.width > 0 && rect.height > 0;
    };
    const label = element => String(element.innerText || element.textContent || element.getAttribute('aria-label') || '').replace(/\\s+/g, ' ').trim();
    const matches = Array.from(document.querySelectorAll('button,[role="button"],[role="menuitem"],[role="option"],[role="tab"],label'))
      .filter(isVisible)
      .filter(element => !element.disabled && element.getAttribute('aria-disabled') !== 'true')
      .filter(element => label(element) === wanted);
    if (matches.length !== 1) return { count: matches.length, clicked: false };
    matches[0].scrollIntoView({ block: 'center', inline: 'center' });
    matches[0].click();
    return { count: 1, clicked: true };
  })()`;
}

function buildPreferredClickScript(labels) {
  return `(() => {
    const labels = ${JSON.stringify(labels)};
    const isVisible = element => {
      const style = getComputedStyle(element);
      const rect = element.getBoundingClientRect();
      return style.display !== 'none' && style.visibility !== 'hidden' && rect.width > 0 && rect.height > 0;
    };
    const isHitTestable = element => {
      const rect = element.getBoundingClientRect();
      const viewportWidth = Number(globalThis.innerWidth) || Number(document.documentElement?.clientWidth) || 0;
      const viewportHeight = Number(globalThis.innerHeight) || Number(document.documentElement?.clientHeight) || 0;
      if (viewportWidth > 0 && viewportHeight > 0) {
        const left = Math.max(0, Number(rect.left) || 0);
        const top = Math.max(0, Number(rect.top) || 0);
        const right = Math.min(viewportWidth, Number(rect.right) || (left + Number(rect.width || 0)));
        const bottom = Math.min(viewportHeight, Number(rect.bottom) || (top + Number(rect.height || 0)));
        if (right <= left || bottom <= top) return false;
        if (typeof document.elementFromPoint === 'function') {
          const hit = document.elementFromPoint((left + right) / 2, (top + bottom) / 2);
          if (!hit) return false;
          const containsHit = typeof element.contains === 'function' && element.contains(hit);
          const hitContains = typeof hit.contains === 'function' && hit.contains(element);
          if (hit !== element && !containsHit && !hitContains) return false;
        }
      }
      return true;
    };
    const isAriaHidden = element => {
      if (element.getAttribute?.('aria-hidden') === 'true') return true;
      const parent = typeof element.closest === 'function' ? element.closest('[aria-hidden="true"]') : null;
      return Boolean(parent);
    };
    const label = element => String(element.innerText || element.textContent || element.getAttribute('aria-label') || '').replace(/\\s+/g, ' ').trim();
    const controls = Array.from(document.querySelectorAll('button,[role="button"],[role="menuitem"],[role="option"],[role="tab"]'))
      .filter(isVisible)
      .filter(element => !element.disabled && element.getAttribute('aria-disabled') !== 'true')
      .filter(element => !isAriaHidden(element));
    for (const wanted of labels) {
      const rawMatches = controls.filter(element => label(element) === wanted);
      if (rawMatches.length === 0) continue;
      let matches = rawMatches.filter(isHitTestable);
      if (matches.length > 1) {
        const keyboardReachable = matches.filter(element => Number(element.tabIndex) >= 0);
        if (keyboardReachable.length === 1) matches = keyboardReachable;
      }
      if (matches.length === 0) continue;
      if (matches.length !== 1) return { count: matches.length, rawCount: rawMatches.length, clicked: false, label: wanted };
      matches[0].scrollIntoView({ block: 'center', inline: 'center' });
      matches[0].click();
      return { count: 1, rawCount: rawMatches.length, clicked: true, label: wanted };
    }
    return { count: 0, rawCount: 0, clicked: false };
  })()`;
}

function buildImageInputExpression() {
  return `(() => {
    const inputs = Array.from(document.querySelectorAll('input[type="file"]'));
    return inputs.find(input => /image/i.test(String(input.accept || ''))) || null;
  })()`;
}

module.exports = {
  DoubaoPageActions,
  DoubaoControlError,
  DOUBAO_VIDEO_WORKSPACE_URL,
  buildSetPromptScript,
  buildExactClickScript,
  buildPreferredClickScript,
  buildImageInputExpression
};
