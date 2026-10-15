function determineSubmissionOutcome({ prompt, before = {}, after = {}, networkEvidence = null } = {}) {
  const networkIdentities = cleanIdentities(networkEvidence?.identities);
  if (networkEvidence?.accepted === true && networkIdentities.length > 0) {
    return {
      status: 'accepted',
      submissionId: chooseSubmissionId(networkIdentities),
      identities: networkIdentities
    };
  }

  const beforeIds = new Set(flattenNodeIdentities(before.identityNodes));
  const promptNeedle = normalizeText(prompt).slice(0, 120);
  const newNodes = (Array.isArray(after.identityNodes) ? after.identityNodes : []).filter(node => {
    const ids = cleanIdentities(node?.identities);
    return ids.some(id => !beforeIds.has(id));
  });
  const promptBound = newNodes.filter(node => {
    if (!promptNeedle) return false;
    return normalizeText(node?.text).includes(promptNeedle);
  });

  if (promptBound.length === 1) {
    const identities = cleanIdentities(promptBound[0].identities);
    if (identities.length > 0) {
      return {
        status: 'accepted',
        submissionId: chooseSubmissionId(identities),
        identities
      };
    }
  }
  if (promptBound.length > 1) return { status: 'unknown' };

  if (newNodes.length === 0 && /提交失败|发送失败|未发送成功|请求失败.{0,12}重试|生成请求失败/.test(normalizeText(after.visibleText))) {
    return { status: 'not_accepted' };
  }
  return { status: 'unknown' };
}

function flattenNodeIdentities(nodes) {
  return (Array.isArray(nodes) ? nodes : []).flatMap(node => cleanIdentities(node?.identities));
}

function cleanIdentities(values) {
  return [...new Set((Array.isArray(values) ? values : []).map(value => String(value || '').trim()).filter(value => value && value.length <= 160))];
}

function chooseSubmissionId(identities) {
  return identities.find(value => /^(msg|message)[_:-]?/i.test(value))
    || identities.find(value => /message/i.test(value))
    || identities.find(value => /^(task|generation)[_:-]?/i.test(value))
    || identities[0];
}

function normalizeText(value) {
  return String(value || '').replace(/\s+/g, ' ').trim();
}

module.exports = { determineSubmissionOutcome, chooseSubmissionId, cleanIdentities };
