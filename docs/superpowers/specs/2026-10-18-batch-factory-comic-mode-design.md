# Batch Factory Comic Mode Design

## Goal

Add a comic production mode to Batch Factory so one novel can independently
produce static vertical comic pages as well as its existing video output.

## Scope

The source novel remains shared. Video and comic production are separate from
asset extraction onward. Existing video data, prompts, constraints, generated
videos, and controls must remain unchanged.

## Production Modes

### Video mode

Keep the existing video asset prompts, quality constraints, visual limits,
negative prompts, storyboard/video prompts, and video-generation flow.

### Comic mode

Comic mode has its own book-scoped data:

- character, scene, and prop asset prompts;
- generated asset images;
- comic quality constraint prompt;
- comic visual-limit prompt;
- comic negative prompt;
- a comic prompt for each existing storyboard item;
- the generated complete comic-page image.

No comic field may overwrite its video-mode counterpart, and no video field
may be reinterpreted as a comic field.

## Comic Prompt Contract

Each comic prompt produces one complete vertical `3:4` comic page directly
from an image model. It must include the chosen comic assets, page-wide art
style, panel layout, panel-by-panel actions and expressions, narration boxes,
speech bubbles, and their required Chinese text. The system does not compose
panels or overlay text after image generation.

The prompt must exclude video-specific timing, camera movement, audio, and
video generation instructions. It includes the comic quality, visual-limit,
and negative-prompt settings selected for that book.

## User Interface

1. Asset settings gains a mode switch: `视频资产` and `漫画资产`.
2. Constraint settings gains a mode switch. Comic mode exposes `漫画画质约束`,
   `漫画画面限制`, and `漫画负面提示词`.
3. The storyboard dialog gains `视频提示词` and `漫画提示词` tabs. The comic tab
   allows generation, editing, saving, and direct image generation for the
   selected page.
4. The production row exposes comic prompt generation and comic-page image
   generation without replacing its current video controls.

## Data and Compatibility

The existing `visualPrompt` field remains untouched for compatibility. Comic
data uses distinct persisted keys and storage records. Existing books load
with comic mode empty and continue to work in video mode without migration of
their saved content.

## Error Handling

Comic prompt generation requires comic assets and a configured text model.
Comic-page generation requires a non-empty comic prompt and a configured image
model. Errors identify the missing prerequisite and keep all previously saved
comic and video data intact.

## Verification

Tests cover mode isolation, comic prompt construction, comic asset selection,
comic constraint inclusion, and the visible labels/actions in the Batch
Factory page. Local browser verification confirms that existing video prompt
controls still work and that comic prompt controls are separately available.
