import { batchFactoryVisibleError } from './batchFactoryBookState.js';

const PROVIDERS = [
  { pattern: /autodl/i, name: 'AutoDL H3' },
  { pattern: /yfai|seedance/i, name: 'YFAI Seedance' },
  { pattern: /gemini/i, name: 'Gemini' }
];

// 上游业务拒绝的标志：provider 自己返回了 code / request_id / rejected submission。
// 只有“没有这些字段”的 capability unavailable 才是真正的能力门控。
export function looksLikeUpstreamRejection(raw) {
  return /rejected submission|code\s*=|request_id\s*=|RequestParametersWrong/i.test(String(raw || ''));
}

export function batchFactoryErrorProvider(raw) {
  return PROVIDERS.find(item => item.pattern.test(String(raw || '')))?.name || '';
}

export function splitBatchFactoryError(error) {
  const raw = String(error || '').trim();
  const provider = batchFactoryErrorProvider(raw);
  const code = (raw.match(/code\s*=\s*([A-Za-z0-9_.-]+)/i) || [])[1] || '';
  const messageText = (raw.match(/message\s*=\s*([\s\S]*?)(?=\s+request_id\s*=|\s*$)/i) || [])[1]?.trim() || '';
  const requestId = (raw.match(/request_id\s*=\s*([A-Za-z0-9_\-]+)/i) || [])[1] || '';
  const upstreamRejected = looksLikeUpstreamRejection(raw);
  const capabilityGate = !upstreamRejected && /capability unavailable|CAPABILITY_NOT_ENABLED|Slice 门控/i.test(raw);
  // capability unavailable 后面经常跟的是真实业务原因（例如“文本模型请求失败：
  // 请在资产设置中选择漫画资产提示词”），只有裸的 Slice 门控才是“能力未开通”。
  // 之前统一渲染成“能力尚未开通”，把可自助修复的原因吞掉了。
  const capabilityReason = capabilityGate
    ? (raw.match(/capability unavailable\s*[:：]\s*([\s\S]+)$/i) || [])[1]?.trim() || ''
    : '';
  // 后端历史记录可能把业务校验包在 ErrConflict 里，渲染成
  // “revision conflict: 请在资产设置中选择漫画资产提示词”。
  // 这里拆开：真正的版本冲突给重试动作，其它业务原因只显示原因本身。
  const revisionConflict = /^revision conflict(?::\s*([\s\S]*))?$/i.exec(raw);

  let summary = '';
  if (revisionConflict) {
    const reason = (revisionConflict[1] || '').trim();
    summary = reason && !/版本已变化|revision conflict/i.test(reason) ? reason : '数据版本冲突，请重试';
  } else if (upstreamRejected) {
    // 能映射的 code 用人话，映射不了保留上游原文，避免把技术串直接糊给用户。
    if (/ref_image_0/i.test(raw)) summary = '上游要求参考图 ref_image_0，本次提交未包含，已被拒绝。';
    else if (messageText) summary = `上游拒绝了本次提交：${messageText}`;
    else summary = '上游拒绝了本次提交，请查看下方原始信息。';
  } else if (capabilityGate) {
    // 漫画资产提示词没选时，后端会包一层 “文本模型请求失败：请在资产设置中选择
    // 漫画资产提示词”。这类问题用户自己能修，不能再显示“能力未开通”。
    if (/请在资产设置中选择漫画资产提示词|漫画资产提示词/.test(raw)) {
      summary = '漫画资产提示词未选择，请在资产设置中选择。';
    } else if (capabilityReason && !/CAPABILITY_NOT_ENABLED|Slice 门控/i.test(capabilityReason)) {
      summary = capabilityReason;
    } else {
      summary = '该能力尚未开通：请管理员启用对应能力（Slice 门控）后重试。';
    }
  } else {
    summary = batchFactoryVisibleError(raw) || '自动化任务执行失败，可重试失败小说。';
  }

  return { raw, provider, code, message: messageText, requestId, upstreamRejected, capabilityGate, summary };
}
