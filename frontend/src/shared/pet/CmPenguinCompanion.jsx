import { useCallback, useEffect, useMemo, useRef, useState, useSyncExternalStore } from 'react';
import { ExternalLink, MessageCircle, ScanSearch, Sparkles, Undo2, X } from 'lucide-react';
import { askAgent, createAgentTask, getAgentTask } from '../api/agent';
import { getConfig } from '../api/config';
import {
  PET_CONTEXT_EVENT,
  PET_EVENT,
  PET_SKILLS_EVENT,
  dispatchPetPreview,
  dispatchPetState,
  normalizePetContext,
  normalizePetState,
  petAtlasRow,
  petFrameCount,
  petLookFrame,
  petSpeech,
  readCmTaskId,
  writeCmTaskId
} from './stacky';
import { parseScriptRevision, classifyPetRequestError } from './scriptCollaboration';
import {
  COMPANION_SPEECH_PRIORITY,
  PET_COMPANION_SETTINGS_EVENT,
  getClickSpeech,
  getCompanionCandidate,
  readCompanionSpeechState
} from './companionSpeech';
import {
  CM_BRIDGE_ACTION_RESULT_EVENT,
  CM_BRIDGE_CONTEXT_EVENT,
  CM_BRIDGE_SELECTION_EVENT,
  dispatchCmAction,
  dispatchCmUndo,
  normalizeCmBridgeContext,
  normalizeCmSelection
} from './cmBridge';
import { cmResponseContract, parseCmActionProposal } from './cmActionProposal';
import './cm-penguin-companion.css';
import { DEFAULT_PET_ID, PET_SELECTION_EVENT, getPetDefinition, petAnimationDelay } from './petCatalog';
import { responsiveCompanionPosition } from './companion-position.js';
import { getPetPageService, subscribePetPageService } from './pageService.js';
import { LittleCompanionSprite } from './LittleCompanionSprite';
import {
  littleCompanionChooseAction,
  littleCompanionFatigueDelay,
  littleCompanionIdleDelay,
  littleCompanionInwardDistance,
  littleCompanionPeekDuration
} from './littleCompanionBehavior.js';
import {
  canLittleCompanionRoam,
  clampLittleCompanionToViewport,
  interpolateLittleCompanionPosition,
  littleCompanionEdgeSide,
  littleCompanionEdgeTarget,
  littleCompanionInwardTarget,
  littleCompanionPeekTarget,
  littleCompanionWalkDuration,
  littleCompanionWalkTarget
} from './littleCompanionRoaming.js';

const bubbleDurationMs = 7000;
const positionStorageKey = 'qiantie:cm-penguin-position';

const createLittleIdleAction = () => ({
  kind: 'idle',
  pose: 'idle',
  durationMs: littleCompanionIdleDelay(Math.random()),
  resetFatigue: false
});

function loadPosition() {
  try {
    const value = JSON.parse(localStorage.getItem(positionStorageKey) || '{}');
    if (Number.isFinite(value.left) && Number.isFinite(value.top)) {
      return responsiveCompanionPosition(value, window.innerWidth, window.innerHeight);
    }
  } catch {
    // Ignore restricted or malformed storage.
  }
  return { left: null, top: null };
}

function savePosition(position) {
  try { localStorage.setItem(positionStorageKey, JSON.stringify(position)); } catch { /* noop */ }
}

function clampPosition(left, top) {
  return responsiveCompanionPosition({ left, top }, window.innerWidth, window.innerHeight);
}

function isMissingTask(error) {
  const status = error?.status || error?.response?.status || error?.details?.status;
  return status === 404 || /\b404\b|not found|不存在/i.test(String(error?.message || error || ''));
}

function messageKey(message, index) {
  return `${message?.createdAt || index}-${message?.role || 'message'}`;
}

function actionCandidatePrompt(action) {
  const patch = action?.patch && typeof action.patch === 'object' ? action.patch : {};
  const payload = action?.payload && typeof action.payload === 'object' ? action.payload : {};
  const value = payload.value && typeof payload.value === 'object' ? payload.value : {};
  const preferred = [
    patch.外形, patch.appearance,
    patch.场景描述, patch.description, patch.场景, patch.scene,
    value.body, patch.body, payload.body
  ].find(item => typeof item === 'string' && item.trim());
  if (preferred) return preferred.trim().slice(0, 1800);
  const entries = Object.entries(patch)
    .filter(([, item]) => typeof item === 'string' && item.trim())
    .slice(0, 8)
    .map(([key, item]) => `${key}：${item}`);
  return entries.join('\n').slice(0, 1800);
}

