const lines = ['窗外是风景', '下一幕由你创造'];

// Render the supplied font itself; stroke-library glyphs would replace its shapes.
export function HomeHandwriting() {
  let index = 0;
  return <h1 className="handwritten supplied-handwriting" aria-label="窗外是风景，下一幕由你创造">
    {lines.map(line => <span className="hand-line" key={line} aria-hidden="true">
      {[...line].map(char => <span className="supplied-glyph" key={char} style={{ '--char-index': index++ }}>{char}</span>)}
    </span>)}
  </h1>;
}

