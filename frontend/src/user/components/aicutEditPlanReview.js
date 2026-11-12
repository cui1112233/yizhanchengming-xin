export function operationReviewDetails(operation = {}) {
 return JSON.stringify({ ...operation, startMs: operation.startMs ?? 0, sourceStartMs: operation.sourceStartMs ?? 0, durationMs: operation.durationMs ?? 0 }, null, 2);
}
export function selectedOperationIds(plan, selected = []) {
 const ids = new Set(selected);
 return (plan?.operations || []).filter(operation => ids.has(operation.id)).map(operation => operation.id);
}
