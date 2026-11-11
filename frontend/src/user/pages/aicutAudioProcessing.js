const channelModes = new Set(['stereo', 'mono', 'left', 'right', 'swap']);
export const audioEffectOptions = Object.freeze([
  { value: 'natural', label: '原声' },
  { value: 'enhance', label: '人声增强' },
  { value: 'bright', label: '清亮' },
  { value: 'bass', label: '低音增强' },
  { value: 'treble', label: '高音增强' },
  { value: 'telephone', label: '电话声' },
  { value: 'radio', label: '收音机' },
  { value: 'megaphone', label: '扩音器' },
  { value: 'robot', label: '机器人声' },
  { value: 'chipmunk', label: '花栗鼠声' },
  { value: 'deep', label: '低沉声' },
  { value: 'sweet', label: '甜美声' },
  { value: 'whisper', label: '耳语声' },
  { value: 'distorted', label: '失真声' },
  { value: 'lofi', label: '复古低保真' },
  { value: 'echo', label: '回声' },
  { value: 'room', label: '房间混响' },
  { value: 'hall', label: '大厅混响' },
  { value: 'plate', label: '板式混响' },
  { value: 'cathedral', label: '教堂混响' },
  { value: 'stadium', label: '体育场混响' },
  { value: 'underwater', label: '水下声' },
  { value: 'vinyl', label: '黑胶唱片' },
]);
export const audioEffectValues = new Set(audioEffectOptions.map(item => item.value));

function clamp(value, minimum, maximum, fallback = 0) {
  const number = Number(value);
  if (!Number.isFinite(number)) return fallback;
  return Math.min(maximum, Math.max(minimum, number));
}

export function normalizeAudioProcessing(value = {}) {
  const channelMode = String(value?.channelMode || 'stereo');
  const requestedTargetLUFS = Number(value?.targetLUFS);
  const targetLUFS = Number.isFinite(requestedTargetLUFS) && requestedTargetLUFS >= -24 && requestedTargetLUFS <= -8 ? requestedTargetLUFS : -16;
  const requestedVoiceEffect = String(value?.voiceEffect || 'natural');
  return {
    denoise: clamp(value?.denoise, 0, 1),
    eqLowDB: clamp(value?.eqLowDB, -12, 12),
    eqMidDB: clamp(value?.eqMidDB, -12, 12),
    eqHighDB: clamp(value?.eqHighDB, -12, 12),
    channelMode: channelModes.has(channelMode) ? channelMode : 'stereo',
    loudnessEnabled: Boolean(value?.loudnessEnabled),
    targetLUFS,
    voiceEffect: audioEffectValues.has(requestedVoiceEffect) ? requestedVoiceEffect : 'natural',
    pitchSemitones: clamp(value?.pitchSemitones, -12, 12)
  };
}

