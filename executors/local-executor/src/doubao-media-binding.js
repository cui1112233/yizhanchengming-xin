class MediaBindingError extends Error {
  constructor(code, message) {
    super(message);
    this.name = 'MediaBindingError';
    this.code = code;
  }
}

function bindExactMedia(submission = {}, candidates = []) {
  const acceptedIdentities = unique([
    submission.submissionId,
    ...(Array.isArray(submission.identities) ? submission.identities : [])
  ].map(normalizeIdentity).filter(Boolean));

  if (acceptedIdentities.length === 0) {
    throw new MediaBindingError('SUBMISSION_IDENTITY_REQUIRED', 'accepted Doubao submission identity is required');
  }

  const matches = (Array.isArray(candidates) ? candidates : []).filter(candidate => {
    const identities = unique([
      candidate.submissionId,
      candidate.messageId,
      candidate.taskId,
      candidate.mediaId,
      ...(Array.isArray(candidate.identities) ? candidate.identities : [])
    ].map(normalizeIdentity).filter(Boolean));
    return identities.some(identity => acceptedIdentities.includes(identity));
  });

  if (matches.length === 0) {
    throw new MediaBindingError('EXACT_MEDIA_NOT_FOUND', 'no generated media can be tied to the accepted Doubao submission');
  }
  if (matches.length > 1) {
    throw new MediaBindingError('EXACT_MEDIA_AMBIGUOUS', 'more than one generated media candidate matches the accepted Doubao submission');
  }
  if (!normalizeIdentity(matches[0].mediaId)) {
    throw new MediaBindingError('MEDIA_ID_REQUIRED', 'matched Doubao media has no stable media identity');
  }
  return matches[0];
}

function normalizeIdentity(value) {
  const text = String(value || '').trim();
  return text || null;
}

function unique(values) {
  return [...new Set(values)];
}

module.exports = { MediaBindingError, bindExactMedia, normalizeIdentity };
