---
title: USS Surveillance - Design Specification
status: final
created: 2026-07-08
updated: 2026-07-08
tokens:
  colors:
    background: '#0a192f'
    surface: '#172a45'
    surfaceHover: '#233554'
    primary: '#00b0ff'
    secondary: '#00e5ff'
    textPrimary: '#e6f1ff'
    textSecondary: '#8892b0'
    success: '#00e676'
    warning: '#ff9100'
    danger: '#ff3d00'
  typography:
    fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif"
    headerFont: "Roboto, -apple-system, sans-serif"
  rounded:
    default: '6px'
    card: '8px'
    badge: '4px'
  spacing:
    containerPadding: '20px'
    elementGap: '12px'
---

# Visual Design Specification: USS Surveillance

Interactive Mockup Reference: [screen-dashboard.html](file:///home/bangnq/wip/uss-surveillance/_bmad-output/planning-artifacts/ux-designs/ux-uss-surveillance-2026-07-08/mockups/screen-dashboard.html)

> [!NOTE]
> In case of any conflict between the visual mockup representation and the specifications written in this document (DESIGN.md), this document wins.

## 1. Brand & Style
USS Surveillance visual design is engineered to project **trust, precision, and high-tech command control**. Drawing inspiration from modern defense and aviation systems, the interface employs a dark blue-grey base palette with glowing cyan and blue accents. The style feels tactical, utilitarian, and clean, minimizing visual clutter so operator Sarah can process telemetry and alerts instantly.

## 2. Colors
We utilize a semantic color hierarchy optimized for dark-mode control rooms:
- **Base Background:** `{tokens.colors.background}` (`#0a192f`) — Minimal light pollution to reduce eye fatigue.
- **Surface Panels:** `{tokens.colors.surface}` (`#172a45`) — Used for sidebar grids, control panels, and header bars.
- **Hover States:** `{tokens.colors.surfaceHover}` (`#233554`) — Interactive component hover feedbacks.
- **Primary Branding / Highlight:** `{tokens.colors.primary}` (`#00b0ff`) — Ocean blue for map boundaries, telemetry gauges, and active navigation.
- **Secondary Branding / Focus:** `{tokens.colors.secondary}` (`#00e5ff`) — Neon cyan for active drones, video overlays, and telemetry focus states.
- **Status Indicators:**
  - **Success:** `{tokens.colors.success}` (`#00e676`) — Green for online, charging, and safe weather.
  - **Warning:** `{tokens.colors.warning}` (`#ff9100`) — Yellow-orange for paused missions or high winds near limit.
  - **Danger:** `{tokens.colors.danger}` (`#ff3d00`) — Red for emergency landing, hardware offline, and critical wind/battery alerts.

## 3. Typography
The system font stack is tuned for telemetry legibility:
- **Primary Font:** `{tokens.typography.fontFamily}` (System Font Stack with Fallback to Roboto)
- **Headers:** `{tokens.typography.headerFont}` (Roboto for high geometric clarity)
- **Telemetry Display:** Monospace fallback is preferred for coordinates and numbers to avoid horizontal jumpiness as values refresh (e.g. `font-family: monospace`).

## 4. Layout & Spacing
Optimized for widescreen desktop displays (1920x1080 and wider):
- **Base Container Padding:** `{tokens.spacing.containerPadding}` (`20px`) — Edge margin for main panels.
- **Element Grid Gap:** `{tokens.spacing.elementGap}` (`12px`) — Grid separation for dashboard sub-cards.
- **Layout Panels:** 3-column docked configuration with fixed drawer widths (280px left sidebar, 320px right sidebar, flexible center mapping workspace).

## 5. Elevation & Depth
In dark mode, depth is represented through color luminance and border highlights rather than heavy drop shadows:
- **Layer 0 (Map & Background):** Lowest luminance (`#0a192f`).
- **Layer 1 (Control Panels & Sidebars):** Higher luminance (`#172a45`).
- **Layer 2 (Toasts & Overlays):** Border highlighted with `{tokens.colors.primary}` and a semi-transparent background to allow map transparency.

## 6. Shapes
Clean, slightly rounded edges consistent with modern Material UI styles:
- **Button / Input Border Radius:** `{tokens.rounded.default}` (`6px`)
- **Card / Video Feed Border Radius:** `{tokens.rounded.card}` (`8px`)
- **Status Badge Border Radius:** `{tokens.rounded.badge}` (`4px`)

## 7. Components
- **Dashboard Telemetry Card:** Uses a container outline matching `{tokens.colors.border-color}` with value highlights in `{tokens.colors.secondary}`.
- **Action Buttons:** Large touch/click surfaces with clear background fills matching status intent (e.g. yellow background for `Pause`, red background for `RTH`).
- **Map Overlays:** Path borders must have at least 2px width in `{tokens.colors.primary}` with 50% opacity overlays.

## 8. Do's and Don'ts
- **DO** use absolute black (`#000000`) for the live video stream placeholder to block bleeding glare.
- **DO** use monospace fonts for live numerical coordinates and telemetry readouts.
- **DON'T** mix bright pastel colors in borders or icons; stick strictly to the semantic color mapping to keep warning alerts distinguishable.
- **DON'T** show the mission replay timeline scrubber while active flight controls are active.
