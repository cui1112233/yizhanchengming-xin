// 插件面板生成任务的取消规则：真正的 AbortController、重复提交拦截、取消与失败区分。

export function isAbortError(error) {
  return error?.name === 'AbortError' || error?.code === 'aborted' || error?.aborted === true;
}

export function cancelledError(message = '已取消') {
  const error = new Error(message);
  error.name = 'AbortError';
  error.aborted = true;
  return error;
}

// 同一面板只允许一个在跑的生成：begin() 返回 null 表示重复提交。
export function createGenerationGate() {
  let controller = null;
  return {
    begin() {
      if (controller) return null;
      controller = new AbortController();
      return controller;
    },
    end(target) {
      if (controller && controller === target) controller = null;
    },
    abort() {
      if (!controller) return false;
      controller.abort();
      controller = null;
      return true;
    },
    get signal() {
      return controller ? controller.signal : null;
    },
    get running() {
      return Boolean(controller);
    }
  };
}

export function validateGenerationInput({ input, modelId, promptId, emptyMessage, promptMessage } = {}) {
  if (!String(input || '').trim()) return { ok: false, message: emptyMessage || '请输入或上传内容' };
  if (!modelId) return { ok: false, message: '请选择模型' };
  if (!promptId) return { ok: false, message: promptMessage || '请选择提示词' };
  return { ok: true, message: '' };
}

// 串行/并发生成。返回结构化结果，由调用方决定是否写入界面：
// 取消 → { ok:false, cancelled:true }，调用方不得写入结果，也不弹错误提示。
export async function executeGeneration({ parts = [], concurrent = false, runPart, signal, onProgress } = {}) {
  const list = Array.isArray(parts) ? parts : [];
  const report = (current, total) => {
    if (typeof onProgress === 'function') onProgress(current, total);
  };
  const ensureLive = () => {
    if (signal?.aborted) throw cancelledError();
  };
  try {
    ensureLive();
    if (!list.length) throw new Error('没有可生成的内容');
    if (concurrent && list.length > 1) {
      let completed = 0;
      const settled = await Promise.allSettled(list.map(async (part, index) => {
        try {
          ensureLive();
          return { index, text: await runPart(part, index, []) };
        } finally {
          completed += 1;
          report(completed, list.length);
        }
      }));
      ensureLive();
      const failedIndex = settled.findIndex(item => item.status === 'rejected');
      const results = settled
        .filter(item => item.status === 'fulfilled')
        .sort((left, right) => left.value.index - right.value.index)
        .map(item => item.value.text);
      if (failedIndex >= 0) {
        // 部分段已成功时必须保留，否则用户需要整批重跑。
        return { ok: false, cancelled: false, results, text: results.join('\n\n').trim(), error: settled[failedIndex].reason, failedIndex };
      }
      return { ok: true, cancelled: false, results, text: results.join('\n\n').trim(), error: null, failedIndex: -1 };
    }
    const results = [];
    for (let index = 0; index < list.length; index += 1) {
      ensureLive();
      try {
        results.push(await runPart(list[index], index, results));
      } catch (error) {
        if (isAbortError(error) || signal?.aborted) throw error;
        // 串行模式下前面段落已经生成完成，失败时保留下来交给调用方写回界面。
        return { ok: false, cancelled: false, results, text: results.join('\n\n').trim(), error, failedIndex: index };
      }
      report(index + 1, list.length);
    }
    return { ok: true, cancelled: false, results, text: results.join('\n\n').trim(), error: null, failedIndex: -1 };
  } catch (error) {
    if (isAbortError(error) || signal?.aborted) {
      return { ok: false, cancelled: true, results: [], text: '', error };
    }
    return { ok: false, cancelled: false, results: [], text: '', error };
  }
}
