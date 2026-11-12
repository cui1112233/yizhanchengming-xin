import { Alert, Button, Modal } from 'antd';
import { Mic } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { createAudioRecording } from './audioRecording.js';
import './audio-recorder.css';

const EMPTY = { status: 'idle', elapsedMs: 0, url: '', error: '', notice: '', canPause: false };
const STATUS = { idle: '准备录音', requesting: '等待麦克风权限', recording: '正在录音', paused: '录音已暂停', stopping: '正在整理录音', ready: '录音已完成', uploading: '正在保存到工程', saved: '已保存到工程', error: '暂时无法录音' };

export default function AudioRecorder({ onUpload, compact = false }) {
  const [open, setOpen] = useState(false);
  const [state, setState] = useState(EMPTY);
  const session = useRef(null);
  useEffect(() => () => { session.current?.dispose(); session.current = null; }, []);
  const show = () => {
    session.current?.dispose();
    session.current = createAudioRecording({ onChange: setState });
    setState(session.current.getState()); setOpen(true);
  };
  const close = () => { session.current?.dispose(); session.current = null; setOpen(false); };
  const save = async () => {
    const recording = session.current;
    if (!recording) return;
    await recording.save(onUpload);
    if (session.current === recording && recording.getState().status === 'saved') close();
  };
  const active = state.status === 'recording' || state.status === 'paused';
  const busy = ['requesting', 'stopping', 'uploading'].includes(state.status);
  const seconds = Math.floor(state.elapsedMs / 1000);
  const elapsed = `${String(Math.floor(seconds / 60)).padStart(2, '0')}:${String(seconds % 60).padStart(2, '0')}`;
  return <>
    <Button size="small" aria-label="麦克风录音" title="麦克风录音" icon={<Mic size={14} />} onClick={show}>{compact ? null : '录音'}</Button>
    <Modal open={open} title="麦克风录音" width={460} onCancel={close} maskClosable={false} footer={<>
      <Button onClick={close}>{state.status === 'uploading' ? '关闭' : '取消'}</Button>
      {state.status === 'ready' && <Button onClick={() => session.current?.start()}>重新录音</Button>}
      {state.status === 'ready' || state.status === 'uploading'
        ? <Button type="primary" loading={state.status === 'uploading'} disabled={!onUpload} onClick={save}>保存到工程</Button>
        : <Button type="primary" disabled={busy} danger={active} onClick={() => active ? session.current?.stop() : session.current?.start()}>{active ? '停止录音' : '开始录音'}</Button>}
    </>}>
      <div className="aicut-audio-recorder">
        <p>点击“开始录音”后才会申请麦克风权限。最长录制 10 分钟，录完可先试听，再保存为本工程的音频素材。</p>
        <div className="aicut-audio-recorder-status" role="status" aria-live="polite"><span className={state.status === 'recording' ? 'is-recording' : ''}>{STATUS[state.status]}</span><time aria-label="录音时长">{elapsed}</time></div>
        {active && state.canPause && <Button onClick={() => state.status === 'paused' ? session.current?.resume() : session.current?.pause()}>{state.status === 'paused' ? '继续录音' : '暂停录音'}</Button>}
        {state.url && <audio controls preload="metadata" src={state.url} aria-label="录音试听" />}
        {state.notice && <Alert type="info" showIcon message={state.notice} />}
        {state.error && <Alert type="error" showIcon message={state.error} />}
        {state.status === 'uploading' && <p>保存正在进行，关闭窗口后仍会继续保存，请勿重复上传。</p>}
      </div>
    </Modal>
  </>;
}
