#!/usr/bin/env bash
# Copilot CLI's half of the TUI refresh signal; the hook fires on both phases, so the
# before one is dropped. Claude Code's half is in .claude/hooks/post-{write,bash}.sh.
grep -q '"phase"[[:space:]]*:[[:space:]]*"after"' || exit 0
mkdir -p .claude/state 2>/dev/null && echo $RANDOM > .claude/state/.tui-refresh 2>/dev/null
exit 0
