export function comicDraftPayload(form, content) {
  const sourceText = String(content ?? '');
  if (!String(form.title || '').trim()) throw new Error('请输入作品名称');
  if (!sourceText.trim()) throw new Error('请直接输入或导入原文内容');
  if ([...sourceText].length > 600000) throw new Error('原文超过 60 万字，请按章拆分后分批提交');
  if (!form.model) throw new Error('请选择可用的文本模型');
  return {
    title: form.title.trim(), sourceText, modelId: form.model,
    promptId: form.promptId, scriptMode: form.scriptMode,
    aspectRatio: form.aspectRatio, assetDegree: form.degree, permissionMode: 'bypass',
  };
}

// 关闭、切换入口、手动输入及新文件会使旧解析失效，迟到的结果不能覆盖新草稿。
export function createSourceImportGate(parse) {
  let revision = 0;
  return {
    cancel() { revision += 1; },
    async load(file) {
      const ticket = ++revision;
      try {
        const content = await parse(file);
        return { current: ticket === revision, content };
      } catch (error) {
        if (ticket !== revision) return { current: false };
        throw error;
      }
    },
  };
}
