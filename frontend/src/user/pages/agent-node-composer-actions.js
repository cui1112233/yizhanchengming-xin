export function compactComposerActionPlan({
  kind,
  source,
  resourceType,
  canRetry = false,
  hasImageCandidates = false,
  hasAudioCandidates = false,
  canPublish = false,
  hasViralCandidate = false
} = {}) {
  const isBatchFactory = source === 'batch_factory';
  const actionPlan = { primary: [], more: [] };

  if (resourceType === 'merge') {
    actionPlan.primary.push('publish-organization');
    if (canPublish) actionPlan.more.push('upload-video-management');
    return actionPlan;
  }

  if (kind === 'script') {
    if (isBatchFactory) {
      actionPlan.primary.push(
        'prompt-templates',
        'extract-assets',
        'generate-storyboard',
        'generate-viral'
      );
      actionPlan.more.push('save-line-count');
      if (hasViralCandidate) actionPlan.more.push('adopt-viral-candidate');
    }
    return actionPlan;
  }

  if (kind === 'image' || kind === 'video') actionPlan.primary.push('reference-assets');
  if (isBatchFactory && (kind === 'image' || kind === 'video' || kind === 'audio')) {
    actionPlan.primary.push('prompt-templates');
  }

  if (kind === 'image' || kind === 'video' || kind === 'audio') {
    actionPlan.more.push('preview-inputs');
  }
  if (kind === 'image' && hasImageCandidates) actionPlan.more.push('confirm-image');
  if (kind === 'audio' && hasAudioCandidates) actionPlan.more.push('confirm-audio');
  if (canRetry) actionPlan.more.push('retry');

  return actionPlan;
}
