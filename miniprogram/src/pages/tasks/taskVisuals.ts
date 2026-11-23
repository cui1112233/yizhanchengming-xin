import type { AggregatedTask, TaskSource } from "../../services/taskSummary";
import privateCustom from "../../assets/functions/private-custom-3d.webp";
import agentCanvas from "../../assets/functions/cm-canvas-3d.webp";
import novelBook from "../../assets/functions/novel-book-v2.webp";
import batchBooks from "../../assets/functions/batch-books-v2.webp";
import comicStudio from "../../assets/functions/comic-studio-mobile.webp";
import storyboardFilm from "../../assets/functions/comic-production-3d.webp";
import videoClapper from "../../assets/functions/clapper-play-v2.webp";
import aiCut from "../../assets/functions/aicut-scissors-v2.webp";

const SOURCE_FALLBACK: Record<TaskSource, string> = {
  script: privateCustom,
  agent: agentCanvas,
  fetch: novelBook,
  batch: batchBooks,
  shuihuo: videoClapper,
};

export function taskVisualAsset(entry: AggregatedTask) {
  const clue = `${entry.visualHint || ""} ${entry.title || ""}`.toLowerCase();
  if (/(人物|场景|图片|图像|image|character|scene|asset)/.test(clue)) return privateCustom;
  if (/(ai-cut|剪辑|裁剪|cut)/.test(clue)) return aiCut;
  if (/(导出|成片|视频|video|export|render)/.test(clue)) return videoClapper;
  if (/(分镜|storyboard|shot)/.test(clue)) return storyboardFilm;
  if (/(漫画|漫剧|comic|animation)/.test(clue)) return comicStudio;
  if (/(续写|小说|书籍|正文|rewrite|novel|writing)/.test(clue)) return novelBook;
  if (/(批量|测书|batch|review)/.test(clue)) return batchBooks;
  return SOURCE_FALLBACK[entry.source];
}
