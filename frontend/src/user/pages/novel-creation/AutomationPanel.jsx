import { Button, Form, Input, InputNumber, Select, Switch, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { novelProjectRequest } from '../../../shared/api/novelCreation.js';
import { getCurrentUsername } from '../../../shared/api/auth.js';
import useWorkspaceScope from './useWorkspaceScope.js';
import { MODE_FIELDS, STATUS, THINKING, availableModel, draftKey, modelOptions, readSaved } from './contracts.js';

export default function AutomationPanel({ projectId, sessionId, mode, input, models, kind = 'schedules', ensureSession }) {
  const scope = useWorkspaceScope(projectId + ':' + sessionId + ':' + mode); const [form] = Form.useForm(); const [items, setItems] = useState([]);
  const [references, setReferences] = useState([]), [notes, setNotes] = useState([]), [styles, setStyles] = useState([]); const [editing, setEditing] = useState('');
  const lock = useRef(false); const [busy, setBusy] = useState(false); const batch = kind === 'batches';
  const creationKey = draftKey(getCurrentUsername(), projectId, sessionId, mode) + ':create:' + kind;
  const [uncertain, setUncertain] = useState(() => readSaved(creationKey, null));
  const [connection, setConnection] = useState('');
  const request = (path, body, options = {}) => novelProjectRequest(projectId, path, body, { signal: scope.signal, ...options });
  const initial = { ...input, prompt: input.content || '', intervalMinutes: 30, dailyChapterCap: 10, maxConsecutiveFailures: 3, totalChapters: 5, enabled: true };
  async function load() { await scope.call(request('/' + kind), r => {
    const rows = r[kind] || []; setItems(rows);
    const pending = readSaved(creationKey, null);
    const fields = ['sessionId', 'prompt', 'modelId', 'thinkingLevel', 'styleProfileId', 'maxConsecutiveFailures', ...(batch ? ['totalChapters'] : ['intervalMinutes', 'dailyChapterCap'])];
    if (pending && rows.some(row => !pending.previousIds.includes(row.id) && fields.every(k => (row[k] ?? '') === (pending.payload[k] ?? '')) && ['modeInputs', 'knowledgeIds', 'noteIds'].every(k => JSON.stringify(row[k] || (k === 'modeInputs' ? {} : [])) === JSON.stringify(pending.payload[k] || (k === 'modeInputs' ? {} : []))))) {
      localStorage.removeItem(creationKey); setUncertain(null); message.success('已核对到原来创建的任务，没有再次提交');
    }
  }); }
  useEffect(() => {
    let timer, failures = 0;
    async function poll() { try { await load(); if (scope.isCurrent()) { failures = 0; setConnection(''); } } catch { if (scope.isCurrent()) { failures++; setConnection('任务列表连接中断，正在重新连接；不会重复提交。'); } } if (scope.isCurrent()) timer = setTimeout(poll, Math.min(4000 * 2 ** failures, 30000)); }
    poll();
    for (const [path, apply, key] of [['materials', setReferences, 'materials'], ['notes', setNotes, 'notes'], ['styles', setStyles, 'styles']]) scope.call(request('/' + path), r => apply(r[key] || [])).catch(() => {});
    return () => clearTimeout(timer);
  }, [projectId, sessionId, mode]);
  async function action(fn) { if (lock.current || !scope.isCurrent()) return; lock.current = true; setBusy(true); try { await fn(); } catch (e) { if (scope.isCurrent()) message.error(e.message); } finally { lock.current = false; if (scope.isCurrent()) setBusy(false); } }
  async function save() {
    if (!editing && readSaved(creationKey, null)) { message.info('请先核对原来的任务创建结果，不会重复发起'); return; }
    await action(async () => {
      const values = await form.validateFields();
      if (!availableModel(models, values.modelId)) { message.warning('所选文本模型当前不可用，请重新选择'); return; }
      const session = sessionId || (await ensureSession())?.id; if (!session || !scope.isCurrent()) return;
      const payload = { ...values, sessionId: session }, actualKey = draftKey(getCurrentUsername(), projectId, session, mode) + ':create:' + kind;
      if (!editing) { const pending = { payload, previousIds: items.map(i => i.id) }; localStorage.setItem(creationKey, JSON.stringify(pending)); localStorage.setItem(actualKey, JSON.stringify(pending)); setUncertain(pending); }
      try { await scope.call(request('/' + kind + (editing ? '/' + editing : ''), payload, { method: editing ? 'PUT' : 'POST' }), r => {
          localStorage.removeItem(creationKey); localStorage.removeItem(actualKey); setUncertain(null);
          const item = batch ? r.batch : r.schedule; setItems(old => editing ? old.map(v => v.id === editing ? item : v) : [item, ...old]);
          setEditing(''); message.success(batch ? '批量写作已创建' : '定时任务已保存，到期后执行');
        });
      } catch (e) { if (scope.isCurrent() && e.status >= 400 && e.status < 500 && ![408, 409, 429].includes(e.status)) { localStorage.removeItem(creationKey); localStorage.removeItem(actualKey); setUncertain(null); } throw e; }
    });
  }
  async function control(item, actionName) { await action(() => scope.call(request(`/${kind}/${item.id}/${actionName}`, {}), r => { const value = batch ? r.batch : r.schedule; setItems(old => old.map(i => i.id === item.id ? value : i)); })); }
  const withMissing = (list, ids) => [...list, ...(ids || []).filter(id => !list.some(i => i.value === id)).map(id => ({ value: id, label: '已选资料（重新导入后可更新）' }))];
  return <div className="novel-panel"><p>{batch ? '每章保存成功后才推进；失败不算完成章数。' : '默认30分钟、每日10章、连续失败3次暂停。保存不会立即生成。'}使用下列完整输入和引用，暂停/继续只改变执行状态。</p>
    {connection && <p role="status">{connection}</p>}
    {uncertain && <div className="novel-notice" role="status">上次创建结果尚未核对，暂停再次创建，避免重复执行。<Button onClick={() => action(load)}>核对原创建请求</Button></div>}
    <Form form={form} layout="vertical" initialValues={initial}>
      <div className="novel-form-grid">{(batch ? [['totalChapters', '总章数', 1, 50]] : [['intervalMinutes', '间隔分钟', 1, 1440], ['dailyChapterCap', '每日章节上限', 1, 100]]).concat([['maxConsecutiveFailures', '连续失败阈值', 1, 10]]).map(([key, label, min, max]) => <Form.Item name={key} key={key} label={label}><InputNumber aria-label={label} min={min} max={max} /></Form.Item>)}</div>
      <Form.Item name="modelId" label="文本模型" rules={[{ required: true, message: '请选择文本模型' }]}><Select aria-label="自动写作模型" options={modelOptions(models, input.modelId)} /></Form.Item>
      <Form.Item name="thinkingLevel" label="思考强度"><Select options={THINKING} /></Form.Item>
      <Form.Item name="prompt" label="每章写作指令" rules={[{ required: true, message: '请输入写作指令' }]}><Input.TextArea aria-label="自动写作指令" rows={3} /></Form.Item>
      {MODE_FIELDS[mode] && <Form.Item name={['modeInputs', MODE_FIELDS[mode][0]]} label={MODE_FIELDS[mode][1]}><Input.TextArea aria-label={'自动写作' + MODE_FIELDS[mode][1]} rows={3} /></Form.Item>}
      <Form.Item name="knowledgeIds" label="参考资料"><Select aria-label="自动写作参考资料" mode="multiple" options={withMissing(references.map(i => ({ value: i.id, label: i.title })), input.knowledgeIds)} /></Form.Item>
      <Form.Item name="noteIds" label="备忘录引用"><Select aria-label="自动写作备忘录" mode="multiple" options={withMissing(notes.map(i => ({ value: i.id, label: i.title })), input.noteIds)} /></Form.Item>
      <Form.Item name="styleProfileId" label="文风"><Select aria-label="自动写作文风" allowClear options={withMissing(styles.map(s => ({ value: s.profile.id, label: s.profile.title })), input.styleProfileId ? [input.styleProfileId] : [])} /></Form.Item>
      {!batch && <Form.Item name="enabled" label="启用计划" valuePropName="checked"><Switch /></Form.Item>}
      <div className="novel-actions"><Button type="primary" loading={busy} disabled={!!uncertain && !editing} onClick={save}>{editing ? '保存定时修改' : batch ? '开始批量续写' : '创建定时任务'}</Button>{editing && <Button onClick={() => { setEditing(''); form.setFieldsValue(initial); }}>取消修改</Button>}</div>
    </Form>
    {items.map(item => <section key={item.id} className="novel-item"><b>{STATUS[item.status] || '等待执行'}</b>
      <p>{batch ? `目标${item.totalChapters}章，已保存${item.completedChapters ?? 0}章，当前第${Math.min((item.completedChapters ?? 0) + 1, item.totalChapters)}章` : `每${item.intervalMinutes}分钟 · 今日${item.dailySavedChapters ?? 0}/${item.dailyChapterCap ?? 10}章 · 累计${item.totalSavedChapters ?? 0}章`}</p>
      <p>连续失败 {item.consecutiveFailures ?? 0}/{item.maxConsecutiveFailures ?? 3}次</p>
      {!batch && <p>下次时间：{item.nextRunAt ? new Date(item.nextRunAt).toLocaleString() : '待安排'}{item.eligibleAt ? ` · 可执行时间：${new Date(item.eligibleAt).toLocaleString()}` : ''}</p>}
      {[item.waitingReason, item.configurationError, item.stoppedReason, item.lastFailure, item.lastError].filter(Boolean).map((reason, i) => <p key={i} role="status">{STATUS[reason] || reason}</p>)}
      <p>{item.prompt}</p><div className="novel-actions">
        {['running', 'waiting', 'not_started'].includes(item.status) && <Button onClick={() => control(item, 'pause')}>暂停</Button>}
        {['paused', 'stopped', 'failed'].includes(item.status) && <Button onClick={() => control(item, 'resume')}>继续</Button>}
        {!['completed', 'stopped'].includes(item.status) && <Button onClick={() => control(item, 'stop')}>停止</Button>}
        {!batch && <><Button onClick={() => action(() => scope.call(request(`/schedules/${item.id}/run`, {}), () => { message.success('已请求立即执行，请查看会话'); load(); }))}>立即执行</Button><Button onClick={() => { setEditing(item.id); form.setFieldsValue(item); }}>修改配置</Button></>}
        <Button onClick={() => action(() => scope.call(request(`/${kind}/${item.id}/logs`), r => setItems(old => old.map(i => i.id === item.id ? { ...i, rounds: r.rounds || [] } : i))))}>刷新逐轮日志</Button>
      </div>
      <ol className="novel-rounds">{(item.rounds || []).map(round => <li key={round.key}>{round.immediate ? '立即执行' : '计划轮次'} · {STATUS[round.state] || (round.state === 'awaiting_confirmation' ? '请确认创作蓝图后继续' : '处理中')} · {round.plannedAt ? new Date(round.plannedAt).toLocaleString() : ''}{round.errorMessage ? ` · ${round.errorMessage}` : ''}</li>)}</ol>
    </section>)}
  </div>;
}