export function audioProcessingPreviewPlan(value = {}) {
  const processing = normalizeAudioProcessing(value);
  let highpassHz = Math.round(40 + processing.denoise * 180);
  let lowpassHz = Math.round(20000 - processing.denoise * 6000);
  let voicePresenceDB = 0;
  let distortionAmount = 0;
  let modulationHz = 0;
  let modulationDepthMS = 0;
  let echoDelayMS = 0;
  let echoFeedback = 0;
  let echoWet = 0;
  let ambienceDelayMS = 0;
  let ambienceFeedback = 0;
  let ambienceWet = 0;
  let voiceCompressor = false;
  if (processing.voiceEffect === 'enhance') {
    highpassHz = Math.max(highpassHz, 80);
    lowpassHz = Math.min(lowpassHz, 14000);
    voicePresenceDB = 4;
    voiceCompressor = true;
  } else if (processing.voiceEffect === 'telephone') {
    highpassHz = 300;
    lowpassHz = 3400;
    voicePresenceDB = 2;
    voiceCompressor = true;
  } else if (processing.voiceEffect === 'robot') {
    highpassHz = 120;
    lowpassHz = 6500;
    distortionAmount = 40;
    modulationHz = 35;
    modulationDepthMS = 1.5;
  } else if (processing.voiceEffect === 'chipmunk') {
    highpassHz = 160;
    lowpassHz = 16000;
    voicePresenceDB = 7;
    modulationHz = 12;
    modulationDepthMS = 4;
  } else if (processing.voiceEffect === 'bright') {
    voicePresenceDB = 3;
    processing.eqHighDB = 4;
  } else if (processing.voiceEffect === 'bass') {
    processing.eqLowDB = 8;
    lowpassHz = 15000;
  } else if (processing.voiceEffect === 'treble') {
    processing.eqHighDB = 7;
    highpassHz = 90;
  } else if (processing.voiceEffect === 'radio') {
    highpassHz = 250;
    lowpassHz = 4500;
    distortionAmount = 22;
    voiceCompressor = true;
  } else if (processing.voiceEffect === 'megaphone') {
    highpassHz = 400;
    lowpassHz = 5000;
    distortionAmount = 36;
    voicePresenceDB = 4;
    voiceCompressor = true;
  } else if (processing.voiceEffect === 'deep') {
    processing.eqLowDB = 7;
    lowpassHz = 10000;
    modulationHz = 5;
    modulationDepthMS = 2;
  } else if (processing.voiceEffect === 'sweet') {
    voicePresenceDB = 5;
    processing.eqHighDB = 3;
    voiceCompressor = true;
  } else if (processing.voiceEffect === 'whisper') {
    highpassHz = 500;
    lowpassHz = 12500;
    processing.eqHighDB = 6;
    voiceCompressor = true;
  } else if (processing.voiceEffect === 'distorted') {
    distortionAmount = 80;
  } else if (processing.voiceEffect === 'lofi') {
    highpassHz = 120;
    lowpassHz = 5200;
    distortionAmount = 14;
  } else if (processing.voiceEffect === 'echo') {
    echoDelayMS = 220;
    echoFeedback = .35;
    echoWet = .4;
  } else if (processing.voiceEffect === 'room') {
    ambienceDelayMS = 28;
    ambienceFeedback = .18;
    ambienceWet = .22;
  } else if (processing.voiceEffect === 'hall') {
    ambienceDelayMS = 55;
    ambienceFeedback = .28;
    ambienceWet = .32;
  } else if (processing.voiceEffect === 'plate') {
    ambienceDelayMS = 38;
    ambienceFeedback = .24;
    ambienceWet = .28;
    processing.eqHighDB = 2;
  } else if (processing.voiceEffect === 'cathedral') {
    ambienceDelayMS = 95;
    ambienceFeedback = .42;
    ambienceWet = .45;
  } else if (processing.voiceEffect === 'stadium') {
    ambienceDelayMS = 125;
    ambienceFeedback = .34;
    ambienceWet = .4;
  } else if (processing.voiceEffect === 'underwater') {
    highpassHz = 40;
    lowpassHz = 1100;
    modulationHz = 1.2;
    modulationDepthMS = 8;
  } else if (processing.voiceEffect === 'vinyl') {
    highpassHz = 120;
    lowpassHz = 7800;
    distortionAmount = 8;
  }
  const compressed = processing.loudnessEnabled || voiceCompressor;
  const pitchRatio = 2 ** (processing.pitchSemitones / 12);
  return {
    highpassHz,
    lowpassHz,
    eqLowDB: processing.eqLowDB,
    eqMidDB: processing.eqMidDB,
    eqHighDB: processing.eqHighDB,
    channelMode: processing.channelMode,
    loudnessEnabled: processing.loudnessEnabled,
    targetLUFS: processing.targetLUFS,
    compressorThresholdDB: processing.loudnessEnabled ? -24 : (voiceCompressor ? -22 : 0),
    compressorRatio: compressed ? 3 : 1,
    makeupGainDB: processing.loudnessEnabled ? clamp(processing.targetLUFS + 16, -8, 8) : (voiceCompressor ? 2 : 0),
    voiceEffect: processing.voiceEffect,
    voicePresenceDB,
    distortionAmount,
    modulationHz,
    modulationDepthMS,
    echoDelayMS,
    echoFeedback,
    echoWet,
    ambienceDelayMS,
    ambienceFeedback,
    ambienceWet,
    pitchSemitones: processing.pitchSemitones,
    pitchRatio
  };
}

