import { mkdirSync, writeFileSync } from "node:fs"
import { dirname, join } from "node:path"

// OpenCode's half of the TUI refresh signal; Claude Code's half is in
// .claude/hooks/post-{write,bash}.sh. Content changes per call because mtime is
// 1s-granular on some filesystems.
export const TuiRefresh = async ({ directory }) => ({
  "tool.execute.after": async () => {
    try {
      const sentinel = join(directory, ".claude/state/.tui-refresh")
      mkdirSync(dirname(sentinel), { recursive: true })
      writeFileSync(sentinel, String(Math.random()))
    } catch {
      // a missed signal just falls back to the 2s tick
    }
  },
})
