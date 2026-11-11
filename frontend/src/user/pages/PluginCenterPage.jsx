import { Button, Input, Modal, Tag } from 'antd';
import { MonitorCog, Search, Sparkles } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { themeDefinition } from '../../shared/styles/theme-catalog.js';
import AssetReferenceWorkbench from '../components/AssetReferenceWorkbench.jsx';
import AISegmentWorkbench from '../components/AISegmentWorkbench.jsx';
import CharacterTurnaroundWorkbench from '../components/CharacterTurnaroundWorkbench.jsx';
import CoverGenerationWorkbench from '../components/CoverGenerationWorkbench.jsx';
import DraftToolWorkbench from '../components/DraftToolWorkbench.jsx';
import LocalToolWorkbench from '../components/LocalToolWorkbench.jsx';
import UpscaleWorkbench from '../components/UpscaleWorkbench.jsx';
import { PLUGIN_CATALOG, filterPlugins, pluginIsUnavailable, pluginLaunchTarget } from './pluginCatalog.js';
import './plugin-center.css';

// 洛水插件市场只有浅色 / 深色两档，这里按当前解析后的主题取同档配色。
function currentTone() {
  const theme = document.documentElement.dataset.theme || 'dark';
  return themeDefinition(theme).colorScheme === 'light' ? 'light' : 'dark';
}

