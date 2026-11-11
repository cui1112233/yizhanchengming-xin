import { Button, Form, Input, List, Popconfirm, Segmented, Select, Slider, Switch, Typography, message } from 'antd';
import { Download, FolderOpen, Image, RefreshCw, Save, Video } from 'lucide-react';
import { useEffect, useState } from 'react';
import { getConfig, saveConfig } from '../../shared/api/config';
import { getCurrentUsername } from '../../shared/api/auth';
import { apiRequest } from '../../shared/api/client';
import { createLocalExecutorPairing as createDoubaoExecutorPairing, getExecutorStatus, getLocalExecutorReleaseManifest, mergeExecutorState } from '../../shared/api/localExecutors.js';
import { useLocalExecutorStatus } from '../../shared/hooks/useLocalExecutorStatus.js';
import {
  createGiantMaterialPairing,
  getGiantExecutorPreference,
  saveGiantExecutorPreference,
  deleteGiantMaterialExecutor
} from '../../shared/api/giantMaterialExecutorPublic';
import { PET_COMPANION_SETTINGS_EVENT, readCompanionSpeechState, writeCompanionSpeechState } from '../../shared/pet/companionSpeech';
import { DEFAULT_PET_ID, dispatchPetSelection, getPetDefinition, getPetOptions, previewPetSelection } from '../../shared/pet/petCatalog';
import { LittleCompanionSprite } from '../../shared/pet/LittleCompanionSprite';
import { ThemePicker } from '../../shared/components/ThemePicker';
import { WindowBackgroundPicker } from '../../shared/components/WindowBackgroundPicker';

const petOptions = getPetOptions();

// 两种执行器共用一套状态说法：未安装 / 需配对 / 离线 / 在线。state 与待配对
// 信息都来自统一状态接口；本地不再用刚生成的配对码叠加状态。
function executorStatusText(group) {
  const devices = Array.isArray(group?.devices) ? group.devices : [];
  const paired = devices.length;
  const online = paired ? devices.filter(item => item?.online).length : 0;
  const effective = mergeExecutorState(group);
  if (effective === 'not_installed') return '未安装：下载客户端后生成配对码';
  if (effective === 'needs_pairing') return '需配对：在客户端输入配对码完成绑定';
  if (effective === 'online') return `在线：${online} 台在线 / 共 ${paired} 台已配对`;
  if (effective === 'offline') return `离线：${paired} 台已配对，请打开客户端`;
  return '状态读取中…';
}

