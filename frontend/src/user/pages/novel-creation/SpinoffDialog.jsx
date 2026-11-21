import { Button, Checkbox, Form, Input, Modal, Select, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { novelRequest, listNovelCreationStyleProfiles } from '../../../shared/api/novelCreation.js';
import { readSaved } from './contracts.js';

import useWorkspaceScope from './useWorkspaceScope.js';

const emptyDraft = { parentBookId: '', title: '', spinoffType: 'sequel', storyIdea: '', referenceText: '', knowledgeBookIds: [] };
export default function SpinoffDialog({ owner, open, viewKey, projects, models, modelId, thinkingLevel, onCancel, onPersisted, onCreated }) {
  const storageKey = 'novel:spinoff:' + owner;
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
      if (!draft.title.trim()) { setError('请填写书名'); return; }
      if (!draft.parentBookId) { setError('请选择正传书籍'); return; }
      if (!(models || []).some(m => m.value === (draft.modelId || modelId))) { setError('请先在对话输入框下方选择可用的文本模型'); return; }
      pending = { draft, payload: { ...draft, title: draft.title.trim(), modelId: draft.modelId || modelId, thinkingLevel, idempotencyKey: crypto.randomUUID() } };
      // Persist the exact request before sending: a lost reply may already have saved it.
      try { localStorage.setItem(storageKey + ':intent', JSON.stringify(pending)); }
      catch { setError('创建恢复记录暂时无法保存，未提交'); return; }
      setIntent(pending);
    }
    flight.current = true; setBusy(true); setError('');
    const submittedView = view.current;
    try {
      const result = await scope.call(novelRequest('/spinoffs', pending.payload, {signal: scope.signal}));
      if (!result || !scope.isCurrent()) return;
      if (!result.book?.id || !result.session?.id || !result.run?.id) throw Error('创建结果不完整，请恢复原创建核对结果');
      // Saving remains real after closing the form. Update facts without taking over newer navigation.
      onPersisted(result); message.success('番外作品已创建，正在整理正传正典与番外设定');
      localStorage.removeItem(storageKey + ':intent'); localStorage.removeItem(storageKey + ':draft');
      setIntent(null); setDraft({ ...emptyDraft });
      if (view.current !== submittedView || !view.current.open) return;
      onCreated(result);
    } catch (e) {
      if (!scope.isCurrent()) return;
      // Only definite rejection unlocks editing. Unknown results keep the original key/input.
      if (e.status >= 400 && e.status < 500 && ![408, 409, 429].includes(e.status)) {
        localStorage.removeItem(storageKey + ':intent'); setIntent(null);
      }
      if (view.current !== submittedView || !view.current.open) return;
      setError(e.message || '创建结果暂未确认，请恢复原创建');
    } finally { flight.current = false; if (scope.isCurrent()) setBusy(false); }
  }
  return <Modal rootClassName="novel-modal novel-spinoff-dialog" open={open} title="番外创作" width="min(640px, calc(100vw - 32px))" onCancel={cancel} footer={<><Button onClick={cancel}>取消</Button><Button type="primary" loading={busy} onClick={submit}>{intent && !busy ? '恢复原创建' : '创建作品'}</Button></>}>
    <p>以已有正传为基础，整理正典并生成番外设定</p>
    <Form layout="vertical">
      <Form.Item label="正传书籍"><Select aria-label="正传书籍" allowClear disabled={frozen} value={draft.parentBookId || undefined} placeholder="选择已有正传" options={projects.filter(p => !['archived', 'purging'].includes(p.status)).map(p => ({ value: p.id, label: p.title || p.id }))} onChange={parentBookId => change({ parentBookId: parentBookId || '' })} /></Form.Item>
      <Form.Item label="书名"><Input aria-label="番外书名" maxLength={255} placeholder="暂定书名" disabled={frozen} value={draft.title} onChange={e => change({ title: e.target.value })} /></Form.Item>
      <Form.Item label="番外类型"><Select aria-label="番外类型" value={draft.spinoffType} disabled={frozen} options={[{value:'sequel',label:'后传'},{value:'prequel',label:'前传'},{value:'parallel',label:'平行线'},{value:'if',label:'If 线'}]} onChange={spinoffType => change({spinoffType})}/></Form.Item>
      <Form.Item label="番外方向"><Input.TextArea aria-label="番外方向" rows={2} placeholder="番外要写什么" disabled={frozen} value={draft.storyIdea} onChange={e => change({storyIdea:e.target.value})}/></Form.Item>
      <Form.Item label="参考原文（可选）"><Input.TextArea aria-label="番外参考原文" rows={3} disabled={frozen} value={draft.referenceText} onChange={e => change({referenceText:e.target.value})}/></Form.Item>
      <Form.Item label="文本模型"><Select aria-label="番外文本模型" disabled={frozen} value={draft.modelId || modelId || undefined} placeholder="选择本系统可用模型" options={models || []} onChange={modelId => change({modelId})}/></Form.Item>
      <Form.Item label="文风设置（多选）"><Checkbox.Group disabled={frozen || stylesLoading} value={draft.knowledgeBookIds} onChange={knowledgeBookIds => change({ knowledgeBookIds })} className="novel-import-style-options">{styles.map(style => <Checkbox key={style.id} value={style.id}>{style.title || style.id}</Checkbox>)}</Checkbox.Group>{stylesLoading ? <p role="status">正在读取文风…</p> : !styles.length && !styleError ? <p>暂无可用文风</p> : null}{styleError && <p role="alert">{styleError}<Button size="small" onClick={loadStyles}>重新读取文风</Button></p>}</Form.Item>
    </Form>
    {error && <p role="alert">{error}</p>}
    {intent && !busy && <p role="status">上次创建结果尚未核对。已保留原资料和请求编号，请恢复原创建。</p>}
  </Modal>;
}
