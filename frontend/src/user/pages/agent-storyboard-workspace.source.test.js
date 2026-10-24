import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const page = fs.readFileSync(path.resolve(dirname, './AgentPageV2.jsx'), 'utf8');
const workspace = fs.readFileSync(path.resolve(dirname, '../components/AgentStoryboardWorkspace.jsx'), 'utf8');
const model = fs.readFileSync(path.resolve(dirname, './agent-storyboard-model.js'), 'utf8');
const storyboardState = fs.readFileSync(path.resolve(dirname, './agent-storyboard-state.js'), 'utf8');
const styles = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');

test('工作流和故事板切换会渲染不同主体而不是只修改地址', () => {
  assert.match(page, /viewState\.activeView === 'workflow'/);
  assert.match(page, /<AgentStoryboardWorkspace/);
  assert.match(page, /onOpenNodeInWorkflow/);
  assert.match(page, /canvasRef\.current\?\.locateNode/);
});

test('故事板提供分类项目素材与独立详情栏', () => {
  assert.match(workspace, /aria-label="故事板项目素材"/);
  assert.match(workspace, /aria-label="故事板详情"/);
  for (const label of ['文本', '音频', '图片', '视频']) assert.match(model, new RegExp(label));
  assert.match(workspace, /materialLibrary\.map/);
  assert.match(workspace, /AuthenticatedAgentImage/);
  assert.match(workspace, /AuthenticatedAgentVideo/);
  assert.match(workspace, /在工作流中打开/);
});

test('故事板消费接口镜头并提供来源定位编辑批量生成和版本列表', () => {
  assert.match(page, /listAgentScriptShots/);
  assert.match(page, /setStoryboardShots/);
  assert.match(page, /shots=\{storyboardShots\}/);
  assert.match(page, /createAgentShotBatchNodes/);
  assert.match(workspace, /locateShotNode/);
  assert.match(workspace, /updateAgentScriptShot/);
  assert.match(workspace, /onCreateShotBatch/);
  assert.match(workspace, /镜头版本/);
  assert.match(workspace, /定位来源/);
});

test('工作流和故事板读取同一份镜头集合，不各自维护镜头副本', () => {
  assert.match(page, /scriptShots=\{storyboardShots\}/);
  assert.match(page, /onScriptShotsChange=\{handleSharedScriptShotsChange\}/);
  assert.match(workspace, /onShotUpdated\?\.\(result\.shot\)/);
  assert.match(model, /currentStoryboardShots/);
});

test('故事板和工作流共享服务端持久任务的项目汇总', () => {
  const storyboardWorkspace = page.slice(
    page.indexOf('<AgentStoryboardWorkspace'),
    page.indexOf('<aside className="agent-conversation-panel"')
  );
  assert.match(page, /getAgentGenerationSummary/);
  assert.match(page, /const \[generationSummary, setGenerationSummary\]/);
  assert.match(storyboardWorkspace, /generationSummary=\{generationSummary\}/);
  assert.match(workspace, /generationSummary = null/);
  assert.match(workspace, /generationSummaryText\(generationSummary\)/);
  assert.match(workspace, /agent-storyboard-generation-summary/);
});

test('镜头批量响应只交给提交时的项目和保存队列', () => {
  const handler = page.match(/const handleStoryboardBatchCreate = useCallback\(async payload => \{([\s\S]*?)\n  \}, \[\]\);/);
  assert.ok(handler, 'page should own the scoped storyboard batch request');
  assert.match(handler[1], /const requestedTaskId = activeTaskIdRef\.current/);
  assert.match(handler[1], /const requestedQueue = canvasSaveQueueRef\.current/);
  assert.match(handler[1], /createAgentShotBatchNodes\(requestedTaskId, payload\)/);
  assert.match(handler[1], /shouldAcceptCanvasRefresh/);
  assert.match(handler[1], /requestedQueue\.acceptServerCanvas/);
  assert.doesNotMatch(handler[1], /canvasSaveQueueRef\.current\?\.acceptServerCanvas/);
  assert.match(workspace, /onCreateShotBatch\?\.\(payload\)/);
});

