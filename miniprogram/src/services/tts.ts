import Taro from "@tarojs/taro";
import { apiRequest } from "./api";
import { onSessionInvalidated, sessionRevision } from './sessionBoundary';
const audioFiles=new Set<string>();
export function releaseTtsAudio(path:string){if(!audioFiles.has(path))return;audioFiles.delete(path);try{Taro.getFileSystemManager().unlink({filePath:path,fail:()=>undefined});}catch{/* cache already cleared */}}
onSessionInvalidated(()=>{for(const file of [...audioFiles])releaseTtsAudio(file);});

export type TtsOptions = {
  voice?: string;
  speed?: number;
  pitch?: number;
  style?: string;
};

export const DEFAULT_TTS_OPTIONS: TtsOptions = { voice: "zh-CN-XiaoxiaoNeural", style: "general", speed: 1.8, pitch: 10 };

export const TTS_VOICES: Array<{ label: string; value: string }> = [
  { label: "晓晓（女声·温柔）", value: "zh-CN-XiaoxiaoNeural" },
  { label: "晓辰（女声·知性）", value: "zh-CN-XiaochenNeural" },
  { label: "云希（男声·清朗）", value: "zh-CN-YunxiNeural" },
  { label: "云扬（男声·阳光）", value: "zh-CN-YunyangNeural" },
  { label: "晓伊（女声·甜美）", value: "zh-CN-XiaoyiNeural" },
  { label: "云健（男声·稳重）", value: "zh-CN-YunjianNeural" },
];

export const TTS_STYLES: Array<{ label: string; value: string }> = [
  { label: "通用", value: "general" },
  { label: "开心", value: "cheerful" },
  { label: "悲伤", value: "sad" },
  { label: "友好", value: "friendly" },
  { label: "聊天", value: "chat" },
];

function finiteNumber(value: unknown) {
  const number = typeof value === "number" ? value : typeof value === "string" && value.trim() ? Number(value) : Number.NaN;
  return Number.isFinite(number) ? number : undefined;
}

/** 只保留服务端明确返回且类型正确的字段，避免把未知内容透传给配音接口。 */
export function normalizeTtsOptions(value: unknown): TtsOptions {
  const input = value && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : {};
  const options: TtsOptions = {};
  if (typeof input.voice === "string" && input.voice.trim()) options.voice = input.voice;
  if (typeof input.style === "string" && input.style.trim()) options.style = input.style;
  const speed = finiteNumber(input.speed);
  if (speed !== undefined) options.speed = speed;
  const pitch = finiteNumber(input.pitch);
  if (pitch !== undefined) options.pitch = pitch;
  return options;
}

/** 读取 `/api/config` 中的默认配音参数；服务端未返回时为空对象，由调用方决定是否使用本地默认值。 */
export async function readTtsOptions(): Promise<TtsOptions> {
  const config = await apiRequest<any>("/api/config", { timeout: 12000 });
  return normalizeTtsOptions(config?.tts);
}

export async function saveTtsOptions(value:TtsOptions):Promise<TtsOptions>{
  const tts=normalizeTtsOptions(value);
  if(!tts.voice||!tts.style||!Number.isFinite(tts.speed)||tts.speed!<0.5||tts.speed!>2||!Number.isFinite(tts.pitch)||tts.pitch!< -50||tts.pitch!>50)throw new Error('配音设置超出支持范围');
  const result=await apiRequest<any>('/api/config',{method:'POST',data:{tts}});
  const saved=normalizeTtsOptions(result?.tts);
  if(Object.entries(tts).some(([key,v])=>saved[key as keyof TtsOptions]!==v))throw new Error('配音默认值未确认保存，请重新读取核对');
  return saved;
}

export async function synthesizeSpeech(input: string, options: TtsOptions = {}): Promise<string> {
  const revision=sessionRevision();
  const text = input.trim();
  if (!text) throw new Error("请输入要配音的文字");
  const bytes = await apiRequest<ArrayBuffer>("/api/tts", {
    method: "POST",
    timeout: 180000,
    data: { input: text, ...options },
    responseType: "arraybuffer",
  });
  if (!(bytes instanceof ArrayBuffer) || !bytes.byteLength) throw new Error("配音接口未返回有效音频");
  if(revision!==sessionRevision())throw new Error('账号已切换，已丢弃旧配音结果');
  const filePath = `${Taro.env.USER_DATA_PATH}/yzcm-tts-${Date.now()}-${Math.random().toString(36).slice(2,8)}.mp3`;
  await new Promise<void>((resolve, reject) => {
    Taro.getFileSystemManager().writeFile({ filePath, data: bytes, success: () => resolve(), fail: reject });
  });
  audioFiles.add(filePath);
  if(revision!==sessionRevision()){releaseTtsAudio(filePath);throw new Error('账号已切换，已清理旧配音缓存');}
  return filePath;
}
