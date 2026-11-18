import { Button, Form, Input, InputNumber, Select, Switch, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { getNovelCreationSessionSnapshot, novelProjectRequest } from '../../../shared/api/novelCreation.js';
import { getCurrentUsername } from '../../../shared/api/auth.js';
import useWorkspaceScope from './useWorkspaceScope.js';
import { MODE_FIELDS, STATUS, THINKING, availableModel, draftKey, modelOptions, readSaved } from './contracts.js';

const READ_ONLY_PLAN_FIELDS = new Set(['id','ownerUsername','projectId','creation','status','createdAt','updatedAt','nextRunAt','lastRunAt','lastRunId','currentRunId','completedChapters','consecutiveFailures','totalSavedChapters','dailySavedChapters','successDay','stoppedReason','lastFailure','waitingReason','eligibleAt','activeRound','rounds','configurationError','lastError']);
function editablePlanInput(plan) { return Object.fromEntries(Object.entries(plan).filter(([key]) => !READ_ONLY_PLAN_FIELDS.has(key))); }

export default function AutomationPanel({ projectId, sessionId, mode, input, models, kind = 'schedules', ensureSession, manageOnly = false }) {
  const scope = useWorkspaceScope(projectId + ':' + sessionId + ':' + mode); const [form] = Form.useForm(); const [items, setItems] = useState([]);
  const [references, setReferences] = useState([]), [notes, setNotes] = useState([]), [styles, setStyles] = useState([]); const [editing, setEditing] = useState(null);
  const lock = useRef(false), editEpoch = useRef(0); const [busy, setBusy] = useState(false); const batch = kind === 'batches';
  const creationKey = draftKey(getCurrentUsername(), projectId, sessionId, mode) + ':create:' + kind;
  const [uncertain, setUncertain] = useState(() => readSaved(creationKey, null));
  const [connection, setConnection] = useState('');
  const request = (path, body, options = {}) => novelProjectRequest(projectId, path, body, { signal: scope.signal, ...options });
  const initial = { ...input, prompt: input.content || '', intervalMinutes: 30, dailyChapterCap: 10, maxConsecutiveFailures: 3, totalChapters: 5, enabled: true };
  async function load() { await scope.call(request('/' + kind), r => setItems(r[kind] || [])); }
  const editorMode = editing ? editing.mode : mode;
  const selectedInput = editing?.plan || input;
  function acknowledge(pending, result) {
    const row = batch ? result.batch : result.schedule;
    if (!row?.id || row.creation?.idempotencyKey !== pending.payload.idempotencyKey) throw Error('创建结果的编号不匹配，原请求已保留，请恢复核对');
    localStorage.removeItem(creationKey); for (const alias of pending.aliases || []) localStorage.removeItem(alias); setUncertain(null);
    setItems(old => [row, ...old.filter(p => p.id !== row.id)]);
  }
  async function recoverCreation() {
    await action(async () => {
      const pending = readSaved(creationKey, null);
      if (!pending?.payload?.idempotencyKey) { message.error('原创建记录没有可核对的创建编号，请先核对原任务；不会再次创建'); return; }
      await scope.call(request('/' + kind, pending.payload), r => { acknowledge(pending, r); message.success('已恢复原创建请求，原任务状态保持不变'); });
    });
  }
  async function editPlan(plan) {
    if (!scope.isCurrent() || lock.current) return;
    const epoch = ++editEpoch.current;
    setEditing({ plan, mode: null, loading: true, error: '' }); form.resetFields(); form.setFieldsValue(plan);
    try { await scope.call(getNovelCreationSessionSnapshot(plan.sessionId, scope.signal), r => {
      if (epoch !== editEpoch.current) return;
      if (r.session?.id !== plan.sessionId || r.session?.projectId !== projectId || !r.session.mode) throw Error('原计划会话不匹配，不能保存修改');
      setEditing({ plan, mode: r.session.mode, loading: false, error: '' });
    }); } catch (e) { if (scope.isCurrent() && epoch === editEpoch.current) setEditing({ plan, mode: null, loading: false, error: '读取原计划会话失败：' + e.message }); }
  }
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
    if (editing && (editing.loading || editing.error || !editing.mode)) return;
    await action(async () => {
      const values = await form.validateFields();
      if (!availableModel(models, values.modelId)) { message.warning('所选文本模型当前不可用，请重新选择'); return; }
      const session = editing ? editing.plan.sessionId : sessionId || (await ensureSession())?.id; if (!session || !scope.isCurrent()) return;
      const payload = editing ? { ...editablePlanInput(editing.plan), ...values, modeInputs: { ...editing.plan.modeInputs, ...values.modeInputs }, sessionId: session } : { ...values, sessionId: session, idempotencyKey: crypto.randomUUID() };
      const actualKey = draftKey(getCurrentUsername(), projectId, session, mode) + ':create:' + kind;
      const pending = { payload, aliases: [creationKey, actualKey] };
      if (!editing) { localStorage.setItem(creationKey, JSON.stringify(pending)); localStorage.setItem(actualKey, JSON.stringify(pending)); setUncertain(pending); }
      try { await scope.call(request('/' + kind + (editing ? '/' + editing.plan.id : ''), payload, { method: editing ? 'PUT' : 'POST' }), r => {
          if (!editing) acknowledge(pending, r);
          else setItems(old => old.map(v => v.id === editing.plan.id ? r.schedule : v));
          if (editing) { form.resetFields(); form.setFieldsValue(initial); }
          setEditing(null); message.success(batch ? '批量写作已创建' : '定时任务已保存，到期后执行');
        });
      } catch (e) { if (!editing && scope.isCurrent() && e.status >= 400 && e.status < 500 && ![408, 409, 429].includes(e.status)) { localStorage.removeItem(creationKey); localStorage.removeItem(actualKey); setUncertain(null); } throw e; }
    });
  }
  async function control(item, actionName) { await action(() => scope.call(request(`/${kind}/${item.id}/${actionName}`, {}), r => { const value = batch ? r.batch : r.schedule; setItems(old => old.map(i => i.id === item.id ? value : i)); })); }
  const withMissing = (list, ids) => [...list, ...(ids || []).filter(id => !list.some(i => i.value === id)).map(id => ({ value: id, label: '已选资料（重新导入后可更新）' }))];
  return <div className="novel-panel"><p>{batch ? '每章保存成功后才推进；失败不算完成章数。' : '默认30分钟、每日10章、连续失败3次暂停。保存不会立即生成。'}使用下列完整输入和引用，暂停/继续只改变执行状态。</p>
    {connection && <p role="status">{connection}</p>}
    {uncertain && <div className="novel-notice" role="status">上次创建结果尚未核对，恢复只使用原创建编号和原始输入，不会重置任务。<Button loading={busy} onClick={recoverCreation}>恢复原创建请求</Button></div>}
    {editing?.loading && <p role="status">正在读取原计划会话与模式…</p>}{editing?.error && <div role="alert">{editing.error}<Button onClick={() => editPlan(editing.plan)}>重新读取原会话</Button></div>}
    {(!manageOnly || editing) && <Form form={form} layout="vertical" initialValues={initial} disabled={!!editing && (editing.loading || !!editing.error)}>
      <div className="novel-form-grid">{(batch ? [['totalChapters', '总章数', 1, 50]] : [['intervalMinutes', '间隔分钟', 1, 1440], ['dailyChapterCap', '每日章节上限', 1, 100]]).concat([['maxConsecutiveFailures', '连续失败阈值', 1, 10]]).map(([key, label, min, max]) => <Form.Item name={key} key={key} label={label}><InputNumber aria-label={label} min={min} max={max} /></Form.Item>)}</div>
      <Form.Item name="modelId" label="文本模型" rules={[{ required: true, message: '请选择文本模型' }]}><Select aria-label="自动写作模型" options={modelOptions(models, selectedInput.modelId)} /></Form.Item>
      <Form.Item name="thinkingLevel" label="思考强度"><Select options={THINKING} /></Form.Item>
      <Form.Item name="prompt" label="每章写作指令" rules={[{ required: true, message: '请输入写作指令' }]}><Input.TextArea aria-label="自动写作指令" rows={3} /></Form.Item>
      {MODE_FIELDS[editorMode] && <Form.Item name={['modeInputs', MODE_FIELDS[editorMode][0]]} label={MODE_FIELDS[editorMode][1]}><Input.TextArea aria-label={'自动写作' + MODE_FIELDS[editorMode][1]} rows={3} /></Form.Item>}
      <Form.Item name="knowledgeIds" label="参考资料"><Select aria-label="自动写作参考资料" mode="multiple" options={withMissing(references.map(i => ({ value: i.id, label: i.title })), selectedInput.knowledgeIds)} /></Form.Item>
      <Form.Item name="noteIds" label="备忘录引用"><Select aria-label="自动写作备忘录" mode="multiple" options={withMissing(notes.map(i => ({ value: i.id, label: i.title })), selectedInput.noteIds)} /></Form.Item>
      <Form.Item name="styleProfileId" label="文风"><Select aria-label="自动写作文风" allowClear options={withMissing(styles.map(s => ({ value: s.profile.id, label: s.profile.title })), selectedInput.styleProfileId ? [selectedInput.styleProfileId] : [])} /></Form.Item>
      {!batch && <Form.Item name="enabled" label="启用计划" valuePropName="checked"><Switch /></Form.Item>}
      <div className="novel-actions"><Button type="primary" loading={busy} disabled={editing ? editing.loading || !!editing.error : !!uncertain} onClick={save}>{editing ? '保存定时修改' : batch ? '开始批量续写' : '创建定时任务'}</Button>{editing && <Button disabled={false} onClick={() => { ++editEpoch.current; setEditing(null); form.resetFields(); form.setFieldsValue(initial); }}>取消修改</Button>}</div>
    </Form>}
    {items.map(item => <section key={item.id} className="novel-item"><b>{STATUS[item.status] || '等待执行'}</b>
      <p>{batch ? `目标${item.totalChapters}章，已保存${item.completedChapters ?? 0}章，当前第${Math.min((item.completedChapters ?? 0) + 1, item.totalChapters)}章` : `每${item.intervalMinutes}分钟 · 今日${item.dailySavedChapters ?? 0}/${item.dailyChapterCap ?? 10}章 · 累计${item.totalSavedChapters ?? 0}章`}</p>
      <p>连续失败 {item.consecutiveFailures ?? 0}/{item.maxConsecutiveFailures ?? 3}次</p>
      {!batch && <p>下次时间：{item.nextRunAt ? new Date(item.nextRunAt).toLocaleString() : '待安排'}{item.eligibleAt ? ` · 可执行时间：${new Date(item.eligibleAt).toLocaleString()}` : ''}</p>}
      {[item.waitingReason, item.configurationError, item.stoppedReason, item.lastFailure, item.lastError].filter(Boolean).map((reason, i) => <p key={i} role="status">{STATUS[reason] || reason}</p>)}
      <p>{item.prompt}</p><div className="novel-actions">
        {['running', 'waiting', 'not_started'].includes(item.status) && <Button onClick={() => control(item, 'pause')}>暂停</Button>}
        {(batch ? ['paused', 'failed'] : ['paused', 'stopped', 'failed']).includes(item.status) && <Button onClick={() => control(item, 'resume')}>继续</Button>}
        {batch && item.status === 'stopped' && <small>此批次已停止，如需继续写作请新建批次。</small>}
        {!['completed', 'stopped'].includes(item.status) && <Button onClick={() => control(item, 'stop')}>停止</Button>}
        {!batch && <><Button onClick={() => action(() => scope.call(request(`/schedules/${item.id}/run`, {}), () => { message.success('已请求立即执行，请查看会话'); load(); }))}>立即执行</Button><Button onClick={() => editPlan(item)}>修改配置</Button></>}
        <Button onClick={() => action(() => scope.call(request(`/${kind}/${item.id}/logs`), r => setItems(old => old.map(i => i.id === item.id ? { ...i, rounds: r.rounds || [] } : i))))}>刷新逐轮日志</Button>
      </div>
      <ol className="novel-rounds">{(item.rounds || []).map(round => <li key={round.key}>{round.immediate ? '立即执行' : '计划轮次'} · {STATUS[round.state] || (round.state === 'awaiting_confirmation' ? '请确认创作蓝图后继续' : '处理中')} · {round.plannedAt ? new Date(round.plannedAt).toLocaleString() : ''}{round.errorMessage ? ` · ${round.errorMessage}` : ''}</li>)}</ol>
    </section>)}
  </div>;
}
