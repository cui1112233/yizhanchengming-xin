export const creativePromptFields = [
  { field: 'script_prompt_id', purpose: 'script', label: '剧本提示词' },
  { field: 'storyboard_prompt_id', purpose: 'storyboard', label: '分镜提示词' },
  { field: 'character_design_prompt_id', purpose: 'character_design', label: '角色设计提示词' },
  { field: 'scene_design_prompt_id', purpose: 'scene_design', label: '场景设计提示词' },
];

export const comicModelConfigFields = ['config_version','aspect_ratio','image_resolution','asset_prompt_type','asset_process_prompt','voice_design_model','show_original_text','character_image_mode','scene_image_mode'];

export function emptyCreativeConfig() {
  return { script_prompt_id: '', storyboard_prompt_id: '', character_design_prompt_id: '', scene_design_prompt_id: '', style_id: null, style_name: '', image_core_instruction: '', image_model_id: '', image_edit_model_id: '' };
}

// 新建时未触碰的指令省略提交，让后端保留系统默认；输入后清空仍为显式空串。
export function creativeCreationDraft() {
  return {...emptyCreativeConfig(),image_core_instruction:undefined,asset_process_prompt:undefined};
}

export function creativeConfigDraft(config) {
  const draft = emptyCreativeConfig();
  for (const field of Object.keys(draft)) if (config?.[field] !== undefined) draft[field] = config[field];
  for (const field of comicModelConfigFields) if (config?.[field] !== undefined) draft[field] = config[field];
  draft.style_id = config?.style?.id || null;
  draft.style_name = config?.style?.name || '';
  return draft;
}

export function creativeConfigPayload(draft, revision) {
  const patch = { expected_revision: revision, style_id: draft.style_id || null };
  for (const { field, label } of creativePromptFields) {
    if (!draft[field]) throw new Error(`请选择${label}`);
    patch[field] = draft[field];
  }
  for (const field of ['image_core_instruction', 'image_model_id', 'image_edit_model_id']) if(draft[field]!==undefined)patch[field] = draft[field] || '';
  for (const field of comicModelConfigFields) if (field !== 'config_version' && draft[field] !== undefined) patch[field] = draft[field];
  return patch;
}

export const comicModelConfigPatchFields = [...comicModelConfigFields.filter(field=>field!=='config_version'),'image_core_instruction','image_model_id','image_edit_model_id','character_design_prompt_id','scene_design_prompt_id'];
export function creativeModelConfigPayload(draft, revision) {
  const patch={expected_revision:revision};
  for(const field of comicModelConfigPatchFields)if(draft[field]!==undefined)patch[field]=draft[field];
  return patch;
}

export function creativePromptOptions(prompts, purpose, selectedId) {
  const rows = prompts.filter(item => String(item.purpose || item.type || '').replaceAll('-', '_') === purpose);
  const options = rows.map(item => {
    const version = String(item.version ?? '').trim();
    const displayVersion = /^[0-9a-f]{8,}$/i.test(version) ? '' : version;
    return { value: item.id, label: `${item.name}${displayVersion ? ` · ${displayVersion}` : ''}` };
  });
  if (selectedId && !rows.some(item => item.id === selectedId)) options.unshift({ value: selectedId, label: `已保存的选择（${selectedId}，当前不可用）`, disabled: true });
  return options;
}

// 同一新建草稿只建立一个作品；配置保存失败后继续补存，不重复新建。
export function createCreativeCreationFlow({ create, load, save }) {
  let project = null, pending = null;
  return {
    getProject() { return project; },
    run(payload, patch, onCreated = () => {}) {
      if (pending) return pending;
      pending = (async () => {
        if (!project) {
          project = await create(payload);
          if (!project?.id) throw new Error('作品创建结果缺少编号，请刷新项目库核对');
          onCreated(project);
        }
        const current = await load(project.id);
        await save(project.id, { ...patch, expected_revision: current.revision });
        return project;
      })().finally(() => { pending = null; });
      return pending;
    },
  };
}


export function comicStoryboardEngine(promptId, fallback = '') {
  return ({
    luoshui_storyboard_v9_30s_307: 'xiaoluo_v9_30s',
    storyboard_xiaoluo_v7_15s: 'xiaoluo_v7_15s',
    storyboard_xiaoyunque_v3: 'xiaoyunque_v3',
  })[promptId] || fallback;
}

export function comicWorkspaceStoryboardEngine(project, boards, current = 'xiaoluo_v7_15s') {
  const configured = comicStoryboardEngine(project?.creative_config?.storyboard_prompt_id);
  if (configured) return configured;
  const available = (boards || []).filter(board => board.shots?.length);
  return available.some(board => board.engine === current) ? current : available[0]?.engine || current;
}

export function comicImageModelSelection(config, { models = [], references = [], override = '', loading = false, loadError = '' } = {}) {
  const editing = references.length > 0;
  const configuredId = editing ? config?.image_edit_model_id || config?.image_model_id || '' : config?.image_model_id || '';
  const modelId = override || configuredId;
  const source = override ? 'temporary' : 'project';
  const options = models.map(item => ({ value: item.id, label: item.displayName || item.name || item.id }));
  if (modelId && !options.some(option => option.value === modelId)) options.unshift({ value: modelId, label: `当前模型（${modelId}，不可用）`, disabled: true });
  let error = loadError;
  if (!error && !loading && !modelId) error = `作品尚未设置${editing ? '图片编辑或图片生成' : '图片生成'}模型，请在项目设置保存，或为本次操作选择模型。`;
  if (!error && !loading && modelId && !models.some(item => item.id === modelId)) error = `${override ? '临时选择的' : '作品保存的'}图片模型「${modelId}」当前不可用，请重新选择；没有自动更换模型。`;
  return { modelId, source, editing, options, error, loading };
}


export function applyCreativeConfigToProject(project, config, revision = project?.revision) {
  return { ...project, ...(config?.aspect_ratio ? {aspect_ratio:config.aspect_ratio} : {}), creative_config: config, revision, art_style: config?.style?.name || '', selected_style: config?.style?.id || '' };
}

export function comicProjectStyleName(project) {
  return project?.creative_config ? project.creative_config.style?.name || '' : project?.art_style || '';
}
