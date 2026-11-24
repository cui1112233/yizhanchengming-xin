import { Alert, Button, Segmented, Select, Space, Tooltip } from 'antd';
import { useEffect, useState } from 'react';
import { getN8ComicPublishStatus, getN8ComicPublishAudios } from '../../../shared/api/batchFactory';
import { audioMatchingModeOptions } from './batchFactoryAudioCategory';
import { DEFAULT_AUDIO_CATEGORIES, chooseN8Audio, producerAccountOptions, readComicPublishDirectoryCache, writeComicPublishDirectoryCache } from './batchFactoryComicPublishAccounts';
function HelpTip({title,children}) { return <Tooltip title={children}><button type="button" className="batch-factory-help-button" aria-label={`查看${title}说明`}>?</button></Tooltip>; }
function describeN8DirectoryFailure(error) {
	const status = Number(error?.status || 0);
	if (status === 401 || status === 403) return { kind: 'auth', message: '登录已失效，请重新登录' };
	if (status >= 500 || status === 0) return { kind: 'unavailable', message: 'N8 目录服务暂时不可用，请稍后重试' };
	const detail = String(error?.message || '').trim();
	return { kind: 'upstream', message: detail || 'N8 目录服务读取失败，请稍后重试' };
}
export function BatchFactoryComicPublishSettingsForm({ value, onChange, singleBook = false, books = [], productionMode = 'comic' }) {
  const comicPublishSettings = value.comicPublishSettings || {};
	const cachedDirectory = readComicPublishDirectoryCache();
	const [producerAccounts, setProducerAccounts] = useState(cachedDirectory?.producerAccounts || []);
	const [organizations, setOrganizations] = useState(cachedDirectory?.organizations || []);
	const [audioCategories, setAudioCategories] = useState(cachedDirectory?.audioCategories?.length ? cachedDirectory.audioCategories : DEFAULT_AUDIO_CATEGORIES);
	const [producerAccountOrganizations, setProducerAccountOrganizations] = useState(cachedDirectory?.producerAccountOrganizations || {});
	const initialAudios = Array.isArray(comicPublishSettings.audioCandidates) && comicPublishSettings.audioCandidates.length
		? comicPublishSettings.audioCandidates
		: (comicPublishSettings.specifiedAudioId ? [{ id: comicPublishSettings.specifiedAudioId, name: comicPublishSettings.specifiedAudioName || `音频 ${comicPublishSettings.specifiedAudioId}`, status: 'ENABLE' }] : []);
	const [audios, setAudios] = useState(initialAudios);
	const [audiosError, setAudiosError] = useState('');
	const [audiosLoading, setAudiosLoading] = useState(false);
	const [producerAccountsLoading, setProducerAccountsLoading] = useState(!cachedDirectory);
	const [producerAccountsWarning, setProducerAccountsWarning] = useState('');
	const [organizationsWarning, setOrganizationsWarning] = useState('');
	const [audioCategoriesWarning, setAudioCategoriesWarning] = useState('');
	const [directoryFailureKind, setDirectoryFailureKind] = useState('');
	const [directoryFailureMessage, setDirectoryFailureMessage] = useState('');
	const [directoryAvailable, setDirectoryAvailable] = useState(null);
	const [publishKeyConfigured, setPublishKeyConfigured] = useState(null);
  const patchComicPublishSettings = next => onChange({
    ...value,
    comicPublishSettings: { ...comicPublishSettings, ...next }
  });
  useEffect(() => {
    let active = true;
    setProducerAccountsLoading(!cachedDirectory);
    getN8ComicPublishStatus({ silent: true, suppressGlobalError: true }).then(response => {
      if (!active) return;
      const nextAccounts = Array.isArray(response?.producerAccounts) ? response.producerAccounts : [];
      const nextOrganizations = Array.isArray(response?.organizations) ? response.organizations : [];
      const nextCategories = Array.isArray(response?.audioCategories) && response.audioCategories.length ? response.audioCategories : (cachedDirectory?.audioCategories || DEFAULT_AUDIO_CATEGORIES);
      const nextAccountOrganizations = response?.producerAccountOrganizations && typeof response.producerAccountOrganizations === 'object' ? response.producerAccountOrganizations : {};
      setProducerAccounts(nextAccounts);
      setOrganizations(nextOrganizations);
      setAudioCategories(nextCategories);
      setProducerAccountOrganizations(nextAccountOrganizations);
      setProducerAccountsWarning(String(response?.producerAccountsWarning || ''));
      setOrganizationsWarning(String(response?.organizationsWarning || ''));
      setAudioCategoriesWarning(String(response?.audioCategoriesWarning || ''));
      setDirectoryAvailable(typeof response?.directoryAvailable === 'boolean' ? response.directoryAvailable : null);
      setPublishKeyConfigured(typeof response?.publishKeyConfigured === 'boolean' ? response.publishKeyConfigured : null);
      setDirectoryFailureKind('');
      setDirectoryFailureMessage('');
      writeComicPublishDirectoryCache({ producerAccounts: nextAccounts, organizations: nextOrganizations, audioCategories: nextCategories, producerAccountOrganizations: nextAccountOrganizations });
    }).catch(error => {
      if (!active) return;
      const failure = describeN8DirectoryFailure(error);
      setDirectoryFailureKind(failure.kind);
      setDirectoryFailureMessage(failure.message);
    }).finally(() => { if (active) setProducerAccountsLoading(false); });
    return () => { active = false; };
  }, []);
  const manualAudioMode = singleBook && comicPublishSettings.audioCategorySelectionMode === 'manual';
  const manualAudioReady = Boolean(singleBook && manualAudioMode && comicPublishSettings.organizationId && comicPublishSettings.audioCategoryId);
  const loadComicPublishAudios = async (organizationId, audioCategoryId, selectedCategory = {}) => {
    if (!organizationId || !audioCategoryId) return;
    setAudiosLoading(true);
    setAudiosError('');
    try {
      const result = await getN8ComicPublishAudios({ organizationId, audioCategoryId }, { silent: true, suppressGlobalError: true });
      const candidates = Array.isArray(result?.audios) ? result.audios : [];
      setAudios(candidates);
      setAudiosError('');
      const selected = chooseN8Audio(candidates, comicPublishSettings.specifiedAudioId);
      patchComicPublishSettings({
        ...selectedCategory,
        audioCandidates: candidates,
        specifiedAudioId: selected?.id || '',
        specifiedAudioName: selected?.name || ''
      });
    } catch (error) {
      setAudiosError(describeN8DirectoryFailure(error).message);
    } finally {
      setAudiosLoading(false);
    }
  };
  // 目录读取成功但上游返回业务错误时，三路 warning 里保存的就是上游 code/message。
  // 目录失败只作为字段下方的候选区提示，不阻塞字段显示与保存。
  const accountDirectoryHint = producerAccountsWarning
    || directoryFailureMessage
    || (directoryFailureKind === 'auth' ? '登录已失效，请重新登录' : '')
    || (directoryAvailable === false ? 'N8 目录服务未配置（服务器缺少 N8 后台登录令牌）' : '');
  return <Space direction="vertical" size={16} style={{ width: '100%' }}>
    <Alert
      type="info"
      showIcon
      message={productionMode === 'text' ? (singleBook ? '当前书文本发布配置' : '文本发布默认设置') : (singleBook ? '当前书漫画发布配置' : '漫画发布默认设置')}
      description={singleBook ? '当前书未单独修改的项目继续继承批次默认值；这里只保存本书的差异。图片顺序、宽高、格式、签名和上传重试由系统自动处理。' : '这里只设置一次，当前批次的每本漫画都会自动继承。图片顺序、宽高、格式、签名和上传重试由系统自动处理。'}
    />
    {publishKeyConfigured === false
      ? <Alert type="info" showIcon message="发布密钥未配置：目录可以读取，但发布/上传不可用。请配置服务器环境变量 QIANTIE_N8_COMIC_OPEN_API_KEY 后重启服务。" />
      : null}
    <section className="batch-factory-engine-card">
      <header><b>素材字段映射</b><HelpTip title="素材字段映射">这里和 N8 素材库的字段保持一致；单书配置未覆盖时，自动继承发布统一设置。</HelpTip></header>
      <div className="batch-factory-engine-card-body batch-factory-engine-grid">
        <label className="batch-factory-engine-field">
          <span><b>文案来源</b></span>
          <Segmented block className="batch-factory-copy-source-segmented" value={comicPublishSettings.copySource || 'original'} options={[{ value: 'original', label: '原创' }, { value: 'bookstore', label: '书城' }, { value: 'third_party', label: '第三方' }]} onChange={copySource => patchComicPublishSettings({ copySource })} />
        </label>
        <label className="batch-factory-engine-field">
          <span><b>素材来源</b></span>
          <Select value={comicPublishSettings.materialSource || 'ORIGINAL'} options={[{ value: 'ORIGINAL', label: '原创制作' }, { value: 'ORIGINAL_TEXT', label: '原文制作' }, { value: 'DERIVATIZATION', label: '衍生制作' }, { value: 'AI_PRE', label: 'AI前帖' }, { value: 'AUTOMATED_CLIPPING', label: '自动化剪辑' }, { value: 'AUTOMATED_CLIPPING_A', label: '自动化剪辑AI文案' }, { value: 'AUTOMATED_CLIPPING_AI_PRE', label: '自动化剪辑AI前贴' }, { value: 'MANGA_VIDEO_AUTOMATED_CLIPPING', label: '漫剧自动化剪辑' }, { value: 'THIRD_PARTY', label: '第三方' }, { value: 'THIRD_PARTY_REMOVE_WATERMARKS', label: '第三方去水印' }, { value: 'TOOL', label: '素材工具' }, { value: 'GROUP_AI', label: '自动化AI' }, { value: 'AUTOMATED_CLIPPING_BLAZING_FAST', label: '自动化3分钟' }, { value: 'AUTOMATED_CLIPPING_MANGA', label: '自动化漫改' }]} onChange={materialSource => patchComicPublishSettings({ materialSource })} />
        </label>
        <div className="batch-factory-engine-field">
          <span><b>制作者管理员账号（创意者）</b><HelpTip title="制作者管理员账号（创意者）">这就是 N8 素材库里的创意者，不再单独保存第二个账号。</HelpTip></span>
          <Select showSearch optionFilterProp="label" loading={producerAccountsLoading} placement="bottomLeft" value={comicPublishSettings.producerAdminName || undefined} options={producerAccountOptions(producerAccounts, comicPublishSettings.producerAdminName, comicPublishSettings.organizationId, producerAccountOrganizations, producerAccountsLoading)} placeholder="请选择制作者管理员账号" getPopupContainer={() => document.body} onChange={producerAdminName => patchComicPublishSettings({ producerAdminName, creativeAdminName: producerAdminName })} />
          {!accountDirectoryHint && producerAccountOptions(producerAccounts, comicPublishSettings.producerAdminName, comicPublishSettings.organizationId, producerAccountOrganizations).length === 0 ? <small>当前组织没有可用的制作者账号，请先选择正确的组织。</small> : null}
          {accountDirectoryHint ? <small className="batch-factory-engine-note">{accountDirectoryHint}</small> : null}
        </div>
        <div className="batch-factory-engine-field">
          <span><b>组织</b></span>
          <Select placement="bottomLeft" value={String(comicPublishSettings.organizationId || '') || undefined} options={organizations.map(item => ({ value: String(item.id), label: `${item.name}（组织 ${item.id}）` }))} placeholder="请选择 N8 组织" getPopupContainer={() => document.body} onChange={organizationId => { const selected = organizations.find(item => String(item.id) === String(organizationId)); setAudios([]); patchComicPublishSettings({ organizationId: String(organizationId || ''), organizationName: selected?.name || '', audioCandidates: [], specifiedAudioId: '', specifiedAudioName: '' }); }} />
          {organizationsWarning ? <small className="batch-factory-engine-note">{organizationsWarning}</small> : null}
        </div>
      </div>
    </section>
    <section className="batch-factory-engine-card">
      <header><b>发布账号与作品匹配</b><HelpTip title="漫画发布统一">账号必须与漫画后台中已启用的管理员账号完全一致。找不到同名作品时，发布前会让你选择一次。</HelpTip></header>
      <div className="batch-factory-engine-card-body batch-factory-engine-grid">
		<label className="batch-factory-engine-field" style={{ gridColumn: '1 / -1' }}>
		  <span className="batch-factory-field-title-inline"><b>音频匹配方式</b><HelpTip title="音频匹配方式">发布时按每本书的男女频和风格自动匹配并随机选择音频；重试会继续使用第一次选中的音频。</HelpTip></span>
		  <Segmented block value={manualAudioMode ? 'manual' : 'auto'} options={audioMatchingModeOptions(singleBook)} onChange={audioCategorySelectionMode => patchComicPublishSettings(audioCategorySelectionMode === 'manual' ? { audioCategorySelectionMode: 'manual' } : { audioCategorySelectionMode: 'auto', audioCategoryId: null, audioCategoryName: '', audioCandidates: [], specifiedAudioId: '', specifiedAudioName: '' })} />
		</label>
		{singleBook ? <div className="batch-factory-engine-field">
          <span><b>启用中的音频分类</b><HelpTip title="启用中的音频分类">请照着漫画后台“音频分类”页面中启用状态的分类名称填写，发布前会同时显示名称和 ID 供你核对。</HelpTip></span>
          <Select
            value={comicPublishSettings.audioCategoryName || ''}
            options={audioCategories.map(item => ({ value: item.name, label: `${item.name}（ID ${item.id}）`, id: item.id }))}
            placeholder="请选择 N8 启用中的音频分类"
            showSearch
            optionFilterProp="label"
            placement="bottomLeft"
            getPopupContainer={() => document.body}
			disabled={!manualAudioMode}
            onChange={audioCategoryName => {
              const selected = audioCategories.find(item => item.name === audioCategoryName);
              const next = { audioCategoryName, audioCategoryId: selected?.id || null, audioCategorySelectionMode: 'manual', audioCandidates: [], specifiedAudioId: '', specifiedAudioName: '' };
              patchComicPublishSettings(next);
              if (singleBook) loadComicPublishAudios(String(comicPublishSettings.organizationId || ''), String(selected?.id || ''), next);
            }}
          />
          {audioCategoriesWarning ? <small className="batch-factory-engine-note">{audioCategoriesWarning}</small> : null}
		</div> : null}
        {singleBook ? <div className="batch-factory-engine-field"><span><b>指定音频</b></span><Select loading={audiosLoading} disabled={!manualAudioReady} showSearch optionFilterProp="label" value={String(comicPublishSettings.specifiedAudioId || '') || undefined} options={audios.map(item => ({ value: String(item.id), label: `${item.name}（ID ${item.id}）` }))} placeholder={manualAudioMode ? (manualAudioReady ? '选择分类后随机带出，可手动更换' : '请先选择组织和音频分类') : '自动模式将在每本书发布时随机选择'} onChange={specifiedAudioId => { const selected = audios.find(item => String(item.id) === String(specifiedAudioId)); patchComicPublishSettings({ specifiedAudioId: selected?.id || '', specifiedAudioName: selected?.name || '', audioCategorySelectionMode: 'manual' }); }} />
          {manualAudioMode && comicPublishSettings.organizationId && comicPublishSettings.audioCategoryId ? <Button size="small" loading={audiosLoading} onClick={() => loadComicPublishAudios(String(comicPublishSettings.organizationId), String(comicPublishSettings.audioCategoryId))}>刷新指定音频</Button> : null}
          {audiosError ? <small className="batch-factory-engine-note">{audiosError}</small> : null}
          {!audiosError && !audios.length && comicPublishSettings.organizationId && comicPublishSettings.audioCategoryId ? <small>当前分类和组织下没有可用指定音频</small> : null}
          {manualAudioMode && comicPublishSettings.specifiedAudioId ? <small>已随机选择，可手动更换；保存后重试不会重新随机。</small> : null}</div> : null}
        <label className="batch-factory-engine-field">
          <span><b>作品匹配方式</b></span>
          <Segmented
            block
            value={comicPublishSettings.workMatchMode || 'title_auto'}
            options={[
              { value: 'title_auto', label: '按书名自动匹配' },
              { value: 'manual', label: '发布时手动选择' }
            ]}
            onChange={workMatchMode => patchComicPublishSettings({ workMatchMode })}
          />
        </label>
      </div>
    </section>
    <Alert type="success" showIcon message={productionMode === 'text' ? '发布时确认文本合集和匹配到的作品。' : (singleBook ? '当前书发布时，只需要确认漫画合集和匹配到的作品。' : '以后发布时，你只需要确认漫画合集和匹配到的作品。')} />
  </Space>;
}

