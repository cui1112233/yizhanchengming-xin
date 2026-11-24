import privateBook from "../../assets/functions/private-custom-reference-3d.webp";
import novelBook from "../../assets/functions/novel-creation-3d.webp";
import batchBooks from "../../assets/functions/batch-review-3d.webp";
import videoClapper from "../../assets/functions/shuihuo-production-3d.webp";
import comicFilm from "../../assets/functions/comic-production-reference-3d.webp";
import aiCut from "../../assets/functions/ai-cut-3d.webp";
import cmCanvas from "../../assets/functions/cm-canvas-3d.webp";
import voiceStudio from "../../assets/functions/voice-studio-3d.webp";
import plugins from "../../assets/functions/plugins-3d.webp";
import type { CatalogModuleKey } from "../../config/modules";

export type FunctionVisual = {
  asset: string;
  label: string;
  description: string;
};

export const FUNCTION_VISUALS: Record<CatalogModuleKey, FunctionVisual> = {
  script: { asset: privateBook, label: "私人定制", description: "按你的想法，生成专属内容" },
  fetch: { asset: novelBook, label: "小说获取", description: "获取、改文与提交" },
  novel: { asset: novelBook, label: "小说创作", description: "从灵感到完整故事" },
  batch: { asset: batchBooks, label: "批量测书", description: "批量获取与生产" },
  shuihuo: { asset: videoClapper, label: "水货生产", description: "漫剧项目与素材生产" },
  comic: { asset: comicFilm, label: "漫画生成", description: "从故事正文到漫画页图" },
  aicut: { asset: aiCut, label: "AI-CUT", description: "智能剪辑与成片" },
  agent: { asset: cmCanvas, label: "CM 画布", description: "多步骤智能协作" },
  voice: { asset: voiceStudio, label: "配音", description: "文本生成自然语音" },
  plugins: { asset: plugins, label: "插件", description: "扩展创作能力" },
};
