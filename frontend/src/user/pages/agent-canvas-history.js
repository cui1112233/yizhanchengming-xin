function clone(value) {
  return value === undefined ? undefined : JSON.parse(JSON.stringify(value));
}

function cleanSnapshot(snapshot) {
  return {
    nodes: (snapshot?.nodes || []).map(node => {
      const { selected: _selected, dragging: _dragging, measured: _measured, ...persistent } = node;
      return clone(persistent);
    }),
    edges: (snapshot?.edges || []).map(edge => {
      const { selected: _selected, ...persistent } = edge;
      return clone(persistent);
    })
  };
}

export function createCanvasHistory(initialSnapshot) {
  let past = [];
  let present = cleanSnapshot(initialSnapshot);
  let future = [];

  return {
    current() {
      return clone(present);
    },
    canUndo() {
      return past.length > 0;
    },
    canRedo() {
      return future.length > 0;
    },
    push(snapshot) {
      past.push(present);
      present = cleanSnapshot(snapshot);
      future = [];
      return clone(present);
    },
    undo() {
      if (!past.length) return clone(present);
      future.unshift(present);
      present = past.pop();
      return clone(present);
    },
    redo() {
      if (!future.length) return clone(present);
      past.push(present);
      present = future.shift();
      return clone(present);
    },
    reset(snapshot) {
      past = [];
      present = cleanSnapshot(snapshot);
      future = [];
      return clone(present);
    }
  };
}
