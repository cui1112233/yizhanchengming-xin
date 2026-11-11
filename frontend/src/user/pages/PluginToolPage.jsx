import { Button, Dropdown, Input, Select, Switch, Upload, message } from 'antd';
import { AlignLeft, Copy, Film, Flame, Image, Save, Trash2, Upload as UploadIcon, UserRound } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { apiRequest } from '../../shared/api/client.js';
import { listAvailableModels, modelSelectOptions } from '../../shared/api/modelCatalog.js';
import { getPluginToolConfig, runPluginTool, extractDocumentText } from '../../shared/api/pluginTools.js';
import PluginPanelShell, { openPluginMarket } from '../components/PluginPanelShell.jsx';
import { PLUGIN_CATALOG, PLUGIN_TOOL_PANELS } from './pluginCatalog.js';
import { createGenerationGate, executeGeneration, validateGenerationInput } from './pluginGeneration.js';
import {
  RESULT_PLACEHOLDER,
  buildDialogueList,
  buildPartPrompt,
  extractScript,
  formatScriptText,
  splitInputByLength
} from './pluginToolLogic.js';

const { TextArea } = Input;

// 标题左侧小图标：洛水 bundle 里小说转剧是 FilmOutline(#14b8a6)、分镜是 ImageOutline(#8b5cf6)、
// 爆款开头是 FlameOutline(#f59430)、角色修改是 PersonOutline(#fb7185)，尺寸 22。
const TITLE_ICONS = {
  'novel-to-drama': Film,
  'storyboard-master': Image,
  'viral-opening': Flame,
  'character-edit': UserRound
};

function currentPromptTool() {
  const pathname = window.location.pathname;
  const entry = PLUGIN_CATALOG.find(plugin => plugin.launch.kind === 'prompt-tool' && plugin.launch.href === pathname);
  if (!entry) return null;
  return { entry, panel: PLUGIN_TOOL_PANELS[entry.launch.toolId] };
}

function downloadFile(name, value, type) {
  const url = URL.createObjectURL(new Blob([value], { type }));
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = name;
  anchor.click();
  URL.revokeObjectURL(url);
}

