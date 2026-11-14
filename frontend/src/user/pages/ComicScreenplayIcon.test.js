import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import Module from 'node:module';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { buildSync } from 'esbuild';
import { originalComicIcons, originalComicProviders, originalComicProviderIndex } from './comic-screenplay-icon-data.js';

const here=path.dirname(fileURLToPath(import.meta.url));
const compiled=buildSync({stdin:{resolveDir:here,loader:'jsx',contents:`
 import {createElement} from 'react';import {renderToStaticMarkup} from 'react-dom/server';
 import Icon,{ComicScreenplayModelIcon} from './ComicScreenplayIcon.jsx';
 export const renderIcon=(name,size)=>renderToStaticMarkup(createElement(Icon,{name,size}));
 export const renderProviders=models=>renderToStaticMarkup(createElement('div',null,...models.map((modelId,index)=>createElement(ComicScreenplayModelIcon,{key:index,modelId}))));
`},bundle:true,write:false,format:'cjs',platform:'node'}).outputFiles[0].text;
const sandbox=new Module(path.join(here,'icon-ssr-fixture.cjs'));sandbox.filename=path.join(here,'icon-ssr-fixture.cjs');sandbox.paths=Module._nodeModulePaths(here);sandbox._compile(compiled,sandbox.filename);
const {renderIcon,renderProviders}=sandbox.exports;
const golden={"CreateOutline": "ad83dda12d660d3ff37db8a445e03a70a704487d38865ffdd3e878ad580dd8f9", "SparklesOutline": "724751ecbf8aca816b8fb74bede0351d8f25a623d29bfc432cfc506c19b5e468", "ArrowBackOutline": "7f676abef473b7c2cc23c0fffd7ad537a4bfdd916903f21443104040ff674ba3", "SettingsOutline": "38c241260ef972c38b7e64ef979d972f3122fb08ef2fced5341a2654701b18da", "ColorPaletteOutline": "bfdc6e8e646e9f0b13917bfcd8dc013b7c9920b99323ad109e4a78450531fa41", "ReorderThreeOutline": "4978bf7c254a15420dac5353697c8c19f68ad9407013ca272f9695742082715b", "FilmOutline": "70a8a5922a81f10f9fa784618c07f911752b536f38040c925d91d8f1cb148619", "PeopleCircleOutline": "8468a626b6d9cedc5dede365d35e5bd198122c7da65b050c63b621d38c0e2f21"};

test('已接入核心SVG几何与原包独立摘录hash完全相等，不拿24视口通用图替代512视口原图',()=>{
 for(const [name,hash] of Object.entries(golden))assert.equal(createHash('sha256').update(JSON.stringify(originalComicIcons[name].svg)).digest('hex'),hash,name);
 assert.equal(originalComicIcons.ReorderThreeOutline.svg.attrs.viewBox,'0 0 512 512');
});

test('静态树只包含可信图形和本地渐变，没有脚本、事件、外链或HTML执行入口',()=>{
 const allowed=new Set(['svg','span','path','circle','rect','ellipse','g','line','polyline','polygon','defs','linearGradient','stop']);
 function check(node){assert.ok(allowed.has(node.tag),node.tag);for(const [name,value] of Object.entries(node.attrs||{})){assert.doesNotMatch(name,/^on|^href$|^xlink:href$/i);if(!name.startsWith('xmlns'))assert.doesNotMatch(String(value),/javascript:|https?:|url\((?!#)/i);}for(const child of node.children||[])check(child);}
 Object.values(originalComicIcons).forEach(item=>check(item.svg));originalComicProviders.forEach(item=>item.tree.forEach(check));
 const source=readFileSync(path.join(here,'ComicScreenplayIcon.jsx'),'utf8');assert.doesNotMatch(source,/dangerouslySetInnerHTML|\beval\(|new Function/);
});

test('22个模型图标分支严格按原顺序匹配，未知名称是灰头像，不是Sparkles',()=>{
 const samples=['claude-3','gpt-4','gemini-pro','deepseek-v3','grok','qwen3','kimi','moonshot','glm4','mistral','doubao','llama','minimax','hunyuan','yi-34b','ernie','spark','perplexity','baichuan','stepfun','cohere','private-model'];
 samples.forEach((sample,index)=>assert.equal(originalComicProviderIndex(sample),index,sample));
 assert.equal(originalComicProviderIndex('o3-mini'),1);assert.equal(originalComicProviderIndex('豆包'),10);assert.equal(originalComicProviderIndex('通义'),5);
 const fallback=renderProviders(['private-model']);assert.match(fallback,/fill="#94a3b8"/);assert.match(fallback,/fill="#fff"/);assert.doesNotMatch(fallback,/Sparkles/);
});

test('React图标维持原viewBox和描边，尺寸由调用点传入且装饰图不抢按钮可访问名称',()=>{
 const markup=renderIcon('ArrowBackOutline',18);assert.match(markup,/viewBox="0 0 512 512"/);assert.match(markup,/width="18"/);assert.match(markup,/height="18"/);assert.match(markup,/stroke-width="48"/);assert.match(markup,/aria-hidden="true"/);assert.match(markup,/focusable="false"/);assert.equal(renderIcon('unknown',16),'');
 const inline=renderIcon('InlineModelDialogClose',14);assert.match(inline,/viewBox="0 0 24 24"/);assert.match(inline,/stroke-width="2.5"/);
});

test('同页多个provider渐变使用useId隔离且所有本地引用都指向各自定义',()=>{
 const markup=renderProviders(['doubao','doubao','minimax','qwen','hunyuan']);const ids=[...markup.matchAll(/\sid="([^"]+)"/g)].map(match=>match[1]);assert.ok(ids.length>1);assert.equal(new Set(ids).size,ids.length);for(const ref of markup.matchAll(/url\(#([^)]+)\)/g))assert.ok(ids.includes(ref[1]),ref[1]);
});
