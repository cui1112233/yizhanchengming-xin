import {beforeEach,expect,test,vi} from 'vitest';
const h=vi.hoisted(()=>({slots:[] as any[],cursor:0,load:null as any,redirect:vi.fn()}));
vi.mock('@tarojs/taro',()=>({default:{redirectTo:h.redirect},useLoad:(fn:any)=>{h.load=fn;}}));
vi.mock('react',()=>({useState:(initial:any)=>{const i=h.cursor++;if(!(i in h.slots))h.slots[i]=initial;return[h.slots[i],(v:any)=>{h.slots[i]=v;}];}}));
vi.mock('@tarojs/components',()=>({View:'View',Text:'Text',Button:'Button'}));
vi.mock('../../theme',()=>({ThemedPage:'ThemedPage'}));
vi.mock('../../components/AmbientStarfield',()=>({default:'AmbientStarfield'}));
vi.mock('../../components/BrandHeader',()=>({default:'BrandHeader'}));
vi.mock('./ScriptWorkspace',()=>({default:'ScriptWorkspace'}));
vi.mock('./FetchWorkspace',()=>({default:'FetchWorkspace'}));
vi.mock('./CmWorkspace',()=>({default:'CmWorkspace'}));
vi.mock('./PanelWorkspace',()=>({default:'PanelWorkspace'}));
vi.mock('./ShuihuoWorkspace',()=>({default:'ShuihuoWorkspace'}));
import Content from './ModuleDetailContent';
const render=()=>{h.cursor=0;return Content();};
beforeEach(()=>{h.slots=[];h.redirect.mockReset().mockResolvedValue({});});
test.each([
 [{key:'script',entryId:'h/a'},'ScriptWorkspace',{historyId:'h/a'}],
 [{key:'booktest',taskId:'book/a'},'FetchWorkspace',{initialTaskId:'book/a',compact:true}],
 [{key:'agent',projectId:'project/a',taskId:'task/a',tab:'tasks'},'CmWorkspace',{projectId:'project/a',taskId:'task/a',initialTab:'tasks'}],
 [{key:'shuihuo',projectId:'water/a'},'ShuihuoWorkspace',{projectId:'water/a'}],
])('subpackage passes the original target parameters to its workspace: %j',(params,type,props)=>{
 render();h.load(params);const result=render();expect(result.type).toBe(type);expect(result.props).toMatchObject(props);
});
test('old novel links preserve project and session when routed to the novel page',()=>{
 render();h.load({key:'novel',projectId:'作品 A',sessionId:'s/a',runId:'run#1'});
 const url=new URL('https://test.invalid'+h.redirect.mock.calls[0][0].url);
 expect(url.pathname).toBe('/comic/pages/novel/index');expect(url.searchParams.get('projectId')).toBe('作品 A');expect(url.searchParams.get('sessionId')).toBe('s/a');expect(url.searchParams.get('runId')).toBe('run#1');
});