export function CmPenguinCompanion({ username, accountSessionKey }) {
  const pageService = useSyncExternalStore(subscribePetPageService, getPetPageService, getPetPageService);
  const serviceIdentity = pageService ? JSON.stringify([pageService.id, pageService.scope, username, accountSessionKey]) : '';
  const [pet, setPet] = useState(() => getPetDefinition(DEFAULT_PET_ID));
  const [state, setState] = useState('idle');
  const [frame, setFrame] = useState(0);
  const [lookFrame, setLookFrame] = useState(null);
  const [position, setPosition] = useState(loadPosition);
  const [chatOpen, setChatOpen] = useState(() => Boolean(pageService?.openRequest));
  const [isDragging, setIsDragging] = useState(false);
  const [pageVisible, setPageVisible] = useState(() => document.visibilityState === 'visible');
  const [reducedMotion, setReducedMotion] = useState(() => window.matchMedia?.('(prefers-reduced-motion: reduce)').matches === true);
  const [viewport, setViewport] = useState(() => ({ width: window.innerWidth, height: window.innerHeight }));
  const [littleAction, setLittleAction] = useState(createLittleIdleAction);
  const [littleDirection, setLittleDirection] = useState(1);
  const [littleSpriteFailed, setLittleSpriteFailed] = useState(false);
  const [messages, setMessages] = useState([]);
  const [question, setQuestion] = useState('');
  const [asking, setAsking] = useState(false);
  const [reply, setReply] = useState('');
  const [companionSpeech, setCompanionSpeech] = useState(null);
  const [companionActive, setCompanionActive] = useState(() => readCompanionSpeechState(username).active);
  const [petTaskId, setPetTaskId] = useState(() => readCmTaskId(username) || null);
  const [bridgeContext, setBridgeContext] = useState(() => normalizeCmBridgeContext({ pagePath: window.location.pathname }));
  const [selection, setSelection] = useState(null);
  const [actionStatus, setActionStatus] = useState(null);
  const [lastUndoToken, setLastUndoToken] = useState('');
  const petContextRef = useRef(normalizePetContext({ pagePath: window.location.pathname }));
  const petTaskIdRef = useRef(petTaskId);
  const skillIdsRef = useRef([]);
  const spriteRef = useRef(null);
  const dragRef = useRef(null);
  const draggedRef = useRef(false);
  const positionRef = useRef(position);
  const roamingFrameRef = useRef(null);
  const littleActionTimerRef = useRef(null);
  const littleActionStartedAtRef = useRef(null);
  const littleActionDurationRef = useRef(0);
  const littleActiveMsRef = useRef(0);
  const littleFatigueMsRef = useRef(littleCompanionFatigueDelay(Math.random()));

  const updatePosition = useCallback(nextValue => {
    setPosition(current => {
      const next = typeof nextValue === 'function' ? nextValue(current) : nextValue;
      positionRef.current = next;
      return next;
    });
  }, []);

  const cancelRoaming = useCallback(() => {
    if (roamingFrameRef.current !== null) window.cancelAnimationFrame(roamingFrameRef.current);
    roamingFrameRef.current = null;
  }, []);

  const cancelLittleActionTimer = useCallback((countElapsed = true) => {
    if (littleActionTimerRef.current !== null) window.clearTimeout(littleActionTimerRef.current);
    if (countElapsed && littleActionStartedAtRef.current !== null) {
      const elapsed = Math.min(
        littleActionDurationRef.current,
        Math.max(0, Date.now() - littleActionStartedAtRef.current)
      );
      littleActiveMsRef.current += elapsed;
    }
    littleActionTimerRef.current = null;
    littleActionStartedAtRef.current = null;
    littleActionDurationRef.current = 0;
  }, []);

  const handleLittleLoadError = useCallback(failed => {
    setLittleSpriteFailed(failed === true);
  }, []);
  const historyRef = useRef(null);
  const speechRef = useRef(null);
  const speechTimerRef = useRef(null);
  const scheduleTimerRef = useRef(null);
  const clickSpeechIndexRef = useRef(-1);
  const requestRef = useRef(0);
  const serviceOpenRef = useRef({ identity: '', request: 0 });

  useEffect(() => {
    const previous = serviceOpenRef.current;
    const request = pageService?.openRequest || 0;
    if (previous.identity !== serviceIdentity) {
      requestRef.current += 1;
      setAsking(false);
      setReply('');
      setChatOpen(request > 0);
      clearCompanionSpeech();
    } else if (request > previous.request) {
      setChatOpen(true);
      clearCompanionSpeech();
    }
    serviceOpenRef.current = { identity: serviceIdentity, request };
  }, [serviceIdentity, pageService?.openRequest]);

  useEffect(() => {
    function handleResize() {
      const nextViewport = { width: window.innerWidth, height: window.innerHeight };
      cancelRoaming();
      cancelLittleActionTimer();
      setViewport(nextViewport);
      updatePosition(current => pet.id === 'little-companion'
        ? clampLittleCompanionToViewport(current, nextViewport)
        : responsiveCompanionPosition(current, nextViewport.width, nextViewport.height));
      if (pet.id === 'little-companion') setLittleAction(createLittleIdleAction());
    }
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [cancelLittleActionTimer, cancelRoaming, pet.id, updatePosition]);

  useEffect(() => {
    const handleVisibility = () => setPageVisible(document.visibilityState === 'visible');
    document.addEventListener('visibilitychange', handleVisibility);
    return () => document.removeEventListener('visibilitychange', handleVisibility);
  }, []);

  useEffect(() => {
    if (typeof window.matchMedia !== 'function') return undefined;
    const query = window.matchMedia('(prefers-reduced-motion: reduce)');
    const handleChange = () => setReducedMotion(query.matches);
    handleChange();
    if (typeof query.addEventListener === 'function') query.addEventListener('change', handleChange);
    else query.addListener?.(handleChange);
    return () => {
      if (typeof query.removeEventListener === 'function') query.removeEventListener('change', handleChange);
      else query.removeListener?.(handleChange);
    };
  }, []);

  useEffect(() => {
    cancelRoaming();
    cancelLittleActionTimer(false);
    littleActiveMsRef.current = 0;
    littleFatigueMsRef.current = littleCompanionFatigueDelay(Math.random());
    setLittleAction(createLittleIdleAction());
    setLittleDirection(1);
    setLittleSpriteFailed(false);
    if (pet.id !== 'little-companion') updatePosition(loadPosition());
    return () => {
      cancelRoaming();
      cancelLittleActionTimer(false);
    };
  }, [cancelLittleActionTimer, cancelRoaming, pet.id, updatePosition]);

  const littleCanAct = canLittleCompanionRoam({
    petId: pet.id,
    state,
    chatOpen,
    dragging: isDragging,
    pageVisible,
    reducedMotion,
    spriteFailed: littleSpriteFailed
  });

  useEffect(() => {
    if (!littleCanAct) {
      cancelLittleActionTimer();
      if (pet.id === 'little-companion' && state === 'idle' && littleAction.kind !== 'idle') {
        setLittleAction(createLittleIdleAction());
      }
      return undefined;
    }
    if (littleAction.kind === 'walk' || littleAction.kind === 'edge-walk' || littleAction.kind === 'inward-walk') {
      return undefined;
    }

    const durationMs = Math.max(0, Number(littleAction.durationMs) || 0);
    littleActionStartedAtRef.current = Date.now();
    littleActionDurationRef.current = durationMs;
    littleActionTimerRef.current = window.setTimeout(() => {
      littleActionTimerRef.current = null;
      littleActionStartedAtRef.current = null;
      littleActionDurationRef.current = 0;

      if (littleAction.resetFatigue) {
        littleActiveMsRef.current = 0;
        littleFatigueMsRef.current = littleCompanionFatigueDelay(Math.random());
      } else {
        littleActiveMsRef.current += durationMs;
      }

      if (littleAction.kind === 'peek') {
        setLittleAction({
          kind: 'inward-walk',
          pose: 'walk',
          durationMs: 0,
          peekSide: littleAction.peekSide,
          inwardDistance: littleCompanionInwardDistance(Math.random()),
          resetFatigue: false
        });
        return;
      }

      if (littleAction.kind === 'idle') {
        setLittleAction(littleCompanionChooseAction({
          activeMs: littleActiveMsRef.current,
          fatigueMs: littleFatigueMsRef.current,
          actionRoll: Math.random(),
          durationRoll: Math.random()
        }));
        return;
      }

      setLittleAction(createLittleIdleAction());
    }, durationMs);

    return () => cancelLittleActionTimer();
  }, [cancelLittleActionTimer, littleAction, littleCanAct, pet.id, state]);

  useEffect(() => {
    const moving = littleAction.kind === 'walk'
      || littleAction.kind === 'edge-walk'
      || littleAction.kind === 'inward-walk';
    if (!littleCanAct || !moving) {
      cancelRoaming();
      return undefined;
    }

    cancelRoaming();
    const rect = spriteRef.current?.getBoundingClientRect();
    const current = positionRef.current;
    const start = Number.isFinite(current?.left) && Number.isFinite(current?.top)
      ? current
      : {
          left: Number.isFinite(rect?.left) ? rect.left : viewport.width - 142,
          top: Number.isFinite(rect?.top) ? rect.top : viewport.height - 148
        };
    const target = littleAction.kind === 'edge-walk'
      ? littleCompanionPeekTarget(start, viewport)
      : littleAction.kind === 'inward-walk'
        ? littleCompanionInwardTarget(start, viewport, littleAction.peekSide, littleAction.inwardDistance)
        : littleCompanionWalkTarget(start, viewport, Math.random(), Math.random());
    const durationMs = littleCompanionWalkDuration(start, target, Math.random());

    positionRef.current = start;
    setLittleDirection(target.direction || (target.left >= start.left ? 1 : -1));

    let startedAt = null;
    function move(timestamp) {
      if (startedAt === null) startedAt = timestamp;
      const progress = durationMs > 0 ? (timestamp - startedAt) / durationMs : 1;
      updatePosition(interpolateLittleCompanionPosition(start, target, progress));
      if (progress < 1) {
        roamingFrameRef.current = window.requestAnimationFrame(move);
        return;
      }

      roamingFrameRef.current = null;
      littleActiveMsRef.current += durationMs;
      if (littleAction.kind === 'inward-walk') {
        setLittleAction(createLittleIdleAction());
        return;
      }

      const edgeSide = littleCompanionEdgeSide(target, viewport);
      if (edgeSide) {
        updatePosition(littleCompanionEdgeTarget(target, viewport, edgeSide));
        setLittleAction({
          kind: 'peek',
          pose: 'peek',
          durationMs: littleCompanionPeekDuration(Math.random()),
          peekSide: edgeSide,
          resetFatigue: false
        });
        return;
      }

      setLittleAction(createLittleIdleAction());
    }
    roamingFrameRef.current = window.requestAnimationFrame(move);
    return cancelRoaming;
  }, [
    cancelRoaming,
    littleAction,
    littleCanAct,
    updatePosition,
    viewport
  ]);

  const effectiveSelection = selection || bridgeContext.selection;
  const contextLabel = effectiveSelection?.label || bridgeContext.page || petContextRef.current.page || '当前页面';
  const canApply = bridgeContext.canApply && bridgeContext.capabilities.length > 0;

  function clearSpeechTimer() {
    if (speechTimerRef.current) window.clearTimeout(speechTimerRef.current);
    speechTimerRef.current = null;
  }

  function clearScheduleTimer() {
    if (scheduleTimerRef.current) window.clearTimeout(scheduleTimerRef.current);
    scheduleTimerRef.current = null;
  }

  function clearCompanionSpeech() {
    clearSpeechTimer();
    speechRef.current = null;
    setCompanionSpeech(null);
  }

  function showSpeech(candidate) {
    if (!candidate?.text || reply) return false;
    if (speechRef.current && candidate.priority < speechRef.current.priority) return false;
    clearSpeechTimer();
    speechRef.current = candidate;
    setCompanionSpeech(candidate);
    if (candidate.priority <= COMPANION_SPEECH_PRIORITY.click) {
      speechTimerRef.current = window.setTimeout(() => {
        speechTimerRef.current = null;
        speechRef.current = null;
        setCompanionSpeech(null);
      }, bubbleDurationMs);
    }
    return true;
  }

  function clearConversation(messageText = '') {
    petTaskIdRef.current = null;
    setPetTaskId(null);
    writeCmTaskId(username, '');
    setMessages([]);
    setActionStatus(null);
    if (messageText) setReply(messageText);
  }

  useEffect(() => {
    requestRef.current += 1;
    const storedTaskId = readCmTaskId(username) || null;
    petTaskIdRef.current = storedTaskId;
    setPetTaskId(storedTaskId);
    setMessages([]);
    setQuestion('');
    setReply('');
    setSelection(null);
    setActionStatus(null);
    setLastUndoToken('');
    setCompanionActive(readCompanionSpeechState(username).active);
    clearScheduleTimer();
    clearCompanionSpeech();
    return () => {
      requestRef.current += 1;
      clearScheduleTimer();
      clearCompanionSpeech();
    };
  }, [username, accountSessionKey]);

  useEffect(() => {
    let active = true;
    getConfig().then(config => {
      if (active) setPet(getPetDefinition(config.pet));
    }).catch(() => {});
    function handlePetSelection(event) {
      if (event.detail?.username && event.detail.username !== username) return;
      setPet(getPetDefinition(event.detail?.pet || event.detail?.petId));
    }
    window.addEventListener(PET_SELECTION_EVENT, handlePetSelection);
    return () => {
      active = false;
      window.removeEventListener(PET_SELECTION_EVENT, handlePetSelection);
    };
  }, [username, accountSessionKey]);

  useEffect(() => {
    function handleState(event) {
      const next = normalizePetState(event.detail?.state);
      if (next !== 'idle') clearCompanionSpeech();
      setState(next);
      setReply('');
      setLookFrame(null);
    }
    window.addEventListener(PET_EVENT, handleState);
    return () => window.removeEventListener(PET_EVENT, handleState);
  }, []);

  useEffect(() => {
    function handleSkills(event) {
      skillIdsRef.current = Array.isArray(event.detail?.skillIds) ? event.detail.skillIds : [];
    }
    window.addEventListener(PET_SKILLS_EVENT, handleSkills);
    return () => window.removeEventListener(PET_SKILLS_EVENT, handleSkills);
  }, []);

  useEffect(() => {
    function handleContext(event) {
      petContextRef.current = normalizePetContext(event.detail);
      const raw = event.detail && typeof event.detail === 'object' ? event.detail : {};
      setBridgeContext(current => normalizeCmBridgeContext({
        ...current,
        page: raw.page || current.page,
        pagePath: raw.pagePath || current.pagePath,
        summary: raw.summary || current.summary,
        selection: raw.selection || current.selection,
        capabilities: raw.capabilities || current.capabilities,
        canApply: typeof raw.canApply === 'boolean' ? raw.canApply : current.canApply
      }));
      if (raw.selection) setSelection(normalizeCmSelection(raw.selection));
    }
    function handleBridge(event) {
      setBridgeContext(normalizeCmBridgeContext(event.detail));
      if (event.detail?.selection) setSelection(normalizeCmSelection(event.detail.selection));
    }
    function handleSelection(event) {
      setSelection(normalizeCmSelection(event.detail));
    }
    window.addEventListener(PET_CONTEXT_EVENT, handleContext);
    window.addEventListener(CM_BRIDGE_CONTEXT_EVENT, handleBridge);
    window.addEventListener(CM_BRIDGE_SELECTION_EVENT, handleSelection);
    return () => {
      window.removeEventListener(PET_CONTEXT_EVENT, handleContext);
      window.removeEventListener(CM_BRIDGE_CONTEXT_EVENT, handleBridge);
      window.removeEventListener(CM_BRIDGE_SELECTION_EVENT, handleSelection);
    };
  }, []);

  useEffect(() => {
    function handleSettings(event) {
      if (event.detail?.username !== username) return;
      const active = event.detail?.active === true;
      setCompanionActive(active);
      if (!active) {
        clearScheduleTimer();
        clearCompanionSpeech();
      }
    }
    window.addEventListener(PET_COMPANION_SETTINGS_EVENT, handleSettings);
    return () => window.removeEventListener(PET_COMPANION_SETTINGS_EVENT, handleSettings);
  }, [username]);

  useEffect(() => {
    function handleActionResult(event) {
      const result = event.detail || {};
      setActionStatus(result.ok ? result.message || '修改已应用。' : result.message || '这次修改没有应用。');
      if (result.ok && result.undoToken) setLastUndoToken(result.undoToken);
      if (result.ok) dispatchPetState('success');
      else dispatchPetState('error');
    }
    window.addEventListener(CM_BRIDGE_ACTION_RESULT_EVENT, handleActionResult);
    return () => window.removeEventListener(CM_BRIDGE_ACTION_RESULT_EVENT, handleActionResult);
  }, []);

  useEffect(() => {
    if (state !== 'idle') {
      clearScheduleTimer();
      return undefined;
    }

    function scheduleSpeech() {
      clearScheduleTimer();
      if (!companionActive || chatOpen || asking || dragRef.current || document.visibilityState !== 'visible') return;
      const stored = readCompanionSpeechState(username);
      const candidate = getCompanionCandidate({
        now: new Date(),
        username,
        storage: window.localStorage,
        active: stored.active,
        visible: true,
        chatOpen: false,
        asking: false,
        dragging: false
      });
      if (candidate) showSpeech(candidate);
      const next = readCompanionSpeechState(username);
      if (!next.active || !Number.isFinite(next.nextIdleAt)) return;
      scheduleTimerRef.current = window.setTimeout(scheduleSpeech, Math.max(1000, next.nextIdleAt - Date.now()));
    }

    function handleVisibilityChange() {
      if (document.visibilityState !== 'visible') {
        getCompanionCandidate({
          now: new Date(),
          username,
          storage: window.localStorage,
          active: companionActive,
          visible: false,
          chatOpen,
          asking,
          dragging: Boolean(dragRef.current)
        });
        clearScheduleTimer();
        clearCompanionSpeech();
        return;
      }
      scheduleSpeech();
    }

    scheduleSpeech();
    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      clearScheduleTimer();
    };
  }, [username, accountSessionKey, state, companionActive, chatOpen, asking]);

  useEffect(() => {
    if (state === 'idle' || state === 'working') return undefined;
    const id = window.setTimeout(() => {
      setState('idle');
      setReply('');
    }, 2400);
    return () => window.clearTimeout(id);
  }, [state]);

  useEffect(() => {
    const delay = petAnimationDelay(pet, state);
    setFrame(0);
    const id = window.setInterval(() => setFrame(current => (current + 1) % petFrameCount(state)), delay);
    return () => window.clearInterval(id);
  }, [pet, state]);

  useEffect(() => {
    function followPointer(event) {
      if (state !== 'idle' || !spriteRef.current || dragRef.current) return;
      const rect = spriteRef.current.getBoundingClientRect();
      setLookFrame(petLookFrame(event.clientX - rect.left - rect.width / 2, event.clientY - rect.top - rect.height / 2));
    }
    window.addEventListener('pointermove', followPointer);
    return () => window.removeEventListener('pointermove', followPointer);
  }, [state]);

  useEffect(() => {
    if (!chatOpen || !historyRef.current) return;
    historyRef.current.scrollTop = historyRef.current.scrollHeight;
  }, [chatOpen, messages, asking]);

  useEffect(() => {
    function handleEscape(event) {
      if (event.key === 'Escape' && chatOpen) closeChat();
    }
    window.addEventListener('keydown', handleEscape);
    return () => window.removeEventListener('keydown', handleEscape);
  }, [chatOpen]);

  async function loadTask(taskId) {
    const requestId = ++requestRef.current;
    try {
      const result = await getAgentTask(taskId);
      if (requestRef.current !== requestId || petTaskIdRef.current !== taskId) return;
      setMessages(Array.isArray(result.task?.messages) ? result.task.messages : []);
    } catch (error) {
      if (requestRef.current !== requestId) return;
      if (isMissingTask(error)) clearConversation('旧对话已经失效，我给你开一个新的。');
      else setReply('聊天记录暂时没读到。');
    }
  }

  async function openChat() {
    setChatOpen(true);
    clearCompanionSpeech();
    if (pageService) return;
    if (petTaskIdRef.current) await loadTask(petTaskIdRef.current);
  }

  function closeChat() {
    requestRef.current += 1;
    clearCompanionSpeech();
    setChatOpen(false);
    setAsking(false);
    spriteRef.current?.focus();
  }

  async function ensureTask(requestId) {
    if (petTaskIdRef.current) return petTaskIdRef.current;
    const result = await createAgentTask();
    if (requestRef.current !== requestId) return null;
    const taskId = result.task?.id;
    if (!taskId) throw new Error(`未能创建 ${pet.displayName} 对话`);
    petTaskIdRef.current = taskId;
    setPetTaskId(taskId);
    writeCmTaskId(username, taskId);
    setMessages(Array.isArray(result.task?.messages) ? result.task.messages : []);
    return taskId;
  }

  async function sendQuestion(prompt) {
    if (pageService) return;
    const content = String(prompt || '').trim();
    if (!content || asking) return;
    const requestId = ++requestRef.current;
    setAsking(true);
    setChatOpen(true);
    setActionStatus(null);
    clearCompanionSpeech();
    dispatchPetState('working');
    try {
      const taskId = await ensureTask(requestId);
      if (!taskId || requestRef.current !== requestId) return;
      const result = await askAgent({
        taskId,
        prompt: content,
        context: {
          ...petContextRef.current,
          page: bridgeContext.page || petContextRef.current.page || window.location.pathname,
          pagePath: bridgeContext.pagePath || petContextRef.current.pagePath || window.location.pathname,
          cmSelection: effectiveSelection,
          cmCapabilities: bridgeContext.capabilities,
          cmResponseContract: cmResponseContract(bridgeContext.capabilities)
        },
        skillIds: skillIdsRef.current,
        suppressGlobalError: true
      });
      if (requestRef.current !== requestId) return;
      const nextTaskId = result.task?.id || taskId;
      petTaskIdRef.current = nextTaskId;
      setPetTaskId(nextTaskId);
      writeCmTaskId(username, nextTaskId);
      setMessages(Array.isArray(result.task?.messages)
        ? result.task.messages
        : current => [...current, result.user, result.assistant].filter(Boolean));
      setQuestion(current => current === content ? '' : current);
      setReply('弄好啦。');
      dispatchPetState('success');
    } catch (error) {
      if (requestRef.current !== requestId) return;
      if (isMissingTask(error)) clearConversation('这段对话失效了，我给你重新开一个。');
      const recovery = classifyPetRequestError(error);
      setReply(recovery.message);
      dispatchPetState('error');
    } finally {
      if (requestRef.current === requestId) setAsking(false);
    }
  }

  function handlePetClick() {
    if (draggedRef.current) {
      draggedRef.current = false;
      return;
    }
    setLookFrame(null);
    if (!chatOpen && companionActive) {
      const text = getClickSpeech(clickSpeechIndexRef.current);
      clickSpeechIndexRef.current = ['戳我干嘛，我有在认真陪你。', '再摸一下也不是不行。', '哼，注意力被你拿走啦。'].indexOf(text);
      showSpeech({ text, priority: COMPANION_SPEECH_PRIORITY.click });
    }
    if (pet.id === 'little-companion') {
      if (!chatOpen) {
        cancelRoaming();
        cancelLittleActionTimer();
        setLittleAction({ kind: 'dance', pose: 'dance', durationMs: 1_220, resetFatigue: false });
      }
      return;
    }
    void openChat();
  }

  function beginDrag(event) {
    if (event.button !== 0) return;
    const rect = spriteRef.current?.getBoundingClientRect();
    if (!rect) return;
    cancelRoaming();
    cancelLittleActionTimer();
    setIsDragging(true);
    clearScheduleTimer();
    dragRef.current = { pointerId: event.pointerId, startX: event.clientX, startY: event.clientY, left: rect.left, top: rect.top };
    draggedRef.current = false;
    event.currentTarget.setPointerCapture?.(event.pointerId);
  }

  function moveDrag(event) {
    const drag = dragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    const dx = event.clientX - drag.startX;
    const dy = event.clientY - drag.startY;
    if (Math.hypot(dx, dy) > 8) draggedRef.current = true;
    if (!draggedRef.current) return;
    const next = clampPosition(drag.left + dx, drag.top + dy);
    updatePosition(next);
  }

  function endDrag(event) {
    const drag = dragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    dragRef.current = null;
    if (draggedRef.current) {
      const manualPosition = positionRef.current;
      if (pet.id === 'little-companion' && Number.isFinite(manualPosition?.left)) {
        const edgeSide = littleCompanionEdgeSide(manualPosition, viewport);
        if (edgeSide) {
          const edgePosition = littleCompanionEdgeTarget(manualPosition, viewport, edgeSide);
          updatePosition(edgePosition);
          savePosition(edgePosition);
          setLittleAction({
            kind: 'peek',
            pose: 'peek',
            durationMs: littleCompanionPeekDuration(Math.random()),
            peekSide: edgeSide,
            resetFatigue: false
          });
        } else {
          savePosition(manualPosition);
          setLittleAction(createLittleIdleAction());
        }
      } else {
        savePosition(manualPosition);
      }
    }
    setIsDragging(false);
    event.currentTarget.releasePointerCapture?.(event.pointerId);
  }

  function openAgent() {
    if (!petTaskId) return;
    window.location.href = `/agent?task=${encodeURIComponent(petTaskId)}`;
  }

  function analyzePage() {
    const subject = effectiveSelection?.label ? `，重点看我当前选中的「${effectiveSelection.label}」` : '';
    void sendQuestion(`请分析当前页面${subject}，告诉我最值得处理的问题；如果适合直接修改，请给我可确认的修改方案。`);
  }

  function applyProposal(proposal) {
    if (!proposal?.actions?.length || !canApply) return;
    setActionStatus('正在交给当前功能区应用…');
    dispatchPetState('working');
    proposal.actions.forEach(action => dispatchCmAction(action, {
      pagePath: bridgeContext.pagePath,
      selection: effectiveSelection,
      taskId: petTaskId
    }));
  }

  const spriteRow = lookFrame && state === 'idle' ? lookFrame.row : petAtlasRow(state);
  const spriteFrame = lookFrame && state === 'idle' ? lookFrame.column : frame;
  const shellStyle = position.left === null
    ? undefined
    : { left: position.left, top: position.top, right: 'auto', bottom: 'auto' };
  const visibleBubble = reply || petSpeech(state) || companionSpeech?.text;
  const renderedMessages = useMemo(() => messages.map((message, index) => {
    const parsed = message.role === 'assistant' ? parseCmActionProposal(message.content) : { text: String(message.content || ''), proposal: null };
    const revision = message.role === 'assistant' && petContextRef.current.scriptOutput ? parseScriptRevision(message.content) : null;
    return { ...message, key: messageKey(message, index), displayText: parsed.text, proposal: parsed.proposal, revision };
  }), [messages]);

  return (
    <div className="cm-penguin-shell" style={shellStyle} aria-live="polite">
      {!chatOpen && visibleBubble ? <div className="cm-penguin-speech">{visibleBubble}</div> : null}

      {pageService ? (
        <section className="cm-penguin-chat cm-penguin-page-service" role="dialog" aria-label={`${pet.displayName} · ${pageService.label}`} hidden={!chatOpen}>
          <header className="cm-penguin-chat-header">
            <div>
              <span className="cm-penguin-status-dot" aria-hidden="true" />
              <strong>{pet.displayName}</strong>
              <span className="cm-penguin-context-chip" title="当前页面服务">{pageService.label}</span>
            </div>
            <div className="cm-penguin-header-actions">
              <button type="button" title="关闭" aria-label={`关闭 ${pet.displayName} 对话`} onClick={closeChat}><X size={16} /></button>
            </div>
          </header>
          <div className="cm-penguin-page-service-body" key={serviceIdentity}>
            {pageService.render()}
          </div>
        </section>
      ) : chatOpen ? (
        <section className="cm-penguin-chat" role="dialog" aria-label={`和 ${pet.displayName} 对话`}>
          <header className="cm-penguin-chat-header">
            <div>
              <span className="cm-penguin-status-dot" aria-hidden="true" />
              <strong>{pet.displayName}</strong>
              <span className="cm-penguin-context-chip" title={effectiveSelection ? '当前选中对象' : '当前页面'}>{contextLabel}</span>
            </div>
            <div className="cm-penguin-header-actions">
              <button type="button" title="分析当前内容" aria-label="分析当前内容" onClick={analyzePage} disabled={asking}><ScanSearch size={16} /></button>
              <button type="button" title="在 CM 画布继续" aria-label="在 CM 画布继续" onClick={openAgent} disabled={!petTaskId}><ExternalLink size={16} /></button>
              <button type="button" title="关闭" aria-label={`关闭 ${pet.displayName} 对话`} onClick={closeChat}><X size={16} /></button>
            </div>
          </header>

          <div className="cm-penguin-chat-history" ref={historyRef}>
            {!renderedMessages.length ? (
              <div className="cm-penguin-empty">
                <Sparkles size={18} />
                <strong>{effectiveSelection?.label ? `我知道你正在看「${effectiveSelection.label}」` : '我在这儿。'}</strong>
                <span>不懂就直接问；想改也可以直接说你的想法。</span>
                <button type="button" onClick={analyzePage}>帮我看看当前内容</button>
              </div>
            ) : renderedMessages.map(message => (
              <article className={`cm-penguin-message is-${message.role}`} key={message.key}>
                <span>{message.displayText}</span>
                {message.revision ? <button className="cm-penguin-inline-action" type="button" onClick={() => dispatchPetPreview(message.revision)}>查看剧本修改</button> : null}
                {message.proposal ? (
                  <div className="cm-penguin-proposal">
                    <strong>{message.proposal.summary || `${pet.displayName} 准备了修改`}</strong>
                    <div className="cm-penguin-proposal-list">
                      {message.proposal.actions.map((action, index) => <span key={`${action.type}-${action.targetId}-${index}`}>{action.label || action.type}</span>)}
                    </div>
                    {message.proposal.actions.map((action, index) => {
                      const candidate = actionCandidatePrompt(action);
                      return candidate ? <div className="cm-penguin-proposal-preview" key={`preview-${action.type}-${action.targetId}-${index}`}>
                        <span>候选提示词</span>
                        <pre>{candidate}</pre>
                      </div> : null;
                    })}
                    <button type="button" disabled={!canApply || asking} title={canApply ? '确认后应用这份候选提示词' : '当前功能区还没有接入直接应用'} onClick={() => applyProposal(message.proposal)}>
                      {canApply ? '应用这份提示词' : '等待功能区接入'}
                    </button>
                  </div>
                ) : null}
              </article>
            ))}
            {asking ? <div className="cm-penguin-thinking"><i /><i /><i /><span>{pet.displayName} 正在想…</span></div> : null}
            {actionStatus ? <div className="cm-penguin-action-status">{actionStatus}{lastUndoToken ? <button type="button" onClick={() => dispatchCmUndo(lastUndoToken)}><Undo2 size={14} />撤销</button> : null}</div> : null}
          </div>

          <form className="cm-penguin-composer" onSubmit={event => { event.preventDefault(); void sendQuestion(question); }}>
                <input value={question} onChange={event => setQuestion(event.target.value)} placeholder={effectiveSelection?.label ? `问问 ${pet.displayName} 关于「${effectiveSelection.label}」…` : `问问 ${pet.displayName}…`} aria-label={`向 ${pet.displayName} 提问`} />
            <button type="submit" disabled={!question.trim() || asking} aria-label="发送">↗</button>
          </form>
        </section>
      ) : null}

      {pet.id === 'little-companion' && !chatOpen ? (
        <button
          className="cm-penguin-chat-trigger"
          type="button"
          title={`和 ${pet.displayName} 聊天`}
          aria-label={`和 ${pet.displayName} 聊天`}
          onClick={event => { event.stopPropagation(); void openChat(); }}
        >
          <MessageCircle size={15} />
        </button>
      ) : null}

      <div
        ref={spriteRef}
        className={`cm-penguin-pet is-${state}${pet.renderer === 'multi-sheet' ? ' is-little-companion' : ''}`}
        style={{ '--cm-row': spriteRow, '--cm-frame': spriteFrame }}
        role="button"
        tabIndex={0}
        aria-label={`${pet.displayName} 宠物，${state}`}
        onPointerDown={beginDrag}
        onPointerMove={moveDrag}
        onPointerUp={endDrag}
        onPointerCancel={endDrag}
        onClick={handlePetClick}
        onKeyDown={event => {
          if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            handlePetClick();
          }
        }}
      >
        {pet.renderer === 'multi-sheet'
          ? <LittleCompanionSprite
              state={state}
              label={pet.displayName}
              direction={littleDirection}
              idleStep={{ pose: littleAction.pose, durationMs: littleAction.durationMs }}
              peekSide={littleAction.kind === 'peek' ? littleAction.peekSide : null}
              onLoadError={handleLittleLoadError}
            />
          : <img className="cm-penguin-atlas" src={pet.spritesheetPath} alt="" draggable="false" style={{ imageRendering: pet.renderMode === 'smooth' ? 'auto' : undefined }} />}
      </div>
    </div>
  );
}

export default CmPenguinCompanion;
