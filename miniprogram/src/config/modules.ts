import type { AppIconName } from "../components/AppIcon";

export type ModuleKey = "script" | "fetch" | "comic" | "panel" | "shuihuo" | "agent" | "batch" | "toolkit" | "aicut" | "voice" | "plugins";
export type CatalogModuleKey = Exclude<ModuleKey, "panel" | "toolkit">;

export type CreatorModule = {
  key: CatalogModuleKey;
  label: string;
  description: string;
  icon: AppIconName;
  group: "创作" | "生产" | "助手";
};

export const MODULES: CreatorModule[] = [
  { key: "script", label: "私人定制", description: "按想法生成专属内容", icon: "file-pen-line", group: "创作" },
  { key: "fetch", label: "小说创作", description: "获取、改文与提交", icon: "book-open", group: "创作" },
  { key: "batch", label: "批量测书", description: "批量获取与生产", icon: "factory", group: "创作" },
  { key: "shuihuo", label: "水货生产", description: "漫剧项目与素材生产", icon: "clapperboard", group: "生产" },
  { key: "comic", label: "漫画生成", description: "从正文生成漫画页图", icon: "sparkles", group: "生产" },
  { key: "aicut", label: "AI-CUT", description: "智能剪辑与成片", icon: "sliders", group: "生产" },
  { key: "agent", label: "CM 画布", description: "多步骤智能协作", icon: "bot", group: "助手" },
  { key: "voice", label: "配音", description: "文本生成语音", icon: "volume", group: "助手" },
  { key: "plugins", label: "插件", description: "扩展创作能力", icon: "settings", group: "助手" },
];

export const DEFAULT_MODULE: ModuleKey = "script";

export function getModule(key: string) {
  return MODULES.find((item) => item.key === key) || MODULES[0];
}

export function isModuleKey(key: string): key is ModuleKey {
  return MODULES.some((item) => item.key === key);
}

export function moduleRoute(key: ModuleKey) {
  if (key === "toolkit" || key === "plugins") return "/comic/pages/index/index";
  if (key === "voice") return "/comic/pages/tts/index";
  if (key === "aicut") return "/pages/module-detail/index?key=agent";
  if (key === "comic") return "/comic/pages/comic-new/index";
  return key === "batch" ? "/pages/batch/index" : `/pages/module-detail/index?key=${key}`;
}
