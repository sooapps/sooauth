# Sooauth Design System

Hot Red, Soft Beige, canvas neutrals, and high-contrast dark surfaces.

## Accent — Hot Red

- **`#FF3B3B`** Hot Red.
  - Hover: `#E03030` (Light) / `#FF5252` (Dark).
  - Active: `#C92525`.
  - Contrast foreground: `#FFFFFF`.
  - Focus ring: `rgba(255, 59, 59, 0.35)` or `2px solid #FF3B3B`.
- Semantic status (ok / warn / error) is distinct from brand:
  - Error: `#DC2626`
  - Success: `#16A34A`
  - Warning: `#D97706`

## Surfaces — Light Mode

| Token | Hex | Use |
| --- | --- | --- |
| canvas | `#FFFFFF` | Page background |
| primary | `#1C1412` | Text, primary headers |
| surface | `#FAFAFA` | Neutral cards and panels |
| surface-warm | `#FFF4E6` | Soft Beige warm cards, notices, explainer panels |
| border | `#E4DFD9` | Hairline borders and dividers |
| muted | `#6B6360` | Secondary / helper text |

## Surfaces — Dark Mode

| Token | Hex | Use |
| --- | --- | --- |
| canvas | `#0E0F11` | Dark page background |
| primary | `#F5F6F6` | White/light text and headers |
| surface | `#1C1D1F` | Dark cards and panels |
| surface-field | `#151618` | Form inputs and elevated cards |
| surface-warm | `#261B18` | Dark warm subtle surface (Soft Beige dark counterpart) |
| border | `#353537` | Subtle dark borders and dividers |
| muted | `#9D9E9F` | Secondary / helper text |

## Typography & Micro-Interactions

- Fonts: "IBM Plex Sans", Inter, system sans-serif.
- Monospace: ui-monospace, "JetBrains Mono", SFMono-Regular.
- Concentric border radius (`outer = inner + padding`).
- Tactile button clicks: `scale(0.96)`.
- Minimum hit target: 44×44px on mobile, 40×40px on desktop.

## Hard Avoids

- Generic AI purple-pink gradients, heavy blurred glassmorphism.
- Neon lime/yellow clashes (avoids confusion with Soobrief).
- Low-contrast light text on bright backgrounds.

## Key Surfaces

- **Hosted Auth Pages (`/auth/sign-in`, `/auth/sign-up`):** Split-screen responsive layout with smooth sliding transitions, brand value props, and dark/light auto-detection.
- **Admin Dashboard:** Ergonomic sidebar, interactive user profile popover with theme switcher, clean topbar, Hot Red active indicators.
- **Marketing Site:** Modern developer-first presentation in Hot Red & Soft Beige.

