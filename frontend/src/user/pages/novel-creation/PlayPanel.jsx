import { Button, Empty, Form, Input, Modal, Select, Spin, Tag, message } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { createNovelCreationPlay, getNovelCreationPlay, listNovelCreationPlays, patchNovelCreationPlay, stepNovelCreationPlay } from '../../../shared/api/novelCreation.js';
import useWorkspaceScope from './useWorkspaceScope.js';

const { TextArea } = Input;
const modeName = { open: '开放模式', guided: '引导模式' };
const visibleTurnCount = 10;

// 互动世界：建世界 → 开场 → 行动 → 推演 → 渲染 → 状态同步。
// 分支运行线、回合重做与版本恢复尚未移植，界面只提供线性推进与设定编辑。
export default function PlayPanel({ projectId }) {
  const scope = useWorkspaceScope(projectId);
  const [worlds, setWorlds] = useState([]), [current, setCurrent] = useState(null), [openingLoading, setOpeningLoading] = useState(false);
  const [loading, setLoading] = useState(false), [busy, setBusy] = useState(false), [errorText, setErrorText] = useState('');
  const [createOpen, setCreateOpen] = useState(false), [editOpen, setEditOpen] = useState(false), [action, setAction] = useState('');
  const [form] = Form.useForm(), [editForm] = Form.useForm();
  const bottom = useRef(null);
  const turns = current?.turns || [], visibleTurns = turns.slice(-visibleTurnCount);

  async function loadList() {
    if (!projectId) return;
    setLoading(true); setErrorText('');
    try {
      await scope.call(listNovelCreationPlays(projectId, scope.signal), r => setWorlds(r.worlds || []));
    } catch (e) {
      if (scope.isCurrent()) setErrorText(e.message || '互动世界列表读取失败');
    } finally {
      if (scope.isCurrent()) setLoading(false);
    }
  }

  async function openWorld(worldId) {
    setOpeningLoading(true); setErrorText('');
    try {
      await scope.call(getNovelCreationPlay(projectId, worldId, scope.signal), r => setCurrent(r.world || null));
    } catch (e) {
      if (scope.isCurrent()) setErrorText(e.message || '互动世界读取失败');
    } finally {
      if (scope.isCurrent()) setOpeningLoading(false);
    }
  }

  useEffect(() => { setCurrent(null); setAction(''); loadList(); }, [projectId]);
  useEffect(() => { if (turns.length) bottom.current?.scrollIntoView({ block: 'nearest' }); }, [turns.length]);

  async function submitCreate() {
    let values;
    try { values = await form.validateFields(); } catch { return; }
    setBusy(true);
    try {
      const result = await createNovelCreationPlay(projectId, values);
      if (!scope.isCurrent()) return;
      setWorlds(old => [result.world, ...old.filter(world => world.id !== result.world.id)]);
      setCurrent(result.world); setCreateOpen(false); form.resetFields();
      message.success(`「${result.world.title}」已开场`);
    } catch (e) {
      if (scope.isCurrent()) message.error(e.message || '创建互动世界失败');
    } finally {
      if (scope.isCurrent()) setBusy(false);
    }
  }

  async function submitAction(value) {
    const text = String(value ?? action).trim();
    if (!text) { message.warning('先写下你的行动'); return; }
    setBusy(true);
    try {
      const result = await stepNovelCreationPlay(projectId, current.id, { input: text });
      if (!scope.isCurrent()) return;
      setCurrent(result.world); setAction('');
      setWorlds(old => [result.world, ...old.filter(world => world.id !== result.world.id)]);
    } catch (e) {
      if (scope.isCurrent()) message.error(e.message || '推进失败');
    } finally {
      if (scope.isCurrent()) setBusy(false);
    }
  }

  function openEdit() {
    editForm.setFieldsValue({
      title: current.title, premise: current.premise, playerPersona: current.playerPersona,
      worldContract: current.worldContract, visualContract: current.visualContract
    });
    setEditOpen(true);
  }

  async function submitEdit() {
    let values;
    try { values = await editForm.validateFields(); } catch { return; }
    setBusy(true);
    try {
      const result = await patchNovelCreationPlay(projectId, current.id, values);
      if (!scope.isCurrent()) return;
      setCurrent(result.world);
      setWorlds(old => old.map(world => world.id === result.world.id ? result.world : world));
      setEditOpen(false);
      message.success('世界设定已更新');
    } catch (e) {
      if (scope.isCurrent()) message.error(e.message || '更新失败');
    } finally {
      if (scope.isCurrent()) setBusy(false);
    }
  }

  if (!projectId) return <Empty description="请先选择作品" />;
  return <div className="novel-panel" aria-label="互动世界">
    <p>每次推进调用三次模型（推演、渲染、状态同步）。分支运行线、回合重做与版本恢复尚未移植。</p>
    <div className="novel-play-toolbar">
      <Button type="primary" onClick={() => setCreateOpen(true)}>新建互动世界</Button>
      <Button onClick={loadList} loading={loading}>刷新列表</Button>
    </div>
    {errorText && <p role="alert">{errorText}</p>}
    {!!worlds.length && <div className="novel-play-worlds">{worlds.map(world => <button type="button" key={world.id} className={current?.id === world.id ? 'active' : ''} onClick={() => openWorld(world.id)}><b>{world.title}</b><small>{modeName[world.mode] || world.mode} · 第 {world.turn} 幕 · {world.turns?.length || 0} 条记录</small></button>)}</div>}
    {loading && !worlds.length && <Spin />}
    {!loading && !worlds.length && <Empty description="还没有互动世界" />}
    {openingLoading && <Spin />}
    {current && <section className="novel-play-stage">
      <header>
        <div><h4>{current.title}</h4><p>{modeName[current.mode] || current.mode} · 第 {current.turn} 幕{current.location ? ` · ${current.location}` : ''}{current.clock ? ` · ${current.clock}` : ''}</p></div>
        <Button size="small" onClick={openEdit}>编辑设定</Button>
      </header>
      <p className="novel-play-premise">{current.premise}</p>
      {!!current.entities?.length && <div className="novel-play-chips">{current.entities.map(entity => <Tag key={entity.name} color="blue">{entity.name}{entity.status ? ` · ${entity.status}` : ''}</Tag>)}</div>}
      {!!current.goals?.length && <p><b>当前目标：</b>{current.goals.join('；')}</p>}
      {turns.length > visibleTurns.length && <p role="status">更早的 {turns.length - visibleTurns.length} 幕已折叠，仅展示最近 {visibleTurnCount} 幕。</p>}
      <div className="novel-play-turns">{visibleTurns.map(turn => <article key={turn.turn} className="novel-play-turn">
        <header><b>第 {turn.turn} 幕</b>{turn.actionKind && <Tag>{turn.actionKind}</Tag>}</header>
        {turn.turn > 0 && <p className="novel-play-input">你：{turn.input}</p>}
        <p className="novel-play-scene">{turn.sceneText}</p>
        {!!turn.suggestedActions?.length && <div className="novel-play-chips">{turn.suggestedActions.map(suggestion => <Button key={suggestion} size="small" disabled={busy} onClick={() => submitAction(suggestion)}>{suggestion}</Button>)}</div>}
      </article>)}</div>
      <div ref={bottom} />
      <div className="novel-play-input-row">
        <TextArea rows={2} value={action} placeholder="写下你这一拍要做什么" onChange={e => setAction(e.target.value)} onPressEnter={e => { if (!e.shiftKey) { e.preventDefault(); submitAction(); } }} />
        <Button type="primary" loading={busy} onClick={() => submitAction()}>推进一幕</Button>
      </div>
      {!!current.facts?.length && <section><h4>已知事实</h4><ul>{current.facts.slice(-8).map(fact => <li key={fact}>{fact}</li>)}</ul></section>}
    </section>}

    <Modal open={createOpen} title="新建互动世界" confirmLoading={busy} onOk={submitCreate} onCancel={() => setCreateOpen(false)} okText="开场" destroyOnHidden>
      <Form form={form} layout="vertical" initialValues={{ mode: 'open' }}>
        <Form.Item name="title" label="世界名称" rules={[{ required: true, message: '请填写世界名称' }]}><Input maxLength={60} /></Form.Item>
        <Form.Item name="premise" label="世界设定" rules={[{ required: true, message: '请填写世界设定' }]}><TextArea rows={4} maxLength={2000} /></Form.Item>
        <Form.Item name="playerPersona" label="玩家身份"><Input maxLength={200} placeholder="例如：失忆的来客" /></Form.Item>
        <Form.Item name="worldContract" label="世界规则"><TextArea rows={2} maxLength={1000} placeholder="例如：超自然力量存在但代价高昂" /></Form.Item>
        <Form.Item name="mode" label="模式"><Select options={[{ value: 'open', label: '开放模式' }, { value: 'guided', label: '引导模式' }]} /></Form.Item>
      </Form>
    </Modal>

    <Modal open={editOpen} title="编辑世界设定" confirmLoading={busy} onOk={submitEdit} onCancel={() => setEditOpen(false)} okText="保存" destroyOnHidden>
      <Form form={editForm} layout="vertical">
        <Form.Item name="title" label="世界名称"><Input maxLength={60} /></Form.Item>
        <Form.Item name="premise" label="世界设定"><TextArea rows={3} maxLength={2000} /></Form.Item>
        <Form.Item name="playerPersona" label="玩家身份"><Input maxLength={200} /></Form.Item>
        <Form.Item name="worldContract" label="世界规则"><TextArea rows={2} maxLength={1000} /></Form.Item>
        <Form.Item name="visualContract" label="视觉约束"><TextArea rows={2} maxLength={1000} /></Form.Item>
      </Form>
    </Modal>
  </div>;
}
