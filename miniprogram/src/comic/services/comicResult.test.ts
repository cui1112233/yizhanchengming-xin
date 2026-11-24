import { beforeEach, describe, expect, test, vi } from "vitest";

const taro = vi.hoisted(() => ({
  downloadFile: undefined as any,
  saveFile: vi.fn(),
  saveImageToPhotosAlbum: vi.fn(),
}));
const api = vi.hoisted(() => ({ apiRequest: vi.fn() }));
const session = vi.hoisted(() => ({ token: "account-a", cookie: "" }));
const media = vi.hoisted(() => ({ loadAuthenticatedMedia: vi.fn() }));

vi.mock("@tarojs/taro", () => ({ default: taro }));
vi.mock("../../services/api", () => api);
vi.mock("../../services/session", () => ({
  getBusinessToken: () => session.token,
  getSessionCookie: () => session.cookie,
}));
vi.mock("../../services/mediaCache", () => media);

import {
  buildPreviewRequest, createComicCollection, downloadComicCollectionZip, downloadComicPagePackage, imageExtension,
  loadComicCollections, loadComicImagePaths, loadComicPages, normalizeComicCollections,
  normalizeComicPages, regenerateComicPage, saveComicImageToAlbum, setPrimaryComicPageImage,
} from "./comicResult";

