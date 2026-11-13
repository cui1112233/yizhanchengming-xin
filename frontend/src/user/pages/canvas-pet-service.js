// This adapter only calls existing owned-project APIs. Prompts live on the server.
export function createCanvasWriteGate() {
  let pending = 0, applying = false;
  return {
    beginWrite() {
      if (applying) throw new Error('正在应用画布方案，请稍后重试');
      pending++;
      let active = true;
      return () => { if (active) { active = false; pending--; } };
    },
    beginApply() {
      if (applying || pending) throw new Error('画布仍在处理素材或剧本，请完成后再应用方案');
      applying = true;
      let active = true;
      return () => { if (active) { active = false; applying = false; } };
    }
  };
}

export function lockCanvasPetInput(target) {
  const events = ['pointerdown', 'click', 'keydown', 'beforeinput', 'paste', 'drop'];
  const options = { capture: true };
  const block = event => { event.preventDefault(); event.stopImmediatePropagation(); };
  for (const type of events) target.addEventListener(type, block, options);
  return () => { for (const type of events) target.removeEventListener(type, block, options); };
}

export async function runCanvasPetRequest(services, { prompt, planId, signal }) {
  const check = () => {
    if (!services.taskId || services.getTaskId() !== services.taskId || signal?.aborted) throw new Error('项目已切换或请求已停止');
  };
  check();
  const saved = await services.flush();
  check();
  if (!Number.isSafeInteger(saved?.revision) || saved.revision < 0) throw new Error('画布尚未可靠保存，请重试');
  const result = planId
    ? await services.approve(services.taskId, planId, saved.revision)
    : await services.ask({ taskId: services.taskId, prompt: String(prompt || '').trim(), signal,
      context: { page: '创作画布', canvasPet: { revision: saved.revision, selectedNodeIds: [...new Set(services.getSelectedIds())].slice(0, 50) }, director: { enabled: true, mode: 'supervised' } } });
  check();
  return result;
}
