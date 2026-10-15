const STABLE_ID_KEYS = new Set([
  'conversationid',
  'messageid',
  'taskid',
  'generationid',
  'mediaid',
  'videoid'
]);

const MEDIA_ID_KEYS = new Set(['mediaid', 'videoid']);

class DoubaoNetworkTracker {
  constructor({ webContents }) {
    if (!webContents?.debugger) throw new Error('Doubao webContents debugger is required');
    this.webContents = webContents;
    this.debugger = webContents.debugger;
    this.attachedByUs = false;
    this.listening = false;
    this.attempt = null;
    this.onMessage = (_event, method, params) => {
      this.handleMessage(method, params).catch(() => {});
    };
  }

  async startAttempt({ prompt }) {
    const promptNeedle = normalizeText(prompt);
    if (!promptNeedle) throw new Error('prompt is required for network acceptance tracking');
    await this.ensureAttached();
    if (!this.listening) {
      this.debugger.on('message', this.onMessage);
      this.listening = true;
    }
    await this.debugger.sendCommand('Network.enable');
    this.attempt = {
      prompt: promptNeedle,
      requests: new Map(),
      responses: new Map(),
      identities: [],
      mediaCandidates: [],
      accepted: false,
      endpoint: null
    };
  }

  async ensureAttached() {
    if (this.debugger.isAttached?.()) return;
    this.debugger.attach('1.3');
    this.attachedByUs = true;
  }

  async handleMessage(method, params = {}) {
    if (!this.attempt) return;
    if (method === 'Network.requestWillBeSent') {
      const request = params.request || {};
      if (!postDataContainsPrompt(request.postData, this.attempt.prompt)) return;
      const identities = extractIdentityEvidence(request.postData);
      this.attempt.requests.set(params.requestId, {
        status: null,
        mimeType: '',
        endpoint: sanitizeNetworkUrl(request.url),
        identities
      });
      this.mergeIdentities(identities);
      return;
    }

    if (method === 'Network.responseReceived') {
      const response = params.response || {};
      const responseMeta = {
        status: Number(response.status) || 0,
        mimeType: String(response.mimeType || ''),
        url: String(response.url || '')
      };
      this.attempt.responses.set(params.requestId, responseMeta);
      const tracked = this.attempt.requests.get(params.requestId);
      if (tracked) {
        tracked.status = responseMeta.status;
        tracked.mimeType = responseMeta.mimeType;
        tracked.endpoint = sanitizeNetworkUrl(responseMeta.url || tracked.endpoint);
        if (tracked.status >= 200 && tracked.status < 300 && tracked.identities.length > 0) {
          this.attempt.accepted = true;
          this.attempt.endpoint = tracked.endpoint;
        }
      }
      return;
    }

    if (method !== 'Network.loadingFinished') return;
    const responseMeta = this.attempt.responses.get(params.requestId);
    if (!responseMeta || !(responseMeta.status >= 200 && responseMeta.status < 300)) return;
    if (!/json|text|javascript/i.test(responseMeta.mimeType || '')) return;

    let text = '';
    try {
      const bodyResult = await this.debugger.sendCommand('Network.getResponseBody', { requestId: params.requestId });
      text = decodeResponseBody(bodyResult);
    } catch {
      return;
    }

    const tracked = this.attempt.requests.get(params.requestId);
    if (tracked) {
      const identities = extractIdentityEvidence(text);
      tracked.identities = unique([...tracked.identities, ...identities]);
      this.mergeIdentities(identities);
      if (this.attempt.identities.length > 0) {
        this.attempt.accepted = true;
        this.attempt.endpoint = tracked.endpoint;
      }
    }

    if (this.attempt.accepted && this.attempt.identities.length > 0) {
      const accepted = new Set(this.attempt.identities);
      for (const candidate of extractMediaCandidates(text)) {
        if (!(candidate.identities || []).some(identity => accepted.has(identity))) continue;
        this.addMediaCandidate(candidate);
      }
    }
  }

  mergeIdentities(values) {
    this.attempt.identities = unique([...this.attempt.identities, ...values]);
  }

  addMediaCandidate(candidate) {
    const key = `${candidate.mediaId}\u0000${candidate.downloadUrl}`;
    const existing = new Set(this.attempt.mediaCandidates.map(item => `${item.mediaId}\u0000${item.downloadUrl}`));
    if (!existing.has(key)) this.attempt.mediaCandidates.push({ ...candidate, identities: [...candidate.identities] });
  }

  getEvidence() {
    if (!this.attempt) return { accepted: false, identities: [] };
    const evidence = {
      accepted: Boolean(this.attempt.accepted),
      identities: [...this.attempt.identities]
    };
    if (this.attempt.endpoint) evidence.endpoint = this.attempt.endpoint;
    return evidence;
  }

  getMediaCandidates() {
    if (!this.attempt) return [];
    return this.attempt.mediaCandidates.map(candidate => ({
      mediaId: candidate.mediaId,
      identities: [...candidate.identities],
      downloadUrl: candidate.downloadUrl
    }));
  }

  stop() {
    this.attempt = null;
    if (this.listening) {
      this.debugger.removeListener?.('message', this.onMessage);
      this.listening = false;
    }
    if (this.attachedByUs && this.debugger.isAttached?.()) {
      try { this.debugger.detach(); } catch {}
    }
    this.attachedByUs = false;
  }
}

