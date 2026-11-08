import { BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY } from './batchFactoryNovelFetchHandoff.js';

// 固定 key：同一个批次访问错误只保留一条提示，避免每次轮询/重渲染叠出多条红字。
export const BATCH_ACCESS_MESSAGE_KEY = 'batch-factory-batch-access-error';
const BATCH_ACCESS_MESSAGE_DURATION = 8;

const FORBIDDEN_TEXT = '该批次属于其他账号，当前账号无数据。';
const NOT_FOUND_TEXT = '该批次不存在，或属于其他账号；当前账号无数据。';
const FALLBACK_TEXT = '读取批量工程失败，请稍后重试。';

/**
 * 把批次详情失败翻译成用户能理解的文案。服务端 403 已经带可读的归属说明，
 * 有文案就沿用；裸 404 由前端补成“已不存在”，不再只显示 not found。
 */
export function batchAccessErrorMessage(error) {
  const code = String(error?.code || '').trim();
  const serverMessage = String(error?.message || '').trim();
  // 旧接口的裸 not found 对用户没有信息量，一律换成可读文案。
  const readable = serverMessage && !/^not found$/i.test(serverMessage) ? serverMessage : '';
  if (code === 'batch_forbidden' || (!code && error?.status === 403)) {
    return readable || FORBIDDEN_TEXT;
  }
  if (code === 'batch_not_found' || (!code && error?.status === 404)) {
    return readable || NOT_FOUND_TEXT;
  }
  return readable || FALLBACK_TEXT;
}

/** 403/404 是确定性结果：保留失效 batch id 只会让每次打开页面都重复 404。 */
export function isStaleBatchAccessError(error) {
  const code = String(error?.code || '').trim();
  return code === 'batch_forbidden' || code === 'batch_not_found' || error?.status === 403 || error?.status === 404;
}

/** 只清理指向失败批次的本地活跃 id，避免把用户手动打开的其他批次一起丢掉。 */
export function forgetActiveBatch(batchId) {
  if (!batchId) return;
  try {
    if (localStorage.getItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY) === batchId) {
      localStorage.removeItem(BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY);
    }
  } catch (_) {
    // 隐私模式下 localStorage 不可用时不阻断提示。
  }
}

/** 供 message.error 使用的单条提示配置：固定 key + 可点击关闭。 */
export function batchAccessNotice(error, key = BATCH_ACCESS_MESSAGE_KEY) {
  return {
    key,
    duration: BATCH_ACCESS_MESSAGE_DURATION,
    content: batchAccessErrorMessage(error)
  };
}
