import { useEffect, useRef, useState } from 'react';
import { Button } from 'antd';
import './upload-sent-button.css';

// 统一上传按钮动效：点击时轻压；提交成功后按钮内容切换为
// 「纸飞机飞出 + 上传成功」，随后自动恢复原状。
// 提交成功时在调用方执行 markSent() 即可。
export function useUploadSentState(resetMs = 1700) {
  const [sent, setSent] = useState(false);
  const timerRef = useRef(null);
  useEffect(() => () => window.clearTimeout(timerRef.current), []);
  function markSent() {
    window.clearTimeout(timerRef.current);
    // 先复位再在下一帧点亮，保证同一按钮连续成功时飞行动画也能重放。
    setSent(false);
    window.requestAnimationFrame(() => {
      setSent(true);
      timerRef.current = window.setTimeout(() => setSent(false), resetMs);
    });
  }
  return { sent, markSent };
}

function PaperPlaneIcon({ className }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      width="15"
      height="15"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M22 2 11 13" />
      <path d="M22 2 15 22 11 13 2 9Z" />
    </svg>
  );
}

export function UploadSentButton({ sent, sentLabel = '上传成功', icon, children, className = '', ...rest }) {
  return (
    <Button className={`${className} batch-factory-upload-sent${sent ? ' is-sent' : ''}`} {...rest}>
      {sent ? (
        <span className="batch-factory-upload-sent-success">
          <PaperPlaneIcon className="batch-factory-upload-sent-success-icon" />
          <span>{sentLabel}</span>
        </span>
      ) : (
        <>
          {icon}
          {children}
        </>
      )}
      {sent ? <PaperPlaneIcon className="batch-factory-upload-sent-plane" /> : null}
    </Button>
  );
}
