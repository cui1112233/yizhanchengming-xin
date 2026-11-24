import { normalizeSensitiveMode, SENSITIVE_MODE_OPTIONS } from './batchFactoryBookConfigRegions.js';
import { reconcileSyncedWebsiteProfiles, websiteProfileOrganization } from './batchFactoryWebsiteProfiles.js';

const fonts = [
  {value:'noto-sans-sc',label:'思源黑体 · 清晰黑体'},
  {value:'ma-shan-zheng',label:'马善政 · 毛笔手写'},
  {value:'lxgw-wenkai',label:'霞鹜文楷 · 楷书 / 钢笔'},
  {value:'long-cang',label:'龙藏体 · 自然手写'},
  {value:'zhi-mang-xing',label:'志莽行书 · 连笔手写'},
  {value:'zcool-kuaile',label:'站酷快乐体 · 圆润活泼'}
];
const options = values => values.map(value => ({ value, label: String(value) }));
const align = [{ value: 'left', label: '左对齐' }, { value: 'center', label: '居中' }, { value: 'right', label: '右对齐' }];
const field = (key, label, type, initial, extra = {}) => ({key, label, type, initial, ...extra});
const solid = { kind: 'color', color: '#111318' };

export const TEXT_PRODUCTION_SECTIONS = [
  { key: 'models', label: '引擎配置', fields: [
    field('textModelId','文本模型','model',''),
    field('sensitiveMode','敏感内容处理','choice','system',{options:SENSITIVE_MODE_OPTIONS,aiHint:'生产阶段会调用文本模型。AI 优化失败会停止生产并显示失败原因。'}),
    field('sensitiveAIModels','AI 修复模型','models',[]),
    field('sensitiveAIOnFailure','取文阶段未处理命中词','choice','review',{options:[{value:'review',label:'打码并标记待复核'},{value:'fail',label:'停止取文'}]}),
    field('highlightReuseEnabled','按关键词标红','switch',false,{hidden:true})
  ]},
  { key:'layout', label:'排版配置', fields:[
    field('textPaginationMode','分页方式','choice','auto',{options:[{value:'auto',label:'智能排版'},{value:'target',label:'目标字数'}]}),
    field('textPageTarget','单页目标字数','number',150,{min:1,max:100000}),
    field('contentRange','内容截取','choice',4000,{options:options([1000,2000,4000,8000,12000,20000,50000,100000])}),
    field('imageAspectRatio','画幅','choice','3:4',{options:options(['3:4','9:16','1:1','4:3','16:9','21:9','3:2','2:3'])}),
    field('textFontId','正文字体','choice','noto-sans-sc',{options:fonts}),
    field('textFontSize','字号','number',44,{min:1,max:256}),
    field('textFontWeight','字重','choice',400,{options:options([100,200,300,400,500,600,700,800,900])}),
    field('textLineHeight','行距倍数','number',1.65,{min:1,max:8,step:0.05}),
    field('textParagraphSpacing','段距（像素）','number',24,{min:0,max:1080}),
    field('textPadding','页面边距（像素）','number',96,{min:0,max:539}),
    field('textAlign','正文对齐','choice','left',{options:align}),
    field('textColor','文字颜色','color','#FFFFFF'),
    field('textHighlightColor','标红颜色','color','#E53935'),
    field('textHighlightRules','标红规则','highlightRules',''),
    field('textHighlightLines','标红行号','input','1,3'),
    field('textMaxHeightRatio','文字最大高度占比','percent',0.75,{min:1,max:100}),
    field('textBackground','整书背景','background',solid),
    field('textCover.layout.fontId','标题字体','choice','noto-sans-sc',{options:fonts}),
    field('textCover.layout.fontSize','标题字号','number',64,{min:1,max:256}),
    field('textCover.layout.fontWeight','标题字重','choice',600,{options:options([100,200,300,400,500,600,700,800,900])}),
    field('textCover.layout.align','标题对齐','choice','center',{options:align}),
    field('textCover.layout.textColor','标题颜色','color','#FFFFFF')
  ]},
  {key:'publish',label:'发布配置',fields:[
    field('comicPublishSettings.producerAdminName','上传人','producer',''),
    field('comicPublishSettings.organizationId','组织归属','organization',''),
    ...[
      ['copySource','original'],['materialSource','ORIGINAL'],['creativeAdminName',''],['organizationName',''],
      ['audioCategorySelectionMode','auto'],['audioCategoryId',null],['audioCategoryName',''],['audioCandidates',[]],
      ['specifiedAudioId',''],['specifiedAudioName',''],['workMatchMode','title_auto']
    ].map(([key,initial])=>field('comicPublishSettings.'+key,key,'publishShared',initial)),
    field('publishSettings.websiteProfileId','网站配置档','profile','',{hidden:true}),
    field('publishSettings.bookCategory','书籍分类','input','',{hidden:true}),
    field('textAutoCollection','自动生成合集','switch',true,{hidden:true}),
    field('textAutoUpload','自动上传','switch',false,{hidden:true}),
    field('textUploadRetryLimit','上传失败重试次数','number',3,{min:0,max:10,hidden:true})
  ]}
];
export const TEXT_PRODUCTION_BOOK_REGIONS = Object.freeze([
  {key:'models',label:'引擎配置',detail:'模型与敏感处理'},
  {key:'layout',label:'排版配置',detail:'字体、颜色与背景'},
  {key:'publish',label:'发布配置',detail:'上传人与发布规则'}
]);
export function textProductionInitialSection(value) {
  return TEXT_PRODUCTION_SECTIONS.some(section=>section.key===value) ? value : 'models';
}
const fields = TEXT_PRODUCTION_SECTIONS.flatMap(section => section.fields);
const profileId = 'publishSettings.websiteProfileId';
const profileCatalog = 'publishSettings.websiteProfiles';
const profileSnapshotKeys = [profileCatalog, 'publishSettings.versionProfile','publishSettings.organization'];
const safeProfileFields = ['id','name','config_id','enabled','source','is_default','platform_id','gender','style','book_category','bookCategory'];
// The sync response also carries account settings and arbitrary upstream
// config_data. Persist only profile metadata and the publish form's fields.
const safeAdvancedFields = [
  'tl5','ziti','zitidx','biaohong','biaohongReuse','biaohong_reuse','keywords',
  'jieyaNum','jieya_num','jieyaAiHead','jieya_ai_head','jieyaSpeed','jieya_speed','jieyaPitch','jieya_pitch',
  'fontColorStyles','font_color_styles','gunpingNum','gunping_num','gunpingSpeed','gunping_speed',
  'jieyaAiHeadHorizontalFlip','jieya_ai_head_horizontal_flip','aiHeadVideoMode','jieya_ai_head_video_mode',
  'book_category','bookCategory'
];
const scalar = value => typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean';
function safeFields(source, keys) {
  const result = {};
  for (const key of keys) {
    const value = source?.[key];
    if (scalar(value) || (Array.isArray(value) && value.every(scalar))) result[key] = value;
  }
  return result;
}
export function textProfileChanges(catalog, selectedId) {
  const profiles = (Array.isArray(catalog) ? catalog : []).filter(profile=>profile?.id).map(profile=>({
    ...safeFields(profile,safeProfileFields),
    organization: websiteProfileOrganization(profile),
    advanced: safeFields(profile.advanced ?? profile.config_data,safeAdvancedFields)
  }));
  const {publishSettings} = reconcileSyncedWebsiteProfiles({websiteProfileId:selectedId},profiles);
  return Object.fromEntries(Object.entries(publishSettings).map(([key,value])=>['publishSettings.'+key,value]));
}
export const TEXT_PRODUCTION_CONFIG_KEYS = Object.freeze(fields.map(item => item.key));
const read = (object, path) => path.split('.').reduce((v,k) => v?.[k], object);
function write(object, keys, value) {
  const [key, ...rest] = keys;
  if (!rest.length) {
    if (value === undefined) delete object[key];
    else object[key] = value;
    return;
  }
  const child = { ...object[key] };
  write(child, rest, value);
  if (Object.keys(child).length) object[key] = child;
  else delete object[key];
}
const legacyKeys = {contentRange:'contentCaptureCharacters',sensitiveMode:'sensitiveAIEnabled',highlightReuseEnabled:'publishSettings.biaohongReuse'};
function explicitValue(settings, key) {
  const v = read(settings,key);
  if (v !== undefined && v !== null) return v;
  const legacy = legacyKeys[key];
  if (!legacy || read(settings,legacy) == null) return undefined;
  return key === 'sensitiveMode' ? normalizeSensitiveMode(settings) : read(settings,legacy);
}
export function textProductionBookOverride(batchPatch = {}, bookPatch = {}) {
  const value = {}, explicitKeys = [];
  for (const {key, initial, from} of fields) {
    const topLevelPolicy = (key === 'highlightReuseEnabled' || key === 'sensitiveMode') && read(batchPatch,key) != null;
    const scoped = topLevelPolicy ? read(bookPatch,key) : explicitValue(bookPatch,key);
    if (scoped !== undefined) explicitKeys.push(key);
    value[key] = scoped ?? explicitValue(batchPatch,key) ?? (from ? value[from] : structuredClone(initial));
  }
  value[profileCatalog] = read(bookPatch,profileCatalog) ?? read(batchPatch,profileCatalog) ?? [];
  return {value,explicitKeys};
}
export const normalizeTextProductionConfig = settings => textProductionBookOverride(settings).value;
export const defaultTextProductionConfig = () => normalizeTextProductionConfig({});

