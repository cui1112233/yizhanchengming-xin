import { beforeEach, expect, test, vi } from 'vitest';
const h=vi.hoisted(()=>({request:vi.fn(),choose:vi.fn(),read:vi.fn()}));
vi.mock('../../services/api',()=>({apiRequest:h.request}));
vi.mock('../../services/batchFactoryV11',()=>({BATCH_FACTORY_BASE:'/api/batch-factory/v12'}));
vi.mock('@tarojs/taro',()=>({default:{chooseMedia:h.choose,getFileSystemManager:()=>({readFile:h.read})}}));
import {saveComicAsset,saveComicPagePrompts,setComicAssetPrimary,replaceComicImage} from './comicEditing';
beforeEach(()=>{vi.clearAllMocks();h.request.mockResolvedValue({});h.choose.mockResolvedValue({tempFiles:[{tempFilePath:'/photo.png',size:100}]});h.read.mockImplementation(options=>options.success({data:'aW1hZ2U='}));});
test('asset changes preserve revision and target the selected book and asset',async()=>{
  await saveComicAsset('batch/a','book/b',{id:'asset/c',revision:7},{name:'林晚',appearance:'米白外套'});
  expect(h.request).toHaveBeenCalledWith('/api/batch-factory/v12/batches/batch%2Fa/books/book%2Fb/assets/asset%2Fc',{method:'PATCH',data:{name:'林晚',appearance:'米白外套',expectedRevision:7}});
});
test('incomplete page prompts cannot overwrite the saved page',()=>{
  expect(()=>saveComicPagePrompts('b','k','v','有效分页','  ')).toThrow('都需要填写');expect(h.request).not.toHaveBeenCalled();
});
test('page prompts use the video id while asset primary uses the asset id',async()=>{
  await saveComicPagePrompts('b','k','video',' 正文 ',' 图片 ');
  await setComicAssetPrimary('b','k','asset','image');
  expect(h.request.mock.calls[0][0]).toContain('/comic-pages/video');expect(h.request.mock.calls[0][1].data).toEqual({comicPrompt:'正文',imagePrompt:'图片'});
  expect(h.request.mock.calls[1][0]).toContain('/assets/asset/images/image/primary');
});
test('album replacement sends data only to the selected page upload endpoint',async()=>{
  await replaceComicImage('b','k',{videoId:'video'});
  expect(h.request).toHaveBeenCalledWith('/api/batch-factory/v12/batches/b/books/k/comic-pages/video/images',expect.objectContaining({method:'POST',data:{dataUrl:'data:image/png;base64,aW1hZ2U='}}));
});
test('oversize image is rejected before file read or network mutation',async()=>{
  h.choose.mockResolvedValue({tempFiles:[{tempFilePath:'/photo.png',size:4*1024*1024}]});
  await expect(replaceComicImage('b','k',{assetId:'a'})).rejects.toThrow('3 MB');expect(h.read).not.toHaveBeenCalled();expect(h.request).not.toHaveBeenCalled();
});