test('故事板只在用户选择素材后打开详情并允许明确关闭', () => {
  assert.match(workspace, /className=\{`agent-storyboard-workspace\$\{selected \? ' has-selection' : ''\}`\}/);
  assert.match(workspace, /selected \? <aside/);
  assert.match(workspace, /aria-label="关闭素材详情"/);
  assert.match(workspace, /selectedKey: ''/);
  assert.match(styles, /agent-storyboard-workspace:not\(\.has-selection\)/);
  assert.match(styles, /agent-storyboard-material-column--left/);
  assert.match(styles, /agent-storyboard-material-column--right/);
});

test('故事板详情直接复用工作流节点的模型参数和设置控件', () => {
  assert.match(workspace, /AgentModelPicker/);
  assert.match(workspace, /AgentNodeParameterPopover/);
  assert.match(workspace, /AgentNodeSettingsPanel/);
  assert.match(workspace, /listAvailableModels/);
  assert.match(workspace, /onNodeConfigChange/);
  assert.match(page, /patchStoryboardNodeConfig/);
  assert.match(page, /onNodeConfigChange=\{handleStoryboardNodeConfigChange\}/);
  assert.match(page, /canvasSaveQueueRef\.current\?\.enqueue\(nextCanvas\)/);
});

test('故事板音频编辑与工作流一致地按操作决定是否需要模型', () => {
  assert.match(workspace, /audioOperationNeedsModel/);
  assert.match(workspace, /selected\.kind !== 'audio' \|\| selectedAudioNeedsModel/);
  assert.match(workspace, /本地音频处理/);
});

test('故事板使用紧凑导演区加右侧独立素材列并在窄屏降为可用单栏', () => {
  assert.match(styles, /agent-canvas-workspace\.is-storyboard/);
  assert.match(styles, /agent-storyboard-workspace/);
  assert.match(styles, /agent-storyboard-library/);
  assert.match(styles, /agent-storyboard-inspector/);
  assert.match(styles, /agent-canvas-workspace\.is-storyboard[^}]*grid-template-columns:\s*minmax\(320px,\s*var\(--storyboard-director-split,\s*31%\)\)\s+minmax\(0,\s*1fr\)/);
  assert.match(styles, /agent-storyboard-material-column--left/);
  assert.match(styles, /agent-storyboard-material-column--right/);
  assert.match(styles, /agent-storyboard-inspector[^}]*position:\s*absolute/);
  assert.match(styles, /@media \(max-width: 1000px\)/);
});

test('故事板素材四分区支持三条分隔线独立调整并双击复位', () => {
  assert.match(workspace, /storyboardSplitRatioFromPointer/);
  assert.match(workspace, /agent-storyboard-grid-splitter--column/);
  assert.match(workspace, /agent-storyboard-grid-splitter--left-row/);
  assert.match(workspace, /agent-storyboard-grid-splitter--right-row/);
  assert.match(workspace, /onPointerDown=\{event => startGridResize\(event, 'pointer'\)\}/);
  assert.match(workspace, /onDoubleClick=\{resetGridSplit\}/);
  assert.match(styles, /--storyboard-column-split/);
  assert.match(styles, /--storyboard-left-row-split/);
  assert.match(styles, /--storyboard-right-row-split/);
  assert.match(styles, /agent-storyboard-material-column--left/);
  assert.match(styles, /agent-storyboard-material-column--right/);
  assert.match(styles, /cursor:\s*col-resize/);
  assert.match(styles, /cursor:\s*row-resize/);
});

test('故事板素材分区使用文件夹图标表达可收起的分组', () => {
  assert.match(workspace, /FolderOpen/);
  assert.match(workspace, /Folder size=\{13\}/);
  assert.match(workspace, /agent-storyboard-section__folder/);
  assert.match(styles, /agent-storyboard-section__folder[^}]*color: var\(--agent-accent\)/);
});

