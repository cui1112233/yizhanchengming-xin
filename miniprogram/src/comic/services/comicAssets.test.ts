import { beforeEach, describe, expect, test, vi } from "vitest";

const h = vi.hoisted(() => ({
  request: vi.fn(),
  media: vi.fn(),
}));

vi.mock("../../services/api", () => ({ apiRequest: h.request }));
vi.mock("../../services/mediaCache", () => ({ loadAuthenticatedMedia: h.media }));
vi.mock("../../services/batchFactoryV11", () => ({ BATCH_FACTORY_BASE: "/api/batch-factory/v12" }));

import {
  comicAssetKindCounts, comicAssetSentence, comicAssetSteps, extractComicAssets, generateComicAssetImages,
  loadComicAssetImagePath, loadComicAssetImages, loadComicAssets, normalizeComicAssetImages,
  normalizeComicAssets, readComicAssetImageTask, retryComicBookStage,
} from "./comicAssets";

beforeEach(() => { vi.clearAllMocks(); });

describe("漫画资产只读展示", () => {
  test("资产列表映射 name/kind/appearance，过滤无 id 项", () => {
    expect(normalizeComicAssets({
      assets: [
        { id: "a1", kind: "character", name: "林岚", appearance: "短发，穿深色风衣", facts: "记者" },
        { id: "a2", kind: "scene", name: "雨夜街口", prompt: "霓虹反光", revision: 0, imageAspectRatio: "", imageModelId: "" },
        { id: "a3", kind: "prop", name: "旧相机" },
        { kind: "character", name: "没有 id" },
      ],
    })).toEqual([
      { id: "a1", kind: "character", name: "林岚", formalName: "", facts: "记者", appearance: "短发，穿深色风衣", prompt: "", revision: 0, imageAspectRatio: "", imageModelId: "" },
      { id: "a2", kind: "scene", name: "雨夜街口", formalName: "", facts: "", appearance: "", prompt: "霓虹反光", revision: 0, imageAspectRatio: "", imageModelId: "" },
      { id: "a3", kind: "prop", name: "旧相机", formalName: "", facts: "", appearance: "", prompt: "", revision: 0, imageAspectRatio: "", imageModelId: "" },
    ]);
  });

  test("资产图映射 content 路径，缺 url 的条目丢弃", () => {
    expect(normalizeComicAssetImages({
      images: [
        { id: "i1", url: "/api/batch-factory/v12/batches/b1/books/k1/assets/a1/images/i1/content", isPrimary: true },
        { id: "i2", url: "" },
      ],
    })).toEqual([{
      id: "i1",
      url: "/api/batch-factory/v12/batches/b1/books/k1/assets/a1/images/i1/content",
      mediaType: "image/png",
      isPrimary: true,
    }]);
  });

  test("定妆图走 mediaCache 的 request 通道，失败静默返回空串", async () => {
    h.media.mockResolvedValueOnce({ path: "/tmp/asset.png" });
    await expect(loadComicAssetImagePath({ id: "i1", url: "/content", mediaType: "image/png", isPrimary: true }))
      .resolves.toBe("/tmp/asset.png");
    expect(h.media).toHaveBeenCalledWith("/content", "png");
    h.media.mockRejectedValueOnce(new Error("downloadFile:fail url not in domain list"));
    await expect(loadComicAssetImagePath({ id: "i2", url: "/content2", mediaType: "image/png", isPrimary: false }))
      .resolves.toBe("");
  });

  test("三行摘要与四步流程条文案照网页版", () => {
    const assets = normalizeComicAssets({
      assets: [
        { id: "a1", kind: "character", name: "林岚" },
        { id: "a2", kind: "character", name: "沈舟" },
        { id: "a3", kind: "scene", name: "雨夜街口" },
      ],
    });
    expect(comicAssetKindCounts(assets)).toEqual({ character: 2, scene: 1, prop: 0 });
    expect(comicAssetSentence(assets[0])).toBe("");
    expect(comicAssetSteps({ assetCount: 3, assetImageCount: 1, promptReadyCount: 0, pagesTotal: 0, imageReadyCount: 0 }))
      .toEqual([
        { key: "assets", label: "提取资产", done: true, detail: "已完成 3 项" },
        { key: "assetImages", label: "资产图", done: false, detail: "已生成 1/3" },
        { key: "prompts", label: "分镜提示词", done: false, detail: "待生成" },
        { key: "images", label: "出图", done: false, detail: "待提示词" },
      ]);
    expect(comicAssetSteps({ assetCount: 0, assetImageCount: 0, promptReadyCount: 0, pagesTotal: 0, imageReadyCount: 0 })[0].detail).toBe("未提取");
    expect(comicAssetSteps({ assetCount: 0, assetImageCount: 0, promptReadyCount: 0, pagesTotal: 0, imageReadyCount: 0 })[1].detail).toBe("待资产");
    expect(comicAssetSteps({ assetCount: 2, assetImageCount: 2, promptReadyCount: 4, pagesTotal: 4, imageReadyCount: 4 }))
      .toEqual([
        { key: "assets", label: "提取资产", done: true, detail: "已完成 2 项" },
        { key: "assetImages", label: "资产图", done: true, detail: "已生成 2/2" },
        { key: "prompts", label: "分镜提示词", done: true, detail: "已生成 4/4 页" },
        { key: "images", label: "出图", done: true, detail: "已出图 4/4" },
      ]);
  });

  test("四个接口走网页版同一路径与载荷", async () => {
    h.request.mockResolvedValueOnce({ assets: [] });
    await loadComicAssets("batch-1", "book-1");
    expect(h.request).toHaveBeenCalledWith("/api/batch-factory/v12/batches/batch-1/books/book-1/assets", { timeout: 30000 });

    h.request.mockResolvedValueOnce({ images: [] });
    await loadComicAssetImages("batch-1", "book-1", "asset-1");
    expect(h.request).toHaveBeenCalledWith("/api/batch-factory/v12/batches/batch-1/books/book-1/assets/asset-1/images", { timeout: 30000 });

    h.request.mockResolvedValueOnce({});
    await extractComicAssets({ batchId: "batch-1", bookId: "book-1", mode: "force", textModelId: "t1" });
    expect(h.request).toHaveBeenCalledWith(
      "/api/batch-factory/v12/batches/batch-1/books/book-1/stages/assets",
      expect.objectContaining({
        method: "POST",
        data: expect.objectContaining({
          requestId: expect.stringMatching(/^miniprogram-\d+-\d+$/),
          mode: "force",
          textModelId: "t1",
        }),
      }),
    );

    h.request.mockResolvedValueOnce({ images: [], taskId: "task-1" });
    await expect(generateComicAssetImages({
      batchId: "batch-1", bookId: "book-1", assetIds: ["asset-1"], modelId: "i1", imageAspectRatio: "9:16",
    })).resolves.toEqual({ images: [], taskId: "task-1", warning: "" });
    expect(h.request).toHaveBeenCalledWith(
      "/api/batch-factory/v12/batches/batch-1/books/book-1/assets/images/generate",
      expect.objectContaining({ data: { assetIds: ["asset-1"], modelId: "i1", imageAspectRatio: "9:16" } }),
    );

    h.request.mockResolvedValueOnce({ status: "processing" });
    await readComicAssetImageTask("batch-1", "book-1", "task-1");
    expect(h.request).toHaveBeenCalledWith(
      "/api/batch-factory/v12/batches/batch-1/books/book-1/assets/images/generate/task-1",
      { timeout: 30000 },
    );

    h.request.mockResolvedValueOnce({});
    await retryComicBookStage({ batchId: "batch-1", bookId: "book-1", textModelId: "t1" });
    expect(h.request).toHaveBeenCalledWith(
      "/api/batch-factory/v12/batches/batch-1/books/book-1/stages/retry",
      expect.objectContaining({
        method: "POST",
        data: expect.objectContaining({
          requestId: expect.stringMatching(/^miniprogram-\d+-\d+$/),
          textModelId: "t1",
        }),
      }),
    );
  });
});

test('empty successful asset-image responses are not accepted as completed work',async()=>{h.request.mockResolvedValueOnce({});await expect(generateComicAssetImages({batchId:'b',bookId:'k',assetIds:['a']})).rejects.toThrow('回执未确认');});
