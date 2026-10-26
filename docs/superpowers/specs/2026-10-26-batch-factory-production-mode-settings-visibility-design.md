# Batch Factory Production Mode Settings Visibility Design

## Goal

Batch Factory must show only the settings that belong to the active production mode. Video-only settings must disappear in comic mode, and comic-only settings must disappear in video mode.

Switching modes changes visibility and execution ownership, but does not delete the inactive mode's saved values. Switching back restores the previous settings.

## Scope

Apply the same visibility rules to:

- unified configuration;
- single-book configuration;
- every configuration section that currently mixes video and comic controls.

The existing `productionMode` value remains the source of truth. No new configuration schema or separate modal is introduced.

## Always Visible

The following controls remain available because they are shared or required to switch modes:

- video/comic production mode selector;
- text model;
- image generation provider and image model;
- image aspect ratio;
- shared asset and constraint section containers, with their contents selected by mode.

## Video Mode

When `productionMode` is `video`, show only the existing video workflow settings:

- video generation provider and video model;
- video aspect ratio and resolution;
- storyboard duration;
- video prompt preset;
- standard video asset extraction prompt;
- standard visual prefix, quality, restriction, and negative prompts;
- audio source;
- TTS voice, style, speed, and pitch where currently available;
- storyboard planning follows audio;
- merge follows audio;
- other existing video-only timing or audio controls.

Do not render comic prompt, comic image derivation, comic asset prompt, comic constraints, style library, comic global reference image, or comic image-description prefix/suffix.

## Comic Mode

When `productionMode` is `comic`, show only comic workflow settings:

- comic page prompt preset;
- image-description derivation preset;
- comic asset extraction prompt;
- comic quality, restriction, and negative constraints;
- style library;
- comic global reference image and its asset/page scopes;
- global image-description prefix and suffix.

Do not render video provider/model, video aspect ratio, video resolution, storyboard duration, video prompt, audio source, TTS controls, or follow-audio controls.

## Mode Switching

The video and comic mode controls remain visible at all times. Only the active mode's detailed controls are rendered.

Changing mode updates the current modal draft immediately:

- inactive values stay in the draft and persisted configuration;
- inactive values are not displayed;
- inactive values are not used by the other production path;
- cancelling the modal discards the mode change and all other draft edits;
- saving persists the selected mode without erasing the inactive mode's configuration.

## Unified Configuration

### Model Configuration

- Video mode shows shared model controls plus video and audio controls.
- Comic mode shows only shared text/image controls.

### Image And Video

- Both mode selectors remain visible.
- The active mode shows its prompt controls.
- The inactive mode shows only its selector, without disabled detail fields.

### AI Reasoning

- Asset prompt selection uses the active mode's catalog slot.
- Constraint selectors use standard categories in video mode and comic categories in comic mode.
- The style library and comic reference controls render only in comic mode.
- The video settings section renders only in video mode.

## Single-Book Configuration

- The engine section keeps the production mode selector visible.
- Model, image, video, and audio controls follow the same visibility matrix as unified configuration.
- Asset settings show one mode-appropriate extraction selector.
- Constraint settings show only the active mode's selectors and supporting controls.
- No inactive-mode card or disabled inactive-mode input is shown.
- Existing inheritance and per-book override behavior remains unchanged.

## Execution Boundary

The backend continues to route production through `productionMode`:

- comic prompt generation must not consume video motion, video prompt, or audio settings;
- video generation must not consume comic style, comic reference, or comic prompt settings.

Existing inactive configuration is preserved for later mode switching, but it is not part of the active generation contract.

## Testing

Automated tests must prove:

- unified video mode contains video/audio controls and excludes comic controls;
- unified comic mode contains comic controls and excludes video/audio controls;
- single-book video mode contains video/audio controls and excludes comic controls;
- single-book comic mode contains comic controls and excludes video/audio controls;
- switching modes preserves the inactive configuration values;
- existing comic and video backend contracts remain isolated by `productionMode`.

Browser acceptance must verify both unified and single-book configuration by switching between video and comic drafts without saving test changes.

## Non-Goals

- changing the persisted settings schema;
- deleting inactive mode settings;
- creating separate video and comic configuration pages;
- changing prompt bodies, model catalogs, generation providers, or publish behavior;
- redesigning unrelated Batch Factory layout.
