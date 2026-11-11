import { Button, Input, Select, Upload, message } from 'antd';
import { Upload as UploadIcon } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { apiRequest } from '../../shared/api/client.js';
import { listAvailableModels, modelSelectOptions } from '../../shared/api/modelCatalog.js';
import { getPluginToolConfig, runPluginTool } from '../../shared/api/pluginTools.js';
import PluginPanelShell from '../components/PluginPanelShell.jsx';
import { createGenerationGate, executeGeneration, validateGenerationInput } from './pluginGeneration.js';

const { TextArea } = Input;

const TOOL_ID = 'ai-rewrite';
// 文案与洛水 v2.9.0 一致：占位输入与结果空态都按原话照抄。
const INPUT_PLACEHOLDER = '请输入要改写的文本内容，或上传txt/word文档...';
const RESULT_PLACEHOLDER = '选择模型和改写类型，输入内容后点击AI生成';

// AI改文：洛水是页面级单栏面板，不是弹窗。这里是整套插件面板的形态样板。
export default function PluginRewritePage() {
  const [input, setInput] = useState('');
  const [output, setOutput] = useState('');
  const [models, setModels] = useState([]);
  const [prompts, setPrompts] = useState([]);
  const [modelId, setModelId] = useState();
  const [promptId, setPromptId] = useState();
  const [loadingConfig, setLoadingConfig] = useState(true);
  const [running, setRunning] = useState(false);
  // 关闭面板或切走路由时真正中止请求。
  const gateRef = useRef(null);
  const mountedRef = useRef(true);
  if (!gateRef.current) gateRef.current = createGenerationGate();

  useEffect(() => () => { mountedRef.current = false; gateRef.current.abort(); }, []);

  useEffect(() => {
    let alive = true;
    (async () => {
      try {
        const [availableModels, promptOptions] = await Promise.all([
          listAvailableModels('text'),
          getPluginToolConfig(TOOL_ID)
        ]);
        if (!alive) return;
        setModels(availableModels);
        setPrompts(promptOptions);
        setModelId(availableModels[0]?.id);
        setPromptId(promptOptions[0]?.id);
      } catch (error) {
        if (alive) message.error(error?.message || '加载模型和改写类型失败');
      } finally {
        if (alive) setLoadingConfig(false);
      }
    })();
    return () => { alive = false; };
  }, []);

  async function readUpload(file) {
    const lowerName = String(file?.name || '').toLowerCase();
    if (!lowerName.endsWith('.txt')) {
      message.warning('浏览器版当前先支持 TXT；Word 文档读取接入后再开放');
      return Upload.LIST_IGNORE;
    }
    try {
      setInput(await file.text());
      message.success('文件读取成功');
    } catch {
      message.error('文件读取失败');
    }
    return Upload.LIST_IGNORE;
  }

  async function generate() {
    if (running) return message.info('正在生成，请勿重复提交');
    const validation = validateGenerationInput({ input, modelId, promptId, emptyMessage: '请输入或上传内容', promptMessage: '请选择改写类型' });
    if (!validation.ok) return message.warning(validation.message);
    const gate = gateRef.current;
    const controller = gate.begin();
    if (!controller) return message.info('正在生成，请勿重复提交');
    const requestWithSignal = (path, options = {}) => apiRequest(path, { ...options, signal: controller.signal });
    setRunning(true);
    setOutput('');
    try {
      const outcome = await executeGeneration({
        parts: [input],
        signal: controller.signal,
        runPart: part => runPluginTool({ toolId: TOOL_ID, input: part, modelId, promptId }, requestWithSignal)
      });
      if (outcome.cancelled) {
        if (mountedRef.current) message.info('已取消');
        return;
      }
      if (!outcome.ok) {
        message.error(outcome.error?.message || '生成失败');
        return;
      }
      setOutput(outcome.text);
      message.success('生成成功');
    } finally {
      gate.end(controller);
      setRunning(false);
    }
  }

  return (
    <PluginPanelShell
      name="AI改文"
      layout="stack"
      result={(
        <>
          <div className="plugin-panel-result-head">
            <strong>改写结果</strong>
            <span>{output.length} 字</span>
          </div>
          <TextArea
            className="plugin-panel-textarea"
            value={output}
            onChange={event => setOutput(event.target.value)}
            placeholder={RESULT_PLACEHOLDER}
            aria-label="改写结果"
          />
        </>
      )}
    >
      <div className="plugin-panel-card">
        <div className="plugin-panel-textarea-wrap">
          <Upload accept=".txt,.doc,.docx" showUploadList={false} beforeUpload={readUpload}>
            <Button className="plugin-panel-upload" icon={<UploadIcon size={15} />}>上传文档</Button>
          </Upload>
          <TextArea
            className="plugin-panel-textarea"
            rows={10}
            value={input}
            onChange={event => setInput(event.target.value)}
            placeholder={INPUT_PLACEHOLDER}
            aria-label="要改写的文本内容"
          />
        </div>

        <div className="plugin-panel-row">
          <Select
            loading={loadingConfig}
            value={modelId}
            onChange={setModelId}
            options={modelSelectOptions(models, 'text')}
            placeholder="选择模型"
            aria-label="选择模型"
          />
          <Select
            loading={loadingConfig}
            value={promptId}
            onChange={setPromptId}
            options={prompts.map(item => ({ value: item.id, label: item.name }))}
            placeholder="选择改写类型"
            aria-label="选择改写类型"
          />
        </div>

        <Button className="plugin-panel-submit" type="primary" block loading={running} onClick={generate}>
          {running ? 'AI生成中...' : 'AI生成'}
        </Button>
      </div>
    </PluginPanelShell>
  );
}
