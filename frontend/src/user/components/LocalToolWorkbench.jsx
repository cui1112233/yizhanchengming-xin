import { Alert, Button, Checkbox, Input, InputNumber, Radio, Tag, message } from 'antd';
import { ArrowLeft, Clapperboard, Copy, Download, FileText, Files, Subtitles, Upload } from 'lucide-react';
import { useMemo, useRef, useState } from 'react';
import {
  downloadTextFile,
  formatSrtTime,
  formatText,
  mergeDocuments,
  mergeSubtitles,
  parseSubtitle,
  readTextFile
} from './localToolLogic.js';
import {
  buildStoryboardJson,
  buildStoryboardPreview,
  parseStoryboardDocument,
  parseStoryboardFiles
} from './storyboardOrganizer.js';

const { TextArea } = Input;

const TOOL_META = {
  'text-format': {
    icon: FileText,
    title: '格式化文本',
    color: '#fbbf24',
    summary: '纯浏览器本地处理：清理空白、规范中文标点、去掉中文之间的多余空格。不上传文件，不调用模型。'
  },
  'subtitle-merge': {
    icon: Subtitles,
    title: '字幕合并',
    color: '#84cc16',
    summary: '读取 SRT / ASS / VTT，按时间轴重排或按文件顺序拼接，去重后统一导出 SRT。全部在本机完成。'
  },
  'document-merge': {
    icon: Files,
    title: '文档合并',
    color: '#14b8a6',
    summary: '批量导入 TXT，按文件名/章节号排序合并，可选章节分隔与重复段落清理。全部在本机完成。'
  },
  'storyboard-organizer': {
    icon: Clapperboard,
    title: '分镜整理',
    color: '#8b5cf6',
    summary: '批量导入分镜文档，按 15 秒或 30 秒分组，导出结构化 JSON。支持表格、镜头标题块和带时间戳的分镜。'
  }
};

const SUBTITLE_ACCEPT = '.srt,.ass,.ssa,.vtt,text/plain';
const TEXT_ACCEPT = '.txt,.md,text/plain';

function nextId(file) {
  return `${file.name}-${file.size}-${Math.random().toString(36).slice(2, 8)}`;
}

function useCopy() {
  const [copied, setCopied] = useState(false);
  const copy = async value => {
    if (!value) return;
    try {
      await navigator.clipboard.writeText(value);
    } catch {
      const holder = document.createElement('textarea');
      holder.value = value;
      holder.style.position = 'fixed';
      holder.style.opacity = '0';
      document.body.appendChild(holder);
      holder.select();
      document.execCommand('copy');
      document.body.removeChild(holder);
    }
    setCopied(true);
    message.success('已复制到剪贴板');
    window.setTimeout(() => setCopied(false), 1600);
  };
  return { copied, copy };
}