// 洛水 4 条文字类是整页面板：内容区整个换掉，单栏上下堆叠，
// 输入区 → 整宽主按钮 → 结果区。规格来自实测控件表，模型与提示词用 V88 自己的库。
export default function PluginToolPage() {
  const tool = useMemo(currentPromptTool, []);
  const panel = tool?.panel;
  const toolId = tool?.entry?.launch?.toolId || '';
  const TitleIcon = TITLE_ICONS[toolId];

  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');
  const [models, setModels] = useState([]);
  const [prompts, setPrompts] = useState([]);
  const [modelId, setModelId] = useState();
  const [promptId, setPromptId] = useState();
  const [loadingConfig, setLoadingConfig] = useState(true);
  const [running, setRunning] = useState(false);
  const [concurrent, setConcurrent] = useState(false);
  const [progress, setProgress] = useState({ current: 0, total: 0 });
  // 卸载/关闭面板/切换工具时真正中止请求，而不是把结果丢掉。
  const gateRef = useRef(null);
  const mountedRef = useRef(true);
  if (!gateRef.current) gateRef.current = createGenerationGate();

  useEffect(() => () => { mountedRef.current = false; gateRef.current.abort(); }, []);

  useEffect(() => {
    if (!toolId) return undefined;
    let alive = true;
    (async () => {
      try {
        const [availableModels, promptOptions] = await Promise.all([
          listAvailableModels('text'),
          getPluginToolConfig(toolId)
        ]);
        if (!alive) return;
        setModels(availableModels);
        setPrompts(promptOptions);
      } catch (error) {
        if (alive) message.error(error?.message || '加载模型和提示词失败');
      } finally {
        if (alive) setLoadingConfig(false);
      }
    })();
    return () => { alive = false; };
  }, [toolId]);

  async function readUpload(file) {
    const lowerName = String(file?.name || '').toLowerCase();
    if (lowerName.endsWith('.txt') || lowerName.endsWith('.md')) {
      try {
        setInput(await file.text());
        message.success('文件读取成功');
      } catch {
        message.error('文件读取失败');
      }
      return Upload.LIST_IGNORE;
    }
    if (lowerName.endsWith('.doc') || lowerName.endsWith('.docx')) {
      try {
        setInput(await extractDocumentText(file));
        message.success('文件读取成功');
      } catch (error) {
        message.error(error?.message || 'Word 文档读取失败');
      }
      return Upload.LIST_IGNORE;
    }
    message.warning('仅支持 TXT / Word 文档');
    return Upload.LIST_IGNORE;
  }

  // 洛水分段生成：每段失败自动重试 1 次，空结果按失败处理；取消不重试。
  async function generatePart(prompt, request = apiRequest) {
    let lastError = null;
    for (let attempt = 0; attempt < 2; attempt += 1) {
      try {
        const text = await runPluginTool({ toolId, input: prompt, modelId, promptId, raw: true }, request);
        if (text.trim()) return text;
        lastError = new Error('AI返回内容为空');
      } catch (error) {
        lastError = error;
        if (error?.name === 'AbortError' || error?.aborted) throw error;
      }
    }
    throw lastError || new Error('生成失败');
  }

  async function generate() {
    // 主按钮点击前不禁用；空输入点击后用提示拦住，文案按工具区分。
    if (running) return message.info('正在生成，请勿重复提交');
    const validation = validateGenerationInput({ input, modelId, promptId, emptyMessage: panel?.emptyInputMessage });
    if (!validation.ok) return message.warning(validation.message);
    const chunkSize = Number(tool?.entry?.launch?.progressChunkSize) || 0;
    const parts = splitInputByLength(input, chunkSize);
    const gate = gateRef.current;
    const controller = gate.begin();
    if (!controller) return message.info('正在生成，请勿重复提交');
    const requestWithSignal = (path, options = {}) => apiRequest(path, { ...options, signal: controller.signal });
    const total = parts.length;
    setRunning(true);
    setOutput('');
    setProgress({ current: 0, total });
    try {
      const outcome = await executeGeneration({
        parts,
        concurrent,
        signal: controller.signal,
        onProgress: (current, done) => setProgress({ current, total: done }),
        runPart: (part, index, previous) => generatePart(
          buildPartPrompt({ full: input, part, index: index + 1, total, history: previous.join('\n\n') }),
          requestWithSignal
        )
      });
      if (outcome.cancelled) {
        // 用户取消不是失败：不弹错误，只在面板还在时提示「已取消」。
        if (mountedRef.current) message.info('已取消');
        return;
      }
      if (!outcome.ok) {
        const kept = Array.isArray(outcome.results) ? outcome.results.length : 0;
        if (kept > 0) {
          // 已经生成好的段落必须写回界面，别让用户整批重跑。
          setOutput(outcome.text);
          const index = Number(outcome.failedIndex) + 1;
          const reason = outcome.error?.message ? `：${outcome.error.message}` : '';
          message.error(concurrent
            ? `第 ${index} 段失败，已生成的 ${kept} 段已保留，可重试剩余段${reason}`
            : `第 ${index} 段失败，前 ${kept} 段已保留，可重试剩余段${reason}`);
          return;
        }
        message.error(outcome.error?.message || '生成失败');
        return;
      }
      setOutput(outcome.text);
      message.success(total > 1 ? `生成成功，共${total}段` : '生成成功');
    } finally {
      gate.end(controller);
      setRunning(false);
    }
  }

  function copyOutput() {
    if (!output) return message.warning('没有可复制的内容');
    if (panel.formatMode === 'script') {
      const extracted = extractScript(output);
      const content = extracted.success ? extracted.content : output;
      navigator.clipboard.writeText(content)
        .then(() => message.success(extracted.success ? `已提取并复制，共${extracted.scriptCount}批${extracted.lineCount}行` : '已复制到剪贴板'))
        .catch(() => message.error('复制失败'));
      return;
    }
    navigator.clipboard.writeText(output)
      .then(() => message.success('已复制到剪贴板'))
      .catch(() => message.error('复制失败'));
  }

  function clearOutput() {
    setOutput('');
    message.success('已清空');
  }

  function formatOutput() {
    if (!output) return message.warning('没有可格式化的内容');
    if (panel.formatMode === 'script') {
      const lines = formatScriptText(output);
      setOutput(lines.join('\n'));
      message.success(`格式化完成，共处理 ${lines.length} 行`);
      return;
    }
    const lines = output.split('\n').filter(line => line.trim());
    setOutput(lines.join('\n'));
    message.success(`格式化完成，共${lines.length}行`);
  }

  function saveTxt() {
    if (!output) {
      message.warning('没有可保存的内容');
      return;
    }
    downloadFile(`${panel.saveFileName}.txt`, output, 'text/plain;charset=utf-8');
    message.success('保存成功');
  }

  function saveJson() {
    if (!output) {
      message.warning('没有可保存的内容');
      return;
    }
    const items = buildDialogueList(output);
    if (!items.length) {
      message.warning('未找到有效的剧本内容，请确保剧本格式正确');
      return;
    }
    downloadFile(`${panel.saveFileName}.json`, JSON.stringify(items, null, 2), 'application/json;charset=utf-8');
    message.success(`保存成功，共${items.length}条台词`);
  }

  if (!tool) return null;

  const saveButton = (
    <Button
      type="text"
      size="small"
      icon={<Save size={15} />}
      disabled={!output}
      aria-label="保存"
      title="保存"
    />
  );

  return (
    <PluginPanelShell
      name={panel.name}
      heading={panel.heading}
      authorized={panel.authorized}
      icon={TitleIcon ? <TitleIcon size={22} color={panel.iconColor} /> : null}
      iconColor={panel.iconColor}
      primaryColor={panel.buttonColor}
      onClose={openPluginMarket}
      layout="stack"
      result={(
        <>
          <div className="plugin-panel-result-head">
            <strong>{panel.resultTitle}</strong>
            {panel.showWordCount && output ? <span className="plugin-panel-wordcount">字数：{output.length}</span> : null}
            <div className="plugin-panel-result-tools">
              <Button type="text" size="small" icon={<Copy size={15} />} disabled={!output} aria-label="复制" title="复制" onClick={copyOutput} />
              <Button type="text" size="small" icon={<Trash2 size={15} />} disabled={!output} aria-label="清空" title="清空" onClick={clearOutput} />
              <Button type="text" size="small" icon={<AlignLeft size={15} />} disabled={!output} aria-label={panel.formatLabel} title={panel.formatLabel} onClick={formatOutput} />
              {panel.saveJson ? (
                <Dropdown
                  disabled={!output}
                  trigger={['click']}
                  menu={{
                    items: [{ key: 'txt', label: '保存为 TXT' }, { key: 'json', label: '保存为 JSON' }],
                    onClick: ({ key }) => (key === 'json' ? saveJson() : saveTxt())
                  }}
                >
                  {saveButton}
                </Dropdown>
              ) : (
                <Button type="text" size="small" icon={<Save size={15} />} disabled={!output} aria-label="保存" title="保存" onClick={saveTxt} />
              )}
            </div>
          </div>
          <TextArea
            className="plugin-panel-textarea"
            rows={panel.resultRows}
            value={output}
            onChange={event => setOutput(event.target.value)}
            placeholder={RESULT_PLACEHOLDER}
            aria-label={panel.resultTitle}
          />
        </>
      )}
    >
      <Upload accept=".txt,.doc,.docx" showUploadList={false} beforeUpload={readUpload}>
        <Button className="plugin-panel-upload-inline" icon={<UploadIcon size={15} />}>上传文档</Button>
      </Upload>

      <TextArea
        className="plugin-panel-textarea"
        rows={panel.inputRows}
        style={{ height: panel.inputHeight }}
        value={input}
        onChange={event => setInput(event.target.value)}
        placeholder={panel.inputPlaceholder}
        aria-label="输入内容"
      />

      <div className="plugin-panel-row">
        <Select
          showSearch
          optionFilterProp="label"
          loading={loadingConfig}
          value={modelId}
          onChange={setModelId}
          options={modelSelectOptions(models, 'text')}
          placeholder="选择模型"
          aria-label="选择模型"
        />
        <Select
          showSearch
          optionFilterProp="label"
          loading={loadingConfig}
          value={promptId}
          onChange={setPromptId}
          options={prompts.map(item => ({ value: item.id, label: item.name }))}
          placeholder="选择提示词"
          aria-label="选择提示词"
        />
      </div>

      {panel.concurrent ? (
        <div className="plugin-panel-concurrent">
          <span><strong>并发生成</strong><em>关闭时按分段顺序逐段发送</em></span>
          <Switch size="small" checked={concurrent} disabled={running} onChange={setConcurrent} aria-label="并发生成" />
        </div>
      ) : null}

      <Button className="plugin-panel-submit" type="primary" block loading={running} onClick={generate}>
        {running ? (progress.total > 1 ? `AI生成中 (${progress.current}/${progress.total})` : 'AI生成中...') : 'AI生成'}
      </Button>
    </PluginPanelShell>
  );
}
