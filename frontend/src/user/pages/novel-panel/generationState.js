export function createGenerationState() {
  return { status: 'idle', result: null, error: null, retryable: false, requestId: 0 };
}

export function beginGeneration(state) {
  if (state.status === 'running') return state;
  return { ...state, status: 'running', result: null, error: null, retryable: false, requestId: state.requestId + 1 };
}

export function cancelGeneration(state) {
  if (state.status !== 'running') return state;
  return { ...state, status: 'cancelled', error: null, retryable: true };
}

export function finishGeneration(state, result, error = null) {
  if (state.status !== 'running') return state;
  if (error) return { ...state, status: 'failed', result: null, error, retryable: error.retryable === true };
  return { ...state, status: 'succeeded', result, error: null, retryable: false };
}
