const { DoubaoPageProbe } = require('./doubao-page-probe');
const { DoubaoPageActions } = require('./doubao-page-actions');
const { DoubaoNetworkTracker } = require('./doubao-network-tracker');
const { classifyPageState, detectCapabilities, selectRequestedOptions } = require('./doubao-page-model');
const { determineSubmissionOutcome } = require('./doubao-acceptance');
const { bindExactMedia, MediaBindingError } = require('./doubao-media-binding');
const { downloadMp4WithSession } = require('./doubao-artifact-downloader');

class DoubaoAccountError extends Error {
  constructor(accountState, message) {
    super(message || accountState);
    this.name = 'DoubaoAccountError';
    this.code = `DOUBAO_${String(accountState || 'error').toUpperCase()}`;
    this.accountState = accountState;
  }
}

class DoubaoResultError extends Error {
  constructor(code, message) {
    super(message);
    this.name = 'DoubaoResultError';
    this.code = code;
  }
}

class DoubaoAdapter {
  constructor({
    accountWindows,
    pageProbe = new DoubaoPageProbe(),
    pageActions = new DoubaoPageActions(),
    trackerFactory = webContents => new DoubaoNetworkTracker({ webContents }),
    downloader = downloadMp4WithSession,
    downloadDir = '',
    sleep = delay,
    navigationDelayMs = 500,
    videoWorkspacePollMs = 1000,
    maxVideoWorkspacePolls = 12,
    confirmationDelayMs = 350,
    completionPollMs = 2000,
    maxCompletionPolls = 900
  } = {}) {
    if (!accountWindows || (!accountWindows.getWebContents && !accountWindows.ensureWebContents)) throw new Error('accountWindows is required');
    this.accountWindows = accountWindows;
    this.pageProbe = pageProbe;
    this.pageActions = pageActions;
    this.trackerFactory = trackerFactory;
    this.downloader = downloader;
    this.downloadDir = downloadDir;
    this.sleep = sleep;
    this.navigationDelayMs = navigationDelayMs;
    this.videoWorkspacePollMs = Math.max(0, Number(videoWorkspacePollMs) || 0);
    this.maxVideoWorkspacePolls = Math.max(1, Number(maxVideoWorkspacePolls) || 1);
    this.confirmationDelayMs = confirmationDelayMs;
    this.completionPollMs = completionPollMs;
    this.maxCompletionPolls = maxCompletionPolls;
    this.jobs = new Map();
  }

  async prepare({ job, account, signal }) {
    throwIfAborted(signal);
    if (!job?.id || !account?.id) throw new Error('job and account are required');
    this.dispose(job.id);
    const payload = payloadOf(job);
    const webContents = await this.getAccountWebContents(account.id);
    const workspace = await this.captureVideoWorkspace(webContents, signal);
    const before = workspace.snapshot;
    const selected = selectRequestedOptions(workspace.capabilities, payload);

    if (selected.model) await this.pageActions.clickExactControl(webContents, selected.model);
    if (selected.duration !== null) await this.pageActions.clickExactControl(webContents, `${selected.duration}秒`);
    if (selected.ratio) await this.pageActions.clickExactControl(webContents, selected.ratio);
    if (selected.needsImageUpload) await this.pageActions.setReferenceImages(webContents, selected.images);
    await this.pageActions.setPrompt(webContents, selected.prompt);
    throwIfAborted(signal);

    this.jobs.set(job.id, {
      jobId: job.id,
      accountId: account.id,
      webContents,
      payload,
      selected,
      before,
      tracker: null,
      submission: null
    });
    return selected;
  }

  async getAccountWebContents(accountId) {
    if (typeof this.accountWindows.ensureWebContents === 'function') {
      return this.accountWindows.ensureWebContents(accountId);
    }
    return this.accountWindows.getWebContents(accountId);
  }

  async captureVideoWorkspace(webContents, signal) {
    let snapshot = await this.pageProbe.capture(webContents);
    assertUsableAccount(snapshot);
    let capabilities = detectCapabilities(snapshot);
    if (capabilities.videoGeneration) return { snapshot, capabilities };

    if (isDoubaoWorkHome(snapshot) && typeof this.pageActions.openVideoWorkspace === 'function') {
      const navigated = await this.pageActions.openVideoWorkspace(webContents);
      if (navigated) {
        const workspace = await this.waitForVideoWorkspace(webContents, signal);
        snapshot = workspace.snapshot;
        capabilities = workspace.capabilities;
        if (capabilities.videoGeneration) return workspace;
      }
    }

    if (typeof this.pageActions.openCreationWorkspace === 'function') {
      const clicked = await this.pageActions.openCreationWorkspace(webContents);
      if (clicked && this.navigationDelayMs > 0) await this.sleep(this.navigationDelayMs, signal);
      throwIfAborted(signal);
      snapshot = await this.pageProbe.capture(webContents);
      assertUsableAccount(snapshot);
      capabilities = detectCapabilities(snapshot);
      if (capabilities.videoGeneration) return { snapshot, capabilities };
    }

    if (typeof this.pageActions.openVideoMode === 'function') {
      const clicked = await this.pageActions.openVideoMode(webContents);
      if (clicked && this.navigationDelayMs > 0) await this.sleep(this.navigationDelayMs, signal);
      throwIfAborted(signal);
      snapshot = await this.pageProbe.capture(webContents);
      assertUsableAccount(snapshot);
      capabilities = detectCapabilities(snapshot);
      if (clicked && !capabilities.videoGeneration) {
        const workspace = await this.waitForVideoWorkspace(webContents, signal);
        snapshot = workspace.snapshot;
        capabilities = workspace.capabilities;
      }
    }
    return { snapshot, capabilities };
  }

