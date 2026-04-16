# Makefile — ai-squad repo root
# Targets for building, testing, and maintaining the ai-squad platform itself.
# For targets in your project template, see template/Makefile.
.PHONY: build-tui test lint sdlc-report sdlc-rotate-logs clean help

## Build the squad TUI binary
build-tui:
	@echo "Building squad TUI..."
	@cd tui && go build -o ../squad .
	@echo "Built: ./squad"

## Cross-compile squad for all platforms
build-tui-all:
	@mkdir -p dist
	GOOS=darwin  GOARCH=amd64  go build -C tui -o ../dist/squad-darwin-amd64  .
	GOOS=darwin  GOARCH=arm64  go build -C tui -o ../dist/squad-darwin-arm64  .
	GOOS=linux   GOARCH=amd64  go build -C tui -o ../dist/squad-linux-amd64   .
	GOOS=linux   GOARCH=arm64  go build -C tui -o ../dist/squad-linux-arm64   .
	GOOS=windows GOARCH=amd64  go build -C tui -o ../dist/squad-windows-amd64.exe .
	@echo "Binaries in dist/"

## Run the test suite for this repo
test:
	@bash tests/cli/test_ai_squad.sh

## Lint Go TUI code
lint:
	@cd tui && gofmt -e . >/dev/null && echo "gofmt: clean" || (echo "gofmt: issues found" && exit 1)

## Print per-story agent invocation counts and estimated costs
## Reads .claude/logs/skills.jsonl; safe to run offline (shows cached data)
sdlc-report:
	@if [ ! -f .claude/logs/skills.jsonl ]; then \
		echo "No skills.jsonl found at .claude/logs/skills.jsonl"; \
		echo "Run some agent workflows first."; \
		exit 0; \
	fi
	@echo "=== AI-Squad SDLC Cost Report ==="
	@echo ""
	@python3 scripts/sdlc-report.py .claude/logs/skills.jsonl 2>/dev/null || \
		python3 -c " \
import json, sys, collections; \
lines = [json.loads(l) for l in open('.claude/logs/skills.jsonl') if l.strip()]; \
by_story = collections.defaultdict(list); \
[by_story[l.get('story','unknown')].append(l) for l in lines]; \
print(f'Stories: {len(by_story)}  Total invocations: {len(lines)}'); \
[print(f'  {s}: {len(v)} invocations') for s,v in sorted(by_story.items())] \
"

## Rotate .claude/logs/ — keep last 5 compressed, delete older
## Safe to run any time; also triggered by post-merge hook
sdlc-rotate-logs:
	@bash scripts/sdlc/rotate-logs.sh

## Remove built binaries
clean:
	@rm -f squad dist/squad-*
	@echo "Cleaned."

## Show available targets
help:
	@echo ""
	@echo "  make build-tui          Build squad TUI binary"
	@echo "  make build-tui-all      Cross-compile for darwin/linux/windows"
	@echo "  make test               Run test suite (218 tests)"
	@echo "  make lint               Lint TUI Go code"
	@echo "  make sdlc-report        Print cost + invocation report"
	@echo "  make sdlc-rotate-logs   Rotate .claude/logs/ (keep last 5)"
	@echo "  make clean              Remove built binaries"
	@echo ""