export function SettingsPage({ themePreference = 'system', onThemeChange, windowBackgroundId = 'none', windowBackgroundUrl = '', onWindowBackgroundChange, onWindowBackgroundUpload }) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [savingSection, setSavingSection] = useState(null);
  const [restoring, setRestoring] = useState(false);
  const [listing, setListing] = useState(false);
  const [restoreReport, setRestoreReport] = useState(null);
  const [fileList, setFileList] = useState(null);
  const [giantMaterialExecutors, setGiantMaterialExecutors] = useState([]);
  const [loadingGiantMaterialExecutors, setLoadingGiantMaterialExecutors] = useState(false);
  const [giantLatestVersions, setGiantLatestVersions] = useState({ windows: null, macos: null });
  const [giantPairing, setGiantPairing] = useState(null);
  const [giantPairingBusy, setGiantPairingBusy] = useState(false);
  const [giantPreference, setGiantPreference] = useState({ preferredOs: 'windows' });
  const [pairing, setPairing] = useState(null);
  const [executorOverview, setExecutorOverview] = useState(null);
  const [localExecutorRelease, setLocalExecutorRelease] = useState({ windows: '', macos: '', windowsUrl: '', macosUrl: '' });
  const [executorTestBusy, setExecutorTestBusy] = useState('');
  const [executorTestResult, setExecutorTestResult] = useState(null);
  const [companionActive, setCompanionActive] = useState(() => readCompanionSpeechState(getCurrentUsername()).active);
  const username = getCurrentUsername();
  const { executors: localExecutors, loading: loadingExecutors, refresh: refreshLocalExecutors } = useLocalExecutorStatus({ enabled: true, intervalMs: 30000 });
  const soundEnabled = Form.useWatch('soundEnabled', form);
  const soundVolume = Form.useWatch('soundVolume', form);
  const soundVolumePercent = Number.isFinite(soundVolume) ? Math.round(soundVolume) : 60;

  useEffect(() => {
    let alive = true;
    getLocalExecutorReleaseManifest().then(value => {
      if (alive) setLocalExecutorRelease(value);
    }).catch(() => {});
    return () => { alive = false; };
  }, []);

  useEffect(() => {
    setCompanionActive(readCompanionSpeechState(username).active);
  }, [username]);

  async function loadLocalExecutors() {
    try {
      await refreshLocalExecutors();
    } catch (error) {
      message.error(error.message || '读取本地执行器失败');
    }
  }

  async function loadGiantMaterialExecutors() {
    setLoadingGiantMaterialExecutors(true);
    try {
      const result = await apiRequest('/api/shuihuo-production/giant-material-executors', { suppressGlobalError: true });
      setGiantMaterialExecutors(Array.isArray(result.executors) ? result.executors : []);
      try {
        const prefResp = await getGiantExecutorPreference();
        setGiantPreference({ preferredOs: prefResp?.preferredOs || 'windows' });
      } catch (_) { /* 偏好接口可能尚未就绪，下一周期重试 */ }
    } catch (error) {
      message.error(error.message || '读取巨量素材执行器失败');
    } finally { setLoadingGiantMaterialExecutors(false); }
  }

  // 生成配对码后立即刷新统一状态接口，让"需配对"在 30 秒轮询之前就生效。
  async function refreshExecutorOverview() {
    try { setExecutorOverview(await getExecutorStatus()); } catch (_) { /* 保留上一份快照 */ }
  }

  useEffect(() => {
    let alive = true;
    const load = () => getExecutorStatus().then(value => { if (alive) setExecutorOverview(value); }).catch(() => {});
    load();
    const timer = window.setInterval(load, 30000);
    return () => { alive = false; window.clearInterval(timer); };
  }, []);

  useEffect(() => { loadGiantMaterialExecutors(); }, []);

  useEffect(() => {
    let alive = true;
    Promise.all(['windows', 'macos'].map(async platform => {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 15000);
      let response;
      try {
        response = await fetch(`/downloads/giant-material-executor/releases/${platform}/latest.json`, { signal: controller.signal });
      } finally {
        clearTimeout(timeoutId);
      }
      const data = response.ok ? await response.json() : null;
      return [platform, data?.version ? String(data.version) : null];
    }))
      .then(entries => {
        if (alive) setGiantLatestVersions(Object.fromEntries(entries));
      })
      .catch(() => {});
    return () => { alive = false; };
  }, []);

  async function createGiantMaterialExecutorPairing() {
    setGiantPairingBusy(true);
    try {
      const result = await createGiantMaterialPairing();
      setGiantPairing(result);
      await refreshExecutorOverview();
      message.success('巨量素材配对码已生成，请在 Windows 执行器首次启动时输入');
    } catch (error) {
      message.error(error.message || '生成巨量素材配对码失败');
    } finally { setGiantPairingBusy(false); }
  }

  async function createLocalExecutorPairing() {
    try {
      const result = await createDoubaoExecutorPairing();
      setPairing(result);
      await refreshExecutorOverview();
      await refreshLocalExecutors();
      message.success('配对码已生成，请在本地执行器中输入');
    } catch (error) { message.error(error.message || '生成配对码失败'); }
  }

  useEffect(() => () => {
    if (executorTestResult?.previewUrl) URL.revokeObjectURL(executorTestResult.previewUrl);
  }, [executorTestResult?.previewUrl]);

  async function runLocalExecutorTest(type) {
    setExecutorTestBusy(type);
    if (executorTestResult?.previewUrl) URL.revokeObjectURL(executorTestResult.previewUrl);
    setExecutorTestResult({ type, status: 'queued' });
    try {
      const created = await apiRequest('/api/shuihuo-production/local-executor-tests', {
        method: 'POST',
        body: JSON.stringify({
          type,
          prompt: type === 'image' ? '一只蓝色小鸟站在樱花枝上' : '一只蓝色小鸟站在樱花枝上，轻轻扇动翅膀，镜头稳定',
          ratio: type === 'image' ? '1:1' : '16:9',
          duration: type === 'video' ? 5 : 0
        }),
        suppressGlobalError: true
      });
      let current = created;
      for (let attempt = 0; attempt < 450; attempt += 1) {
        current = await apiRequest(`/api/shuihuo-production/local-executor-tests/${created.jobId}`, { suppressGlobalError: true });
        setExecutorTestResult({ type, ...current });
        if (current.status === 'succeeded') {
          const blob = await apiRequest(current.resultUrl, { responseType: 'blob', suppressGlobalError: true });
          const previewUrl = URL.createObjectURL(blob);
          setExecutorTestResult({ type, ...current, previewUrl });
          message.success(`${type === 'image' ? '图片' : '视频'}测试成功`);
          return;
        }
        if (current.status === 'failed' || current.status === 'cancelled') throw new Error(current.error || '本地执行器测试失败');
        await new Promise(resolve => setTimeout(resolve, 2000));
      }
      throw new Error('本地执行器测试等待超时');
    } catch (error) {
      setExecutorTestResult(current => ({ ...current, type, status: 'failed', error: error.message || '测试失败' }));
      message.error(error.message || '本地执行器测试失败');
    } finally {
      setExecutorTestBusy('');
      loadLocalExecutors();
    }
  }

  const changeGiantPreference = async (value) => {
    const previous = giantPreference?.preferredOs || 'windows';
    if (value === previous) return;
    setGiantPreference({ preferredOs: value });
    try {
      const saved = await saveGiantExecutorPreference(value);
      setGiantPreference({ preferredOs: saved?.preferredOs || value });
    } catch (error) {
      setGiantPreference({ preferredOs: previous });
      message.error('偏好保存失败，请重试');
    }
  };

  const removeGiantExecutor = async (executorId) => {
    await deleteGiantMaterialExecutor(executorId);
    message.success('设备已删除，未完成任务已重新排队');
    await loadGiantMaterialExecutors();
  };

  useEffect(() => {
    let alive = true;
    setLoading(true);
    getConfig()
      .then(config => {
        if (!alive) return;
        form.setFieldsValue({
          storageRoot: config.storageRoot || '',
          productionRetentionDays: [3, 7, 14, 30].includes(Number(config.productionRetentionDays)) ? Number(config.productionRetentionDays) : 7,
          petId: getPetDefinition(config.pet).id,
          soundEnabled: config.notifications?.soundEnabled !== false,
          soundVolume: Number.isFinite(config.notifications?.soundVolume) ? config.notifications.soundVolume : 60,
          petVisible: config.notifications?.petVisible !== false
        });
      })
      .catch(error => message.error(error.message || '读取设置失败'))
      .finally(() => {
        if (alive) setLoading(false);
      });
    return () => { alive = false; };
  }, [form]);

  async function saveSection(section) {
    let values;
    try {
      if (section === 'workspace') {
        const result = await form.validateFields(['petId', 'soundEnabled', 'soundVolume', 'petVisible']);
        values = {
          pet: result.petId,
          notifications: {
            soundEnabled: result.soundEnabled !== false,
            soundVolume: Number.isFinite(result.soundVolume) ? result.soundVolume : 60,
            petVisible: result.petVisible !== false
          }
        };
      } else if (section === 'storage') {
        const result = await form.validateFields(['storageRoot', 'productionRetentionDays']);
        values = {
          storageRoot: result.storageRoot || '',
          productionRetentionDays: [3, 7, 14, 30].includes(Number(result.productionRetentionDays)) ? Number(result.productionRetentionDays) : 7
        };
      }
    } catch (error) {
      if (!error?.errorFields) message.error(error.message || '保存失败');
      return;
    }

    setSavingSection(section);
    try {
      const saved = await saveConfig(values);
      if (section === 'workspace') {
        window.dispatchEvent(new CustomEvent('qiantie:notifications-updated', { detail: saved.notifications }));
        dispatchPetSelection(saved.pet || values.pet);
      }
      message.success(`${({ workspace: '工作台设置', storage: '本地存储' })[section]}已保存`);
    } catch (error) {
      message.error(error.message || '保存失败');
    } finally {
      setSavingSection(null);
    }
  }

  async function handleRestore() {
    setRestoring(true);
    setRestoreReport(null);
    try {
      const report = await apiRequest('/api/storage/restore', { method: 'POST' });
      setRestoreReport(report);
      message.success(`恢复完成：剧本并入 ${report.scriptResults?.added ?? 0} 条`);
    } catch (error) {
      message.error(error.message || '恢复失败');
    } finally {
      setRestoring(false);
    }
  }

  async function handleList() {
    setListing(true);
    setFileList(null);
    try {
      const list = await apiRequest('/api/storage/list');
      setFileList(list);
    } catch (error) {
      message.error(error.message || '读取文件清单失败');
    } finally {
      setListing(false);
    }
  }

  const giantPreferenceHint = (() => {
    const selected = giantPreference?.preferredOs || 'windows';
    const devices = giantMaterialExecutors.filter(item => String(item.os || '').toLowerCase() === selected);
    const label = selected === 'darwin' ? 'macOS' : 'Windows';
    const fallback = selected === 'darwin' ? 'Windows' : 'macOS';
    if (!devices.length) return `当前没有绑定的 ${label} 设备，任务仍会交给 ${fallback}`;
    if (devices.some(item => item.recentFailureAt)) return `${label} 刚读取失败（冷却中），任务暂交给 ${fallback}`;
    if (!devices.some(item => item.online)) return `${label} 当前离线，任务仍会交给 ${fallback}`;
    return '';
  })();

  return (
    <div className="utility-page settings-page">
      <div className="settings-workbench">
        <div className="settings-heading">
          <div>
            <Typography.Title level={3}>工作台设置</Typography.Title>
            <Typography.Paragraph>管理前贴桌面宠物、工作提醒、本地视频执行器与服务器归档。模型密钥在个人中心「API 配置」统一配置。</Typography.Paragraph>
          </div>
          <span className="settings-status">本账号配置</span>
        </div>
      <Form
        className="settings-form"
        form={form}
        layout="vertical"
        disabled={loading}
        initialValues={{ storageRoot: '', productionRetentionDays: 7, petId: DEFAULT_PET_ID, soundEnabled: true, soundVolume: 60, petVisible: true }}
      >
        <section className="settings-section settings-theme-section" aria-labelledby="settings-theme-title">
          <div>
            <h2 id="settings-theme-title">外观主题</h2>
            <p>选择后立即切换。登录账号会自动同步到其他设备。</p>
          </div>
          <ThemePicker value={themePreference} onChange={onThemeChange} variant="grid" />
          <WindowBackgroundPicker value={windowBackgroundId} customUrl={windowBackgroundUrl} onChange={onWindowBackgroundChange} onUpload={onWindowBackgroundUpload} />
        </section>
        <div className="settings-top-grid">
        <section className="settings-section settings-model-section" aria-labelledby="settings-model-title">
          <div>
            <h2 id="settings-model-title">工作台与桌面宠物</h2>
            <p>管理桌面宠物和工作完成提醒。</p>
          </div>
          <Form.Item label="前贴宠物" name="petId">
            <Select options={petOptions} onChange={value => previewPetSelection(value)} />
          </Form.Item>
          <Form.Item label="提示音" name="soundEnabled" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Typography.Paragraph type="secondary">用于在剧本人物/场景提取完成、剧本创作完成时提醒；提取或生成失败（如网络、404、鉴权错误）时播放警示音。</Typography.Paragraph>
          <Form.Item label="提示音音量">
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <Form.Item name="soundVolume" noStyle>
                <Slider min={0} max={100} step={1} value={soundVolume} disabled={!soundEnabled} tooltip={{ formatter: value => `${value}%` }} style={{ flex: 1 }} />
              </Form.Item>
              <Typography.Text>{soundVolumePercent}%</Typography.Text>
            </div>
          </Form.Item>
          <Form.Item label="显示桌面宠物" name="petVisible" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Typography.Paragraph type="secondary">控制右下角桌面宠物是否显示，关闭后可减少界面干扰。</Typography.Paragraph>
          <PetPreview form={form} />
          <Form.Item label="宠物主动说话" valuePropName="checked">
            <Switch checked={companionActive} onChange={checked => {
              setCompanionActive(checked);
              writeCompanionSpeechState(username, { ...readCompanionSpeechState(username), active: checked, nextIdleAt: 0 });
              window.dispatchEvent(new CustomEvent(PET_COMPANION_SETTINGS_EVENT, { detail: { active: checked, username } }));
            }} />
          </Form.Item>
          <div className="settings-section-footer settings-workspace-actions">
            <Button type="primary" icon={<Save size={16} strokeWidth={1.8} aria-hidden="true" />} onClick={() => saveSection('workspace')} loading={savingSection === 'workspace'}>保存工作台与宠物</Button>
          </div>
        </section>

        <section className="settings-section settings-storage-section" aria-labelledby="settings-storage-title">
          <div>
            <h2 id="settings-storage-title">服务器归档文件夹</h2>
            <p>指定服务器保存剧本、小说与制作工程的归档目录；不会改变网页用户的浏览器下载位置。</p>
          </div>
          <Form.Item label="服务器归档路径" name="storageRoot" extra="必须是服务器可访问的绝对路径；留空表示关闭服务器归档。">
            <Input placeholder="例如 /data/qiantie-archive" />
          </Form.Item>
          <Form.Item label="制作文件保留时长" name="productionRetentionDays" extra="制作时先保存在服务器；上传 TOS 并校验成功后，本地副本按所选时长清理。未上传或校验失败的文件绝不删除。">
            <Select options={[3, 7, 14, 30].map(days => ({ value: days, label: `${days} 天` }))} />
          </Form.Item>
          <Typography.Paragraph type="secondary">默认保留 7 天；磁盘达到 75% 时会优先清理已在 TOS 校验成功的最旧本地副本，降到 65% 后停止。头像、账号、项目配置和 TOS 正式文件不受影响。</Typography.Paragraph>
          <div className="settings-section-footer settings-storage-actions">
            <Button type="primary" icon={<Save size={16} strokeWidth={1.8} aria-hidden="true" />} onClick={() => saveSection('storage')} loading={savingSection === 'storage'}>保存存储设置</Button>
            <Button icon={<RefreshCw size={16} strokeWidth={1.8} aria-hidden="true" />} onClick={handleRestore} loading={restoring}>恢复</Button>
            <Button icon={<FolderOpen size={16} strokeWidth={1.8} aria-hidden="true" />} onClick={handleList} loading={listing}>查看文件清单</Button>
          </div>
          {restoreReport && (
            <div style={{ marginTop: 16 }}>
              <Typography.Text strong>恢复报告</Typography.Text>
              <ul style={{ margin: '8px 0 0', paddingLeft: 20, color: 'var(--legacy-muted)', fontSize: 13, lineHeight: 1.7 }}>
                <li>私人定制：找到 {restoreReport.scriptResults?.found ?? 0} 个，新并入历史 {restoreReport.scriptResults?.added ?? 0} 条</li>
                <li>小说获取：{restoreReport.novelFetch?.found ?? 0} 个文件</li>
                <li>改编小说：{restoreReport.novelAdapt?.found ?? 0} 个文件</li>
                <li>制作工程：{restoreReport.projects?.found ?? 0} 个项目</li>
              </ul>
              {Array.isArray(restoreReport.errors) && restoreReport.errors.length > 0 && (
                <Typography.Paragraph type="danger" style={{ margin: '8px 0 0', fontSize: 13 }}>错误：{restoreReport.errors.join('；')}</Typography.Paragraph>
              )}
            </div>
          )}
          {fileList && (
            <div style={{ marginTop: 16 }}>
              <Typography.Text strong>文件清单</Typography.Text>
              <List
                size="small"
                style={{ marginTop: 8 }}
                dataSource={[
                  { key: 'scriptResults', title: '私人定制', files: fileList.scriptResults || [] },
                  { key: 'novelFetch', title: '小说获取', files: fileList.novelFetch || [] },
                  { key: 'novelAdapt', title: '改编小说', files: fileList.novelAdapt || [] },
                  { key: 'projects', title: '制作工程', projects: fileList.projects || [] }
                ]}
                renderItem={group => (
                  <List.Item>
                    <div style={{ width: '100%' }}>
                      <Typography.Text strong>{group.title}（{group.projects ? group.projects.length : group.files.length}）</Typography.Text>
                      <List
                        size="small"
                        dataSource={group.projects || group.files}
                        locale={{ emptyText: '（空）' }}
                        renderItem={item => (
                          <List.Item style={{ padding: '2px 0', borderBottom: 'none' }}>
                            {group.projects
                              ? `${item.name}/（${(item.files || []).length} 个文件）`
                              : `${item.name}（${item.size}B）`}
                          </List.Item>
                        )}
                      />
                    </div>
                  </List.Item>
                )}
              />
            </div>
          )}
        </section>
        </div>

        <section className="settings-section settings-executor-section" aria-labelledby="settings-executor-overview-title">
          <div>
            <h2 id="settings-executor-overview-title">执行器状态</h2>
            <p>两种执行器都运行在自己的电脑上：豆包负责图片/视频，巨量素材负责 OCR 读取正文。状态口径统一为：未安装、需配对、离线、在线。</p>
          </div>
          <div className="settings-executor-layout">
            <div className="settings-executor-status-grid">
              <div className="settings-executor-status">
                <div className="settings-executor-icon"><Video size={20} /></div>
                <div>
                  <strong>豆包（图片/视频）</strong>
                  <span>{executorStatusText({ devices: executorOverview?.local?.devices ?? localExecutors, state: executorOverview?.local?.state, pairing: executorOverview?.local?.pairing })}</span>
                </div>
              </div>
              <div className="settings-executor-status">
                <div className="settings-executor-icon"><Video size={20} /></div>
                <div>
                  <strong>巨量素材（OCR）</strong>
                  <span>{executorStatusText({ devices: executorOverview?.giant?.devices ?? giantMaterialExecutors, state: executorOverview?.giant?.state, pairing: executorOverview?.giant?.pairing })}</span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="settings-section settings-executor-section" aria-labelledby="settings-executor-title">
          <div>
            <h2 id="settings-executor-title">一战晟铭执行器</h2>
            <p>在自己的电脑上使用豆包生成图片和视频。下载安装后，用一次性配对码连接，之后会自动接收网页任务。</p>
          </div>
          <div className="settings-executor-layout">
            {pairing ? <p className="settings-executor-pairing">请在一战晟铭执行器中输入配对码：<strong>{pairing.code}</strong>（10 分钟内有效，只能使用一次）</p> : null}
            <div className="settings-executor-actions settings-executor-actions-doubao">
              <div className="settings-executor-action-row settings-executor-action-row-top">
                <Button icon={<RefreshCw size={16} strokeWidth={1.8} aria-hidden="true" />} onClick={loadLocalExecutors} loading={loadingExecutors}>刷新状态</Button>
                <Button type="primary" onClick={createLocalExecutorPairing}>生成配对码</Button>
              </div>
              <div className="settings-executor-action-row settings-executor-action-row-downloads">
                <Button icon={<Download size={16} strokeWidth={1.8} aria-hidden="true" />} href={localExecutorRelease.macosUrl || '/downloads/local-executor/doubao-relay_1.0.0_universal.dmg'}>下载 macOS 客户端{localExecutorRelease.macos ? ` ${localExecutorRelease.macos}` : ''}</Button>
                <Button type="primary" icon={<Download size={16} strokeWidth={1.8} aria-hidden="true" />} href={localExecutorRelease.windowsUrl || '/downloads/local-executor/一战晟铭执行器-安装版.exe'}>下载 Windows 执行器{localExecutorRelease.windows ? ` ${localExecutorRelease.windows}` : ''}</Button>
              </div>
            </div>
            {localExecutors.length > 0 ? (
              <ul className="settings-local-executor-devices">
                {localExecutors.map(item => (
                  <li key={item.id}>
                    <span className={`settings-device-dot ${item.online ? 'is-online' : ''}`} aria-hidden="true" />
                    <strong>{item.name || '未命名设备'}</strong>
                    <span>{item.os || '未知系统'} · 版本 {item.version || '未知'}</span>
                    <span className="settings-device-seen">{item.online ? '在线' : formatLastSeen(item.lastSeenAt)}</span>
                  </li>
                ))}
              </ul>
            ) : null}
            <div className="settings-executor-test">
              <div className="settings-executor-test-actions">
                <Button icon={<Image size={16} />} loading={executorTestBusy === 'image'} disabled={Boolean(executorTestBusy) || !localExecutors.some(item => item.online)} onClick={() => runLocalExecutorTest('image')}>测试图片生成</Button>
                <Button icon={<Video size={16} />} loading={executorTestBusy === 'video'} disabled={Boolean(executorTestBusy) || !localExecutors.some(item => item.online)} onClick={() => runLocalExecutorTest('video')}>测试视频生成</Button>
              </div>
              {executorTestResult ? <div className={`settings-executor-test-result is-${executorTestResult.status || 'queued'}`}>
                <strong>{executorTestResult.status === 'succeeded' ? '测试成功' : executorTestResult.status === 'failed' ? '测试失败' : '正在通过本地执行器生成'}</strong>
                {executorTestResult.error ? <span>{executorTestResult.error}</span> : null}
                {executorTestResult.previewUrl && executorTestResult.mediaType?.startsWith('image/') ? <img src={executorTestResult.previewUrl} alt="本地执行器图片测试结果" /> : null}
                {executorTestResult.previewUrl && executorTestResult.mediaType === 'video/mp4' ? <video src={executorTestResult.previewUrl} controls playsInline /> : null}
              </div> : null}
            </div>
          </div>
        </section>

        <section className="settings-section settings-executor-section" aria-labelledby="settings-giant-material-executor-title">
          <div>
            <h2 id="settings-giant-material-executor-title">巨量素材执行器</h2>
            <p>Windows 与 macOS 本地 OCR 执行器。首次绑定一次，之后会在后台自动连接，不需要重复配对。</p>
          </div>
          <div className="settings-executor-layout">
            <div className="settings-giant-preference">
              <span className="settings-giant-preference-label">优先读取平台</span>
              <Segmented size="small" value={giantPreference?.preferredOs || 'windows'}
                options={[{ label: 'Windows', value: 'windows' }, { label: 'macOS', value: 'darwin' }]}
                onChange={changeGiantPreference} />
              {giantPreferenceHint ? <span className="settings-giant-preference-hint">{giantPreferenceHint}</span> : null}
            </div>
            {giantMaterialExecutors.length > 0 && (
              <ul style={{ listStyle: 'none', margin: '4px 0 0', padding: 0, display: 'flex', flexDirection: 'column', gap: 6, width: '100%' }}>
                {giantMaterialExecutors.map(item => {
                  const platform = item.os === 'darwin' ? 'macos' : item.os === 'windows' ? 'windows' : null;
                  const latestVersion = platform ? giantLatestVersions[platform] : null;
                  const outdated = Boolean(latestVersion) && Boolean(item.version) && item.version !== latestVersion;
                  return (
                    <li key={item.id} style={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: 8, fontSize: 13 }}>
                      <span aria-hidden="true" style={{ width: 8, height: 8, borderRadius: '50%', flexShrink: 0, background: item.online ? '#52c41a' : '#5b6b79' }} />
                      <strong>{item.name || '未命名设备'}</strong>
                      <span style={{ color: 'var(--legacy-muted, #9bb1c0)' }}>{item.os === 'darwin' ? 'macOS' : item.os === 'windows' ? 'Windows' : item.os || '未知系统'}</span>
                      <span style={{ color: 'var(--legacy-muted, #9bb1c0)' }}>版本 {item.version || '未知（旧版）'}</span>
                      {outdated && <span style={{ color: '#d48806' }}>可更新至 {latestVersion}</span>}
                      <span style={{ marginLeft: 'auto', color: 'var(--legacy-muted, #9bb1c0)' }}>{item.online ? '在线' : formatLastSeen(item.lastSeenAt)}</span>
                      <Popconfirm title="确定删除这台设备吗？未完成的任务会自动重新排队。"
                        okText="删除" cancelText="取消" onConfirm={() => removeGiantExecutor(item.id)}>
                        <Button size="small" type="text" danger>删除</Button>
                      </Popconfirm>
                    </li>
                  );
                })}
              </ul>
            )}
            <div className="settings-executor-actions settings-executor-actions-giant">
              <div className="settings-executor-action-row settings-executor-action-row-top">
                <Button icon={<RefreshCw size={16} strokeWidth={1.8} aria-hidden="true" />} onClick={loadGiantMaterialExecutors} loading={loadingGiantMaterialExecutors}>刷新状态</Button>
                <Button onClick={createGiantMaterialExecutorPairing} loading={giantPairingBusy}>生成配对码</Button>
              </div>
              <div className="settings-executor-action-row settings-executor-action-row-downloads">
                <Button type="primary" icon={<Download size={16} strokeWidth={1.8} aria-hidden="true" />} href="/downloads/giant-material-executor/GiantMaterialExecutor-windows-x64.zip">下载 Windows 执行器{giantLatestVersions.windows ? ` ${giantLatestVersions.windows}` : ''}</Button>
                <Button icon={<Download size={16} strokeWidth={1.8} aria-hidden="true" />} href="/downloads/giant-material-executor/GiantMaterialExecutor-macos-universal.zip">下载 macOS 执行器{giantLatestVersions.macos ? ` ${giantLatestVersions.macos}` : ''}</Button>
              </div>
            </div>
            {giantPairing?.code ? <p className="settings-executor-pairing">配对码：<strong>{giantPairing.code}</strong>（10 分钟内有效）。请在执行器首次启动窗口或本机配对页中输入；绑定成功后这里会显示在线状态。</p> : null}
            <p className="settings-executor-pairing">Windows 解压后双击 GiantMaterialExecutor.exe；macOS 解压后打开 GiantMaterialExecutor.app。首次绑定一次后会在后台自动连接；有新版本时，请在本机执行器中点击“立即更新”，不会替其他电脑远程更新。</p>
          </div>
        </section>

      </Form>
      </div>
    </div>
  );
}

