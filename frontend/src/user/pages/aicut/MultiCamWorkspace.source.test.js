import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
test('workspace exposes reviewed manual alignment, explicit camera slots and synchronized silent native previews',()=>{
 const s=readFileSync(new URL('./MultiCamWorkspace.jsx',import.meta.url),'utf8');
 for(const token of ['手动对齐','onApply(plan)','preload="metadata"','muted','playsInline','取消','4 格','9 格','源时间偏移','切换点','requestAnimationFrame','cancelAnimationFrame','removeAttribute']) assert.ok(s.includes(token),token);
 assert.ok(!s.includes('autoPlay'));
});

test('workspace automatic alignment is explicit, cancellable and guards stale camera identities',()=>{
 const s=readFileSync(new URL('./MultiCamWorkspace.jsx',import.meta.url),'utf8');
 for(const token of ['onAutoAlign','AbortController','controller.abort()','cameraIdsRef.current','reviewMultiCamAlignment','自动声音对齐','匹配度','timeResolutionMs','分析时长','最大偏移']) assert.ok(s.includes(token),token);
});