function FileDrop({
  accept,
  multiple = true,
  hint,
  files,
  onAdd,
  onRemove,
  onClear
}) {
  const inputRef = useRef(null);
  const [dragging, setDragging] = useState(false);
  const [reading, setReading] = useState(false);

  async function ingest(fileList) {
    const picked = Array.from(fileList || []);
    if (!picked.length) return;
    setReading(true);
    try {
      const items = await Promise.all(picked.map(async file => ({
        id: nextId(file),
        name: file.name,
        size: file.size,
        content: await readTextFile(file)
      })));
      onAdd(items);
    } catch (error) {
      message.error(`文件读取失败：${error?.message || '未知错误'}`);
    } finally {
      setReading(false);
    }
  }

  return (
    <div
      className={`local-tool-drop${dragging ? ' is-dragging' : ''}`}
      onDragOver={event => {
        event.preventDefault();
        setDragging(true);
      }}
      onDragLeave={() => setDragging(false)}
      onDrop={event => {
        event.preventDefault();
        setDragging(false);
        ingest(event.dataTransfer?.files);
      }}
    >
      <div className="local-tool-drop-head">
        <span>已选 {files.length} 个文件</span>
        <div className="local-tool-drop-actions">
          <Button size="small" icon={<Upload size={14} />} loading={reading} onClick={() => inputRef.current?.click()}>选择文件</Button>
          {files.length ? <Button size="small" type="text" onClick={onClear}>清空</Button> : null}
        </div>
      </div>
      <label className="local-tool-drop-hint">
        <input
          ref={inputRef}
          type="file"
          accept={accept}
          multiple={multiple}
          hidden
          onChange={event => {
            const picked = event.target.files;
            event.target.value = '';
            ingest(picked);
          }}
        />
        {hint}
      </label>
      {files.length ? (
        <ul className="local-tool-file-list">
          {files.map(file => (
            <li key={file.id}>
              <span className="local-tool-file-name" title={file.name}>{file.name}</span>
              <span className="local-tool-file-meta">{file.meta}</span>
              <Button size="small" type="text" danger onClick={() => onRemove(file.id)}>移除</Button>
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}

function ResultToolbar({ text, fileName, withBom }) {
  const { copied, copy } = useCopy();
  return (
    <div className="local-tool-actions">
      <Button size="small" icon={copied ? null : <Copy size={14} />} onClick={() => copy(text)} disabled={!text}>
        {copied ? '已复制' : '复制结果'}
      </Button>
      <Button
        size="small"
        icon={<Download size={14} />}
        disabled={!text}
        onClick={() => downloadTextFile(fileName, text, 'text/plain;charset=utf-8', withBom)}
      >
        下载文件
      </Button>
      <span>共 {text ? text.length : 0} 字符</span>
    </div>
  );
}

/* ------------------------------------------------------------------ */

function TextFormatPanel() {
  const [input, setInput] = useState('');
  const [result, setResult] = useState(null);
  const [options, setOptions] = useState({
    trimLines: true,
    collapseBlank: true,
    removeBlank: false,
    normalizePunctuation: true,
    removeCjkSpaces: true,
    indentParagraphs: false
  });
  const fileRef = useRef(null);

  function toggle(key, value) {
    setOptions(current => ({ ...current, [key]: value }));
  }

  async function importFile(file) {
    if (!file) return;
    try {
      setInput(await readTextFile(file));
      message.success('文件已读入输入框');
    } catch (error) {
      message.error(`文件读取失败：${error?.message || '未知错误'}`);
    }
  }

  function run() {
    if (!input.trim()) return message.warning('请先粘贴或导入文本');
    setResult(formatText(input, options));
  }

  return (
    <>
      <div className="local-tool-panel">
        <div className="local-tool-field">
          <div className="local-tool-field-head">
            <span>原始文本</span>
            <Button size="small" icon={<Upload size={14} />} onClick={() => fileRef.current?.click()}>导入 txt / md</Button>
            <input
              ref={fileRef}
              type="file"
              accept={TEXT_ACCEPT}
              hidden
              onChange={event => {
                const file = event.target.files?.[0];
                event.target.value = '';
                importFile(file);
              }}
            />
          </div>
          <TextArea rows={10} value={input} onChange={event => setInput(event.target.value)} placeholder="粘贴需要整理的文本，或从本机导入 txt / md" />
        </div>

        <div className="local-tool-options">
          <Checkbox checked={options.trimLines} onChange={event => toggle('trimLines', event.target.checked)}>去掉每行首尾空白</Checkbox>
          <Checkbox checked={options.normalizePunctuation} onChange={event => toggle('normalizePunctuation', event.target.checked)}>中文标点规范化（仅改紧挨中文的半角标点）</Checkbox>
          <Checkbox checked={options.removeCjkSpaces} onChange={event => toggle('removeCjkSpaces', event.target.checked)}>去掉中文之间的空格</Checkbox>
          <Checkbox checked={options.collapseBlank} onChange={event => toggle('collapseBlank', event.target.checked)}>连续空行收成一个</Checkbox>
          <Checkbox checked={options.removeBlank} onChange={event => toggle('removeBlank', event.target.checked)}>删除全部空行</Checkbox>
          <Checkbox checked={options.indentParagraphs} onChange={event => toggle('indentParagraphs', event.target.checked)}>每段前加两个全角空格</Checkbox>
        </div>

        <div className="local-tool-actions">
          <Button type="primary" onClick={run}>开始处理</Button>
          <span>只在本机浏览器内运算，不发起网络请求。</span>
        </div>
      </div>

      {result ? (
        <div className="local-tool-result">
          <div className="local-tool-stats">
            <span>字符 <strong>{result.stats.charsBefore}</strong> → <strong>{result.stats.charsAfter}</strong></span>
            <span>行数 <strong>{result.stats.linesBefore}</strong> → <strong>{result.stats.linesAfter}</strong></span>
            <span>被改动的行 <strong>{result.stats.changedLines}</strong></span>
          </div>
          <TextArea rows={10} readOnly value={result.text} />
          <ResultToolbar text={result.text} fileName="格式化文本.txt" />
        </div>
      ) : null}
    </>
  );
}

/* ------------------------------------------------------------------ */

function SubtitleMergePanel() {
  const [files, setFiles] = useState([]);
  const [mode, setMode] = useState('sort');
  const [gapMs, setGapMs] = useState(500);
  const [dedupe, setDedupe] = useState(true);
  const [withBom, setWithBom] = useState(true);
  const [result, setResult] = useState(null);

  const descriptions = useMemo(() => files.map(file => {
    const parsed = parseSubtitle(file.content, { fileName: file.name });
    return { format: parsed.format, cues: parsed.cues.length };
  }), [files]);

  function addFiles(items) {
    setFiles(current => [...current, ...items]);
  }

  function updateOffset(id, value) {
    setFiles(current => current.map(file => (file.id === id ? { ...file, offsetMs: value } : file)));
  }

  function run() {
    if (!files.length) return message.warning('请先选择字幕文件');
    const merged = mergeSubtitles(files, { mode, gapMs, dedupe });
    if (!merged.cues.length) return message.warning('没有解析到有效字幕，请确认文件格式');
    setResult(merged);
  }

  return (
    <>
      <div className="local-tool-panel">
        <FileDrop
          accept={SUBTITLE_ACCEPT}
          hint="拖入或选择 .srt / .ass / .ssa / .vtt，可多选；GBK 与 UTF-8 编码会自动识别。"
          files={files.map((file, index) => ({
            id: file.id,
            name: file.name,
            meta: `${descriptions[index]?.format?.toUpperCase() || '-'} · ${descriptions[index]?.cues ?? 0} 条`
          }))}
          onAdd={addFiles}
          onRemove={id => setFiles(current => current.filter(file => file.id !== id))}
          onClear={() => setFiles([])}
        />

        {files.length ? (
          <div className="local-tool-offsets">
            <div className="local-tool-offsets-title">单文件时间偏移（秒）</div>
            {files.map((file, index) => (
              <label key={file.id} className="local-tool-offset-row">
                <span title={file.name}>{file.name}</span>
                <InputNumber
                  size="small"
                  min={-36000}
                  max={36000}
                  step={0.1}
                  value={(file.offsetMs || 0) / 1000}
                  onChange={value => updateOffset(file.id, Math.round((Number(value) || 0) * 1000))}
                  addonAfter="秒"
                  style={{ width: 140 }}
                />
                <em>{descriptions[index]?.cues ?? 0} 条</em>
              </label>
            ))}
          </div>
        ) : null}

        <div className="local-tool-field">
          <span>合并方式</span>
          <Radio.Group value={mode} onChange={event => setMode(event.target.value)}>
            <Radio value="sort">按时间轴重排（多份文件属于同一条时间线）</Radio>
            <Radio value="concat">按文件顺序首尾拼接（分集字幕合成一条）</Radio>
          </Radio.Group>
        </div>

        {mode === 'concat' ? (
          <div className="local-tool-field local-tool-inline">
            <span>文件之间留白</span>
            <InputNumber min={0} max={60000} step={100} value={gapMs} onChange={value => setGapMs(Number(value) || 0)} addonAfter="毫秒" />
          </div>
        ) : null}

        <div className="local-tool-options">
          <Checkbox checked={dedupe} onChange={event => setDedupe(event.target.checked)}>去掉重复字幕（文本相同且时间重叠 ≥ 80%）</Checkbox>
          <Checkbox checked={withBom} onChange={event => setWithBom(event.target.checked)}>导出时带 UTF-8 BOM（Windows 播放器/记事本兼容）</Checkbox>
        </div>

        <div className="local-tool-actions">
          <Button type="primary" onClick={run}>合并并导出 SRT</Button>
          <span>结果只包含字幕文本与时间轴，不含样式；ASS 的字体、颜色、特效不会保留。</span>
        </div>
      </div>

      {result ? (
        <div className="local-tool-result">
          <Alert
            type="success"
            showIcon
            message={`合并完成：${result.stats.cues} 条字幕，总时长 ${formatSrtTime(result.stats.durationMs)}`}
            description={`读取 ${result.stats.files} 个文件；去重 ${result.stats.duplicates} 条；丢弃时间非法 ${result.stats.dropped} 条。`}
          />
          <TextArea rows={10} readOnly value={result.text} />
          <ResultToolbar text={result.text} fileName="合并字幕.srt" withBom={withBom} />
        </div>
      ) : null}
    </>
  );
}

/* ------------------------------------------------------------------ */

const ORDER_OPTIONS = [
  { value: 'natural', label: '按文件名自然排序（part2 在 part10 前）' },
  { value: 'chapter', label: '按识别到的章节号排序（无章节号的排最后）' },
  { value: 'upload', label: '保持添加顺序' }
];

function DocumentMergePanel() {
  const [files, setFiles] = useState([]);
  const [order, setOrder] = useState('natural');
  const [separator, setSeparator] = useState('title');
  const [collapseBlank, setCollapseBlank] = useState(true);
  const [dedupeParagraphs, setDedupeParagraphs] = useState(false);
  const [result, setResult] = useState(null);

  function run() {
    if (!files.length) return message.warning('请先选择 TXT 文件');
    const merged = mergeDocuments(files, { order, separator, collapseBlank, dedupeParagraphs });
    if (!merged.stats.files) return message.warning('没有可合并的内容');
    setResult(merged);
  }

  return (
    <>
      <div className="local-tool-panel">
        <FileDrop
          accept={TEXT_ACCEPT}
          hint="拖入或选择 .txt / .md，可多选；按文件名里的章节号或自然顺序排序。"
          files={files.map(file => ({
            id: file.id,
            name: file.name,
            meta: `${file.content.length} 字符`
          }))}
          onAdd={items => setFiles(current => [...current, ...items])}
          onRemove={id => setFiles(current => current.filter(file => file.id !== id))}
          onClear={() => setFiles([])}
        />

        <div className="local-tool-field">
          <span>排序方式</span>
          <Radio.Group value={order} onChange={event => setOrder(event.target.value)} options={ORDER_OPTIONS} />
        </div>

        <div className="local-tool-field">
          <span>文档之间</span>
          <Radio.Group
            value={separator}
            onChange={event => setSeparator(event.target.value)}
            options={[
              { value: 'title', label: '插入 === 文件名 === 标题行' },
              { value: 'blank', label: '只空一行' },
              { value: 'none', label: '不加分隔' }
            ]}
          />
        </div>

        <div className="local-tool-options">
          <Checkbox checked={collapseBlank} onChange={event => setCollapseBlank(event.target.checked)}>连续空行收成一个</Checkbox>
          <Checkbox checked={dedupeParagraphs} onChange={event => setDedupeParagraphs(event.target.checked)}>去掉完全重复的段落</Checkbox>
        </div>

        <div className="local-tool-actions">
          <Button type="primary" onClick={run}>合并文档</Button>
          <span>只合并文本内容，不保留原文件的排版样式。</span>
        </div>
      </div>

      {result ? (
        <div className="local-tool-result">
          <div className="local-tool-stats">
            <span>文件 <strong>{result.stats.files}</strong> 个</span>
            <span>结果 <strong>{result.stats.chars}</strong> 字符</span>
          </div>
          <div className="local-tool-tags">
            {result.stats.perFile.map(file => (
              <Tag key={file.name} color={file.chapter ? 'green' : 'default'}>
                {file.name}{file.chapter ? ` · 第${file.chapter}章` : ' · 未识别章节号'}
              </Tag>
            ))}
          </div>
          <TextArea rows={10} readOnly value={result.text} />
          <ResultToolbar text={result.text} fileName="合并文档.txt" />
        </div>
      ) : null}
    </>
  );
}

/* ------------------------------------------------------------------ */

const STORYBOARD_ACCEPT = '.txt,.md,.srt,.ass,.csv,text/plain';

function StoryboardOrganizerPanel() {
  const [files, setFiles] = useState([]);
  const [pasted, setPasted] = useState('');
  const [target, setTarget] = useState(15);
  const [defaultShotSeconds, setDefaultShotSeconds] = useState(3);
  const [result, setResult] = useState(null);

  const counts = useMemo(() => files.map(file => (
    parseStoryboardDocument(file.content, { fileName: file.name, defaultShotSeconds }).length
  )), [files, defaultShotSeconds]);

  function run() {
    const sources = [...files];
    if (pasted.trim()) sources.push({ id: 'pasted', name: '粘贴内容.txt', content: pasted });
    if (!sources.length) return message.warning('请先导入分镜文档或粘贴分镜内容');
    const shots = parseStoryboardFiles(sources, { defaultShotSeconds });
    if (!shots.length) return message.warning('没有识别到分镜，请检查文档里的镜头编号、时长和画面描述');
    const json = buildStoryboardJson(shots, { targetSeconds: target });
    setResult({ json, preview: buildStoryboardPreview(shots, { targetSeconds: target }) });
  }

  const jsonText = result ? JSON.stringify(result.json, null, 2) : '';

  return (
    <>
      <div className="local-tool-panel">
        <FileDrop
          accept={STORYBOARD_ACCEPT}
          hint="拖入或选择分镜文档（.txt / .md / .srt / .csv），可多选；表格、镜头标题块、数字序号和带时间戳的格式都能识别。"
          files={files.map((file, index) => ({
            id: file.id,
            name: file.name,
            meta: `${counts[index]} 镜`
          }))}
          onAdd={items => setFiles(current => [...current, ...items])}
          onRemove={id => setFiles(current => current.filter(file => file.id !== id))}
          onClear={() => setFiles([])}
        />

        <div className="local-tool-field">
          <span>或者直接粘贴分镜文本</span>
          <TextArea rows={5} value={pasted} onChange={event => setPasted(event.target.value)} placeholder={'示例：\n镜头1（3秒）\n画面：雨夜街道，霓虹倒影\n运镜：缓慢推进\n台词：别回头'} />
        </div>

        <div className="local-tool-field">
          <span>分组时长</span>
          <Radio.Group value={target} onChange={event => setTarget(event.target.value)}>
            <Radio value={15}>15 秒一组</Radio>
            <Radio value={30}>30 秒一组</Radio>
          </Radio.Group>
        </div>

        <div className="local-tool-field local-tool-inline">
          <span>未标注时长的镜头按</span>
          <InputNumber min={0.5} max={60} step={0.5} value={defaultShotSeconds} onChange={value => setDefaultShotSeconds(Number(value) || 3)} addonAfter="秒" />
          <span>计算</span>
        </div>

        <div className="local-tool-actions">
          <Button type="primary" onClick={run}>整理并导出 JSON</Button>
          <span>单镜超过目标时长时独占一组；不会丢镜头。</span>
        </div>
      </div>

      {result ? (
        <div className="local-tool-result">
          <div className="local-tool-stats">
            <span>共 <strong>{result.json.totalShots}</strong> 镜</span>
            <span>总时长 <strong>{result.json.totalSeconds}</strong> 秒</span>
            <span>分为 <strong>{result.json.groupCount}</strong> 组（每组 {result.json.targetSeconds} 秒）</span>
          </div>
          <div className="local-tool-groups">
            {result.json.groups.map(group => (
              <article key={group.index} className="local-tool-group">
                <header>
                  <strong>第 {group.index} 组</strong>
                  <span>{group.startSeconds}s - {group.endSeconds}s · {group.durationSeconds} 秒 · {group.shots.length} 镜</span>
                </header>
                <ol>
                  {group.shots.map(shot => (
                    <li key={`${group.index}-${shot.index}`}>
                      <em>{shot.index}</em>
                      <span>{shot.content || shot.dialogue || '（无画面描述）'}</span>
                      <small>{shot.durationSeconds}s</small>
                    </li>
                  ))}
                </ol>
              </article>
            ))}
          </div>
          <TextArea rows={8} readOnly value={jsonText} />
          <ResultToolbar text={jsonText} fileName={`分镜整理-${result.json.targetSeconds}秒.json`} />
        </div>
      ) : null}
    </>
  );
}

/* ------------------------------------------------------------------ */

export default function LocalToolWorkbench({ tool, onBack }) {
  const meta = TOOL_META[tool] || TOOL_META['text-format'];
  const Icon = meta.icon;
  return (
    <section className="local-tool">
      <div className="local-tool-head">
        <Button size="small" icon={<ArrowLeft size={15} />} onClick={onBack}>返回插件市场</Button>
        <span className="local-tool-title">
          <Icon size={18} style={{ color: meta.color }} aria-hidden="true" />
          {meta.title}
        </span>
        <Tag color="green">本地执行</Tag>
      </div>
      <p className="local-tool-summary">{meta.summary}</p>

      {tool === 'subtitle-merge' ? <SubtitleMergePanel /> : null}
      {tool === 'document-merge' ? <DocumentMergePanel /> : null}
      {tool === 'storyboard-organizer' ? <StoryboardOrganizerPanel /> : null}
      {!['subtitle-merge', 'document-merge', 'storyboard-organizer'].includes(tool) ? <TextFormatPanel /> : null}
    </section>
  );
}
