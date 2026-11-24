import Taro from '@tarojs/taro';
import { Button, Input, Text, View } from '@tarojs/components';
import { useEffect, useRef, useState } from 'react';
import BrandHeader from '../../../components/BrandHeader';
import AppIcon from '../../../components/AppIcon';
import { ThemedPage } from '../../../theme';
import { apiRequest } from '../../../services/api';
import { createAICutClient, type CutRecord } from '../../../services/aicut';
import { onSessionInvalidated } from '../../../services/sessionBoundary';
import CutEditor from './CutEditor';
import './style.scss';

export default function AICutPage() {
  const [projects, setProjects] = useState<CutRecord[]>([]), [project, setProject] = useState<CutRecord | null>(null);
  const [query, setQuery] = useState(''), [busy, setBusy] = useState(''), [error, setError] = useState('');
  const [loaded, setLoaded] = useState(false), [dialog, setDialog] = useState<CutRecord | null>(null);
  const [name, setName] = useState(''), [ratio, setRatio] = useState('9:16');
  const alive = useRef(true), version = useRef(0), lock = useRef(false);
  const client = createAICutClient(apiRequest);
  const run = async (label: string, work: () => Promise<void>) => {
    if (lock.current) return; lock.current = true; const ticket = version.current; setBusy(label); setError('');
    try { await work(); } catch (cause: any) { if (alive.current && ticket === version.current) setError(cause.message || cause.errMsg || '操作失败，请重试'); }
    finally { if (alive.current && ticket === version.current) setBusy(''); lock.current = false; }
  };
  const reload = () => run('正在读取工程', async () => { const ticket = version.current; const list = await client.list(); if (alive.current && ticket === version.current) { setProjects(list); setLoaded(true); } });
  useEffect(() => {
    void reload();
    const off = onSessionInvalidated(() => { version.current++; setProjects([]); setProject(null); setDialog(null); setName(''); setLoaded(false); setBusy(''); setError('账号已切换，请返回功能页重新打开'); });
    return () => { alive.current = false; version.current++; off(); };
  }, []);
  const confirm = () => run('正在保存工程', async () => {
    if (!dialog) return;
    const ticket = version.current;
    if (dialog.id) await client.rename(dialog.id, name);
    else {
      const created = await client.create(name, ratio);
      if (alive.current && ticket === version.current) { setDialog(null); setProject(created); }
    }
    if (!alive.current || ticket !== version.current) return;
    setDialog(null); const list = await client.list();
    if (alive.current && ticket === version.current) { setProjects(list); setLoaded(true); }
  });
  const visible = projects.filter(item => String(item.title || '').toLocaleLowerCase().includes(query.trim().toLocaleLowerCase())).sort((a, b) => (Date.parse(b.updatedAt || b.createdAt) || 0) - (Date.parse(a.updatedAt || a.createdAt) || 0));
  return <ThemedPage className="mp-page cut-page" dynamicGlass={false}>
    {project ? <CutEditor key={project.id} project={project} onBack={() => { setProject(null); void reload(); }} /> : <>
      <BrandHeader title="AI-CUT" eyebrow="剪辑工作区" showBack />
      <View className="cut-content">
        <View className="cut-heading"><View><Text className="cut-title">我的剪辑</Text><Text className="cut-caption">把素材，剪成你的故事</Text></View><Button id="cut-create" className="cut-primary" disabled={Boolean(busy)} onClick={() => { setName(''); setRatio('9:16'); setDialog({}); }}><AppIcon name="plus" size={16} />新建</Button></View>
        <View className="cut-search"><AppIcon name="search" size={17} /><Input id="cut-search" value={query} onInput={e => setQuery(e.detail.value)} placeholder="搜索工程名称" confirmType="search" /></View>
        <View className="cut-heading cut-small"><Text>全部工程 · {loaded ? projects.length : '—'}</Text><Button className="cut-link" disabled={Boolean(busy)} onClick={() => void reload()}>刷新</Button></View>
        {busy ? <Text className="cut-caption">{busy}…</Text> : null}
        {error ? <View className="cut-error"><Text>{error}</Text><Button disabled={Boolean(busy)} onClick={() => void reload()}>重新读取</Button></View> : null}
        {!error && loaded && !visible.length ? <View className="cut-empty"><AppIcon name="clapperboard" size={34} /><Text>{query ? '没有找到这个工程' : '你的下一部作品，从这里开始'}</Text><Text className="cut-caption">{query ? '试试其他名称' : '新建工程，添加素材后开始剪辑'}</Text></View> : null}
        <View className="cut-projects">{visible.map(item => <View className="cut-project" key={item.id}>
          <Button className="cut-cover" disabled={Boolean(busy)} onClick={() => setProject(item)}><AppIcon name="clapperboard" size={36} /><Text className="cut-ratio">{item.aspectRatio}</Text><Text className="cut-open">继续剪辑 →</Text></Button>
          <View className="cut-project-copy"><Button className="cut-project-name" disabled={Boolean(busy)} onClick={() => setProject(item)}>{item.title || '未命名剪辑'}</Button><Text className="cut-caption">{(item.updatedAt || item.createdAt) ? new Date(item.updatedAt || item.createdAt).toLocaleDateString() : '时间未返回'}</Text><Button className="cut-link" disabled={Boolean(busy)} onClick={() => { setName(item.title || ''); setDialog(item); }}>重命名</Button></View>
        </View>)}</View>
      </View>
    </>}
    {dialog ? <View className="cut-overlay"><View className="cut-mask" onClick={() => { if (!busy) setDialog(null); }} /><View className="cut-sheet"><View className="cut-handle" /><Text className="cut-title">{dialog.id ? '重命名工程' : '新建剪辑工程'}</Text><Text className="cut-caption">工程名称</Text><Input id="cut-project-title" className="cut-input" maxlength={120} value={name} onInput={e => setName(e.detail.value)} placeholder="给这个故事起个名字" disabled={Boolean(busy)} />{!dialog.id ? <><Text className="cut-caption">画面比例</Text><View className="cut-ratios">{['9:16', '16:9', '1:1'].map(item => <Button key={item} className={ratio === item ? 'is-active' : ''} disabled={Boolean(busy)} onClick={() => setRatio(item)}>{item}</Button>)}</View></> : null}{error ? <Text className="cut-error">{error}</Text> : null}<View className="cut-actions"><Button disabled={Boolean(busy)} onClick={() => setDialog(null)}>取消</Button><Button className="cut-primary" loading={Boolean(busy)} disabled={Boolean(busy) || !name.trim()} onClick={() => void confirm()}>{dialog.id ? '保存名称' : '创建并进入剪辑'}</Button></View></View></View> : null}
  </ThemedPage>;
}