function extractIdentityEvidence(input) {
  if (input === null || input === undefined) return [];
  if (typeof input === 'string') {
    const text = input.slice(0, 2_000_000);
    try {
      return extractIdentityEvidence(JSON.parse(text));
    } catch {
      const ids = [];
      const re = /["']?(conversation[_-]?id|message[_-]?id|task[_-]?id|generation[_-]?id|media[_-]?id|video[_-]?id)["']?\s*[:=]\s*["']([^"'\s,}]{1,160})["']/gi;
      let match;
      while ((match = re.exec(text))) ids.push(match[2]);
      return unique(ids);
    }
  }

  const ids = [];
  walk(input, (key, value) => {
    if (!STABLE_ID_KEYS.has(normalizeKey(key))) return;
    if (typeof value !== 'string' && typeof value !== 'number') return;
    const identity = String(value).trim();
    if (identity && identity.length <= 160) ids.push(identity);
  });
  return unique(ids);
}

function extractMediaCandidates(input) {
  let value = input;
  if (typeof input === 'string') {
    try { value = JSON.parse(input.slice(0, 2_000_000)); }
    catch { return []; }
  }
  const candidates = [];
  collectMediaCandidates(value, candidates);
  return dedupeCandidates(candidates);
}

function collectMediaCandidates(value, output) {
  if (Array.isArray(value)) {
    for (const item of value) collectMediaCandidates(item, output);
    return;
  }
  if (!value || typeof value !== 'object') return;

  const directIdentities = [];
  let mediaId = null;
  const urls = [];
  for (const [key, child] of Object.entries(value)) {
    const normalized = normalizeKey(key);
    if (STABLE_ID_KEYS.has(normalized) && (typeof child === 'string' || typeof child === 'number')) {
      const identity = String(child).trim();
      if (identity && identity.length <= 160) {
        directIdentities.push(identity);
        if (MEDIA_ID_KEYS.has(normalized)) mediaId = identity;
      }
    }
    if (typeof child === 'string') {
      const score = videoUrlScore(normalized);
      if (score > 0 && /^https?:\/\//i.test(child.trim())) urls.push({ url: child.trim(), score });
    }
  }
  const selectedUrl = selectStrongVideoUrl(urls);
  if (mediaId && directIdentities.length > 0 && selectedUrl) {
    output.push({ mediaId, identities: unique(directIdentities), downloadUrl: selectedUrl });
  }

  for (const child of Object.values(value)) collectMediaCandidates(child, output);
}

function videoUrlScore(normalizedKey) {
  if (!normalizedKey || /(cover|poster|thumb|thumbnail|avatar|image)/.test(normalizedKey)) return 0;
  if (/(original|origin|source|download).*(url|uri)|(url|uri).*(original|origin|source|download)/.test(normalizedKey)) return 5;
  if (/^(downloadurl|originalurl|originurl|sourceurl)$/.test(normalizedKey)) return 5;
  if (/video.*(url|uri)|(url|uri).*video/.test(normalizedKey)) return 4;
  if (/playback.*(url|uri)/.test(normalizedKey)) return 3;
  return 0;
}

function selectStrongVideoUrl(urls) {
  if (!urls.length) return null;
  const topScore = Math.max(...urls.map(item => item.score));
  const top = unique(urls.filter(item => item.score === topScore).map(item => item.url));
  return top.length === 1 ? top[0] : null;
}

function dedupeCandidates(candidates) {
  const seen = new Set();
  const out = [];
  for (const candidate of candidates) {
    const key = `${candidate.mediaId}\u0000${candidate.downloadUrl}`;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push(candidate);
  }
  return out;
}

function walk(value, visit) {
  if (Array.isArray(value)) {
    for (const item of value) walk(item, visit);
    return;
  }
  if (!value || typeof value !== 'object') return;
  for (const [key, child] of Object.entries(value)) {
    visit(key, child);
    walk(child, visit);
  }
}

function normalizeKey(key) {
  return String(key || '').toLowerCase().replace(/[^a-z0-9]/g, '');
}

function postDataContainsPrompt(postData, prompt) {
  const text = String(postData || '');
  if (!text) return false;
  if (normalizeText(text).includes(prompt)) return true;
  try {
    return normalizeText(decodeURIComponent(text)).includes(prompt);
  } catch {
    return false;
  }
}

function decodeResponseBody(result = {}) {
  const body = String(result.body || '');
  if (!result.base64Encoded) return body.slice(0, 2_000_000);
  return Buffer.from(body, 'base64').toString('utf8').slice(0, 2_000_000);
}

function sanitizeNetworkUrl(value) {
  try {
    const url = new URL(String(value || ''));
    return `${url.origin}${url.pathname}`;
  } catch {
    return String(value || '').split(/[?#]/, 1)[0];
  }
}

function normalizeText(value) {
  return String(value || '').replace(/\s+/g, ' ').trim();
}

function unique(values) {
  return [...new Set(values.filter(Boolean))];
}

module.exports = {
  DoubaoNetworkTracker,
  extractIdentityEvidence,
  extractMediaCandidates,
  sanitizeNetworkUrl,
  postDataContainsPrompt,
  decodeResponseBody,
  videoUrlScore,
  selectStrongVideoUrl
};
