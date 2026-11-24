import Taro from '@tarojs/taro';
import {sessionRevision} from './sessionBoundary';
export async function pickWaterMedia(kind:'image'|'video'|'audio'){
 const revision=sessionRevision();let path='',size=0,name='';try{if(kind==='audio'){const r=await Taro.chooseMessageFile({count:1,type:'file',extension:['mp3','wav','m4a','aac']});const file=r.tempFiles[0];if(file){path=file.path;size=file.size;name=file.name;}}else{const r=await Taro.chooseMedia({count:1,mediaType:[kind],sourceType:['album','camera']});const file=r.tempFiles[0];if(file){path=file.tempFilePath;size=file.size;name=path.split('/').pop()||'';}}}catch(c:any){if(/cancel/i.test(c.errMsg||''))return null;throw c;}
 if(revision!==sessionRevision())throw Error('账号已切换，请重新选择素材');if(!path)return null;if(size>8*1024*1024)throw Error('当前手机导入单个素材最多8MiB，请先压缩素材');
 const extension=(name.match(/\.([a-z0-9]+)$/i)?.[1]||'').toLowerCase(),mime=({png:'image/png',jpg:'image/jpeg',jpeg:'image/jpeg',webp:'image/webp',mp4:'video/mp4',mov:'video/quicktime',mp3:'audio/mpeg',wav:'audio/wav',m4a:'audio/mp4',aac:'audio/aac'} as any)[extension];if(!mime?.startsWith(kind+'/'))throw Error('文件格式不匹配，请选择支持的图片、视频或音频');
 const r:any=await new Promise((resolve,reject)=>Taro.getFileSystemManager().readFile({filePath:path,encoding:'base64',success:resolve,fail:reject}));if(revision!==sessionRevision())throw Error('账号已切换，未上传素材');if(typeof r.data!=='string'||r.data.length>Math.ceil(8*1024*1024/3)*4)throw Error('素材太大，未上传');return{kind,name,dataUrl:`data:${mime};base64,${r.data}`};
}