function distortionCurve(amount = 0, samples = 256) {
  const curve = new Float32Array(samples);
  const strength = Math.max(0, Number(amount) || 0);
  for (let index = 0; index < samples; index += 1) {
    const x = index * 2 / (samples - 1) - 1;
    curve[index] = strength === 0 ? x : ((3 + strength) * x * 20 * Math.PI / 180) / (Math.PI + strength * Math.abs(x));
  }
  return curve;
}

function connectChannelMode(context, input, mode) {
  if (mode === 'stereo') {
    input.connect(context.destination);
    return [];
  }
  const splitter = context.createChannelSplitter(2);
  const merger = context.createChannelMerger(2);
  input.connect(splitter);
  if (mode === 'mono') {
    const left = context.createGain();
    const right = context.createGain();
    left.gain.value = .5;
    right.gain.value = .5;
    splitter.connect(left, 0);
    splitter.connect(right, 1);
    left.connect(merger, 0, 0);
    left.connect(merger, 0, 1);
    right.connect(merger, 0, 0);
    right.connect(merger, 0, 1);
    merger.connect(context.destination);
    return [splitter, left, right, merger];
  }
  const sourceLeft = mode === 'right' || mode === 'swap' ? 1 : 0;
  const sourceRight = mode === 'left' || mode === 'swap' ? 0 : 1;
  splitter.connect(merger, sourceLeft, 0);
  splitter.connect(merger, sourceRight, 1);
  merger.connect(context.destination);
  return [splitter, merger];
}

