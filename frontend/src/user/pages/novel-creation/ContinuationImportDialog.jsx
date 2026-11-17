import { Button, Checkbox, Form, Input, Modal, Select, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { importNovelCreationChapters, listNovelCreationStyleProfiles } from '../../../shared/api/novelCreation.js';
import { readSaved } from './contracts.js';
import { splitContinuationChapters } from './continuationImport.js';
import useWorkspaceScope from './useWorkspaceScope.js';

const emptyDraft = { bookId: '', title: '', content: '', knowledgeBookIds: [] };
export default function ContinuationImportDialog({ owner, open, viewKey, projects, onCancel, onPersisted, onImported }) {
  const storageKey = 'novel:continuation-import:' + owner;
  const scope = useWorkspaceScope(owner);
  const [intent, setIntent] = useState(() => readSaved(storageKey + ':intent', null));
  const [draft, setDraft] = useState(() => readSaved(storageKey + ':intent', null)?.draft || { ...emptyDraft, ...readSaved(storageKey + ':draft') });
  const [styles, setStyles] = useState([]), [stylesLoading, setStylesLoading] = useState(false), [styleError, setStyleError] = useState('');
  const [busy, setBusy] = useState(false), [error, setError] = useState('');
  const flight = useRef(false), styleSequence = useRef(0);
  const view = useRef();
  if (!view.current || view.current.open !== open || view.current.key !== viewKey) view.current = { open, key: viewKey };
  const frozen = !!intent || busy;
  function cancel() { view.current = { open: false, key: viewKey }; onCancel(); }
  async function loadStyles() {
    const sequence = ++styleSequence.current;
    setStylesLoading(true); setStyleError('');
    try { await scope.call(listNovelCreationStyleProfiles(scope.signal), result => { if (sequence === styleSequence.current) setStyles(result.styles || []); }); }
    catch (e) { if (scope.isCurrent() && sequence === styleSequence.current) setStyleError(e.message || '文风列表暂时无法读取'); }
    finally { if (scope.isCurrent() && sequence === styleSequence.current) setStylesLoading(false); }
  }
  useEffect(() => { if (open) loadStyles(); }, [open, owner]);
  function change(patch) {
    if (frozen) return;
    const next = { ...draft, ...patch };
    setDraft(next); setError('');
    try { localStorage.setItem(storageKey + ':draft', JSON.stringify(next)); }
    catch { setError('草稿暂时无法保存，请保留正文后重试'); }
  }
  async function submit() {
    if (!scope.isCurrent() || flight.current) return;
    let pending = intent;
    if (!pending) {
      if (!draft.bookId && !draft.title.trim()) { setError('请输入新书书名'); return; }
      const chapters = splitContinuationChapters(draft.content);
      if (!chapters.length) { setError('请输入章节正文'); return; }
      pending = { draft, payload: { ...(draft.bookId ? { bookId: draft.bookId } : { title: draft.title.trim() }), chapters, knowledgeBookIds: [...draft.knowledgeBookIds], idempotencyKey: crypto.randomUUID() } };
      // Persist the exact request before sending: a lost reply may already have saved it.
      try { localStorage.setItem(storageKey + ':intent', JSON.stringify(pending)); }
      catch { setError('导入恢复记录暂时无法保存，请保留正文后重试'); return; }
      setIntent(pending);
    }
    flight.current = true; setBusy(true); setError('');
    const submittedView = view.current;
    try {
      const result = await scope.call(importNovelCreationChapters(pending.payload, scope.signal));
      if (!result || !scope.isCurrent()) return;
      // Saving remains real after closing the form. Update facts without taking over newer navigation.
      onPersisted(result); message.success('章节已导入');
      if (view.current !== submittedView || !view.current.open) return;
      localStorage.removeItem(storageKey + ':intent'); localStorage.removeItem(storageKey + ':draft');
      setIntent(null); setDraft({ ...emptyDraft });
      onImported(result);
    } catch (e) {
      if (!scope.isCurrent() || view.current !== submittedView || !view.current.open) return;
      // Only definite rejection unlocks editing. Unknown results keep the original key/input.
      if (e.status >= 400 && e.status < 500 && ![408, 409, 429].includes(e.status)) {
        localStorage.removeItem(storageKey + ':intent'); setIntent(null);
      }
      setError(e.message || '导入暂未完成，请稍后恢复原导入');
    } finally { flight.current = false; if (scope.isCurrent()) setBusy(false); }
  }
  return <Modal rootClassName="novel-modal novel-continuation-import" open={open} title="续写创作" width="min(640px, calc(100vw - 32px))" onCancel={cancel} footer={<><Button onClick={cancel}>取消</Button><Button type="primary" loading={busy} onClick={submit}>{intent && !busy ? '恢复原导入' : '导入章节'}</Button></>}>
    <p>导入已有章节，接续创作后续内容</p>
    <Form layout="vertical">
      <Form.Item label="导入到书籍（可选，不选则新建）"><Select aria-label="导入到书籍（可选，不选则新建）" allowClear disabled={frozen} value={draft.bookId || undefined} placeholder="选择已有书籍或留空新建" options={projects.filter(p => !['archived', 'purging'].includes(p.status)).map(p => ({ value: p.id, label: p.title || p.id }))} onChange={bookId => change({ bookId: bookId || '' })} /></Form.Item>
      {!draft.bookId && <Form.Item label="新书书名"><Input aria-label="新书书名" placeholder="暂定书名" disabled={frozen} value={draft.title} onChange={e => change({ title: e.target.value })} /></Form.Item>}
      <Form.Item label="章节正文"><Input.TextArea aria-label="章节正文" rows={8} placeholder="粘贴章节，建议以「第1章」等标题分段" disabled={frozen} value={draft.content} onChange={e => change({ content: e.target.value })} /></Form.Item>
      <Form.Item label="文风设置（多选）"><Checkbox.Group disabled={frozen || stylesLoading} value={draft.knowledgeBookIds} onChange={knowledgeBookIds => change({ knowledgeBookIds })} className="novel-import-style-options">{styles.map(style => <Checkbox key={style.id} value={style.id}>{style.title || style.id}</Checkbox>)}</Checkbox.Group>{stylesLoading ? <p role="status">正在读取文风…</p> : !styles.length && !styleError ? <p>暂无可用文风</p> : null}{styleError && <p role="alert">{styleError}<Button size="small" onClick={loadStyles}>重新读取文风</Button></p>}</Form.Item>
    </Form>
    {error && <p role="alert">{error}</p>}
    {intent && !busy && <p role="status">上次导入结果尚未核对。已保留原正文和导入编号，请恢复原导入。</p>}
  </Modal>;
}