export default function PluginCenterPage() {
  const [query, setQuery] = useState('');
  const [tone, setTone] = useState(currentTone);
  const [desktopPlugin, setDesktopPlugin] = useState(null);
  const [unavailablePlugin, setUnavailablePlugin] = useState(null);
  const [activeWorkspace, setActiveWorkspace] = useState('');
  const [draftTool, setDraftTool] = useState('');
  const [localTool, setLocalTool] = useState('');
  const [upscaleOpen, setUpscaleOpen] = useState(false);
  const [coverOpen, setCoverOpen] = useState(false);
  const [turnaroundOpen, setTurnaroundOpen] = useState(false);

  useEffect(() => {
    const root = document.documentElement;
    const observer = new MutationObserver(() => setTone(currentTone()));
    observer.observe(root, { attributes: true, attributeFilter: ['data-theme'] });
    setTone(currentTone());
    return () => observer.disconnect();
  }, []);

  const visiblePlugins = useMemo(() => filterPlugins(PLUGIN_CATALOG, { query }), [query]);

  async function openPlugin(plugin) {
    const target = pluginLaunchTarget(PLUGIN_CATALOG, plugin.id);
    if (target?.kind === 'workspace') {
      setActiveWorkspace(target.workspace);
      return;
    }
    if (target?.kind === 'route' || target?.kind === 'prompt-tool') {
      window.history.pushState({}, '', target.href);
      window.dispatchEvent(new PopStateEvent('popstate'));
      return;
    }
    if (target?.kind === 'draft-tool') {
      setDraftTool(target.tool);
      return;
    }
    if (target?.kind === 'local-tool') {
      setLocalTool(target.tool);
      return;
    }
    if (target?.kind === 'upscale-tool') {
      setUpscaleOpen(true);
      return;
    }
    if (target?.kind === 'cover-tool') {
      setCoverOpen(true);
      return;
    }
    if (target?.kind === 'turnaround-tool') {
      setTurnaroundOpen(true);
      return;
    }
    if (target?.kind === 'unavailable') {
      setUnavailablePlugin(plugin);
      return;
    }
    if (target?.kind === 'desktop-tool') {
      setDesktopPlugin(plugin);
      return;
    }
  }

  if (activeWorkspace === 'asset-reference') {
    return <AssetReferenceWorkbench onBack={() => setActiveWorkspace('')} />;
  }

  // AI-分段在洛水里是独立整页（左参数列 + 右处理日志），不能叠在插件列表下方。
  if (activeWorkspace === 'ai-segment') {
    return (
      <main className="plugin-market-page" data-tone={tone}>
        <AISegmentWorkbench onBack={() => setActiveWorkspace('')} />
      </main>
    );
  }

  if (draftTool) {
    return (
      <main className="plugin-market-page" data-tone={tone}>
        <DraftToolWorkbench tool={draftTool} onBack={() => setDraftTool('')} />
      </main>
    );
  }

  if (localTool) {
    return (
      <main className="plugin-market-page" data-tone={tone}>
        <LocalToolWorkbench tool={localTool} onBack={() => setLocalTool('')} />
      </main>
    );
  }

  if (upscaleOpen) {
    return (
      <main className="plugin-market-page" data-tone={tone}>
        <UpscaleWorkbench onBack={() => setUpscaleOpen(false)} />
      </main>
    );
  }

  if (coverOpen) {
    return <CoverGenerationWorkbench onBack={() => setCoverOpen(false)} />;
  }

  if (turnaroundOpen) {
    return <CharacterTurnaroundWorkbench onBack={() => setTurnaroundOpen(false)} />;
  }

  return (
    <main className="plugin-market-page" data-tone={tone}>
      <div className="plugin-market-inner">
        <header className="plugin-market-header">
          <h1>插件市场</h1>
          <p>丰富的AI辅助工具，提升创作效率</p>
          <Input
            className="plugin-market-search"
            allowClear
            prefix={<Search size={16} aria-hidden="true" />}
            value={query}
            onChange={event => setQuery(event.target.value)}
            placeholder="搜索插件名称或功能"
            aria-label="搜索插件名称或功能"
          />
        </header>

        {visiblePlugins.length ? (
          <section className="plugin-market-list" aria-label="插件列表">
            {visiblePlugins.map(plugin => {
              const unavailable = pluginIsUnavailable(plugin);
              return (
                <button key={plugin.id} type="button" className="plugin-market-card" onClick={() => openPlugin(plugin)}>
                  <span className="plugin-market-icon" style={{ backgroundColor: `${plugin.color}22`, color: plugin.color }}>
                    <Sparkles size={20} strokeWidth={1.8} aria-hidden="true" />
                  </span>
                  <span className="plugin-market-copy">
                    <span className="plugin-market-name">
                      <strong>{plugin.name}</strong>
                      {unavailable ? <em className="plugin-market-state">未接入</em> : null}
                    </span>
                    <small>{plugin.summary}</small>
                  </span>
                  {plugin.isNew ? <span className="plugin-market-new">NEW</span> : null}
                </button>
              );
            })}
          </section>
        ) : <div className="plugin-market-empty">没有找到「{query.trim()}」相关插件</div>}
      </div>

      <Modal title={unavailablePlugin?.name || '插件'} open={Boolean(unavailablePlugin)} onCancel={() => setUnavailablePlugin(null)} footer={<Button onClick={() => setUnavailablePlugin(null)}>知道了</Button>}>
        <div className="plugin-placeholder-panel"><span className="plugin-market-icon" style={{ color: unavailablePlugin?.color }}><Sparkles size={28} /></span><h3>{unavailablePlugin?.name}</h3><p>{unavailablePlugin?.summary}</p><div>当前未接入：{unavailablePlugin?.launch?.reason || 'V88 还没有对应的执行链路'}。</div></div>
      </Modal>

      <Modal title={desktopPlugin?.name || '本地工具'} open={Boolean(desktopPlugin)} onCancel={() => setDesktopPlugin(null)} footer={<Button onClick={() => setDesktopPlugin(null)}>知道了</Button>}>
        <div className="plugin-environment-panel"><MonitorCog size={34} aria-hidden="true" /><h3>该功能依赖本地桌面执行器</h3><p>{desktopPlugin?.summary}</p><Tag color="blue">{desktopPlugin?.launch?.environment}</Tag><p>CM 通过 Python 桌面桥接读取本机文件、剪映草稿或媒体；网页不能只靠提示词完成。V88 当前未接入这条链路。</p></div>
      </Modal>
    </main>
  );
}
