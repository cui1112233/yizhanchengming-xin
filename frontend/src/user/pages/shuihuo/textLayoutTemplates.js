// Explicit choices only: never replace a person's settings on opening a form.
export const TEXT_LAYOUT_TEMPLATES = [
 {id:'reading',label:'舒适阅读',values:{textPaginationMode:'auto',textFontId:'noto-sans-sc',textFontSize:40,textFontWeight:400,textLineHeight:1.8,textParagraphSpacing:40,textPadding:112,textMaxHeightRatio:0.75,textAlign:'left',textColor:'#FFFFFF',textBackground:{kind:'color',color:'#111318'}}},
 {id:'handwriting',label:'纸感手写',values:{textPaginationMode:'auto',textFontId:'lxgw-wenkai',textFontSize:44,textFontWeight:400,textLineHeight:1.8,textParagraphSpacing:36,textPadding:104,textMaxHeightRatio:0.75,textAlign:'left',textColor:'#302920',textBackground:{kind:'color',color:'#F6EFDF'}}},
 {id:'clear',label:'清晰黑体',values:{textPaginationMode:'auto',textFontId:'noto-sans-sc',textFontSize:46,textFontWeight:500,textLineHeight:1.65,textParagraphSpacing:32,textPadding:96,textMaxHeightRatio:0.75,textAlign:'left',textColor:'#FFFFFF',textBackground:{kind:'color',color:'#111318'}}},
];

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