export function createAudioProcessingGraph(mediaElement, initialValue = {}) {
  const AudioContextClass = globalThis.AudioContext || globalThis.webkitAudioContext;
  if (!AudioContextClass || !mediaElement) return null;
  const context = new AudioContextClass({ latencyHint: 'interactive' });
  const source = context.createMediaElementSource(mediaElement);
  const highpass = context.createBiquadFilter();
  const lowpass = context.createBiquadFilter();
  const lowEQ = context.createBiquadFilter();
  const midEQ = context.createBiquadFilter();
  const highEQ = context.createBiquadFilter();
  const voicePresence = context.createBiquadFilter();
  const distortion = context.createWaveShaper();
  const modulationDelay = context.createDelay(.05);
  const modulationOscillator = context.createOscillator();
  const modulationGain = context.createGain();
  const dryGain = context.createGain();
  const echoDelay = context.createDelay(.5);
  const echoFeedback = context.createGain();
  const echoWet = context.createGain();
  const ambienceDelay = context.createDelay(.25);
  const ambienceFeedback = context.createGain();
  const ambienceWet = context.createGain();
  const effectMix = context.createGain();
  const compressor = context.createDynamicsCompressor();
  const makeupGain = context.createGain();
  highpass.type = 'highpass';
  lowpass.type = 'lowpass';
  lowEQ.type = 'lowshelf';
  lowEQ.frequency.value = 120;
  midEQ.type = 'peaking';
  midEQ.frequency.value = 1000;
  midEQ.Q.value = 1;
  highEQ.type = 'highshelf';
  highEQ.frequency.value = 8000;
  voicePresence.type = 'peaking';
  voicePresence.frequency.value = 2500;
  voicePresence.Q.value = .8;
  modulationOscillator.type = 'sine';
  modulationOscillator.connect(modulationGain).connect(modulationDelay.delayTime);
  modulationOscillator.start();
  source.connect(highpass).connect(lowpass).connect(lowEQ).connect(midEQ).connect(highEQ).connect(voicePresence).connect(distortion).connect(modulationDelay).connect(compressor).connect(makeupGain);
  makeupGain.connect(dryGain).connect(effectMix);
  makeupGain.connect(echoDelay).connect(echoWet).connect(effectMix);
  echoDelay.connect(echoFeedback).connect(echoDelay);
  makeupGain.connect(ambienceDelay).connect(ambienceWet).connect(effectMix);
  ambienceDelay.connect(ambienceFeedback).connect(ambienceDelay);
  let channelNodes = [];
  const update = (value) => {
    const plan = audioProcessingPreviewPlan(value);
    highpass.frequency.value = plan.highpassHz;
    lowpass.frequency.value = plan.lowpassHz;
    lowEQ.gain.value = plan.eqLowDB;
    midEQ.gain.value = plan.eqMidDB;
    highEQ.gain.value = plan.eqHighDB;
    voicePresence.gain.value = plan.voicePresenceDB;
    distortion.curve = distortionCurve(plan.distortionAmount);
    distortion.oversample = plan.distortionAmount > 0 ? '2x' : 'none';
    const pitchPreviewDepthMS = Math.abs(plan.pitchSemitones) * .16;
    const pitchPreviewHz = Math.abs(plan.pitchSemitones) * .55;
    modulationOscillator.frequency.value = Math.max(.01, plan.modulationHz || pitchPreviewHz || .01);
    modulationDelay.delayTime.value = (plan.modulationDepthMS + pitchPreviewDepthMS) / 1000;
    modulationGain.gain.value = (plan.modulationDepthMS + pitchPreviewDepthMS) / 1000 * (plan.pitchSemitones < 0 ? -1 : 1);
    compressor.threshold.value = plan.compressorThresholdDB;
    compressor.knee.value = plan.compressorRatio > 1 ? 24 : 0;
    compressor.ratio.value = plan.compressorRatio;
    compressor.attack.value = .003;
    compressor.release.value = .25;
    makeupGain.gain.value = 10 ** (plan.makeupGainDB / 20);
    dryGain.gain.value = Math.max(0, 1 - Math.max(plan.echoWet, plan.ambienceWet) * .35);
    echoDelay.delayTime.value = Math.max(.001, plan.echoDelayMS / 1000 || .001);
    echoFeedback.gain.value = plan.echoFeedback;
    echoWet.gain.value = plan.echoWet;
    ambienceDelay.delayTime.value = Math.max(.001, plan.ambienceDelayMS / 1000 || .001);
    ambienceFeedback.gain.value = plan.ambienceFeedback;
    ambienceWet.gain.value = plan.ambienceWet;
    effectMix.disconnect();
    channelNodes.forEach(node => node.disconnect());
    channelNodes = connectChannelMode(context, effectMix, plan.channelMode);
  };
  update(initialValue);
  return {
    update,
    resume: () => context.state === 'suspended' ? context.resume() : Promise.resolve(),
    close: () => {
      try { source.disconnect(); } catch {}
      try { highEQ.disconnect(); } catch {}
      try { voicePresence.disconnect(); } catch {}
      try { distortion.disconnect(); } catch {}
      try { modulationDelay.disconnect(); } catch {}
      try { modulationGain.disconnect(); } catch {}
      try { modulationOscillator.stop(); } catch {}
      try { modulationOscillator.disconnect(); } catch {}
      try { compressor.disconnect(); } catch {}
      try { makeupGain.disconnect(); } catch {}
      try { dryGain.disconnect(); } catch {}
      try { echoDelay.disconnect(); } catch {}
      try { echoFeedback.disconnect(); } catch {}
      try { echoWet.disconnect(); } catch {}
      try { ambienceDelay.disconnect(); } catch {}
      try { ambienceFeedback.disconnect(); } catch {}
      try { ambienceWet.disconnect(); } catch {}
      try { effectMix.disconnect(); } catch {}
      channelNodes.forEach(node => { try { node.disconnect(); } catch {} });
      return context.close();
    }
  };
}
