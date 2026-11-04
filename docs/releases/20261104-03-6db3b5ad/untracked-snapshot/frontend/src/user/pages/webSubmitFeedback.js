// 网站提交结果反馈：只有业务结果真的全部成功才提示成功。
// 后端 /sub... 接口在部分/全部失败时仍返回 HTTP 200（兼容旧前端），
// 因此这里必须读取 ok / failed_groups / results，不能只看 HTTP 状态码。

const FAILED_STATUSES = new Set(['error', 'failed']);

function numberOrZero(raw) {
  const value = Number(raw);
  return Number.isFinite(value) && value > 0 ? value : 0;
}

function resultItems(result) {
  return Array.isArray(result?.results) ? result.results : [];
}

function statusOf(item) {
  return String(item?.status || '').trim().toLowerCase();
}

function failureLabel(item) {
  const book = String(item?.bookId || item?.book_id || item?.id || '').trim();
  const version = String(item?.version || '').trim();
  const reason = String(item?.message || item?.error || '').trim();
  const target = [book, version].filter(Boolean).join('/');
  return target && reason ? `${target}：${reason}` : (target || reason);
}

export function webSubmitFailureLines(result, limit = 3) {
  return resultItems(result)
    .filter(item => FAILED_STATUSES.has(statusOf(item)) || statusOf(item) === 'pending_confirmation')
    .map(failureLabel)
    .filter(Boolean)
    .slice(0, limit);
}

// 用于提交/重试：failed_groups>0 时 warning（有成功也有失败）或 error（全部失败），
// 并列出具体失败版本；只有 failed_groups===0 且确实产生了成功提交才 success。
export function describeWebSubmitExecution(result, successText = '网站提交完成') {
  if (!result || typeof result !== 'object') {
    return { level: 'error', text: '网站提交没有返回有效结果，请稍后重试' };
  }
  const hasOutcome = result.ok !== undefined || result.failed_groups !== undefined || result.failedGroups !== undefined || resultItems(result).length > 0;
  if (!hasOutcome) return { level: 'success', text: successText };

  const failed = numberOrZero(result.failed_groups ?? result.failedGroups);
  const succeeded = numberOrZero(result.success_groups ?? result.successGroups);
  const lines = webSubmitFailureLines(result);
  const detail = lines.length ? `：${lines.join('；')}` : '';

  if (result.ok === false || failed > 0) {
    if (failed > 0) return { level: 'warning', text: `${failed} 个版本未提交成功${detail}` };
    return { level: 'warning', text: '没有产生新的提交：所选版本可能已提交过或没有可提交内容' };
  }
  if (resultItems(result).length > 0 && succeeded === 0) {
    return { level: 'warning', text: '没有产生新的提交：所选版本均已提交过或已跳过' };
  }
  return { level: 'success', text: successText };
}

// 用于“核对”：pending_confirmation 属于等待回读，不是完成。
export function describeWebSubmitReconcile(result, successText = '核对完成') {
  const base = describeWebSubmitExecution(result, successText);
  if (base.level === 'success') return base;
  const results = resultItems(result);
  const pending = results.filter(item => statusOf(item) === 'pending_confirmation').length;
  if (pending > 0) {
    return { level: 'warning', text: `已发起核对，仍有 ${pending} 个版本等待目标站回读确认` };
  }
  return base;
}

// 用于“视频管理系统登录可用”：必须同时看 ok 与会话检查结果，
// 不能因为 HTTP 200 就告诉用户会话有效。
export function describeWebSubmitVisible(result) {
  if (!result || typeof result !== 'object') {
    return { level: 'error', text: '登录状态检查没有返回有效结果，请重新登录并验证' };
  }
  const checks = result.result?.checks || result.checks || {};
  const loginSession = checks.login_session;
  if (result.ok === false || loginSession === false) {
    return { level: 'error', text: '视频管理系统登录已失效：请重新登录并验证后再提交' };
  }
  return { level: 'success', text: '视频管理系统登录可用' };
}
