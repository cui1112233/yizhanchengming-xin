import { Button, Modal } from 'antd';
import { useEffect, useState } from 'react';
import { novelProjectRequest } from '../../../shared/api/novelCreation.js';
import useWorkspaceScope from './useWorkspaceScope.js';

const dimensions = [['narrativeVoice','叙事声音与语气'],['dialogueStyle','对话风格'],['sceneDescription','场景描写特征'],['transitions','转折与衔接手法'],['pacing','节奏特征'],['diction','词汇偏好'],['emotionalExpression','情绪表达方式'],['distinctiveHabits','独特习惯']];
export default function LongImitationContext({ owner, projectId, refreshKey, onLoaded }) {
  const scope = useWorkspaceScope(`${owner}:${projectId}`);
  const [creation, setCreation] = useState(null), [open, setOpen] = useState(false), [error, setError] = useState('');
  const [reference, setReference] = useState(null), [readingReference, setReadingReference] = useState(false), [referenceError, setReferenceError] = useState('');
  async function load() {
    try {
      await scope.call(novelProjectRequest(projectId, '/long-imitation?summary=1', undefined, { signal: scope.signal }), result => { setCreation(result.creation || null); onLoaded(projectId, result.creation || null); setError(''); });
    } catch (e) {
      if (!scope.isCurrent()) return;
      if (e.status === 404) { setCreation(null); onLoaded(projectId, null); } else setError('仿写创建设置暂时无法读取');
    }
  }
  useEffect(() => { load(); }, [owner, projectId, refreshKey]);
  async function readReference() {
    setReadingReference(true); setReferenceError('');
    try { await scope.call(novelProjectRequest(projectId, '/long-imitation', undefined, { signal: scope.signal }), result => setReference(result.creation?.input?.referenceText ?? '')); }
    catch { if (scope.isCurrent()) setReferenceError('参考原文暂时无法读取，已保存内容不会因此被修改。'); }
    finally { if (scope.isCurrent()) setReadingReference(false); }
  }
  if (!creation) return error ? <p role="alert">{error}<Button size="small" onClick={load}>重新读取仿写设置</Button></p> : null;
  const input = creation.input || {}, analysis = creation.analysis;
  return <>
    <div className="novel-imitation-context"><span>仿写长篇 · 每章 {input.targetWordCount?.toLocaleString()} 字 · 目标 {input.targetChapters} 章</span><Button size="small" onClick={() => setOpen(true)}>查看仿写设置与拆解</Button></div>
    <Modal rootClassName="novel-modal novel-imitation-details" open={open} title="仿写设置与拆解" width={760} onCancel={() => setOpen(false)} footer={<Button onClick={() => setOpen(false)}>关闭</Button>}>
      <div className="novel-imitation-details-scroll">
        <h3>已保存的创建设置</h3><dl>{[['暂定标题',input.title || '留空，由蓝图自动命名'],['每章字数',input.targetWordCount],['目标章节数',input.targetChapters],['读者偏好',input.readerPreferences],['原创故事想法',input.storyIdea],['仿写方向',input.selectedDirection || '采用拆解后的方向一'],['补充信息',input.supplementaryInfo]].map(([label,value]) => value !== undefined && value !== '' && <div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>
        <p>参考原文 {creation.referenceCharacters?.toLocaleString() || '已保存'} 字元 · {creation.chunkCount ?? '已保存'} 个参考分块</p>
        <p>关联文风 {(creation.styles || []).map(style => style.title || style.id).join('、') || '无'}；备忘录 {(creation.memos || []).map(memo => memo.title || memo.id).join('、') || '无'}</p>
        {reference === null ? <Button loading={readingReference} onClick={readReference}>查看已保存参考原文</Button> : <pre className="novel-imitation-reference">{reference}</pre>}
        {referenceError && <p role="alert">{referenceError}</p>}
        {error && <p role="alert">{error}<Button size="small" onClick={load}>重新读取仿写设置</Button></p>}
        {creation.requirements && <section>
          <h3>长篇要求提炼</h3>
          <p>较长的创作要求或文风已分段整理。后续蓝图和写作使用以下已保存摘要，原始输入仍完整保留。</p>
          <pre className="novel-imitation-reference">{creation.requirements.summary}</pre>
        </section>}
        <h3>文风与故事拆解</h3>
        {!analysis ? <p>拆解结果尚未完成。蓝图生成后可在这里查看实际结果。</p> : <>
          <dl>{dimensions.map(([key,label]) => <div key={key}><dt>{label}</dt><dd>{analysis.dimensions?.[key] || '暂无结果'}</dd></div>)}</dl>
          <h4>情绪曲线</h4>{(analysis.emotionCurve || []).map((point,index) => <p key={index}><b>{point.stage} · {point.emotion}</b><br/>{point.reason}</p>)}
          <h4>核心价值观</h4><ul>{(analysis.coreValues || []).map((value,index) => <li key={index}>{value}</li>)}</ul>
          <h4>原创故事方向</h4>{(analysis.directions || []).map((direction,index) => <section key={index}><b>方向{index+1}：{direction.title}</b><p>{direction.summary}</p></section>)}
          <p><b>本次采用：</b>{analysis.selectedDirection}</p>
        </>}
      </div>
    </Modal>
  </>;
}
