import { normalizeSensitiveMode, SENSITIVE_MODE_OPTIONS } from './batchFactoryBookConfigRegions.js';

const fonts = [{ value: 'noto-sans-sc', label: 'Noto Sans SC · 黑体' }, { value: 'ma-shan-zheng', label: 'Ma Shan Zheng · 手写体' }];
const options = values => values.map(value => ({ value, label: String(value) }));
const align = [{ value: 'left', label: '左对齐' }, { value: 'center', label: '居中' }, { value: 'right', label: '右对齐' }];
const field = (key, label, type, initial, extra = {}) => ({key, label, type, initial, ...extra});
const solid = { kind: 'color', color: '#111318' };

export const TEXT_PRODUCTION_SECTIONS = [
  { key: 'models', label: '模型配置', fields: [
    field('textModelId','文本模型','model',''),
    field('sensitiveMode','敏感内容处理','choice','system',{options:SENSITIVE_MODE_OPTIONS}),
    field('sensitiveAIModels','AI 修复模型','models',[]),
    field('sensitiveAIOnFailure','失败时策略','choice','review',{options:[{value:'review',label:'自动打码并继续'},{value:'fail',label:'判为失败'}]}),
    field('highlightReuseEnabled','按关键词标红','switch',false)
  ]},
  { key:'layout', label:'排版配置', fields:[
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
    field('textMaxHeightRatio','文字最大高度占比','percent',0.75,{min:1,max:100}),
    field('textBackground','整书背景','background',solid),
    field('textCover.background','封面背景','background',null,{from:'textBackground'}),
    field('textCover.layout.fontId','标题字体','choice',null,{options:fonts,from:'textFontId'}),
    field('textCover.layout.fontSize','标题字号','number',null,{min:1,max:256,from:'textFontSize'}),
    field('textCover.layout.fontWeight','标题字重','choice',null,{options:options([100,200,300,400,500,600,700,800,900]),from:'textFontWeight'}),
    field('textCover.layout.align','标题对齐','choice',null,{options:align,from:'textAlign'}),
    field('textCover.layout.textColor','标题颜色','color',null,{from:'textColor'})
  ]},
  {key:'publish',label:'发布配置',fields:[
    field('comicPublishSettings.producerAdminName','上传人','producer',''),
    field('comicPublishSettings.organizationId','组织归属','organization',''),
    field('publishSettings.websiteProfileId','网站配置档','profile',''),
    field('publishSettings.bookCategory','书籍分类','input',''),
    field('textAutoCollection','自动生成合集','switch',true),
    field('textAutoUpload','自动上传','switch',false),
    field('textUploadRetryLimit','上传失败重试次数','number',3,{min:0,max:10})
  ]}
];
const fields = TEXT_PRODUCTION_SECTIONS.flatMap(section => section.fields);
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
    if (!TEXT_PRODUCTION_CONFIG_KEYS.includes(key)) continue;
    set(key,value);
    if (key === 'contentRange') set('contentCaptureCharacters',value);
    else if (legacyKeys[key] && explicitValue(current,legacyKeys[key]) !== undefined)
      set(legacyKeys[key],undefined);
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
