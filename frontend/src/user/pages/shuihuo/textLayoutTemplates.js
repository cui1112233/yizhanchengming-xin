// Explicit choices only: never replace a person's settings on opening a form.
export const TEXT_LAYOUT_TEMPLATES = [
 {id:'reading',label:'舒适阅读',values:{textPaginationMode:'auto',textFontId:'noto-sans-sc',textFontSize:40,textFontWeight:400,textLineHeight:1.8,textParagraphSpacing:40,textPadding:112,textMaxHeightRatio:0.75,textAlign:'left',textColor:'#FFFFFF',textBackground:{kind:'color',color:'#111318'}}},
 {id:'handwriting',label:'纸感手写',values:{textPaginationMode:'auto',textFontId:'long-cang',textFontSize:44,textFontWeight:400,textLineHeight:1.8,textParagraphSpacing:36,textPadding:104,textMaxHeightRatio:0.75,textAlign:'left',textColor:'#302920',textBackground:{kind:'color',color:'#F6EFDF'}}},
 {id:'clear',label:'清晰黑体',values:{textPaginationMode:'auto',textFontId:'lxgw-wenkai',textFontSize:46,textFontWeight:500,textLineHeight:1.65,textParagraphSpacing:32,textPadding:96,textMaxHeightRatio:0.75,textAlign:'left',textColor:'#FFFFFF',textBackground:{kind:'color',color:'#111318'}}},
];

export function textColorScheme(mode){
 return mode==='dark'?{textColor:'#222222',textBackground:{kind:'color',color:'#F6EFDF'}}:{textColor:'#FFFFFF',textBackground:{kind:'color',color:'#111318'}};
}
function luminance(hex){
 if(!/^#[0-9a-f]{6}$/i.test(hex || ''))return null;
 const rgb=[1,3,5].map(i=>parseInt(hex.slice(i,i+2),16)/255).map(v=>v<=0.04045?v/12.92:((v+0.055)/1.055)**2.4);
 return rgb[0]*0.2126+rgb[1]*0.7152+rgb[2]*0.0722;
}
function contrast(a,b){return (Math.max(a,b)+0.05)/(Math.min(a,b)+0.05);}
export function hasLowTextContrast(color,background){
 if(background?.id)return false; // Image contrast is measured by the real preview renderer.
 const a=luminance(color),b=luminance(background?.color || '#111318');
 return a!==null && b!==null && contrast(a,b)<4.5;
}
export function readableTextColor(background){
 const bg=luminance(background?.color || '#111318') ?? 0;
 return contrast(1,bg)>=contrast(luminance('#222222'),bg)?'#FFFFFF':'#222222';
}

export const TEXT_PUBLISH_OVERRIDE_GROUPS = [
 ['文案来源',['copySource']],['素材来源',['materialSource']],
 ['制作者账号',['producerAdminName','creativeAdminName']],
 ['组织',['organizationId','organizationName']],
 ['音频匹配',['audioCategorySelectionMode','audioCategoryId','audioCategoryName','audioCandidates','specifiedAudioId','specifiedAudioName']],
 ['作品匹配',['workMatchMode']],
].map(([label,keys])=>({label,keys:keys.map(key=>'comicPublishSettings.'+key)}));

export function publishOverrideRestores(explicitKeys){
 return TEXT_PUBLISH_OVERRIDE_GROUPS.filter(group=>group.keys.some(key=>explicitKeys.includes(key)));
}