test('没有镜头时不占用素材区高度，读取和错误状态仍可见', () => {
  assert.match(workspace, /const showShotSection = shotsLoading \|\| Boolean\(shotsError\) \|\| shotSection\.items\.length > 0/);
  assert.match(workspace, /\{showShotSection \? <section className="agent-storyboard-shots"/);
  assert.doesNotMatch(workspace, /当前剧本还没有镜头/);
});

test('素材分区标题支持指针拖动重排并避免拖动后误触收起', () => {
  assert.match(workspace, /sectionDragRef/);
  assert.match(workspace, /sectionClickSuppressionRef/);
  assert.match(workspace, /data-section-kind=\{section\.kind\}/);
  assert.match(workspace, /onPointerDown=\{event => startSectionDrag\(event, section\.kind\)\}/);
  assert.match(workspace, /onPointerMove=\{moveSectionDrag\}/);
  assert.match(workspace, /onPointerUp=\{event => finishSectionDrag\(event\)\}/);
  assert.match(workspace, /document\.elementsFromPoint\(event\.clientX, event\.clientY\)/);
  assert.match(workspace, /sectionClickSuppressionRef\.current = true/);
  assert.match(workspace, /onStoryboardStateChange\?\.\(\{ layout: \{ sectionOrder \} \}\)/);
});

test('故事板素材支持同分区直接拖拽排序并保留按钮操作', () => {
  assert.match(workspace, /reorderStoryboardMaterial/);
  assert.match(workspace, /storyboardPointerDropTarget/);
  assert.match(workspace, /agent-storyboard-item__drag/);
  assert.match(workspace, /aria-label=\{`拖动排序 \$\{item\.title\}`\}/);
  assert.match(workspace, /onPointerDown=\{event => startItemDrag\(event, item, section\.items\)\}/);
  assert.match(workspace, /onPointerMove=\{moveItemDrag\}/);
  assert.match(workspace, /onPointerUp=\{event => finishItemDrag\(event\)\}/);
  assert.match(workspace, /data-storyboard-key=\{item\.key\}/);
  assert.match(workspace, /上移素材/);
  assert.match(workspace, /下移素材/);
});

test('故事板提供真实素材上传入口并把新素材写回当前项目', () => {
  assert.match(workspace, /uploadAgentAsset/);
  assert.match(workspace, /添加素材/);
  assert.match(workspace, /type="file"/);
  assert.match(workspace, /onAssetsCreated/);
  assert.match(page, /taskId=\{activeTaskId\}/);
  assert.match(page, /onAssetsCreated=\{handleWorkspaceAssetsCreated\}/);
});

test('故事板批量模式提供全选、加入对话和同类型分类', () => {
  assert.match(workspace, /批量操作/);
  assert.match(workspace, /全选当前结果/);
  assert.match(workspace, /批量加入对话/);
  assert.match(workspace, /storyboardBatchCategoryPatch/);
  assert.match(workspace, /toggleStoryboardBatchSelection/);
});

test('故事板素材可以加入当前对话并随发送请求提交', () => {
  assert.match(workspace, /onAddToConversation/);
  assert.match(workspace, /加入对话/);
  assert.match(page, /selectedAssets: assetContextForRequest/);
  assert.match(page, /agent-composer-asset-attachments/);
  assert.match(page, /ConversationAssetPreview/);
  assert.match(page, /removeStoryboardConversationAsset/);
});

test('故事板素材可直接拖到对话输入框作为引用', () => {
  assert.match(workspace, /draggable=\{!batchMode && Boolean\(conversationAssetID\)\}/);
  assert.match(workspace, /writeStoryboardAssetDrag/);
  assert.match(workspace, /onDragStart=\{event => startConversationAssetDrag\(event, item\)\}/);
  assert.match(page, /readStoryboardAssetDrop/);
  assert.match(page, /onDrop=\{handleConversationAssetDrop\}/);
  assert.match(page, /is-asset-drop-target/);
  assert.match(styles, /is-asset-drop-target/);
});

test('未关联节点的项目素材可直接加入工作流并定位到新节点', () => {
  assert.match(workspace, /onAddToWorkflow/);
  assert.match(workspace, /加入工作流/);
  assert.match(page, /addStoryboardAssetToWorkflow/);
  assert.match(page, /createNodeAt/);
  assert.match(page, /pendingWorkflowNodeRef\.current = nodeId/);
  assert.match(page, /openNodeInWorkflow\(nodeId\)/);
});

test('故事板提供 Liblib 风格二级分类并把选择和排序写回同一画布', () => {
  for (const label of ['剧本', '其他', '角色', '场景', '物品', '视频片段', '成片']) {
    assert.match(storyboardState, new RegExp(label));
  }
  assert.match(workspace, /storyboardState/);
  assert.match(workspace, /onStoryboardStateChange/);
  assert.match(workspace, /素材分类/);
  assert.match(workspace, /上移素材/);
  assert.match(workspace, /下移素材/);
  assert.match(page, /storyboard: nextStoryboard/);
  assert.match(page, /viewState\.storyboard/);
  assert.match(styles, /agent-storyboard-section__filters/);
});