function formatLastSeen(value) {
  if (!value) return '从未在线';
  const time = new Date(value).getTime();
  if (!Number.isFinite(time)) return '从未在线';
  const diff = Date.now() - time;
  if (diff < 60 * 1000) return '1 分钟内在线';
  if (diff < 60 * 60 * 1000) return `${Math.floor(diff / (60 * 1000))} 分钟前在线`;
  if (diff < 24 * 60 * 60 * 1000) return `${Math.floor(diff / (60 * 60 * 1000))} 小时前在线`;
  return new Date(time).toLocaleString();
}

function PetPreview({ form }) {
  const petId = Form.useWatch('petId', form) || DEFAULT_PET_ID;
  const pet = getPetDefinition(petId);
  return (
    <div className="settings-pet-preview" aria-label={`当前前贴宠物 ${pet.displayName}`}>
      <div className="settings-pet-frame">
        {pet.renderer === 'multi-sheet'
          ? <LittleCompanionSprite state="idle" preview label={pet.displayName} />
          : <img className="settings-pet-atlas" src={pet.spritesheetPath} alt={pet.displayName} style={{ imageRendering: pet.renderMode === 'smooth' ? 'auto' : undefined }} />}
      </div>
      <div>
        <Typography.Text strong>{pet.displayName}</Typography.Text>
        <Typography.Paragraph type="secondary">{pet.description}</Typography.Paragraph>
      </div>
    </div>
  );
}

export default SettingsPage;
