---
version: alpha
name: Withgauge
description: |
  Gauge's design system embodies a bold, forward-thinking aesthetic built for
  clarity and decisiveness in AI-driven marketing. The visual identity combines
  sharp geometry with vibrant neon accents, creating a contemporary interface
  that feels both technical and approachable. The palette pivots around a deep
  navy primary (#051B23) grounded against a lime-yellow hairline accent
  (#E8FF3F), establishing immediate visual hierarchy through bold colour
  contrast rather than subtle gradation. The typography employs monospace and
  geometric sans-serif fonts to reinforce precision and modernity. Whitespace is
  generous, sections are clearly delineated by pastel bands, and interactive
  elements command attention through high-saturation fills and sharp
  corners—nothing curves or softens the edges.
source:
  url: "https://withgauge.com"
  pagesAnalyzed: 1
  extractedAt: 2026-09-05
  tokensMeasured: true
colors:
  primary: "#051B23"
  canvas: "#FFFFFF"
  surface: "#F5F5F5"
  surface-alt: "#E3EFD7"
  on-primary: "#FFFFFF"
  ink: "#051B23"
  hairline: "#E8FF3F"
  neutral-1: "#C4C4C4"
typography:
  display-lg:
    fontFamily: pxGrotesk
    fontSize: 57.6px
    fontWeight: 400
    lineHeight: 1.2
    letterSpacing: -2.3px
  display-md:
    fontFamily: pxGrotesk
    fontSize: 48px
    fontWeight: 400
    lineHeight: 1.1
    letterSpacing: -0.96px
  display-md-strong:
    fontFamily: pxGrotesk
    fontSize: 48px
    fontWeight: 700
    lineHeight: 1.1
    letterSpacing: -0.96px
  heading-md:
    fontFamily: pxGrotesk
    fontSize: 32px
    fontWeight: 400
    lineHeight: 1.2
    letterSpacing: -0.96px
  heading-sm:
    fontFamily: pxGrotesk
    fontSize: 20px
    fontWeight: 700
    lineHeight: 1.4
    letterSpacing: 0px
  heading-xs:
    fontFamily: switzer
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.28px
  body-xl:
    fontFamily: switzer
    fontSize: 20px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.32px
  body-lg:
    fontFamily: switzer
    fontSize: 18px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.32px
  body-md:
    fontFamily: switzer
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.32px
  body-sm:
    fontFamily: switzer
    fontSize: 15px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.32px
  body-xs:
    fontFamily: switzer
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.43
    letterSpacing: -0.32px
  body-xs-strong:
    fontFamily: switzer
    fontSize: 14px
    fontWeight: 600
    lineHeight: 1.5
    letterSpacing: -0.32px
  button-xl:
    fontFamily: pxGrotesk
    fontSize: 20px
    fontWeight: 400
    lineHeight: 1.4
    letterSpacing: -0.6px
  button-lg:
    fontFamily: switzer
    fontSize: 18px
    fontWeight: 700
    lineHeight: 1.5
    letterSpacing: -0.32px
  caption:
    fontFamily: switzer
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.32px
  code-md:
    fontFamily: pxGroteskMono
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.32px
    textTransform: uppercase
  code-sm:
    fontFamily: pxGroteskMono
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.32px
    textTransform: uppercase
  code-xs:
    fontFamily: pxGroteskMono
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.33
    letterSpacing: -0.32px
  code-xs-uppercase:
    fontFamily: pxGroteskMono
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.32px
    textTransform: uppercase
  code-xs-loose:
    fontFamily: pxGroteskMono
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.12px
    textTransform: uppercase
rounded:
  none: 0px
  full: 9999px
spacing:
  xxs: 4px
  xs: 8px
  sm: 12px
  md: 16px
  lg: 20px
  xl: 24px
  xxl: 32px
  xxxl: 40px
  section: 48px
  band: 64px
borderWidths:
  thin: 1px
shadows:
  sm: "rgba(0, 0, 0, 0.1) 0px 10px 15px -3px, rgba(0, 0, 0, 0.1) 0px 4px 6px -4px"
  md: "rgba(5, 27, 35, 0.2) 2px 2px 0px 0px inset, rgba(5, 27, 35, 0.1) 4px 4px 0px 0px inset"
  lg: "rgba(255, 255, 255, 0.3) 1px 1px 0px 0px inset, rgba(5, 27, 35, 0.08) -1px -1px 0px 0px inset, rgba(5, 27, 35, 0.18) 2px 2px 0px 0px"
  xl: "rgba(0, 0, 0, 0.1) 0px 20px 25px -5px, rgba(0, 0, 0, 0.1) 0px 8px 10px -6px"
elevationStrategy: layered-micro
themes:
  derived: dark   # the other theme is the site's measured palette
  light:
    bg: "#FFFFFF"
    surface: "#F5F5F5"
    surfaceRaised: "#EBECED"
    text: "#051B23"
    textMuted: "#5D6B70"
    border: "#E8FF3F"
    accent: "#051B23"
    accentFg: "#FFFFFF"
    focusRing: "#051B23"
    elevation: shadow
  dark:
    bg: "#0E0F11"
    surface: "#1C1D1F"
    surfaceRaised: "#29292B"
    text: "#F5F6F6"
    textMuted: "#9D9E9F"
    border: "#353537"
    accent: "#1989B2"
    accentFg: "#0B0B0C"
    focusRing: "#146E8E"
    elevation: "border+surface"
gradients:
  - context: section
    kind: linear
    value: "linear-gradient(in oklab, rgb(245, 245, 245) 0%, rgb(227, 239, 215) 100%)"
components:
  button-filled:
    typography: "{typography.code-md}"
    textColor: "{colors.ink}"
    border: "1px solid {colors.primary}"
    height: 50px
    padding: "12px 24px 12px 24px"
    backgroundColor: "{colors.hairline}"
  button-filled-lg:
    typography: "{typography.code-sm}"
    textColor: "{colors.ink}"
    height: 70px
    padding: "12px 24px 12px 24px"
    backgroundColor: "{colors.hairline}"
  button-outline:
    typography: "{typography.code-md}"
    textColor: "{colors.ink}"
    border: "1px solid {colors.primary}"
    height: 50px
    padding: "12px 24px 12px 24px"
    backgroundColor: "{colors.canvas}"
  button-text:
    typography: "{typography.code-xs-uppercase}"
    textColor: "oklab(0.208566 -0.0229038 -0.0226818 / 0.55)"
    height: 30px
    padding: "6px 14px 6px 14px"
  button-text-2:
    typography: "{typography.code-xs-uppercase}"
    textColor: "{colors.ink}"
    height: 30px
    padding: "6px 14px 6px 14px"
  card:
    typography: "{typography.body-md}"
    textColor: "{colors.ink}"
    border: "1px solid {colors.primary}"
    padding: "16px 16px 16px 16px"
    backgroundColor: "{colors.surface}"
  card-sm:
    typography: "{typography.body-md}"
    textColor: "{colors.ink}"
    padding: "24px 24px 24px 24px"
    backgroundColor: "{colors.canvas}"
  navigation:
    typography: "{typography.body-md}"
    textColor: "{colors.ink}"
    height: 70px
  footer:
    typography: "{typography.body-md}"
    textColor: "{colors.ink}"
    backgroundColor: "{colors.surface-alt}"
  link:
    typography: "{typography.body-sm}"
    textColor: "{colors.ink}"
  link-2:
    typography: "{typography.body-md}"
    textColor: "{colors.ink}"
states:
  other-hover:
    target: other
    state: hover
    borderColor: "rgba(5, 27, 35, 0.3)"
  other-focus-visible:
    target: other
    state: focus-visible
    outlineWidth: 2px
  button-hover:
    target: button
    state: hover
    textColor: "{colors.on-primary}"
  button-focus-visible:
    target: button
    state: focus-visible
    outlineColor: currentcolor
  nav-focus-visible:
    target: nav
    state: focus-visible
    outlineColor: currentcolor
  nav-hover:
    target: nav
    state: hover
    backgroundColor: "oklch(from currentcolor l c h / 0.1)"
breakpoints:
  - width: 375
    containerWidth: 375
    gridColumns: 4
    navLinksVisible: 32
    menuToggleVisible: true
    headingPx: 40
    bodyPx: 16
    sectionPaddingX: 0
  - width: 768
    containerWidth: 768
    gridColumns: 3
    navLinksVisible: 32
    menuToggleVisible: true
    headingPx: 46
    bodyPx: 16
    sectionPaddingX: 0
  - width: 1024
    containerWidth: 1024
    gridColumns: 6
    navLinksVisible: 32
    menuToggleVisible: true
    headingPx: 58
    bodyPx: 16
    sectionPaddingX: 0
  - width: 1280
    containerWidth: 1280
    gridColumns: 6
    navLinksVisible: 36
    menuToggleVisible: true
    headingPx: 58
    bodyPx: 16
    sectionPaddingX: 0
  - width: 1440
    containerWidth: 1440
    gridColumns: 6
    navLinksVisible: 36
    menuToggleVisible: true
    headingPx: 58
    bodyPx: 16
    sectionPaddingX: 0
coverage:
  statesFound: 33
  gradientsFound: 1
  rolesUnassigned: 1
  archetypesUnnamed: 0
  archetypesDetected: 0
  responsiveMeasured: true
  stylesheetsBlocked: false
  semanticRampDeclared: false
---

# Design System Inspired by Gauge

## 1. Visual Theme & Atmosphere

Gauge's design system embodies a bold, forward-thinking aesthetic built for clarity and decisiveness in AI-driven marketing. The visual identity combines sharp geometry with vibrant neon accents, creating a contemporary interface that feels both technical and approachable. The palette pivots around a deep navy primary (`{colors.primary}` — `#051B23`) grounded against a lime-yellow hairline accent (`{colors.hairline}` — `#E8FF3F`), establishing immediate visual hierarchy through bold colour contrast rather than subtle gradation. The typography employs monospace and geometric sans-serif fonts to reinforce precision and modernity. Whitespace is generous, sections are clearly delineated by pastel bands, and interactive elements command attention through high-saturation fills and sharp corners—nothing curves or softens the edges.

**Key Characteristics**
- Sharp, rectilinear geometry; no rounded corners on interactive elements
- Bold neon-to-navy contrast for visual separation
- Generous, intentional whitespace organized into distinct bands
- Monospace typography paired with humanist sans-serif for variety
- Flat elevation strategy using only inset and decorative shadows
- Gradient section backgrounds transitioning from neutral to soft pastels
- High-contrast call-to-action styling with solid fills and no decorative effects

## 2. Color Palette & Roles

### Primary
- **Primary / Brand** (`{colors.primary}` — `#051B23`): Deep navy used for primary CTAs, brand mark, headings, and primary text; also signals active states and borders on interactive elements.

### Accent Colors
- **Hairline / Neon Accent** (`{colors.hairline}` — `#E8FF3F`): Bright lime-yellow reserved for CTA fill backgrounds and prominent accent borders; demands immediate attention.

### Neutral Scale
- **Canvas / On Primary** (`{colors.canvas}` — `#FFFFFF`): Default page background and label colour on brand surfaces; highest contrast surface.
- **Surface** (`{colors.surface}` — `#F5F5F5`): Standard card and panel background; subtle tonal separation from canvas.
- **Surface Alt** (`{colors.surface-alt}` — `#E3EFD7`): Soft sage-green used for alternating section bands and decorative depth; introduces warmth without reducing legibility.
- **Neutral Decorative** (`{colors.neutral-1}` — `#C4C4C4`): Unassigned decorative role; appears in minimal measure, no primary function assigned.

### Semantic / Status
No error, success, warning, or info colours are declared in the site's markup. Status communication relies on text and layout hierarchy rather than chromatic coding.

## 3. Typography Rules

### Font Family
**Primary:** pxGroteskMono — monospace typeface for precision and technical character.
**Secondary:** Switzer — geometric sans-serif humanist typeface for body and narrative text.
**Fallback:** system-ui, -apple-system, sans-serif

### Hierarchy

| Role | Font | Size | Weight | Line Height | Letter Spacing | Notes |
|---|---|---|---|---|---|---|
| Display / Heading (XL) | Switzer | 58px | 400 | 1.5 | −0.32px | Largest headings, hero section |
| Heading (MD–LG) | Switzer | 46px | 400 | 1.5 | −0.32px | Section headings, mid-level prominence |
| Heading (XS) | Switzer | 14px | 400 | 1.5 | −0.28px | Small structured headings, labels |
| Body (XL) | Switzer | 20px | 400 | 1.5 | −0.32px | Large body copy, feature text |
| Body (LG) | Switzer | 18px | 400 | 1.5 | −0.32px | Standard generous body copy |
| Body (MD) | Switzer | 16px | 400 | 1.5 | −0.32px | Default body, card text |
| Body (SM) | Switzer | 15px | 400 | 1.5 | −0.32px | Compact body, dense content |
| Body (XS) | Switzer | 14px | 400 | 1.5 | −0.32px | Links, metadata, fine print |
| Body (XS Strong) | Switzer | 14px | 600 | 1.5 | −0.32px | Emphasized labels, strong inline text |
| Button (LG) | pxGroteskMono | 18px | 700 | 1.5 | −0.32px | Large primary and secondary buttons |
| Button (MD) | pxGroteskMono | 16px | 400 | 1.5 | −0.32px | Standard button text |
| Caption | Switzer | 12px | 400 | 1.5 | −0.32px | Captions, helper text, timestamps |

### Principles
- **Tight tracking throughout:** Consistent −0.32px letter-spacing on all Switzer roles maintains a modern, compact appearance and prevents excessive air in long-form text.
- **Monospace for action:** Button text uses pxGroteskMono at bold weight to create tactile, precise call-to-action messaging distinct from body hierarchy.
- **Line height constancy:** All roles maintain 1.5 line height for predictable vertical rhythm and accessible readability.
- **Weight discipline:** Only two weights in use (400 regular and 600/700 bold); no intermediate weights introduce unnecessary variation.

## 4. Component Stylings

### Buttons

**Filled (Primary CTA)**
- Background: `{colors.hairline}` (`#E8FF3F`)
- Text color: `{colors.primary}` (`#051B23`)
- Border: `1px solid {colors.primary}` (`#051B23`)
- Padding: `12px 24px`
- Height: `50px`
- Font: pxGroteskMono, `16px`, weight 400, line-height `24px`
- Border radius: `0px` (sharp corners)
- Box shadow: none
- Hover state: `backgroundColor` shifts to `rgba(232, 255, 63, 0.8)` (slightly desaturated lime), `color` remains primary ink

**Filled Large**
- Background: `{colors.hairline}` (`#E8FF3F`)
- Text color: `{colors.primary}` (`#051B23`)
- Border: `0px` (no border)
- Padding: `12px 24px`
- Height: `70px`
- Font: pxGroteskMono, `14px`, weight 400, line-height `21px`
- Border radius: `0px`
- Box shadow: none
- Hover state: similar to filled, opacity reduced to `0.8`

**Outline (Secondary)**
- Background: `{colors.canvas}` (`#FFFFFF`)
- Text color: `{colors.primary}` (`#051B23`)
- Border: `1px solid {colors.primary}` (`#051B23`)
- Padding: `12px 24px`
- Height: `50px`
- Font: pxGroteskMono, `16px`, weight 400, line-height `24px`
- Border radius: `0px`
- Box shadow: none
- Hover state: `backgroundColor` becomes `var(--color-charcoal)` (darkened primary), `color` turns white; border colour updates to `rgba(5, 27, 35, 0.3)`

**Text (Ghost/Link button)**
- Background: `rgba(0, 0, 0, 0)` (transparent)
- Text color: `oklab(0.208566 -0.0229038 -0.0226818 / 0.55)` (muted primary at 55% opacity)
- Border: `0px` (none)
- Padding: `6px 14px`
- Height: `30px`
- Font: pxGroteskMono, `12px`, weight 400, line-height `18px`
- Border radius: `0px`
- Box shadow: none
- Hover state: `opacity` shifts to `1`, full colour appears; `color` may transition to primary or neon accent depending on context

**Text Alt (Dark Ghost)**
- Background: `rgba(0, 0, 0, 0)` (transparent)
- Text color: `{colors.primary}` (`#051B23`)
- Border: `0px`
- Padding: `6px 14px`
- Height: `30px`
- Font: pxGroteskMono, `12px`, weight 400, line-height `18px`
- Border radius: `0px`
- Box shadow: none
- Hover state: `color` lightens or `opacity` adjusts for distinction

### Cards & Containers

**Card Default (Large)**
- Background: `{colors.surface}` (`#F5F5F5`)
- Text color: `{colors.primary}` (`#051B23`)
- Border: `1px solid {colors.primary}` (`#051B23`)
- Padding: `{spacing.md}` (`16px`)
- Font: Switzer, `16px`, weight 400, line-height `24px`
- Border radius: `0px`
- Box shadow: none
- Width: ~1200px; height: ~484px (responsive, content-driven)

**Card Default Small**
- Background: `{colors.canvas}` (`#FFFFFF`)
- Text color: `{colors.primary}` (`#051B23`)
- Border: `0px` (none)
- Padding: `{spacing.xl}` (`24px`)
- Font: Switzer, `16px`, weight 400, line-height `24px`
- Border radius: `0px`
- Box shadow: none
- Width: ~425px; height: `204px`
- Hover state: may receive `backgroundColor` tint via `color-mix(in oklab, var(--color-ice-blue) 10%, transparent)` for subtle lift

### Inputs & Forms

No form-specific input data was extracted from the site. See Known Gaps.

### Navigation

**Navigation Default**
- Background: `rgba(0, 0, 0, 0)` (transparent)
- Text color: `{colors.primary}` (`#051B23`)
- Border: `0px`
- Font: Switzer, `16px`, weight 400, line-height `24px`
- Height: `70px`
- Padding: `0px`
- Border radius: `0px`
- Box shadow: none
- Hover state: `backgroundColor` transitions to `oklch(from currentcolor l c h / 0.1)` (tinted wash of current text colour) or `rgba(0, 0, 0, 0)` depending on link role
- Focus state: `outline-color: currentcolor`, `outline-width: 2px`

### Footer

**Footer Default**
- Background: `{colors.surface-alt}` (`#E3EFD7`)
- Text color: `{colors.primary}` (`#051B23`)
- Border: `0px`
- Font: Switzer, `16px`, weight 400, line-height `24px`
- Padding: `0px` (structure inherited from sections)
- Border radius: `0px`
- Box shadow: none
- Width: full viewport
- Height: responsive (measured at ~912px on a standard desktop section)

### Links

**Link Default**
- Color: `{colors.primary}` (`#051B23`)
- Font: Switzer, `15px`, weight 400, line-height `22.5px`
- Text decoration: underline (inherited from link semantics)
- Background: transparent
- Border: `0px`
- Padding: `0px`
- Hover state: `color` may shift to accent or hover variant; `text-decoration: none` may apply
- Focus state: `outline-color: currentcolor`, `outline-width: 2px`

## 5. Layout Principles

### Spacing System

**Base Unit:** `{spacing.xs}` = `8px`

**Scale:**
- `{spacing.xxs}` = `4px` — micro spacing for tight UI elements (icon margins, badge padding)
- `{spacing.xs}` = `8px` — button label padding, list item gutters
- `{spacing.sm}` = `12px` — card internal spacing, small gap between related elements
- `{spacing.md}` = `16px` — default padding on cards and sections, content margins
- `{spacing.lg}` = `20px` — spacing between logical groups
- `{spacing.xl}` = `24px` — generous padding on feature cards and containers
- `{spacing.xxl}` = `32px` — gap between major content blocks
- `{spacing.xxxl}` = `40px` — vertical rhythm between hero and following section
- `{spacing.section}` = `48px` — padding around full-width sections (top/bottom)
- `{spacing.band}` = `64px` — largest section separation, used for major layout boundaries

**Usage Context:**
- **Micro UI:** buttons, badges, inline elements use `xxs` to `xs`
- **Cards:** internal padding uses `md` to `xl`
- **Sections:** vertical padding uses `section` (`48px`) to `band` (`64px`)
- **Horizontal gutters:** zero padding on containers to edge; content inside uses `md` to `xl` depending on viewport

### Grid & Container

- **Max width:** `1440px` at largest viewport
- **Column strategy:** Measured at 6-column grid at 1024px and above; 3 columns at 768px; 4 columns at 375px (though mobile often renders single-column content stacks)
- **Section patterns:** Full-width colour bands (canvas, surface, surface-alt) alternate to create visual rhythm; no fixed column count drives layout—content flows naturally and uses breakpoint-driven reflow
- **Container padding-x:** `0px` (no horizontal padding enforced on body/container; padding applied to children as needed)

### Whitespace Philosophy

Whitespace is deliberate and abundant. Large section breaks (`{spacing.band}` — `64px` vertically) separate major content zones. Horizontal whitespace is generated by the viewport width; at mobile, content fills edge-to-edge with internal padding preserving breath. Cards and panels are isolated by colour change rather than void, creating a layered but not cluttered appearance. Text maintains loose tracking (`−0.32px`) and 1.5 line height to ensure air between lines and words.

### Border Radius Scale

- `{rounded.none}` = `0px` — applied to all interactive elements (buttons, cards, images): buttons, cards, images all use sharp corners
- `{rounded.full}` = `9999px` — not observed in measured component roles; registered as a fallback scale value but not activated in the system as measured

The system defaults to sharp geometry. No pill-shaped or rounded variants are in use.

### Border Widths

- **Thin:** `1px` — standard border on buttons (outline style) and cards; used for subtle delineation and contrast

## 6. Depth & Elevation

| Level | Treatment | Use |
|---|---|---|
| Flat (Base) | No shadow | Default cards, surfaces, body backgrounds |
| Dropdown (SM) | `rgba(0, 0, 0, 0.1) 0px 10px 15px -3px, rgba(0, 0, 0, 0.1) 0px 4px 6px -4px` | Floating dropdowns, popovers, modest lift |
| Custom (MD) | `rgba(5, 27, 35, 0.2) 2px 2px 0px 0px inset, rgba(5, 27, 35, 0.1) 4px 4px 0px 0px inset` | Custom branded inset depth, visual emboss effect |
| Custom (LG) | `rgba(255, 255, 255, 0.3) 1px 1px 0px 0px inset, rgba(5, 27, 35, 0.08) -1px -1px 0px 0px inset, rgba(5, 27, 35, 0.18) 2px 2px 0px 0px` | Enhanced custom depth with highlight and shadow layers |
| Custom (XL) | `rgba(0, 0, 0, 0.1) 0px 20px 25px -5px, rgba(0, 0, 0, 0.1) 0px 8px 10px -6px` | Maximum ambient drop shadow for modals and overlays |

**Elevation Philosophy:** Gauge employs a **layered-micro** strategy. Rather than a progressive ladder of shadows, the system uses stacked near-transparent layers (often inset) to create barely-perceptible depth. The "custom" shadows introduce inset light and dark accents to simulate embossing, while dropdown and overlay shadows use modest ambient blur for separation. No element is truly "floating" in the traditional sense—instead, elevation is communicated through colour change (surface bands) and subtle shadow layering.

### Opacity Levels

The system defines opacity at the following steps for hover, disabled, and overlay states:

- `0.40` — deeply faded, minimal presence (disabled state or very subtle overlay)
- `0.45` — light presence
- `0.50` — medium reduced opacity, often used for secondary text or muted layers
- `0.70` — moderately visible, approaching full saturation
- `0.80` — near-full visibility, used for hover states on filled elements
- `0.90` — nearly full, for gentle interaction feedback

## 7. Do's and Don'ts

### Do
- **Use sharp corners.** All interactive elements (buttons, cards, images) employ `0px` border radius. Maintain geometric crispness as a defining trait.
- **Leverage colour bands for section separation.** Alternate between `{colors.canvas}`, `{colors.surface}`, and `{colors.surface-alt}` to delineate content zones without borders.
- **Prioritize contrast.** The neon hairline (`#E8FF3F`) is reserved for high-importance CTAs; pair it with the primary ink (`#051B23`) for maximum legibility.
- **Apply negative letter-spacing consistently.** All typography maintains `−0.32px` tracking; this tight, modern feel is foundational.
- **Use inset shadows sparingly for emboss effects.** Custom shadow layers create subtle 3D depth without compromising flat modern aesthetics.
- **Organize with substantial spacing.** Use `{spacing.section}` (`48px`) and `{spacing.band}` (`64px`) to clearly separate major sections and reduce visual noise.
- **Default to monospace for interactive labels.** Buttons and form actions use pxGroteskMono for precision and distinction from body text.

### Don't
- **Avoid rounded corners on any primary interactive element.** Even mild border radius contradicts the system's geometric foundation.
- **Do not mix too many opacity levels.** Stick to the defined scale (0.40, 0.45, 0.50, 0.70, 0.80, 0.90) to maintain predictability.
- **Never apply shadow to the primary flat surface.** Base cards and panels are flat; reserve shadows for floating or lifted states only.
- **Avoid introducing new colours outside the defined palette.** Decorative gradients are allowed (e.g. section band gradient), but all solid fills must come from primary, neutral, or accent roles.
- **Do not use the neon hairline for body text or low-priority elements.** It is reserved for CTAs and accent borders; overuse dilutes impact.
- **Never mix font families within a single semantic role.** pxGroteskMono and Switzer have distinct jobs; don't interchange them.
- **Avoid excessive white space within cards or buttons.** Padding is `{spacing.md}` (`16px`) minimum; stretch it to `{spacing.xl}` (`24px`) only on generous feature cards.

## 8. Responsive Behavior

### Breakpoints

| Breakpoint | Viewport Width | Columns | Key Changes | Notes |
|---|---|---|---|---|
| Mobile (XS) | 375px | 4 | Full-width content stacks; nav links hidden (menu toggle appears); largest heading `40px` | Single-column effective layout for stacked cards |
| Tablet (MD) | 768px | 3 | Content begins to organise into columns; nav links still hidden; heading grows to `46px`; body remains `16px` | Mid-column grid activates |
| Desktop (LG) | 1024px | 6 | Full navigation links visible; layout stabilises at 6-column grid; heading `58px`; body `16px` | Standard desktop view |
| Desktop (XL) | 1280px | 6 | Minor refinements; nav link count increases to 36 visible items; heading maintains `58px` | Wider viewport with more nav options |
| Desktop (2XL) | 1440px | 6 | Full max-width container at `1440px`; nav stabilises at 36 links; heading `58px`; body `16px` | Ultimate max width achieved |

**Collapse Point:** Between 768px and 1024px, the primary navigation collapses to a menu toggle; at 1024px and above, nav links are always visible.

### Touch Targets

Minimum interactive touch target size is `50px` in height for buttons and `70px` for large button variants. Navigation items maintain `70px` height to accommodate thumb-friendly interaction on mobile and tablet viewports. Links embedded in body text use default text line height (`1.5`), which at `16px` body yields `24px` vertical space—acceptable for mouse interaction but ideally wrapped in a larger tap zone on touch devices.

### Collapsing Strategy

- **Heading sizes:** Shrink from `58px` (desktop) → `46px` (tablet) → `40px` (mobile)
- **Body copy:** Stable at `16px` across all breakpoints; reduces only in captions to `12px`
- **Padding:** Sections maintain consistent internal padding (`{spacing.md}` to `{spacing.xl}`); viewport-driven reflow handles content stacking, not padding reduction
- **Grid:** 6 columns (desktop) → 3 columns (tablet) → 4 columns (mobile); content reorganises within column boundaries rather than reflow triggering a complete layout redraw
- **Navigation:** Link visibility toggles at the breakpoint; menu toggle replaces visible nav at 768px and below
- **Card width:** Large cards shrink to fit viewport; small cards stack vertically on mobile but maintain `{spacing.md}` internal padding

## 9. Agent Prompt Guide

### Quick Color Reference

- **Primary CTA fill:** Hairline (`{colors.hairline}` — `#E8FF3F`)
- **Primary CTA text & border:** Primary (`{colors.primary}` — `#051B23`)
- **Background (default):** Canvas (`{colors.canvas}` — `#FFFFFF`)
- **Card / panel background:** Surface (`{colors.surface}` — `#F5F5F5`)
- **Section band background:** Surface Alt (`{colors.surface-alt}` — `#E3EFD7`)
- **Heading / primary text:** Primary (`{colors.primary}` — `#051B23`)
- **Body text:** Primary (`{colors.primary}` — `#051B23`)
- **Borders:** Primary (`{colors.primary}` — `#051B23`) at `1px`
- **Accents / highlights:** Hairline (`{colors.hairline}` — `#E8FF3F`)

### Iteration Guide

1. **All corners are sharp.** Border radius is `0px` for buttons, cards, and images. No rounding except in edge-case decorative elements.
2. **Button text uses pxGroteskMono at bold.** Primary fills use `16px` weight 400 (`button-lg` at `18px` weight 700). Secondary (outline) uses matching monospace. Ghost buttons use thin weight or opacity reduction.
3. **Neon yellow (`#E8FF3F`) is reserved for primary CTA fills only.** Do not apply it to body, borders (except accent borders), or secondary elements.
4. **Sections are separated by colour, not whitespace alone.** Alternate background colours between canvas, surface, and surface-alt to create rhythm and reduce reliance on shadow or subtle borders.
5. **Spacing uses the `8px` base unit.** All padding, margin, and gap values are multiples of 8: `4px`, `8px`, `12px`, `16px`, `20px`, `24px`, `32px`, `40px`, `48px`, `64px`.
6. **Typography maintains `−0.32px` letter-spacing across all Switzer roles.** This tight tracking is a visual signature; do not increase letter-spacing on any body text.
7. **Line height is consistently `1.5` for all roles** (except pxGroteskMono button text, which also uses 1.5 on a per-line basis). This ensures predictable vertical rhythm.
8. **Shadows are minimal and mostly inset.** Use drop shadows (SM/XL levels) only for floating elements (dropdowns, modals). Most cards and panels are flat.
9. **Mobile collapse point is 1024px.** Navigation becomes a toggle menu at 768px and below; all other layout reflows are content-driven, not menu-driven.
10. **Opacity reductions are limited to 6 steps:** 0.40, 0.45, 0.50, 0.70, 0.80, 0.90. Use these for hover states, disabled states, and subtle overlays; avoid interpolating intermediate values.

## 10. Known Gaps

- **No semantic status colours.** The site does not declare error, success, warning, or info states in its CSS or HTML structure. Any UI requiring status indication must rely on text, icons, or layout hierarchy rather than chromatic coding.
- **Form inputs not measured.** No input, textarea, select, or checkbox styles were extracted. If a form is added, input styling must be defined separately and aligned with button and card principles (sharp corners, primary border colour, monospace labels).
- **Interaction states partially observed.** Only hover and focus-visible states are recorded from the extracted stylesheet. Other states such as `:active`, `:disabled`, `:invalid`, or `:readonly` were not explicitly captured; implementation may require interpolation from the patterns shown (opacity reduction, border colour shift, text colour dimming).
- **One colour role unassigned.** `{colors.neutral-1}` (`#C4C4C4`) was measured but has no declared semantic role. It appears minimally and may be decorative or reserved for future use; do not assume a job for it.
- **No dark mode or theme switching.** The extraction covers a single light theme. No dark-mode, high-contrast, or alt-theme variants were observed or measured.
- **Animation and motion not recorded.** Transitions, animations, and microinteractions (spin, fade, slide) were not captured in the static design token export. Refer to live site for behaviour details.
- **Icon sizing and styling not specified.** Icon colours and sizing within buttons, cards, and navigation are inferred from measured font sizes but are not explicitly documented; align icons to the height of adjacent text (`14px`, `16px`, `18px`) and use primary ink colour by default.
- **Gradient section band is the only declared gradient.** The system is predominantly flat and solid; only `linear-gradient(in oklab, rgb(245, 245, 245) 0%, rgb(227, 239, 215) 100%)` bridges Surface and Surface Alt. No other gradients are in use.
- **Container max-width applied at the template layer, not CSS class.** Measurement shows `1440px` max width achieved, but no explicit `max-width` rule was extracted for a universal container class; assume this is enforced by template layout, not a reusable utility.
- **One page analysed.** This extraction covers the Gauge homepage only. Authenticated surfaces, product dashboards, or secondary pages were not visited; their design patterns may diverge from this specification.