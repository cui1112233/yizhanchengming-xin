import test from 'node:test';
import assert from 'node:assert/strict';
import {availableComicImageModels,comicImageModelFamily,comicImageModelGroups,comicImageModelChange} from './comic-screenplay-model-options.js';

const models=[
  {id:'legacy-kling',modelId:'gpt-image-2-c',displayName:'自己的C通道',kind:'image',capabilities:{generationModes:['text_to_image','image_to_image']}},
  {id:'b',modelId:'[b]gpt-image-2',kind:'image',capabilities:{generationModes:['text_to_image']}},
  {id:'channel',modelId:'gpt-image-2-2k',kind:'image'},
  {id:'disabled',modelId:'FLUX',kind:'image',available:false},
  {id:'text',modelId:'Gemini',kind:'text'},
];
test('图片目录仅选自己可用的图片模型，按后端能力区分生图改图',()=>{
  assert.deepEqual(availableComicImageModels(models).map(m=>m.id),['legacy-kling','b','channel']);
  assert.deepEqual(availableComicImageModels(models,'edit').map(m=>m.id),['legacy-kling','channel']);
  assert.equal(availableComicImageModels([{id:'x',kind:'image',capabilities:{acceptsImageReferences:false}}],'edit').length,0);
});
test('分类来自真实modelId，不把旧目录id或其他软件渠道冒充实际模型',()=>{
  assert.deepEqual(comicImageModelFamily(models[0]),{key:'gpt-image',label:'GPT 图片'});
  const groups=comicImageModelGroups(availableComicImageModels(models));
  assert.deepEqual(groups,[{key:'all',label:'全部',count:3},{key:'gpt-image',label:'GPT 图片',count:3}]);
  assert.deepEqual(comicImageModelGroups([]),[{key:'all',label:'全部',count:0}]);
  assert.equal(comicImageModelFamily({modelId:'my-private-model'}).label,'my-private-model');
});
test('生图选择只在模型支持改图时同步，不覆盖独立改图选择或持久ID',()=>{
  assert.deepEqual(comicImageModelChange('image_model_id','legacy-kling',models),{image_model_id:'legacy-kling',image_edit_model_id:'legacy-kling'});
  assert.deepEqual(comicImageModelChange('image_model_id','b',models),{image_model_id:'b'});
  assert.deepEqual(comicImageModelChange('image_edit_model_id','channel',models),{image_edit_model_id:'channel'});
});
