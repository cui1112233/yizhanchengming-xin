const { ApiError } = require('./api-client');

class AmbiguousAcceptanceError extends Error {
  constructor(message = 'submission acceptance is still unknown') {
    super(message);
    this.name = 'AmbiguousAcceptanceError';
  }
}

class JobRunner {
  constructor({ api, token, accountPool, adapter, leaseRenewIntervalMs = 20000, retryDelayMs = 1000, maxSubmitAttempts = 3, maxRecoveryAttempts = 3, maxCompletionAttempts = 3, maxDownloadAttempts = 3, maxUploadAttempts = 3 }) {
    this.api = api;
    this.token = token;
    this.accountPool = accountPool;
    this.adapter = adapter;
    this.leaseRenewIntervalMs = leaseRenewIntervalMs;
    this.retryDelayMs = retryDelayMs;
    this.maxSubmitAttempts = maxSubmitAttempts;
    this.maxRecoveryAttempts = maxRecoveryAttempts;
    this.maxCompletionAttempts = maxCompletionAttempts;
    this.maxDownloadAttempts = maxDownloadAttempts;
    this.maxUploadAttempts = maxUploadAttempts;
  }

  async runClaim(claim) {
    const job = claim?.job;
    if (!job?.id) throw new Error('claim job is required');
    const lease = { leaseToken: claim.leaseToken, leaseGeneration: claim.leaseGeneration };
    const controller = new AbortController();
    let leaseError = null;
    let timer = null;
    if (this.leaseRenewIntervalMs > 0) {
      timer = setInterval(async () => {
        try {
          await this.api.renew(this.token, job.id, lease);
        } catch (error) {
          leaseError = error;
          controller.abort(error);
        }
      }, this.leaseRenewIntervalMs);
      timer.unref?.();
    }

    let account = null;
    let accepted = false;
    let accountReleaseState = 'available';
    try {
      account = this.accountPool.acquire({ jobId: job.id });
      if (!account) {
        await this.api.release(this.token, job.id, lease, 'no available local doubao account');
        return { released: true, reason: 'no_account' };
      }

      await this.api.progress(this.token, job.id, lease, 'preparing');
      this.throwLeaseError(leaseError);
      await this.adapter.prepare({ job, account, signal: controller.signal });
      this.throwLeaseError(leaseError);

      let submission = null;
      for (let attempt = 1; attempt <= this.maxSubmitAttempts; attempt++) {
        await this.api.progress(this.token, job.id, lease, 'submitting');
        this.throwLeaseError(leaseError);
        const outcome = await this.adapter.submit({ job, account, signal: controller.signal, attempt });
        this.throwLeaseError(leaseError);

        if (outcome?.status === 'accepted') {
          submission = outcome;
          break;
        }
        if (outcome?.status === 'unknown') {
          await this.api.progress(this.token, job.id, lease, 'acceptance_unknown');
          let recovered = outcome;
          for (let recovery = 1; recovery <= this.maxRecoveryAttempts; recovery++) {
            recovered = await this.adapter.recoverAcceptance({ job, account, signal: controller.signal, recovery });
            this.throwLeaseError(leaseError);
            if (recovered?.status !== 'unknown') break;
            await delay(this.retryDelayMs, controller.signal);
          }
          if (recovered?.status === 'accepted') {
            submission = recovered;
            break;
          }
          if (recovered?.status === 'not_accepted') {
            await this.api.release(this.token, job.id, lease, 'submission verified not accepted after recovery');
            return { released: true, reason: 'not_accepted_after_unknown' };
          }
          throw new AmbiguousAcceptanceError();
        }
        if (outcome?.status !== 'not_accepted') {
          throw new AmbiguousAcceptanceError('adapter returned an invalid acceptance status');
        }
        if (attempt < this.maxSubmitAttempts) await delay(this.retryDelayMs, controller.signal);
      }

      if (!submission?.submissionId) {
        await this.api.release(this.token, job.id, lease, 'submission not accepted after bounded retries');
        return { released: true, reason: 'not_accepted' };
      }

      await this.api.acceptance(this.token, job.id, lease, { accountId: account.id, submissionId: submission.submissionId });
      accepted = true;
      await this.api.progress(this.token, job.id, lease, 'generating');

      const completion = await retrySameOperation(
        () => this.adapter.waitForCompletion({ job, account, submissionId: submission.submissionId, signal: controller.signal }),
        this.maxCompletionAttempts,
        this.retryDelayMs,
        controller.signal
      );
      this.throwLeaseError(leaseError);

      await this.api.progress(this.token, job.id, lease, 'downloading');
      const artifact = await retrySameOperation(
        () => this.adapter.fetchArtifact({ job, account, completion, signal: controller.signal }),
        this.maxDownloadAttempts,
        this.retryDelayMs,
        controller.signal
      );
      this.throwLeaseError(leaseError);
      if (!artifact?.filePath) throw new Error('adapter did not return local artifact filePath');

      await this.api.progress(this.token, job.id, lease, 'uploading');
      const uploaded = await retrySameOperation(
        () => this.api.uploadArtifact(this.token, job.id, lease, artifact.filePath),
        this.maxUploadAttempts,
        this.retryDelayMs,
        controller.signal
      );
      this.throwLeaseError(leaseError);
      if (!uploaded?.artifactId) throw new Error('server did not return artifactId');

      await this.api.result(this.token, job.id, lease, uploaded.artifactId);
      return { succeeded: true, artifactId: uploaded.artifactId };
    } catch (error) {
      if (isAccountHoldState(error?.accountState)) accountReleaseState = error.accountState;
      if (isCancellation(error)) throw error;
      if (error instanceof AmbiguousAcceptanceError) {
        await safeCall(() => this.api.fail(this.token, job.id, lease, { code: 'ACCEPTANCE_UNKNOWN', message: error.message }));
        throw error;
      }
      if (accepted) {
        await safeCall(() => this.api.fail(this.token, job.id, lease, { code: 'ACCEPTED_FLOW_FAILED', message: error.message || String(error) }));
      } else {
        await safeCall(() => this.api.release(this.token, job.id, lease, error.message || 'pre-acceptance failure'));
      }
      throw error;
    } finally {
      if (timer) clearInterval(timer);
      await safeCall(() => this.adapter.dispose?.(job.id));
      if (account) this.accountPool.release(account.id, accountReleaseState);
    }
  }

  throwLeaseError(error) {
    if (error) throw error;
  }
}

function isAccountHoldState(state) {
  return state === 'human_verification' || state === 'quota_exhausted' || state === 'auth_required';
}

function isCancellation(error) {
  return error instanceof ApiError && error.status === 409 && /cancel/i.test(error.message || '');
}

async function retrySameOperation(fn, attempts, waitMs, signal) {
  let lastError;
  for (let i = 1; i <= attempts; i++) {
    try { return await fn(); }
    catch (error) {
      lastError = error;
      if (i < attempts) await delay(waitMs, signal);
    }
  }
  throw lastError;
}

async function safeCall(fn) { try { return await fn(); } catch { return undefined; } }

function delay(ms, signal) {
  if (!ms) return Promise.resolve();
  return new Promise((resolve, reject) => {
    if (signal?.aborted) return reject(signal.reason || new Error('aborted'));
    const timer = setTimeout(resolve, ms);
    signal?.addEventListener('abort', () => { clearTimeout(timer); reject(signal.reason || new Error('aborted')); }, { once: true });
  });
}

module.exports = { JobRunner, AmbiguousAcceptanceError, retrySameOperation, isCancellation, isAccountHoldState };
