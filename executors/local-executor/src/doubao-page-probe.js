class DoubaoPageProbe {
  async capture(webContents) {
    if (!webContents?.executeJavaScript) throw new Error('Doubao webContents is required');
    const raw = await webContents.executeJavaScript(buildSnapshotScript(), false);
    return sanitizeSnapshot(raw);
  }
}

function buildSnapshotScript() {
  return `(() => {
    const isVisible = element => {
      if (!element || !(element instanceof Element)) return false;
      const style = getComputedStyle(element);
      const rect = element.getBoundingClientRect();
      return style.visibility !== 'hidden' && style.display !== 'none' && rect.width > 0 && rect.height > 0;
    };
    const clean = value => String(value || '').replace(/\\s+/g, ' ').trim();
    const textOf = element => clean(element.innerText || element.textContent || '');
    const collectIds = element => {
      const ids = [];
      let node = element;
      for (let depth = 0; node && depth < 6; depth++, node = node.parentElement) {
        for (const attr of Array.from(node.attributes || [])) {
          if (!/^data-/i.test(attr.name)) continue;
          const key = attr.name.toLowerCase();
          if (!/(message|task|conversation|media|video|generation).*(id)|data-id$/.test(key)) continue;
          const value = clean(attr.value);
          if (value && value.length <= 160) ids.push(value);
        }
      }
      return Array.from(new Set(ids)).slice(0, 16);
    };

    const controls = Array.from(document.querySelectorAll('button,[role="button"],[role="menuitem"],[role="option"],[role="tab"],label'))
      .filter(isVisible)
      .slice(0, 240)
      .map(element => ({
        role: element.getAttribute('role') || element.tagName.toLowerCase(),
        text: textOf(element).slice(0, 240),
        aria: clean(element.getAttribute('aria-label')).slice(0, 240),
        title: clean(element.getAttribute('title')).slice(0, 240),
        disabled: Boolean(element.disabled || element.getAttribute('aria-disabled') === 'true')
      }));

    const promptInputs = Array.from(document.querySelectorAll('textarea,[contenteditable="true"]'))
      .filter(isVisible)
      .slice(0, 32)
      .map(element => ({
        kind: element.tagName.toLowerCase() === 'textarea' ? 'textarea' : 'contenteditable',
        placeholder: clean(element.getAttribute('placeholder') || element.getAttribute('data-placeholder')).slice(0, 240),
        aria: clean(element.getAttribute('aria-label')).slice(0, 240)
      }));

    const fileInputs = Array.from(document.querySelectorAll('input[type="file"]'))
      .slice(0, 32)
      .map(element => ({
        accept: clean(element.getAttribute('accept')).slice(0, 240),
        multiple: Boolean(element.multiple)
      }));

    const identityNodes = Array.from(document.querySelectorAll('[data-message-id],[data-task-id],[data-conversation-id],[data-media-id],[data-video-id],[data-generation-id]'))
      .filter(isVisible)
      .slice(0, 160)
      .map(element => ({
        identities: collectIds(element),
        text: textOf(element).slice(0, 500)
      }))
      .filter(item => item.identities.length > 0);

    const videos = Array.from(document.querySelectorAll('video'))
      .filter(isVisible)
      .slice(0, 64)
      .map(video => {
        const identities = collectIds(video);
        const mediaId = identities.find(value => value) || null;
        const src = clean(video.currentSrc || video.src || '');
        let srcKind = 'none';
        if (/^https?:/i.test(src)) srcKind = 'https';
        else if (/^blob:/i.test(src)) srcKind = 'blob';
        return { mediaId, identities, srcKind };
      });

    return {
      visibleText: clean(document.body?.innerText || '').slice(0, 20000),
      controls,
      promptInputs,
      fileInputs,
      identityNodes,
      videos
    };
  })()`;
}

function sanitizeSnapshot(raw = {}) {
  return {
    visibleText: String(raw.visibleText || '').slice(0, 20000),
    controls: limit(raw.controls, 200).map(item => ({
      role: short(item?.role),
      text: short(item?.text),
      aria: short(item?.aria),
      title: short(item?.title),
      disabled: Boolean(item?.disabled)
    })),
    promptInputs: limit(raw.promptInputs, 32).map(item => ({
      kind: short(item?.kind),
      placeholder: short(item?.placeholder),
      aria: short(item?.aria)
    })),
    fileInputs: limit(raw.fileInputs, 32).map(item => ({
      accept: short(item?.accept),
      multiple: Boolean(item?.multiple)
    })),
    identityNodes: limit(raw.identityNodes, 160).map(item => ({
      identities: limit(item?.identities, 16).map(identity).filter(Boolean),
      text: String(item?.text || '').slice(0, 500)
    })),
    videos: limit(raw.videos, 64).map(item => ({
      mediaId: identity(item?.mediaId),
      identities: limit(item?.identities, 16).map(identity).filter(Boolean),
      srcKind: ['https', 'blob', 'none'].includes(item?.srcKind) ? item.srcKind : 'none'
    }))
  };
}

function limit(value, max) {
  return (Array.isArray(value) ? value : []).slice(0, max);
}

function short(value) {
  return String(value || '').replace(/\s+/g, ' ').trim().slice(0, 240);
}

function identity(value) {
  const text = String(value || '').trim();
  return text && text.length <= 160 ? text : null;
}

module.exports = { DoubaoPageProbe, buildSnapshotScript, sanitizeSnapshot };
