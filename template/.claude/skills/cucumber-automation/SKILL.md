---
name: cucumber-automation
description: "Extract Gherkin scenarios from story markdown files into runnable .feature files and generate step definition stubs. Use when syncing stories to test suites."
---

# SKILL: cucumber-automation

## Purpose

Extract Gherkin scenarios from story markdown files into runnable `.feature` files, and generate step definition stubs for the project's test stack. Keeps `tests/features/` in sync with `docs/stories/` as the single source of Gherkin truth.

## Supported Stacks

| Stack | Step file extension | Framework |
|---|---|---|
| TypeScript / Node.js | `.steps.ts` | `@cucumber/cucumber` |
| JavaScript / Node.js | `.steps.js` | `@cucumber/cucumber` |
| Python | `_steps.py` | `behave` |
| Java | `Steps.java` | Cucumber-JVM |

Detect the stack from the repo root:
```bash
if [ -f "package.json" ] && grep -q "@cucumber/cucumber" package.json 2>/dev/null; then
  STACK="typescript"
elif [ -f "requirements.txt" ] && grep -q "behave" requirements.txt 2>/dev/null; then
  STACK="python"
elif [ -f "pom.xml" ] || [ -f "build.gradle" ]; then
  STACK="java"
else
  STACK="unknown"
fi
```

## Extract Gherkin from a Story File

Story files contain Gherkin in fenced code blocks tagged with `gherkin`:

```bash
STORY_FILE="docs/stories/user-auth-reset-password-20250416143000-0001.md"
EPIC="user-auth"

# Extract content of ```gherkin ... ``` blocks
python3 - <<'EOF'
import re, sys

story = open(sys.argv[1]).read()
blocks = re.findall(r'```gherkin\n(.*?)```', story, re.DOTALL)
if not blocks:
    print("NO_GHERKIN", file=sys.stderr)
    sys.exit(1)
print('\n\n'.join(blocks))
EOF "$STORY_FILE"
```

## Write the Feature File

```bash
STORY_FILE="docs/stories/user-auth-reset-password-20250416143000-0001.md"
FEATURE_DIR="tests/features"
EPIC="user-auth"

mkdir -p "$FEATURE_DIR/$EPIC"
FEATURE_FILE="$FEATURE_DIR/$EPIC/reset-password.feature"

# Extract and write (idempotent — overwrites if Gherkin changed)
python3 - <<'EOF'
import re, sys, os

story_path, feature_path = sys.argv[1], sys.argv[2]
story = open(story_path).read()
blocks = re.findall(r'```gherkin\n(.*?)```', story, re.DOTALL)
if not blocks:
    print(f"[cucumber] NO_GHERKIN in {story_path}", file=sys.stderr)
    sys.exit(1)

os.makedirs(os.path.dirname(feature_path), exist_ok=True)
with open(feature_path, 'w') as f:
    f.write('\n\n'.join(b.rstrip() for b in blocks))
    f.write('\n')

print(f"[cucumber] WRITE  {feature_path}  scenarios={len(re.findall(r'^  Scenario', ''.join(blocks), re.MULTILINE))}")
EOF "$STORY_FILE" "$FEATURE_FILE"
```

## Generate Step Definition Stubs (TypeScript)

```bash
FEATURE_FILE="tests/features/user-auth/reset-password.feature"
STEPS_DIR="tests/step-definitions/user-auth"
mkdir -p "$STEPS_DIR"

python3 - <<'EOF'
import re, sys, os

feature_path = sys.argv[1]
steps_dir = sys.argv[2]
feature = open(feature_path).read()

# Collect unique step texts
steps = re.findall(r'^\s+(Given|When|Then|And|But) (.+)$', feature, re.MULTILINE)
seen = set()
unique = []
for keyword, text in steps:
    norm = re.sub(r'"[^"]+"', '"{string}"', text)
    norm = re.sub(r'\b\d+\b', '{int}', norm)
    if norm not in seen:
        seen.add(norm)
        unique.append((keyword if keyword not in ('And','But') else 'Given', norm, text))

# Build TypeScript file
feature_name = os.path.basename(feature_path).replace('.feature','')
out_path = os.path.join(steps_dir, f"{feature_name}.steps.ts")

if os.path.exists(out_path):
    print(f"[cucumber] SKIP  {out_path}  (already exists)")
    sys.exit(0)

lines = [
    "import { Given, When, Then } from '@cucumber/cucumber';",
    "",
]
for keyword, pattern, original in unique:
    fn = keyword.lower()
    # build parameter list from placeholders
    params = []
    if '{string}' in pattern:
        params += [f"arg{i}: string" for i in range(pattern.count('{string}'))]
    if '{int}' in pattern:
        params += [f"n{i}: number" for i in range(pattern.count('{int}'))]
    param_str = ', '.join(params)
    escaped = pattern.replace("'", "\\'")
    lines += [
        f"// {original}",
        f"{fn}('{escaped}', async ({param_str}) => {{",
        f"  // TODO: implement",
        f"  throw new Error('Pending: {escaped}');",
        f"}});",
        "",
    ]

with open(out_path, 'w') as f:
    f.write('\n'.join(lines))

print(f"[cucumber] STUBS  {out_path}  steps={len(unique)}")
EOF "$FEATURE_FILE" "$STEPS_DIR"
```