  async waitForVideoWorkspace(webContents, signal) {
    let snapshot = null;
    let capabilities = {};
    for (let poll = 0; poll < this.maxVideoWorkspacePolls; poll++) {
      throwIfAborted(signal);
      snapshot = await this.pageProbe.capture(webContents);
      assertUsableAccount(snapshot);
      capabilities = detectCapabilities(snapshot);
      if (capabilities.videoGeneration) return { snapshot, capabilities };
      if (poll + 1 < this.maxVideoWorkspacePolls && this.videoWorkspacePollMs > 0) {
        await this.sleep(this.videoWorkspacePollMs, signal);
      }
    }
    return { snapshot: snapshot || {}, capabilities };
  }

  async submit({ job, account, signal }) {
    const state = this.requirePrepared(job, account);
    throwIfAborted(signal);
    if (state.submission?.status === 'accepted') return { ...state.submission };

    state.tracker?.stop?.();
    state.tracker = this.trackerFactory(state.webContents);
    await state.tracker.startAttempt({ prompt: state.selected.prompt });
    await this.pageActions.submit(state.webContents);
    if (this.confirmationDelayMs > 0) await this.sleep(this.confirmationDelayMs, signal);
    throwIfAborted(signal);
    await this.pageActions.confirmNormal(state.webContents);
    const after = await this.pageProbe.capture(state.webContents);
    assertUsableAccount(after);
    const outcome = determineSubmissionOutcome({
      prompt: state.selected.prompt,
      before: state.before,
      after,
      networkEvidence: state.tracker.getEvidence?.()
    });
    if (outcome.status === 'accepted') state.submission = { ...outcome };
    if (outcome.status === 'not_accepted') {
      state.tracker.stop?.();
      state.tracker = null;
      state.before = after;
    }
    return outcome;
  }

  async recoverAcceptance({ job, account, signal }) {
    const state = this.requirePrepared(job, account);
    throwIfAborted(signal);
    if (state.submission?.status === 'accepted') return { ...state.submission };
    if (!state.tracker) return { status: 'unknown' };
    const after = await this.pageProbe.capture(state.webContents);
    assertUsableAccount(after);
    const outcome = determineSubmissionOutcome({
      prompt: state.selected.prompt,
      before: state.before,
      after,
      networkEvidence: state.tracker.getEvidence?.()
    });
    if (outcome.status === 'accepted') state.submission = { ...outcome };
    if (outcome.status === 'not_accepted') {
      state.tracker.stop?.();
      state.tracker = null;
      state.before = after;
    }
    return outcome;
  }

  async waitForCompletion({ job, account, submissionId, signal }) {
    const state = this.requireAccepted(job, account, submissionId);
    for (let poll = 0; poll < this.maxCompletionPolls; poll++) {
      throwIfAborted(signal);
      const snapshot = await this.pageProbe.capture(state.webContents);
      assertUsableAccount(snapshot);
      const candidates = mergeMediaCandidates(
        state.tracker?.getMediaCandidates?.() || [],
        snapshotVideoCandidates(snapshot)
      );
      try {
        return bindExactMedia(state.submission, candidates);
      } catch (error) {
        if (!(error instanceof MediaBindingError) || error.code !== 'EXACT_MEDIA_NOT_FOUND') throw error;
      }
      if (poll + 1 < this.maxCompletionPolls) await this.sleep(this.completionPollMs, signal);
    }
    throw new DoubaoResultError('RESULT_TIMEOUT', 'exact Doubao video did not become available before timeout');
  }

