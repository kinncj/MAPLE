# SKILL: wireframe

## Purpose

Generate low-fidelity wireframes from user story files. Output is deterministic — given the same story and layout hints, the same wireframe structure is produced. Three output formats: ASCII (default), SVG, HTML. Human approval is required before the wireframe feeds downstream mockup work.

## Inputs

| Field | Source | Example |
|---|---|---|
| `story_file` | path to story markdown | `docs/stories/auth-reset-0001.md` |
| `format` | `ascii` \| `svg` \| `html` | `ascii` |
| `ui_components` | derived from story Gherkin | form, button, error message |
| `stack` | `project.config.yaml` | `react-mantine` |

## Outputs

| File | Location |
|---|---|
| `<story-id>.wireframe.md` | `docs/design/wireframes/` |
| `<story-id>.wireframe.svg` | `docs/design/wireframes/` (if format=svg) |
| `<story-id>.wireframe.html` | `docs/design/wireframes/` (if format=html) |

## ASCII Wireframe Primitives

Use these consistently across all wireframes:

```
┌─────────────────────────────────┐   ← container / card
│  [Label]  [Input Field      ]   │   ← label + text input
│  [Button: Primary Action    ]   │   ← primary button
│  [Button: Secondary]            │   ← secondary button
│  ○ Option A  ○ Option B         │   ← radio group
│  ☐ Checkbox label               │   ← checkbox
│  ▼ Dropdown / Select            │   ← select / combobox
│  ──────────────────────         │   ← divider
│  ⚠ Error message text           │   ← validation error
│  ✓ Success confirmation         │   ← success state
└─────────────────────────────────┘

[Nav: Logo | Item 1 | Item 2 | CTA]  ← navigation bar
[ Sidebar  ][     Main Content    ]  ← two-column layout
[  Col 1  ][  Col 2  ][  Col 3  ]   ← three-column grid
[         Full-width Banner         ]← hero / header band
```

## ASCII Wireframe — Example (Password Reset)

```
┌──────────────────────────────────────┐
│            Reset Password            │
│                                      │
│  Email                               │
│  [                              ]    │
│                                      │
│  ⚠ No account found for this email  │  ← error state
│                                      │
│  [Button: Send Reset Link       ]    │
│                                      │
│  ← Back to Login                     │
└──────────────────────────────────────┘
```

## Generate a Wireframe File

```bash
STORY_FILE="docs/stories/auth-reset-password-20250416143000-0001.md"
STORY_ID=$(python3 -c "
import re
m = re.search(r'^id:\s*[\"\'](.*?)[\"\']', open('$STORY_FILE').read(), re.MULTILINE)
print(m.group(1) if m else 'unknown')
")
OUTPUT="docs/design/wireframes/${STORY_ID}.wireframe.md"
mkdir -p docs/design/wireframes
```

Write the wireframe file:

```markdown
---
story_id: "{story_id}"
story_file: "{story_file}"
format: ascii
status: draft          # draft | approved | rejected
approved_by: null
approved_at: null
---

## Wireframe: {story title}

### Layout

{ASCII wireframe here}

### States

**Default state:**
{ASCII wireframe — initial}

**Error state:**
{ASCII wireframe — validation error}

**Success state:**
{ASCII wireframe — confirmation}

### Interaction Notes

- Tab order: {list of focusable elements in tab sequence}
- Primary action: {describe}
- Error handling: {describe visible error states}

### Approval

- [ ] Approved by product owner
- [ ] Approved by UX lead (if applicable)
```

## SVG Wireframe

When `format=svg`, generate minimal SVG using rectangles and text:

```bash
cat > "docs/design/wireframes/${STORY_ID}.wireframe.svg" <<'SVGEOF'
<svg xmlns="http://www.w3.org/2000/svg" width="400" height="300" font-family="monospace" font-size="13">
  <!-- Container -->
  <rect x="10" y="10" width="380" height="280" fill="none" stroke="#333" stroke-width="1.5" rx="4"/>
  <!-- Title -->
  <text x="200" y="35" text-anchor="middle" font-weight="bold">Screen Title</text>
  <!-- Input -->
  <rect x="30" y="60" width="340" height="28" fill="none" stroke="#666" stroke-width="1"/>
  <text x="40" y="79" fill="#999">Label</text>
  <!-- Button -->
  <rect x="30" y="110" width="340" height="32" fill="#333" rx="3"/>
  <text x="200" y="131" text-anchor="middle" fill="#fff">Primary Action</text>
</svg>
SVGEOF
```

Extend the SVG template per story layout. Keep it structural, not aesthetic — no colors other than neutral grays, no styling that anticipates the visual identity.

## HTML Wireframe

When `format=html`, generate a minimal HTML skeleton with placeholder structure:

```bash
cat > "docs/design/wireframes/${STORY_ID}.wireframe.html" <<'HTMLEOF'
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Wireframe: {story title}</title>
  <style>
    * { box-sizing: border-box; font-family: monospace; }
    body { background: #f5f5f5; display: flex; justify-content: center; padding: 2rem; }
    .frame { background: white; border: 1.5px solid #333; border-radius: 4px; padding: 1.5rem; width: 400px; }
    .input { border: 1px solid #666; padding: .4rem; width: 100%; margin-bottom: .5rem; }
    .btn-primary { background: #333; color: white; border: none; padding: .5rem 1rem; width: 100%; cursor: pointer; }
    .error { color: #c0392b; font-size: .85rem; }
  </style>
</head>
<body>
  <div class="frame">
    <h2>{Screen Title}</h2>
    <!-- Layout here -->
  </div>
</body>
</html>
HTMLEOF
```

## Approval Gate

A wireframe must be marked `status: approved` before `mockup` or `ui-mockup-builder` proceeds:

```bash
python3 -c "
import re, sys
path = sys.argv[1]
text = open(path).read()
m = re.search(r'^status:\s*(\w+)', text, re.MULTILINE)
status = m.group(1) if m else 'draft'
if status != 'approved':
    print(f'BLOCKED: wireframe {path} is not approved (status={status})')
    sys.exit(1)
print('approved')
" "docs/design/wireframes/${STORY_ID}.wireframe.md"
```

## Failure Modes

| Condition | Action |
|---|---|
| Story has no Gherkin | Generate skeleton wireframe with placeholder states. Log `NO_GHERKIN — skeleton only`. |
| `docs/design/wireframes/` missing | Create it. |
| Wireframe exists and is `approved` | Do not overwrite. Log `SKIP — approved wireframe exists`. |
| Wireframe exists and is `draft` | Overwrite only if story Gherkin has changed. |
| `format` not recognised | Default to `ascii`. Log warning. |

## Logging

```
[wireframe] CREATE   docs/design/wireframes/auth-reset-0001.wireframe.md  format=ascii
[wireframe] SKIP     docs/design/wireframes/auth-reset-0001.wireframe.md  (approved — locked)
[wireframe] BLOCKED  auth-reset-0001  status=draft — needs approval before mockup
```
