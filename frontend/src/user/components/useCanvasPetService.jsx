import { useEffect, useRef } from 'react';
import { getCurrentUsername } from '../../shared/api/auth.js';
import { registerPetPageService, openPetPageService } from '../../shared/pet/pageService.js';
import CanvasPetPanel from './CanvasPetPanel.jsx';

export function useCanvasPetService({ taskId, scope: providedScope, canvas, selectedIds, flush, acceptCanvas, canApply = true, beginApply, editSelection, locateSelection }) {
  const username = getCurrentUsername();
  const scope = providedScope || `${username}:canvas:${taskId}`;
  const runtime = useRef(null);
  const current = { taskId, scope, selectedIds, flush, acceptCanvas, canApply, beginApply, editSelection, locateSelection };
  runtime.current = current;
  useEffect(() => { runtime.current = current; return () => { if (runtime.current === current) runtime.current = null; }; });
  const cleanup = useRef(null);
  useEffect(() => () => { cleanup.current?.(); cleanup.current = null; }, [scope]);
  useEffect(() => {
    if (!taskId || typeof flush !== 'function' || typeof acceptCanvas !== 'function') return;
    const check = () => runtime.current?.scope === scope ? runtime.current : null;
    const services = {
      getTaskId: () => check()?.taskId,
      getSelectedIds: () => check()?.selectedIds || [],
      checkCanApply: () => { if (!check()?.canApply) throw Error('卡片正在生成或处理，请结束后再应用画布方案'); },
      beginApply: () => check()?.beginApply?.(),
      flush: () => { const current = check(); if (!current) throw Error('项目已切换'); return current.flush(); },
      acceptCanvas: next => { const current = check(); if (!current) throw Error('项目已切换'); return current.acceptCanvas(next); }
    };
    cleanup.current = registerPetPageService({ id: 'canvas-agent', scope, label: 'AI 创作画布',
      render: () => <CanvasPetPanel key={scope} taskId={taskId} selectedCount={selectedIds.length} nodeCount={canvas?.nodes?.length || 0} revision={Number(canvas?.revision) || 0} services={services}
        onEditSelection={() => check()?.editSelection()} onLocateSelection={() => check()?.locateSelection()} />
    });
  }, [scope, taskId, canvas, selectedIds, flush, acceptCanvas]);
  return () => openPetPageService('canvas-agent');
}
