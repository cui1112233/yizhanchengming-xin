# Agent Canvas Energy Flow Design

## Goal

Give only the Agent canvas connections a moderately visible electric-current
feel. Keep the V88 dark palette, the existing coral default, all node styles,
toolbars, and the canvas background unchanged.

## Visual Rules

- Every connection keeps its existing solid base line and direction arrow.
- When `edgeAnimated` is enabled, a brighter short dash travels from source to
  target above the base line. The dash uses the selected connection color.
- The moving dash has a restrained glow so it reads as energy, rather than a
  second thick line or a flickering electric arc.
- Animation duration remains 1.1 seconds and uses linear movement so its
  direction is predictable.
- `prefers-reduced-motion` and the existing `edgeAnimated: false` setting both
  remove the moving dash and glow while retaining the static line and arrow.

## Scope Boundary

- Modify only the custom canvas edge rendering and its scoped canvas CSS.
- Do not change nodes, groups, toolbars, popovers, canvas background, stored
  data, APIs, migrations, or project settings.
- Existing color, width, and animation settings keep their current meaning.

## Verification

- Add a source-level regression assertion for the new energy-flow class.
- Run the focused canvas Node tests and production frontend build.
- Rebuild the local Go-embedded runtime and inspect an existing authenticated
  canvas project at `localhost:4001`.
- Confirm animated edges are visible with the default coral setting and that
  turning off motion leaves a static directed connection.
