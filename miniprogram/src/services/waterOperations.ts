import {onSessionInvalidated} from './sessionBoundary';

// Kept outside page instances so returning to a page cannot repeat an uncertain paid write.
export const waterUnknownOperations=new Set<string>();
onSessionInvalidated(()=>waterUnknownOperations.clear());
