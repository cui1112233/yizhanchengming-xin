import { beforeEach, describe, expect, test, vi } from "vitest";

const api = vi.hoisted(() => ({ apiRequest: vi.fn() }));
const taro = vi.hoisted(() => ({
  env: { USER_DATA_PATH: "/user-data" },
  getFileSystemManager: vi.fn(),
}));
vi.mock("./api", () => api);
vi.mock("@tarojs/taro", () => ({ default: taro }));

import { normalizeTtsOptions, readTtsOptions, saveTtsOptions, synthesizeSpeech } from "./tts";

describe("TTS 配音服务", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    taro.getFileSystemManager.mockReturnValue({
      writeFile: ({ success }: { success: () => void }) => success(),
    });
  });

  test("只保留服务端明确返回的合法配音字段", () => {
    expect(normalizeTtsOptions({ voice: "v", style: "s", speed: "1.8", pitch: 10, unknown: "x" })).toEqual({ voice: "v", style: "s", speed: 1.8, pitch: 10 });
    expect(normalizeTtsOptions({ voice: "", speed: "fast", pitch: null })).toEqual({});
    expect(normalizeTtsOptions(null)).toEqual({});
  });

  test("读取配置只透传类型正确的 tts 字段", async () => {
    api.apiRequest.mockResolvedValue({ tts: { voice: "zh-CN-YunxiNeural", speed: 1.2, extra: true } });
    await expect(readTtsOptions()).resolves.toEqual({ voice: "zh-CN-YunxiNeural", speed: 1.2 });
    expect(api.apiRequest).toHaveBeenCalledWith("/api/config", { timeout: 12000 });
  });

  test("空文本不发请求", async () => {
    await expect(synthesizeSpeech("   ")).rejects.toThrow("请输入要配音的文字");
    expect(api.apiRequest).not.toHaveBeenCalled();
  });

  test("音频返回非法时不写文件", async () => {
    api.apiRequest.mockResolvedValue({});
    await expect(synthesizeSpeech("你好")).rejects.toThrow("配音接口未返回有效音频");
    expect(taro.getFileSystemManager).not.toHaveBeenCalled();
  });

  test("成功时写入本地并返回文件路径", async () => {
    api.apiRequest.mockResolvedValue(new ArrayBuffer(8));
    const path = await synthesizeSpeech("你好", { voice: "zh-CN-XiaoxiaoNeural" });
    expect(path).toMatch(/^\/user-data\/yzcm-tts-\d+-[a-z0-9]+\.mp3$/);
    expect(api.apiRequest).toHaveBeenCalledWith("/api/tts", expect.objectContaining({
      method: "POST", responseType: "arraybuffer", data: { input: "你好", voice: "zh-CN-XiaoxiaoNeural" },
    }));
  });
});

test('default preference save sends only tts and reads back the saved preference',async()=>{
 api.apiRequest.mockResolvedValue({tts:{voice:'v',speed:1.2,pitch:-10,style:'general'}});
 await expect(saveTtsOptions({voice:'v',speed:1.2,pitch:-10,style:'general'})).resolves.toMatchObject({pitch:-10});
 expect(api.apiRequest).toHaveBeenCalledWith('/api/config',{method:'POST',data:{tts:{voice:'v',speed:1.2,pitch:-10,style:'general'}}});
});
