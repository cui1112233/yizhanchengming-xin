import {expect,test,vi} from 'vitest';
import {loadTaskSummary,taskDestination} from './taskSummary';
test('overview reads project lists without requesting every water project task', async () => {
 const request=vi.fn(async (path:string):Promise<unknown> => {
  if(path==='/api/history')return {entries:[{id:'same',title:'Earlier',createdAt:'2026-11-18'}]};
  if(path==='/api/shuihuo-production/projects')return {projects:[{id:'same',name:'Latest',updatedAt:'2026-11-19'}]};
  throw new Error('Unavailable');
 });
 const summary=await loadTaskSummary(request,{projectsOnly:true});
 expect(request).toHaveBeenCalledTimes(7);
 expect(summary.entries.map(x=>x.source)).toEqual(['shuihuo','script']);
 expect(taskDestination(summary.entries[0])).toBe('/comic/pages/module-detail/index?key=shuihuo&projectId=same');
 expect(summary.sources.filter(x=>x.state==='error')).toHaveLength(5);
});
test('missing status stays unknown and invalid lists are reported separately', async () => {
 const summary=await loadTaskSummary(async path=>path==='/api/history'?{entries:[{id:'one',title:'Draft'}]}:{},{projectsOnly:true});
 expect(summary.entries[0].status).toBe('unknown');expect(summary.sources.filter(x=>x.state==='error')).toHaveLength(6);
});