describe("漫画结果服务", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    session.token = "account-a";
    session.cookie = "";
    taro.downloadFile = undefined;
    media.loadAuthenticatedMedia.mockImplementation(async (_url: string, extension: string) => ({ path: `/tmp/${extension}.png`, fromCache: false }));
  });

  test.each([['prompts','download-zip'],['douyin','douyin-zip']] as const)('导出 %s 使用本书现有下载接口，不提交发布', async (kind,suffix) => {
    taro.saveFile.mockResolvedValue({savedFilePath:'/saved/comic.zip'});
    const result=await downloadComicPagePackage('batch/a','book/b',kind);
    expect(media.loadAuthenticatedMedia).toHaveBeenCalledWith(`/api/batch-factory/v12/batches/batch%2Fa/books/book%2Fb/comic-pages/${suffix}`,'zip');
    expect(result.path).toBe('/saved/comic.zip');expect(api.apiRequest).not.toHaveBeenCalled();
  });

  test("按服务端顺序编号并提取主图，丢弃缺字段的图片", () => {
    const pages = normalizeComicPages({
      pages: [
        { id: "p1", videoId: "v1", status: "ready", aspectRatio: "9:16", comicPrompt: "prompt", imagePrompt: "image prompt", images: [{ id: "i1", url: "/a", isPrimary: true, mediaType: "image/png" }, { id: "i2", url: "/b", isPrimary: false, mediaType: "image/jpeg" }, { id: "", url: "/c" }] },
        { id: "p2", videoId: "v2", status: "generating", images: [] },
      ],
    });
    expect(pages).toHaveLength(2);
    expect(pages[0]).toMatchObject({ ordinal: 1, pageId: "p1", videoId: "v1", status: "ready", imagePrompt: "image prompt", primaryImageId: "i1" });
    expect(pages[0].images).toHaveLength(2);
    expect(pages[1]).toMatchObject({ ordinal: 2, primaryImageId: "", images: [] });
  });

  test("按媒体类型和地址推断图片扩展名", () => {
    expect(imageExtension("image/png")).toBe("png");
    expect(imageExtension("image/webp")).toBe("webp");
    expect(imageExtension("image/jpeg")).toBe("jpg");
    expect(imageExtension("", "/x/content.png?token=1")).toBe("png");
    expect(imageExtension("", "/x/content")).toBe("png");
  });

  test("预览请求始终落在 9 张以内的窗口", () => {
    const paths = Array.from({ length: 12 }, (_, index) => `/tmp/p${index}.png`);
    expect(buildPreviewRequest(paths, 0).urls).toHaveLength(9);
    expect(buildPreviewRequest(paths, 0).current).toBe("/tmp/p0.png");
    const late = buildPreviewRequest(paths, 10);
    expect(late.urls).toEqual(["/tmp/p9.png", "/tmp/p10.png", "/tmp/p11.png"]);
    expect(late.current).toBe("/tmp/p10.png");
    expect(buildPreviewRequest(paths, 10, 20).urls).toHaveLength(3);
  });

  test("读取漫画页使用 book 级路径并解析合集", async () => {
    api.apiRequest.mockResolvedValue({ pages: [{ id: "p1", videoId: "v1", images: [] }], readiness: {} });
    const pages = await loadComicPages("batch-1", "book-1");
    expect(api.apiRequest).toHaveBeenCalledWith("/api/batch-factory/v12/batches/batch-1/books/book-1/comic-pages", { timeout: 30000 });
    expect(pages[0].videoId).toBe("v1");

    api.apiRequest.mockResolvedValue({ collections: [{ id: "c1", bookId: "book-1", status: "ready", items: [{ ordinal: 2, imageId: "i2", url: "/b" }, { ordinal: 1, imageId: "i1", url: "/a" }] }] });
    const collections = await loadComicCollections("batch-1", "book-1");
    expect(collections[0].items.map((item) => item.ordinal)).toEqual([1, 2]);
    expect(normalizeComicCollections({ collections: [{ id: "" }] })).toEqual([]);
  });

  test("创建合集只发送 videoIds，并按 ordinal 排序回执", async () => {
    api.apiRequest.mockResolvedValue({ collection: { id: "c1", bookId: "book-1", status: "ready", items: [{ ordinal: 2, imageId: "i2" }, { ordinal: 1, imageId: "i1" }] } });
    const collection = await createComicCollection("batch-1", "book-1", ["v1", "v2"]);
    expect(api.apiRequest).toHaveBeenCalledWith(
      "/api/batch-factory/v12/batches/batch-1/books/book-1/comic-collections",
      { method: "POST", timeout: 60000, data: { videoIds: ["v1", "v2"] } },
    );
    expect(collection.items.map((item) => item.imageId)).toEqual(["i1", "i2"]);
  });

  test("候选图设为主图调用线上漫画页接口", async () => {
    api.apiRequest.mockResolvedValue({});
    await setPrimaryComicPageImage("batch-1", "book-1", "page-1", "image-2");
    expect(api.apiRequest).toHaveBeenCalledWith(
      "/api/batch-factory/v12/batches/batch-1/books/book-1/comic-pages/page-1/images/image-2/primary",
      { method: "PUT", data: {} },
    );
  });

  test("重新生成单页发送线上需要的提示词、模型和画幅", async () => {
    api.apiRequest.mockResolvedValue({});
    await regenerateComicPage("batch-1", "book-1", "video-1", {
      comicPrompt: "导演描述", imagePrompt: "出图描述", modelId: "image-model", imageProvider: "personal_api", aspectRatio: "3:4",
    });
    expect(api.apiRequest).toHaveBeenCalledWith(
      "/api/batch-factory/v12/batches/batch-1/books/book-1/comic-pages/video-1/generate",
      { method: "POST", timeout: 180000, data: { comicPrompt: "导演描述", imagePrompt: "出图描述", modelId: "image-model", imageProvider: "personal_api", aspectRatio: "3:4" } },
    );
  });

  test("图片逐张下载，失败项保留空值", async () => {
    media.loadAuthenticatedMedia
      .mockResolvedValueOnce({ path: "/tmp/a.png", fromCache: false })
      .mockRejectedValueOnce(new Error("下载失败"));
    const paths = await loadComicImagePaths({
      ordinal: 1, pageId: "p1", videoId: "v1", status: "ready", aspectRatio: "", comicPrompt: "", imagePrompt: "",
      images: [{ id: "a", url: "/a", isPrimary: true, mediaType: "image/png" }, { id: "b", url: "/b", isPrimary: false, mediaType: "image/png" }],
      primaryImageId: "a",
    });
    expect(paths.a).toBe("/tmp/a.png");
    expect(paths.b).toBeUndefined();
  });

  test("合集 ZIP 走 request 拉取并 saveFile 持久化", async () => {
    media.loadAuthenticatedMedia.mockResolvedValue({ path: "/tmp/comic-collection.zip", fromCache: false });
    taro.saveFile.mockResolvedValue({ savedFilePath: "/saved/comic-collection.zip" });
    const file = await downloadComicCollectionZip("batch-1", "book-1", "c1");
    expect(media.loadAuthenticatedMedia).toHaveBeenCalledWith(expect.stringContaining("/comic-collections/c1/download"), "zip");
    expect(taro.downloadFile).toBeUndefined();
    expect(file.path).toBe("/saved/comic-collection.zip");
    expect(file.url).toContain("/comic-collections/c1/download");
  });

  test("合集拉取失败时给出可重试的提示", async () => {
    media.loadAuthenticatedMedia.mockRejectedValue(new Error("网络中断"));
    await expect(downloadComicCollectionZip("batch-1", "book-1", "c1")).rejects.toThrow(/合集下载失败/);
  });

  test("相册权限被拒时抛出引导文案", async () => {
    taro.saveImageToPhotosAlbum.mockRejectedValue({ errMsg: "saveImageToPhotosAlbum:fail auth deny" });
    await expect(saveComicImageToAlbum("/tmp/a.png")).rejects.toThrow(/权限/);
    taro.saveImageToPhotosAlbum.mockResolvedValue({});
    await expect(saveComicImageToAlbum("/tmp/a.png")).resolves.toBeUndefined();
    await expect(saveComicImageToAlbum("")).rejects.toThrow(/尚未下载完成/);
  });
});
