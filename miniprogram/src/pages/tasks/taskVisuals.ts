import type { AggregatedTask, TaskSource } from "../../services/taskSummary";
import privateCustom from "../../assets/functions/private-custom-reference-3d.webp";
import agentCanvas from "../../assets/functions/cm-canvas-3d.webp";
import novelBook from "../../assets/functions/novel-creation-3d.webp";
import batchBooks from "../../assets/functions/batch-review-3d.webp";
import comicStudio from "../../assets/functions/comic-studio-mobile.webp";
import storyboardFilm from "../../assets/functions/comic-production-3d.webp";
import videoClapper from "../../assets/functions/shuihuo-production-3d.webp";
import aiCut from "../../assets/functions/ai-cut-3d.webp";

const SOURCE_FALLBACK: Record<TaskSource, string> = {
  script: privateCustom,
  agent: agentCanvas,
  fetch: novelBook,
  batch: batchBooks,
  shuihuo: videoClapper,
  novel:novelBook,
  aicut:aiCut,
};

export function taskVisualAsset(entry: AggregatedTask) {
  const moduleKey = entry.moduleKey || entry.source;
  const typedClue = `${entry.kind || ""} ${entry.taskType || ""} ${entry.stage || ""} ${entry.visualHint || ""}`.toLowerCase();
  if (/(人物|场景|图片|图像|image|character|scene|asset)/.test(typedClue)) return privateCustom;
  if (/(ai-cut|剪辑|裁剪|cut)/.test(typedClue)) return aiCut;
  if (/(分镜|storyboard|shot)/.test(typedClue)) return storyboardFilm;
  if (/(漫画|漫剧|comic|animation)/.test(typedClue)) return comicStudio;
  if (/(导出|成片|视频|video|export|render|merge)/.test(typedClue)) return videoClapper;
  if (/(续写|小说|书籍|正文|rewrite|novel|writing)/.test(typedClue)) return novelBook;
  if (/(批量|测书|batch|review)/.test(typedClue)) return batchBooks;

  if (moduleKey === "script") return SOURCE_FALLBACK.script;
  if (moduleKey === "agent") return SOURCE_FALLBACK.agent;
  if (moduleKey === "fetch") return SOURCE_FALLBACK.fetch;
  const titleClue = `${entry.title || ""}`.toLowerCase();
  if (/(人物|场景|图片|图像|image|character|scene|asset)/.test(titleClue)) return privateCustom;
  if (/(ai-cut|剪辑|裁剪|cut)/.test(titleClue)) return aiCut;
  if (/(分镜|storyboard|shot)/.test(titleClue)) return storyboardFilm;
  if (/(漫画|漫剧|comic|animation)/.test(titleClue)) return comicStudio;
  if (/(导出|成片|视频|video|export|render|merge)/.test(titleClue)) return videoClapper;
  return SOURCE_FALLBACK[moduleKey];
}
