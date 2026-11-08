import errorCodeContract from './apiErrorCodes.json' with { type: 'json' };

// 错误码与中文说明只维护这一份：后端返回稳定的 code，前端按 code 显示
// “缺什么 → 去哪配 → 怎么开”。旧接口只返回 error 文案时仍按原样显示。
const ERROR_MESSAGES = Object.fromEntries(errorCodeContract.codes.map(item => [item.code, item.message]));

const CODE_PATTERN = /^[A-Z][A-Z0-9_]{2,}$/;

export function apiErrorCode(error) {
  const explicit = String(error?.code || '').trim();
  if (explicit) return explicit;
  const message = String(error?.message || '').trim();
  return CODE_PATTERN.test(message) ? message : '';
}

export function apiErrorMessage(error, fallback = '') {
  // 统一请求层超时：与业务失败区分开，提示用户可以直接重试。
  if (error?.code === 'timeout' || error?.name === 'TimeoutError') {
    return '请求超时：服务端没有在限定时间内返回，请点击重试。';
  }
  const code = apiErrorCode(error);
  const friendly = (code && ERROR_MESSAGES[code])
    || String(error?.message || '').trim()
    || code
    || fallback
    || '请求失败。';
  // 服务器返回的 502/503 网页错误会是一整段 HTML，直接显示会刷屏；换成一句人话。
  if (/<html[\s>]|502 Bad Gateway|503 Service/i.test(friendly)) {
    return '服务器通讯短暂中断（正在更新或重启），请稍后重试。';
  }
  return friendly;
}

export { ERROR_MESSAGES };
