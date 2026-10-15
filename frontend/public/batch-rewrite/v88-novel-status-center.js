'use strict';

(function initFactory(rootFactory) {
  const api = rootFactory();
  if (typeof module !== 'undefined' && module.exports) module.exports = api;
  if (typeof window !== 'undefined') window.V88NovelStatusCenter = api;
  if (typeof document !== 'undefined') api.install();
})(function factory() {
  const SOURCE_IDS = ['batchStatus', 'siteSubmitStatus', 'v78PreviewStatus', 'webLoginResult'];
  const CENTER_ID = 'v88NovelStatusCenter';
  const STYLE_ID = 'v88NovelStatusCenterStyle';

  function clean(value) {
    return String(value == null ? '' : value).replace(/\r/g, '').trim();
  }

  function localizeUserText(value) {
    return clean(value)
      .replace(/(^|[^0-9])121\s*网站提交(?=[^0-9]|$)/g, '$1视频管理系统提交')
      .replace(/(^|[^0-9])121\s*登录(?=[^0-9]|$)/g, '$1视频管理系统登录');
  }

  function numberAfter(text, labels) {
    for (const label of labels) {
      const re = new RegExp(label + '\\s*[：:]?\\s*(\\d+)', 'i');
      const match = text.match(re);
      if (match) return Number(match[1]) || 0;
    }
    return null;
  }

  function failureTotal(text) {
    const labels = [
      '系统规则\\/敏感词处理失败',
      '接口抓取失败',
      'AI文案生成失败',
      '提交失败',
      '处理失败'
    ];
    let sum = 0;
    let found = false;
    for (const label of labels) {
      const n = numberAfter(text, [label]);
      if (n != null) {
        found = true;
        sum += n;
      }
    }
    if (found) return sum;
    const generic = text.match(/(?:失败|错误|异常)\s*[：:]?\s*(\d+)/g) || [];
    return generic.reduce((total, part) => total + (Number((part.match(/\d+/) || ['0'])[0]) || 0), 0);
  }

  function firstMeaningfulLine(text) {
    return text.split(/\n+/).map(part => part.trim()).find(Boolean) || '';
  }

  function summarizeStatus(value) {
    const text = localizeUserText(value);
    if (!text) return { text: '', tone: 'idle' };

    const valid = numberAfter(text, ['有效任务']);
    const originalOk = numberAfter(text, ['原文处理成功']);
    const aiDocs = numberAfter(text, ['AI文案']);
    const failures = failureTotal(text);

    if (/处理完成/.test(text) && [valid, originalOk, aiDocs].some(v => v != null)) {
      const parts = ['处理完成'];
      if (valid != null) parts.push(`有效任务 ${valid}`);
      if (originalOk != null) parts.push(`原文成功 ${originalOk}`);
      if (aiDocs != null) parts.push(`AI文案 ${aiDocs}`);
      parts.push(`失败 ${failures}`);
      return { text: parts.join(' · '), tone: failures > 0 ? 'error' : 'success' };
    }

    const first = firstMeaningfulLine(text).replace(/\s+/g, ' ').slice(0, 96);
    const nonZeroFailure = failureTotal(text) > 0;
    if (nonZeroFailure || /(^|[^0-9])(失败|错误|异常|超时)(?!\s*[：:]?\s*0)/.test(first)) {
      return { text: first, tone: 'error' };
    }
    if (/正在|处理中|排队|提交中|验证中|生成中|抓取中|加载中|已提交.*排队/.test(first)) {
      return { text: first, tone: 'working' };
    }
    if (/待确认|警告|等待/.test(first)) return { text: first, tone: 'warning' };
    if (/完成|成功|已登录|已保存|已就绪/.test(first)) return { text: first, tone: 'success' };
    return { text: first, tone: 'info' };
  }

  function statusClass(tone) {
    const known = new Set(['idle', 'info', 'working', 'success', 'warning', 'error']);
    return `v88-novel-status-${known.has(tone) ? tone : 'info'}`;
  }

  function cssText() {
    return `
#${CENTER_ID}{position:absolute;left:50%;top:50%;transform:translate(-50%,-50%);z-index:20;max-width:min(900px,68vw);min-height:24px;display:flex;align-items:center;justify-content:center;text-align:center;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;font-size:13px;font-weight:600;letter-spacing:.1px;pointer-events:none;border:0!important;outline:0!important;box-shadow:none!important;background:transparent!important}
#${CENTER_ID}.v88-novel-status-idle{color:transparent}
#${CENTER_ID}.v88-novel-status-info{color:#a9b8cc}
#${CENTER_ID}.v88-novel-status-working{color:#4da3ff}
#${CENTER_ID}.v88-novel-status-success{color:#48c774}
#${CENTER_ID}.v88-novel-status-warning{color:#d9a441}
#${CENTER_ID}.v88-novel-status-error{color:#ef6a6a}
#${CENTER_ID}.v88-novel-status-working::before{content:'';width:7px;height:7px;border-radius:50%;margin-right:7px;background:currentColor;animation:v88NovelStatusPulse 1.1s ease-in-out infinite}
@keyframes v88NovelStatusPulse{0%,100%{opacity:.35;transform:scale(.8)}50%{opacity:1;transform:scale(1)}}
#batchStatus{border:0!important;outline:0!important;box-shadow:none!important;background:transparent!important}
`;
  }

  function ensureStyle(doc) {
    if (!doc || !doc.head || doc.getElementById(STYLE_ID)) return;
    const style = doc.createElement('style');
    style.id = STYLE_ID;
    style.textContent = cssText();
    doc.head.appendChild(style);
  }

  function exactNovelTitle(el) {
    return el && clean(el.textContent) === '小说获取' && el.children.length <= 2;
  }

  function findNovelHeaderHost(doc) {
    if (!doc || !doc.querySelectorAll) return null;
    const candidates = Array.from(doc.querySelectorAll('h1,h2,h3,strong,span,div')).filter(exactNovelTitle);
    for (const title of candidates) {
      let node = title.parentElement;
      for (let depth = 0; node && depth < 5; depth += 1, node = node.parentElement) {
        const rect = typeof node.getBoundingClientRect === 'function' ? node.getBoundingClientRect() : { width: 0, height: 0 };
        if (rect.width >= 520 && rect.height > 28 && rect.height <= 100) return node;
      }
    }
    return null;
  }

  function ensureCenterInDocument(doc) {
    if (!doc || !doc.body) return null;
    ensureStyle(doc);
    let center = doc.getElementById(CENTER_ID);
    if (center) return center;

    const host = findNovelHeaderHost(doc) || doc.body;
    center = doc.createElement('div');
    center.id = CENTER_ID;
    center.className = statusClass('idle');
    center.setAttribute('role', 'status');
    center.setAttribute('aria-live', 'polite');

    if (host === doc.body) {
      center.style.position = 'fixed';
      center.style.top = '16px';
    } else {
      try {
        const view = doc.defaultView;
        if (view && view.getComputedStyle(host).position === 'static') host.style.position = 'relative';
      } catch (_) {}
    }
    host.appendChild(center);
    neutralizeRedBars(doc, host, center);
    return center;
  }

  function parseRgb(value) {
    const m = String(value || '').match(/rgba?\((\d+)\D+(\d+)\D+(\d+)/i);
    return m ? [Number(m[1]), Number(m[2]), Number(m[3])] : null;
  }

  function isRedish(value) {
    const rgb = parseRgb(value);
    return !!(rgb && rgb[0] >= 170 && rgb[1] <= 130 && rgb[2] <= 130 && rgb[0] > rgb[1] * 1.25);
  }

  function neutralizeRedBars(doc, host, center) {
    if (!doc || !host || !doc.defaultView) return;
    const nodes = [host, ...Array.from(host.querySelectorAll ? host.querySelectorAll('div,span,section') : [])];
    for (const el of nodes) {
      if (!el || el === center) continue;
      let rect;
      let style;
      try { rect = el.getBoundingClientRect(); style = doc.defaultView.getComputedStyle(el); } catch (_) { continue; }
      if (!rect || rect.width < 240 || rect.height < 12 || rect.height > 64) continue;
      if (![style.borderTopColor, style.borderLeftColor, style.outlineColor].some(isRedish)) continue;
      el.style.setProperty('border-color', 'transparent', 'important');
      el.style.setProperty('outline', 'none', 'important');
      el.style.setProperty('box-shadow', 'none', 'important');
      if (!clean(el.textContent)) el.style.setProperty('background', 'transparent', 'important');
    }
  }

  function accessibleDocuments() {
    const docs = [];
    if (typeof document !== 'undefined') docs.push(document);
    try {
      if (typeof window !== 'undefined' && window.parent && window.parent !== window && window.parent.document) docs.unshift(window.parent.document);
    } catch (_) {}
    return Array.from(new Set(docs));
  }

  function renderStatus(value) {
    const summary = summarizeStatus(value);
    for (const doc of accessibleDocuments()) {
      const center = ensureCenterInDocument(doc);
      if (!center) continue;
      center.textContent = summary.text;
      center.className = statusClass(summary.tone);
      const host = center.parentElement;
      if (host) neutralizeRedBars(doc, host, center);
    }
    return summary;
  }

  function localizeTextNodes(doc) {
    if (!doc || !doc.createTreeWalker || !doc.body) return;
    const view = doc.defaultView;
    const showText = view && view.NodeFilter ? view.NodeFilter.SHOW_TEXT : 4;
    const walker = doc.createTreeWalker(doc.body, showText);
    let node;
    while ((node = walker.nextNode())) {
      const parent = node.parentElement;
      if (parent && ['SCRIPT', 'STYLE', 'TEXTAREA', 'PRE', 'CODE'].includes(parent.tagName)) continue;
      const next = localizeUserText(node.nodeValue || '');
      if (next !== node.nodeValue) node.nodeValue = next;
    }
  }

  function bestSource(doc) {
    for (const id of SOURCE_IDS) {
      const el = doc && doc.getElementById ? doc.getElementById(id) : null;
      if (el && clean(el.textContent)) return el;
    }
    return null;
  }

  function syncFromSources() {
    if (typeof document === 'undefined') return;
    const source = bestSource(document);
    if (source) renderStatus(source.textContent);
  }

  function installObserver(doc) {
    if (!doc || !doc.documentElement || typeof MutationObserver === 'undefined') return;
    const observer = new MutationObserver(mutations => {
      let shouldSync = false;
      for (const mutation of mutations) {
        const target = mutation.target && mutation.target.nodeType === 1 ? mutation.target : mutation.target && mutation.target.parentElement;
        if (target && SOURCE_IDS.some(id => target.id === id || (target.closest && target.closest(`#${id}`)))) shouldSync = true;
      }
      localizeTextNodes(doc);
      if (shouldSync) syncFromSources();
      const center = doc.getElementById(CENTER_ID);
      const host = center && center.parentElement;
      if (center && host) neutralizeRedBars(doc, host, center);
    });
    observer.observe(doc.documentElement, { childList: true, subtree: true, characterData: true });
  }

  function install() {
    const run = () => {
      for (const doc of accessibleDocuments()) {
        ensureCenterInDocument(doc);
        localizeTextNodes(doc);
        installObserver(doc);
      }
      syncFromSources();
    };
    if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', run, { once: true });
    else run();
  }

  return {
    localizeUserText,
    summarizeStatus,
    statusClass,
    failureTotal,
    cssText,
    isRedish,
    install
  };
});