## Generate Step Definition Stubs (Python / behave)

```bash
FEATURE_FILE="tests/features/user_auth/reset_password.feature"
STEPS_DIR="tests/features/user_auth"
mkdir -p "$STEPS_DIR"

python3 - <<'EOF'
import re, sys, os

feature_path = sys.argv[1]
steps_dir = sys.argv[2]
feature = open(feature_path).read()

steps = re.findall(r'^\s+(Given|When|Then|And|But) (.+)$', feature, re.MULTILINE)
seen = set()
unique = []
for keyword, text in steps:
    norm = re.sub(r'"[^"]+"', '"{text}"', text)
    norm = re.sub(r'\b\d+\b', '{n:d}', norm)
    if norm not in seen:
        seen.add(norm)
        unique.append((keyword if keyword not in ('And','But') else 'given', norm, text))

feature_name = os.path.basename(feature_path).replace('.feature','')
out_path = os.path.join(steps_dir, f"{feature_name}_steps.py")

if os.path.exists(out_path):
    print(f"[cucumber] SKIP  {out_path}  (already exists)")
    sys.exit(0)

lines = ["from behave import given, when, then", ""]
for keyword, pattern, original in unique:
    fn = keyword.lower()
    escaped = pattern.replace('"', '\\"')
    lines += [
        f"# {original}",
        f'@{fn}(u"{escaped}")',
        f"def step_impl(context):",
        f'    raise NotImplementedError(u"STEP: {fn} {escaped}")',
        "",
    ]

with open(out_path, 'w') as f:
    f.write('\n'.join(lines))

print(f"[cucumber] STUBS  {out_path}  steps={len(unique)}")
EOF "$FEATURE_FILE" "$STEPS_DIR"
```

## Batch Sync (all stories → all feature files)

```bash
find docs/stories -name "*.md" ! -name "_template.md" | while read -r story; do
  EPIC=$(python3 -c "
import re
m = re.search(r'^epic:\s*[\"\'](.*?)[\"\']', open('$story').read(), re.MULTILINE)
print(m.group(1) if m else 'unknown')
")
  SLUG=$(basename "$story" | sed 's/-[0-9]\{14\}-[0-9]\{4\}\.md$//')
  FEATURE_FILE="tests/features/$EPIC/$SLUG.feature"
  # extract-and-write step (see above) ...
  echo "[cucumber] SYNC  $story → $FEATURE_FILE"
done
```

## Failure Modes

| Condition | Action |
|---|---|
| Story file has no `gherkin` fenced block | Log `NO_GHERKIN`. Skip. Do not create empty feature file. |
| Feature file already exists with manual edits | Compare checksums. If changed: log `MANUAL_EDIT — skip overwrite`. Never overwrite manual edits. |
| Step file already exists | Skip generation. Log `SKIP`. Agents implement stubs; generator does not regenerate. |
| Stack is `unknown` | Log warning. Write `.feature` file only; skip step stubs. |
| Gherkin parse error (malformed block) | Log the offending line. Skip the file. Do not attempt partial extraction. |

## Logging

```
[cucumber] WRITE   tests/features/user-auth/reset-password.feature  scenarios=3
[cucumber] STUBS   tests/step-definitions/user-auth/reset-password.steps.ts  steps=7
[cucumber] SKIP    tests/step-definitions/user-auth/login.steps.ts  (already exists)
[cucumber] NO_GHERKIN  docs/stories/auth-spike-20250416143000-0002.md  (spike — expected)
```
