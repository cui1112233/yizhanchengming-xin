import {beforeEach,expect,test,vi} from 'vitest';
import {moduleRoute,legacyModuleDetailDestination} from '../../config/modules';
import appConfig from '../../app.config';
const h=vi.hoisted(()=>({load:null as any,redirect:vi.fn(),slots:[] as any[],cursor:0}));
vi.mock('@tarojs/taro',()=>({default:{redirectTo:h.redirect},useLoad:(fn:any)=>{h.load=fn;}}));
vi.mock('@tarojs/components',()=>({Text:'Text',View:'View',Button:'Button'}));
vi.mock('react',()=>({useState:(value:any)=>{const i=h.cursor++;if(!(i in h.slots))h.slots[i]=value;return[h.slots[i],(v:any)=>{h.slots[i]=v;}];},useRef:(value:any)=>{const i=h.cursor++;if(!(i in h.slots))h.slots[i]={current:value};return h.slots[i];}}));
vi.mock('../../theme',()=>({ThemedPage:'ThemedPage'}));
// A compatibility route must never import any heavy workspace implementation.
vi.mock('./ScriptWorkspace',()=>{throw Error('main package still imports ScriptWorkspace');});
vi.mock('./FetchWorkspace',()=>{throw Error('main package still imports FetchWorkspace');});
vi.mock('./CmWorkspace',()=>{throw Error('main package still imports CmWorkspace');});
vi.mock('./ShuihuoWorkspace',()=>{throw Error('main package still imports ShuihuoWorkspace');});
import LegacyEntry from './index';
const render=()=>{h.cursor=0;return LegacyEntry();};
const nodes=(v:any):any[]=>!v||typeof v!=='object'?[]:[v,...[v.props?.children].flat(Infinity).flatMap(nodes)];
beforeEach(()=>{h.slots=[];h.redirect.mockReset().mockResolvedValue({});});
test('new functional entry directly opens a registered subpackage page',()=>{
 for(const key of ['script','booktest','shuihuo','agent'] as const)expect(moduleRoute(key)).toBe(`/comic/pages/module-detail/index?key=${key}`);
 expect(appConfig.subPackages.find(p=>p.root==='comic')?.pages).toContain('pages/module-detail/index');
});
test('legacy entry preserves Unicode, reserved characters and all task parameters',()=>{
 const params={key:'agent',taskId:'a/b?x=1',projectId:'作品 A',historyId:'h#1',entryId:'old',tab:'tasks'};
 const result=legacyModuleDetailDestination(params);const query=new URL('https://test.invalid'+result).searchParams;
 expect(result.startsWith('/comic/pages/module-detail/index?')).toBe(true);
 for(const [key,value]of Object.entries(params))expect(query.get(key)).toBe(value);
 render();h.load(params);expect(h.redirect).toHaveBeenCalledWith({url:result});
});
test('failed compatibility navigation can be retried with the same destination',async()=>{
 h.redirect.mockRejectedValueOnce(Error('下载失败'));render();h.load({key:'script',historyId:'saved'});await Promise.resolve();await Promise.resolve();
 const button=nodes(render()).find(v=>v.type==='Button');expect(button).toBeDefined();button.props.onClick();
 expect(h.redirect).toHaveBeenCalledTimes(2);expect(h.redirect.mock.calls[1]).toEqual(h.redirect.mock.calls[0]);
});
