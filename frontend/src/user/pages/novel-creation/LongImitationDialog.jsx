import { Button, Checkbox, Input, InputNumber, Modal, Select, Spin } from 'antd';
import { Pencil, Upload, X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { novelRequest } from '../../../shared/api/novelCreation.js';
import { availableModel } from './contracts.js';
import { emptyLongImitationDraft, imitationChapters, imitationChapterCharacters, longImitationCharacters, recentImitationChapters, selectedImitationText, validateLongImitationDraft } from './longImitationInputs.js';
import { readLongImitationDraft, writeLongImitationDraft } from './longImitationDraftStore.js';
import useWorkspaceScope from './useWorkspaceScope.js';
import './LongImitationDialog.css';

const preferences = ['玄幻', '都市', '仙侠', '科幻', '历史', '悬疑', '言情', '武侠', '奇幻', '灵异', '末世', '系统'];
function Field({ label, children, required, extra }) {
  return <div className="long-imitation-field"><div className="long-imitation-field-heading"><label>{label}{required && <span className="long-imitation-required"> *</span>}</label>{extra}</div>{children}</div>;
}
function Presets({ values, value, onChange, disabled, unit = '' }) {
  return <div className="long-imitation-presets">{values.map(number => <button type="button" key={number} disabled={disabled} className={value === number ? 'selected' : ''} aria-pressed={value === number} onClick={() => onChange(number)}>{number === 10000 ? '1万' : number}{unit}</button>)}</div>;
}

export default function LongImitationDialog({ owner, open, viewKey, models, modelId, thinkingLevel, onCancel, onPersisted, onCreated }) {
  const scope = useWorkspaceScope(owner);
  const [draft, setDraft] = useState(emptyLongImitationDraft), [intent, setIntent] = useState(null);
  const [ready, setReady] = useState(false), [cacheError, setCacheError] = useState(''), [error, setError] = useState('');
  const [busy, setBusy] = useState(false), [uploadBusy, setUploadBusy] = useState(false);
  const [styles, setStyles] = useState([]), [memos, setMemos] = useState([]), [metadataError, setMetadataError] = useState(''), [metadataBusy, setMetadataBusy] = useState(false);
  const [statistics, setStatistics] = useState(null), [previewError, setPreviewError] = useState(''), [previewRetry, setPreviewRetry] = useState(0);
  const [pickerOpen, setPickerOpen] = useState(false), [pickerKeys, setPickerKeys] = useState([]);
  const fileInput = useRef(), fileRequest = useRef(), fileSequence = useRef(0), metadataSequence = useRef(0), flight = useRef(false), saveTimer = useRef(), view = useRef(), draftLoaded = useRef(false);
  const current = useRef({ draft, intent });
  current.current = { draft, intent };
  if (!view.current || view.current.key !== viewKey || view.current.open !== open) view.current = { key: viewKey, open };
  const locked = !ready || !!intent || busy || uploadBusy;
  const chapters = imitationChapters(draft.files || []), selected = chapters.filter(chapter => draft.chapterKeys.includes(chapter.key));
  const characters = longImitationCharacters(draft.referenceText);
  const recentPresets = [3,5,10].filter(number => number===3 ? chapters.length>=3 : number===10 ? chapters.length>10 : true);
  const selectedRecent = recentPresets.find(number=>JSON.stringify(recentImitationChapters(chapters,number))===JSON.stringify(draft.chapterKeys));

  async function loadDraft() {
    draftLoaded.current = false; setReady(false); setCacheError('');
    try {
      const saved = await scope.call(readLongImitationDraft(owner));
      if (!scope.isCurrent()) return;
      const restored = { draft: { ...emptyLongImitationDraft(), ...(saved?.draft || {}) }, intent: saved?.intent || null };
      current.current = restored; draftLoaded.current = true;
      setDraft(restored.draft); setIntent(restored.intent); setReady(true);
    } catch { if (scope.isCurrent()) setCacheError('创作草稿暂时无法读取，请重试；不会覆盖原草稿。'); }
  }
  useEffect(() => {
    loadDraft();
    const flush = () => {
      clearTimeout(saveTimer.current);
      // An in-flight submit already queued the immutable intent. Never
      // overwrite it with an older draft during navigation.
      if (draftLoaded.current && !flight.current) writeLongImitationDraft(owner, structuredClone(current.current)).catch(() => {});
    };
    window.addEventListener('pagehide', flush);
    return () => { window.removeEventListener('pagehide', flush); flush(); fileRequest.current?.abort(); };
  }, [owner]);
  async function saveCurrent() {
    try { await writeLongImitationDraft(owner, current.current); if (scope.isCurrent()) setCacheError(''); }
    catch { if (scope.isCurrent()) setCacheError('草稿暂时无法保存，请保留原文并重试保存。'); }
  }
  function change(patch) {
    if (locked) return;
    const next = { ...current.current.draft, ...patch };
    current.current = { draft: next, intent: null }; setDraft(next); setError('');
    clearTimeout(saveTimer.current); saveTimer.current = setTimeout(saveCurrent, 250);
  }
  async function loadMetadata() {
    const sequence = ++metadataSequence.current;
    setMetadataBusy(true); setMetadataError('');
    const results = await Promise.allSettled([
      novelRequest('/style-profiles', undefined, { signal: scope.signal }),
      novelRequest('/long-imitation/memos', undefined, { signal: scope.signal })
    ]);
    if (!scope.isCurrent() || sequence !== metadataSequence.current) return;
    if (results[0].status === 'fulfilled') setStyles(results[0].value.styles || []);
    if (results[1].status === 'fulfilled') setMemos(results[1].value.memos || []);
    if (results.some(result => result.status === 'rejected')) setMetadataError('文风或备忘录暂时无法读取，已选引用仍会保留，请重试。');
    else if (results[1].value.truncated) setMetadataError('备忘录只显示前 500 条。');
    setMetadataBusy(false);
  }
  useEffect(() => { if (open) loadMetadata(); }, [owner, open]);
  useEffect(() => {
    setStatistics(null); setPreviewError('');
    if (!open || characters < 100 || draft.referenceText.length > 2_000_000) return;
    const controller = new AbortController();
    const timer = setTimeout(() => {
      scope.call(novelRequest('/long-imitation/preview', { referenceText: draft.referenceText }, { signal: AbortSignal.any([scope.signal, controller.signal]) }), result => { if (!controller.signal.aborted) setStatistics(result.statistics); }).catch(() => {
        if (scope.isCurrent() && !controller.signal.aborted) setPreviewError('文风统计暂时无法读取，可重试；参考原文已保留。');
      });
    }, 400);
    return () => { clearTimeout(timer); controller.abort(); };
  }, [owner, open, draft.referenceText, previewRetry]);

  function cancel() {
    view.current = { key: viewKey, open: false }; fileSequence.current++; fileRequest.current?.abort(); setUploadBusy(false); setPickerOpen(false); onCancel();
  }
  function selectChapters(keys) { change({ chapterKeys: keys, referenceText: selectedImitationText(chapters, keys) }); }
  async function upload(event) {
    const picked = [...(event.target.files || [])]; event.target.value = '';
    if (!picked.length || locked) return;
    const prior = current.current.draft.files || [];
    const unique = picked.filter(file => !prior.some(old => old.filename === file.name && old.size === file.size && old.lastModified === file.lastModified));
    if (!unique.length) { setError('所选文件已经读取，已跳过重复文件。'); return; }
    if (unique.length + prior.length > 8) { setError('一次创作最多支持 8 个参考文件'); return; }
    if (unique.some(file => !/\.(txt|docx)$/i.test(file.name) || file.size > 20 * 1024 * 1024)) { setError('只支持 TXT / DOCX，每个文件最多 20 MiB'); return; }
    const sequence = ++fileSequence.current, submittedView = view.current;
    const controller = new AbortController(); fileRequest.current = controller; setUploadBusy(true); setError('');
    try {
      const body = new FormData(); unique.forEach(file => body.append('files', file));
      const result = await scope.call(novelRequest('/long-imitation/parse', body, { signal: AbortSignal.any([scope.signal, controller.signal]) }));
      if (!result || sequence !== fileSequence.current || submittedView !== view.current) return;
      const added = (result.files || []).map((file, index) => {
        const source = unique.find(item => item.name === file.filename) || unique[index];
        return { ...file, key: crypto.randomUUID(), size: source?.size, lastModified: source?.lastModified, chapters: (file.chapters || []).map((chapter, n) => ({ ...chapter, id: chapter.id || String(n) })) };
      });
      const files = [...prior, ...added], all = imitationChapters(files), keys = recentImitationChapters(all, 5);
      if (added.length) {
        const next = { ...current.current.draft, files, chapterKeys: keys, referenceText: selectedImitationText(all, keys) };
        current.current = { draft: next, intent: null }; setDraft(next); clearTimeout(saveTimer.current); await saveCurrent();
      }
      const failures = (result.failures || []).map(item => `${item.filename}：${item.error || item.message || '读取失败'}`);
      if (failures.length || !added.length) setError(failures.join('；') || '未能读取文件内容，原参考原文已保留。');
    } catch (e) { if (scope.isCurrent() && !controller.signal.aborted && sequence === fileSequence.current) setError(e.message || '文件读取失败，原参考原文已保留'); }
    finally { if (scope.isCurrent() && sequence === fileSequence.current) setUploadBusy(false); else if (scope.isCurrent() && !view.current.open) setUploadBusy(false); }
  }
  function removeFile(key) {
    if (locked) return;
    const files = draft.files.filter(file => file.key !== key), all = imitationChapters(files);
    const keys = draft.chapterKeys.filter(id => all.some(chapter => chapter.key === id));
    change({ files, chapterKeys: keys, ...(all.length ? { referenceText: selectedImitationText(all, keys) } : {}) });
  }
  function makePayload(value) {
    return {
      idempotencyKey: crypto.randomUUID(), modelId, thinkingLevel,
      title: value.title.trim(), referenceText: value.referenceText, storyIdea: value.storyIdea.trim(), selectedDirection: value.selectedDirection.trim(),
      targetWordCount: value.targetWordCount, targetChapters: value.targetChapters, readerPreferences: value.readerPreferences.trim(), supplementaryInfo: value.supplementaryInfo.trim(),
      knowledgeBookIds: [...value.knowledgeBookIds], memoReferences: value.memoKeys.map(key => { const [projectId, noteId] = JSON.parse(key); return { projectId, noteId }; }),
      referenceSelection: { files: value.files.map(file => ({ filename: file.filename, chapterIds: file.chapters.filter(chapter => value.chapterKeys.includes(`${file.key}:${chapter.id}`)).map(chapter => chapter.id) })) }
    };
  }
  async function finish(result, submittedView) {
    if (!result?.project?.id || !result.session?.id || !result.run?.id || !result.creation) throw Error('创建结果尚不完整，请核对原创建请求');
    onPersisted?.(result);
    await writeLongImitationDraft(owner, null);
    if (!scope.isCurrent()) return;
    current.current = { draft: emptyLongImitationDraft(), intent: null }; setDraft(current.current.draft); setIntent(null); setError('');
    if (submittedView === view.current && view.current.open) onCreated(result);
  }
  async function submit(resubmit = false) {
    if (!ready || flight.current || !scope.isCurrent()) return;
    if (intent && !resubmit) { await recover(); return; }
    let pending = intent;
    if (!pending) {
      const problem = validateLongImitationDraft(current.current.draft, availableModel(models, modelId) ? modelId : '');
      if (problem) { setError(problem); return; }
      try { pending = { payload: makePayload(current.current.draft), allowResubmit: false }; }
      catch { setError('备忘录选择无效，请重新选择'); return; }
    }
    flight.current = true; setBusy(true); setError(''); clearTimeout(saveTimer.current);
    const submittedView = view.current;
    try {
      await writeLongImitationDraft(owner, { draft: current.current.draft, intent: pending });
      if (!scope.isCurrent()) return;
      current.current.intent = pending; setIntent(pending);
      const result = await scope.call(novelRequest('/long-imitation', pending.payload, { signal: scope.signal }));
      if (result) await finish(result, submittedView);
    } catch (e) {
      if (!scope.isCurrent()) return;
      if (e.status >= 400 && e.status < 500 && ![408, 409, 429].includes(e.status)) {
        current.current.intent = null; setIntent(null);
        try { await writeLongImitationDraft(owner, current.current); } catch { setCacheError('恢复记录尚未更新，请保留原文并重试保存。'); }
      }
      setError(e.message || '创建结果暂未确认，请核对原请求');
    } finally { flight.current = false; if (scope.isCurrent()) setBusy(false); }
  }
  async function recover() {
    if (!intent || flight.current || !scope.isCurrent()) return;
    flight.current = true; setBusy(true); setError(''); const submittedView = view.current;
    try {
      const result = await scope.call(novelRequest(`/long-imitation/requests/${encodeURIComponent(intent.payload.idempotencyKey)}`, undefined, { signal: scope.signal }));
      if (result) await finish(result, submittedView);
    } catch (e) {
      if (!scope.isCurrent()) return;
      if (e.status === 404) {
        const pending = { ...intent, allowResubmit: true }; current.current.intent = pending; setIntent(pending); await saveCurrent();
        setError('尚未找到创建结果。可以明确重新提交原请求，编号和参数保持不变。');
      } else setError(e.message || '结果暂时无法核对，请稍后重试');
    } finally { flight.current = false; if (scope.isCurrent()) setBusy(false); }
  }
  const input = (name, label, rows, placeholder) => rows ? <Input.TextArea aria-label={label} rows={rows} disabled={locked} value={draft[name]} onChange={e => change({ [name]: e.target.value })} placeholder={placeholder}/> : <Input aria-label={label} disabled={locked} value={draft[name]} onChange={e => change({ [name]: e.target.value })} placeholder={placeholder}/>;
  return <>
    <Modal rootClassName="novel-modal novel-long-imitation" open={open} centered width="min(640px, calc(100vw - 32px))" title={<div className="long-imitation-title"><span><Pencil size={17}/></span><div><h2>仿写创作</h2><p>参考原文风格，创作全新原创故事</p></div></div>} onCancel={cancel} footer={<><Button onClick={cancel}>取消</Button>{intent?.allowResubmit && <Button disabled={busy} onClick={() => submit(true)}>重新提交原请求</Button>}<Button type="primary" loading={busy} disabled={!ready || uploadBusy} onClick={() => submit()}>{intent ? '核对创建结果' : '创建作品'}</Button></>}>
      {error && <p className="long-imitation-error" role="alert">{error}</p>}
      <div className="long-imitation-scroll">
        {!ready && !cacheError && <Spin aria-label="读取仿写草稿"/>}
        {cacheError && <p className="long-imitation-error" role="alert">{cacheError}<Button size="small" onClick={ready ? saveCurrent : loadDraft}>{ready ? '重试保存草稿' : '重新读取草稿'}</Button></p>}
        <Field label="暂定标题（可选）">{input('title', '暂定标题（可选）', 0, '留空则自动生成')}</Field>
        <Field label="参考原文" required extra={<span className={`long-imitation-count${characters >= 100 ? ' valid' : ''}`}>{characters.toLocaleString()} 字</span>}>
          {input('referenceText', '参考原文', 6, '粘贴参考小说全文（至少 100 字），或上传 txt / docx 自动填入')}
          <small>粘贴或上传要仿写的参考原文，提交后拆解情绪曲线与核心价值观</small>
        </Field>
        <Field label="上传参考文件（可选）">
          <input ref={fileInput} type="file" className="long-imitation-file-input" aria-label="上传参考文件" multiple accept=".txt,.docx" onChange={upload} disabled={locked}/>
          <Button className="long-imitation-file-button" icon={<Upload size={14}/>} disabled={locked} loading={uploadBusy} onClick={() => fileInput.current?.click()}>选择文件（docx / txt）</Button>
          {draft.files.map(file => <div className="long-imitation-attachment" key={file.key}><span>已读取 · {file.filename}</span><Button aria-label={`移除参考文件${file.filename}`} type="text" size="small" disabled={locked} icon={<X size={13}/>} onClick={() => removeFile(file.key)}/></div>)}
          <small>最多 8 个文件，每个不超过 20 MiB；参考内容合计最多 200 万字元。</small>
        </Field>
        {chapters.length > 1 && <Field label="选择参考章节"><div className="long-imitation-presets">{recentPresets.map(number => <button type="button" disabled={locked} key={number} className={selectedRecent===number ? 'selected' : ''} onClick={() => selectChapters(recentImitationChapters(chapters, number))}>最近{number}章</button>)}<button type="button" disabled={locked} onClick={() => { setPickerKeys([...draft.chapterKeys]); setPickerOpen(true); }}>选择章节</button></div><small>共识别 {chapters.length} 章 · 已选择 {selected.length} 章，共 {selected.reduce((sum, chapter) => sum + imitationChapterCharacters(chapter), 0).toLocaleString()} 字</small></Field>}
        {statistics && <section className="long-imitation-statistics"><b>文风统计指纹</b><div>{[['平均句长', statistics.avgSentenceLength, '字'], ['句长波动', statistics.sentenceLengthStdDev, ''], ['段落长度', statistics.avgParagraphLength, '字'], ['词汇多样性', Math.round((statistics.vocabularyDiversity || 0) * 100), '%']].map(([label, value, unit]) => <span key={label}>{label}<strong>{value}{unit}</strong></span>)}</div>{statistics.topPatterns?.length > 0 && <p>高频句首：{statistics.topPatterns.join('、')}</p>}{statistics.rhetoricalFeatures?.length > 0 && <p>修辞特征：{statistics.rhetoricalFeatures.join('、')}</p>}<small>本地算法统计，不消耗模型额度；提交后进行八维文风分析。</small></section>}
        {previewError && <p role="alert">{previewError}<Button size="small" onClick={() => setPreviewRetry(n => n + 1)}>重试统计</Button></p>}
        <div className="long-imitation-number-grid">
          <Field label="每章字数"><InputNumber aria-label="每章字数" value={draft.targetWordCount} min={1000} max={50000} step={500} controls={false} disabled={locked} onChange={targetWordCount => change({ targetWordCount })}/><Presets values={[3000, 4000, 5000, 6000, 8000, 10000]} value={draft.targetWordCount} onChange={targetWordCount => change({ targetWordCount })} disabled={locked}/><small>可自由填写每章字数，也可点选常用字数</small></Field>
          <Field label="目标章节数"><InputNumber aria-label="目标章节数" value={draft.targetChapters} min={1} max={500} step={10} controls={false} disabled={locked} onChange={targetChapters => change({ targetChapters })}/><Presets values={[50, 100, 150, 200, 300]} value={draft.targetChapters} onChange={targetChapters => change({ targetChapters })} disabled={locked}/><small>长篇默认 100 章，可按需调整</small></Field>
        </div>
        <Field label="读者偏好">{input('readerPreferences', '读者偏好', 0, '填写题材偏好，如：玄幻、都市')}<div className="long-imitation-presets">{preferences.map(value => <button type="button" key={value} disabled={locked} onClick={() => { const values = draft.readerPreferences.trim().split(/[、,，\s]+/).filter(Boolean); if (!values.includes(value)) change({ readerPreferences: [...values, value].join('、') }); }}>{value}</button>)}</div><small>可自由填写，点选标签会自动追加</small></Field>
        <Field label="原创故事想法（可选）">{input('storyIdea', '原创故事想法（可选）', 3, '剧情、主角、核心冲突（原创剧情）')}</Field>
        <Field label="仿写方向（可选）">{input('selectedDirection', '仿写方向（可选）', 2, '留空则默认采用拆书后的方向一；可填“方向二：…”')}</Field>
        <Field label="补充信息（可选）">{input('supplementaryInfo', '补充信息（可选）', 3, '填写额外要求，如人设、情节走向、必须出现的桥段、禁忌点等')}</Field>
        <Field label="文风设置（多选）"><Checkbox.Group disabled={locked || metadataBusy} value={draft.knowledgeBookIds} onChange={knowledgeBookIds => change({ knowledgeBookIds })}>{styles.map(style => <Checkbox key={style.id} value={style.id}>{style.title || style.id}</Checkbox>)}</Checkbox.Group>{!metadataBusy && !styles.length && <small>暂无可用文风</small>}</Field>
        <Field label="关联备忘录（可选）"><Select aria-label="关联备忘录（可选）" mode="multiple" disabled={locked || metadataBusy} value={draft.memoKeys} options={memos.map(memo => ({ value: JSON.stringify([memo.projectId, memo.id]), label: memo.title || memo.id }))} placeholder="选择备忘录作为参考素材" onChange={memoKeys => change({ memoKeys })}/>{draft.memoKeys.length > 0 && <small>已选 {draft.memoKeys.length} 条备忘录，内容将作为参考素材注入</small>}</Field>
        {metadataBusy && <p role="status">正在读取文风和备忘录…</p>}
        {metadataError && <p role="alert">{metadataError}<Button size="small" onClick={loadMetadata}>重新读取引用</Button></p>}
        {intent && <p role="status" className="long-imitation-notice">原创建请求已保留，核对期间不会再次调用模型。关闭后可重新进入继续核对。</p>}
      </div>
    </Modal>
    <Modal rootClassName="novel-modal long-imitation-chapter-picker" open={pickerOpen && open} title="选择参考章节" onCancel={() => setPickerOpen(false)} width={640} onOk={() => { selectChapters(pickerKeys); setPickerOpen(false); }} okText="确认选择" cancelText="取消">
      <div className="long-imitation-picker-tools"><Button size="small" onClick={() => setPickerKeys(chapters.map(chapter => chapter.key))}>全选</Button><Button size="small" onClick={() => setPickerKeys([])}>清空</Button><span>已选 {pickerKeys.length} / {chapters.length} 章</span></div>
      <div className="long-imitation-picker-list">{chapters.map(chapter => <Checkbox checked={pickerKeys.includes(chapter.key)} onChange={event => { const checked=event.target.checked; setPickerKeys(current => checked ? [...new Set([...current,chapter.key])] : current.filter(key=>key!==chapter.key)); }} key={chapter.key}><span>{chapter.title}</span><small>{imitationChapterCharacters(chapter).toLocaleString()} 字 · {chapter.filename}</small></Checkbox>)}</div>
    </Modal>
  </>;
}
