import Taro, { useDidHide, useDidShow } from '@tarojs/taro';
import { Button, Progress, Text, Video, View } from '@tarojs/components';
import { useEffect, useRef, useState } from 'react';
import { apiRequest } from '../../../services/api';
import { createAICutClient, type CutRecord } from '../../../services/aicut';
import { downloadCutExport } from '../../../services/aicutMedia';

const labels: Record<string, string> = { queued: '排队中', running: '正在导出', succeeded: '导出完成', failed: '导出失败', canceled: '已取消', cancel_requested: '正在取消' };
export default function CutExports({ projectId, timelineId, revision, canExport = true }: { projectId: string; timelineId: string; revision: number; canExport?: boolean }) {
  const [jobs, setJobs] = useState<CutRecord[]>([]), [error, setError] = useState(''), [busy, setBusy] = useState(''), [preview, setPreview] = useState('');
  const [loaded, setLoaded] = useState(false);
  const alive = useRef(true), visible = useRef(true), lock = useRef(false), read = useRef(0), timer = useRef<ReturnType<typeof setTimeout>>();
  const uncertain = useRef(false);
  const client = createAICutClient(apiRequest);
  const stop = () => { if (timer.current) clearTimeout(timer.current); timer.current = undefined; read.current++; };
  const refresh = async () => {
    stop(); const ticket = read.current;
    try {
      const value = await client.jobs(projectId, timelineId);
      if (!alive.current || !visible.current || ticket !== read.current) return;
      const scoped = value.filter(job => (job.timelineId || job.options?.timelineId || 'default') === timelineId);
      setJobs(scoped); setLoaded(true);
      if (scoped.some(job => ['queued', 'running', 'cancel_requested'].includes(job.state))) timer.current = setTimeout(() => void refresh(), 3000);
    } catch (cause: any) { if (alive.current && ticket === read.current) setError(cause.message || '任务读取失败，请刷新'); }
  };
  useEffect(() => { void refresh(); return () => { alive.current = false; stop(); }; }, []);
  useDidHide(() => { visible.current = false; stop(); setPreview(''); });
  useDidShow(() => { visible.current = true; void refresh(); });
  const run = async (label: string, work: () => Promise<void>) => {
    if (lock.current) return; lock.current = true; setBusy(label); setError('');
    try { await work(); } catch (cause: any) { if (alive.current) setError(cause.message || cause.errMsg || '操作失败'); }
    finally { lock.current = false; if (alive.current) setBusy(''); }
  };
  const exporting = jobs.some(job => ['queued', 'running', 'cancel_requested'].includes(job.state));
  return <View className="cut-export">
    <View className="cut-export-preset"><Text>MP4 · 1080P · 30fps</Text><Text className="cut-caption">标准质量 · 导出当前已保存的时间线</Text></View>
    {!canExport ? <Text className="cut-caption">当前时间线没有片段，可查看导出记录；添加并保存片段后才能开始导出。</Text> : null}
    <Button id="cut-export-start" className="cut-primary" disabled={!canExport || !loaded || Boolean(busy) || exporting || uncertain.current} onClick={() => void run('提交导出', async () => {
      if (!(await Taro.showModal({ title: '开始导出视频？', content: '将使用当前已保存的工程生成视频，其他未保存内容不会包含。', confirmText: '开始导出', cancelText: '暂不导出' })).confirm || !alive.current) return;
      uncertain.current = true;
      await client.export(projectId, timelineId, revision);
      if (alive.current) { uncertain.current = false; await refresh(); }
    })}>{exporting ? '视频正在导出' : '开始导出'}</Button>
    {uncertain.current && !busy ? <Text className="cut-caption">提交结果尚未确认，请刷新任务列表核对。重新打开面板前不会重复提交。</Text> : null}
    {error ? <Text className="cut-error">{error}</Text> : null}
    <View className="cut-heading cut-small"><Text>导出记录</Text><Button className="cut-link" disabled={Boolean(busy)} onClick={() => { setError(''); void refresh(); }}>刷新状态</Button></View>
    {!loaded && !error ? <Text className="cut-caption">正在读取真实导出任务…</Text> : loaded && !jobs.length ? <Text className="cut-caption">这条时间线还没有导出记录</Text> : null}
    {jobs.map(job => <View className="cut-job" key={job.id}><View className="cut-heading cut-small"><Text>{String(job.format || '').toUpperCase()} · {labels[job.state] || '状态待确认'}</Text>{Number.isFinite(job.progress) ? <Text>{Math.max(0, Math.min(100, job.progress))}%</Text> : null}</View>
      {Number.isFinite(job.progress) ? <Progress percent={Math.max(0, Math.min(100, job.progress))} activeColor="#38c9bd" backgroundColor="#42505a" strokeWidth={4} /> : null}
      {job.error ? <Text className="cut-error">{job.error}</Text> : null}
      <Text className="cut-caption">{job.createdAt ? new Date(job.createdAt).toLocaleString() : '任务时间未返回'}</Text>
      {['queued', 'running'].includes(job.state) ? <Button disabled={Boolean(busy)} onClick={() => void run('取消导出', async () => { if (!(await Taro.showModal({ title: '取消这次导出？', content: '工程和素材会保留。', confirmText: '取消导出' })).confirm || !alive.current) return; await client.cancel(projectId, timelineId, job.id); if (alive.current) await refresh(); })}>取消这次导出</Button> : null}
      {job.state === 'succeeded' ? <View className="cut-actions"><Button disabled={Boolean(busy)} onClick={() => void run('读取成片', async () => { const file = await downloadCutExport(projectId, timelineId, job.id, String(job.format || 'mp4')); if (alive.current) setPreview(file); })}>预览成片</Button><Button disabled={Boolean(busy) || job.format !== 'mp4'} onClick={() => void run('保存到相册', async () => { const file = await downloadCutExport(projectId, timelineId, job.id, 'mp4'); if (!alive.current) return; await Taro.saveVideoToPhotosAlbum({ filePath: file }); await Taro.showToast({ title: '已保存到相册', icon: 'success' }); })}>保存到相册</Button></View> : null}
      {job.state === 'failed' ? <Text className="cut-caption">修正工程后，可点击上方“开始导出”重试。</Text> : null}
    </View>)}
    {preview ? <Video id="cut-export-preview" src={preview} controls /> : null}
    {busy ? <Text className="cut-caption">正在{busy}…</Text> : null}
  </View>;
}