// Changes use dotted UI paths. The settings API replaces top-level values;
// keep only the book's own sibling overrides when serializing a nested group.
export function updateTextProductionSettings(current = {}, changes = {}) {
  const next = {...current}, touched = new Set();
  const set = (key,v) => {const keys=key.split('.');touched.add(keys[0]);write(next,keys,v);};
  for (const [key,value] of Object.entries(changes)) {
    if (!TEXT_PRODUCTION_CONFIG_KEYS.includes(key) && !profileSnapshotKeys.includes(key)) continue;
    set(key,value);
    if (key === 'contentRange') set('contentCaptureCharacters',value);
    else if (legacyKeys[key] && explicitValue(current,legacyKeys[key]) !== undefined)
      set(legacyKeys[key],undefined);
  }
  if (Object.hasOwn(changes,profileId) && changes[profileId] === undefined) {
    for (const key of profileSnapshotKeys) set(key,undefined);
  }
  const patch = {}, restoreKeys = [];
  for (const k of touched) {
    if (Object.hasOwn(next,k)) patch[k] = next[k];
    else if (Object.hasOwn(current,k)) restoreKeys.push(k);
  }
  return {patch,restoreKeys};
}
export function applyTextProductionChanges(current, changes) {
  const {patch,restoreKeys} = updateTextProductionSettings(current,changes);
  const next = {...current,...patch};
  for (const key of restoreKeys) delete next[key];
  return next;
}