  async fetchArtifact({ job, account, completion, signal }) {
    const state = this.requireAccepted(job, account, stateSubmissionId(this.jobs.get(job?.id)));
    throwIfAborted(signal);
    const candidates = mergeMediaCandidates(
      state.tracker?.getMediaCandidates?.() || [],
      completion ? [completion] : []
    );
    const media = bindExactMedia(state.submission, candidates);
    const downloadUrl = String(media.downloadUrl || media.originalUrl || '').trim();
    if (!/^https?:\/\//i.test(downloadUrl)) {
      throw new DoubaoResultError('OFFICIAL_DOWNLOAD_NOT_READY', 'exact Doubao media has no official http(s) download URL yet');
    }
    const artifact = await this.downloader({
      webContents: state.webContents,
      url: downloadUrl,
      downloadDir: this.downloadDir,
      jobId: job.id,
      mediaId: media.mediaId,
      media,
      signal
    });
    if (!artifact?.filePath) throw new DoubaoResultError('DOWNLOAD_FILE_MISSING', 'Doubao downloader did not return a local MP4 file');
    state.tracker?.stop?.();
    state.tracker = null;
    return artifact;
  }

  dispose(jobId) {
    const id = String(jobId || '').trim();
    const state = this.jobs.get(id);
    state?.tracker?.stop?.();
    this.jobs.delete(id);
  }

  requirePrepared(job, account) {
    const state = this.jobs.get(job?.id);
    if (!state || state.accountId !== account?.id) throw new Error('Doubao job was not prepared on this account');
    return state;
  }

  requireAccepted(job, account, submissionId) {
    const state = this.requirePrepared(job, account);
    if (state.submission?.status !== 'accepted') throw new DoubaoResultError('SUBMISSION_NOT_ACCEPTED', 'Doubao submission is not accepted');
    if (submissionId && state.submission.submissionId !== submissionId) {
      throw new DoubaoResultError('SUBMISSION_ID_MISMATCH', 'Doubao submission identity changed unexpectedly');
    }
    return state;
  }
}

function isDoubaoWorkHome(snapshot = {}) {
  const text = String(snapshot.visibleText || '').replace(/\s+/g, ' ').trim();
  if (/今天有什么工作要处理/.test(text)) return true;
  return /豆包\s*工作/.test(text) && /(新工作任务|定时任务|技能.{0,8}连接器)/.test(text);
}

function assertUsableAccount(snapshot) {
  const state = classifyPageState(snapshot);
  if (state === 'available') return;
  const messages = {
    human_verification: 'Doubao requires human verification',
    quota_exhausted: 'Doubao daily video quota is exhausted',
    auth_required: 'Doubao account login is required'
  };
  throw new DoubaoAccountError(state, messages[state] || `Doubao account is unavailable: ${state}`);
}

function snapshotVideoCandidates(snapshot = {}) {
  return (Array.isArray(snapshot.videos) ? snapshot.videos : []).map(video => ({
    mediaId: video.mediaId,
    identities: Array.isArray(video.identities) ? video.identities : [],
    srcKind: video.srcKind
  }));
}

function mergeMediaCandidates(...groups) {
  const byMedia = new Map();
  const loose = [];
  for (const candidate of groups.flatMap(group => Array.isArray(group) ? group : [])) {
    if (!candidate) continue;
    const mediaId = String(candidate.mediaId || '').trim();
    if (!mediaId) {
      loose.push(candidate);
      continue;
    }
    const previous = byMedia.get(mediaId);
    if (!previous) {
      byMedia.set(mediaId, { ...candidate });
      continue;
    }
    byMedia.set(mediaId, {
      ...previous,
      ...candidate,
      identities: [...new Set([...(previous.identities || []), ...(candidate.identities || [])])],
      downloadUrl: candidate.downloadUrl || previous.downloadUrl,
      originalUrl: candidate.originalUrl || previous.originalUrl
    });
  }
  return [...byMedia.values(), ...loose];
}

function payloadOf(job = {}) {
  if (job.payload && typeof job.payload === 'object' && !Buffer.isBuffer(job.payload)) return job.payload;
  if (typeof job.payload === 'string') {
    try { return JSON.parse(job.payload); } catch { throw new Error('job payload is invalid JSON'); }
  }
  return {};
}

function stateSubmissionId(state) {
  return state?.submission?.submissionId || '';
}

function throwIfAborted(signal) {
  if (signal?.aborted) throw signal.reason || new Error('aborted');
}

function delay(ms, signal) {
  if (!ms) return Promise.resolve();
  return new Promise((resolve, reject) => {
    if (signal?.aborted) return reject(signal.reason || new Error('aborted'));
    const timer = setTimeout(resolve, ms);
    signal?.addEventListener('abort', () => {
      clearTimeout(timer);
      reject(signal.reason || new Error('aborted'));
    }, { once: true });
  });
}

module.exports = {
  DoubaoAdapter,
  DoubaoAccountError,
  DoubaoResultError,
  isDoubaoWorkHome,
  assertUsableAccount,
  snapshotVideoCandidates,
  mergeMediaCandidates,
  payloadOf
};
