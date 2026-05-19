#!/usr/bin/env python3

import argparse
import html
import json
import os
import pathlib
import posixpath
import urllib.parse
from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Dict, List, Optional


def now_iso() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace("+00:00", "Z")


def read_json(path: pathlib.Path) -> Dict:
    try:
        with path.open("r", encoding="utf-8") as f:
            return json.load(f)
    except Exception:
        return {}


def write_json(path: pathlib.Path, payload: Dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        json.dump(payload, f, indent=2)
        f.write("\n")


def read_text(path: pathlib.Path) -> str:
    try:
        return path.read_text(encoding="utf-8").strip()
    except Exception:
        return ""


class PortalState:
    def __init__(self, root: pathlib.Path, token_file: pathlib.Path):
        self.root = root
        self.state_dir = root / ".claude" / "state"
        self.maple_json = self.state_dir / "maple.json"
        self.pending_file = self.state_dir / "approval-pending.txt"
        self.feedback_file = self.state_dir / "design-feedback.json"
        self.feedback_log = self.state_dir / "design-feedback.log"
        self.token_file = token_file

    def token(self) -> str:
        return read_text(self.token_file)

    def pipeline(self) -> Dict:
        p = read_json(self.maple_json)
        pending = read_text(self.pending_file)
        feedback = read_json(self.feedback_file)
        return {
            "taffy": p.get("taffy", ""),
            "stage": p.get("stage", ""),
            "status": p.get("status", ""),
            "awaiting_approval": p.get("awaiting_approval", ""),
            "updated_at": p.get("updated_at", ""),
            "approval_pending": pending,
            "feedback": feedback,
        }

    def request_changes(self, message: str) -> Dict:
        stage = read_text(self.pending_file)
        payload = {
            "stage": stage,
            "message": message.strip(),
            "ts": now_iso(),
            "status": "requested_changes",
        }
        write_json(self.feedback_file, payload)
        self.feedback_log.parent.mkdir(parents=True, exist_ok=True)
        with self.feedback_log.open("a", encoding="utf-8") as f:
            f.write(json.dumps(payload) + "\n")
        return payload

    def approve(self) -> Dict:
        stage = read_text(self.pending_file)
        if self.pending_file.exists():
            self.pending_file.unlink()
        return {"approved": True, "stage": stage, "ts": now_iso()}


def normalize_rel(path: str) -> str:
    norm = posixpath.normpath(path.replace("\\", "/"))
    if norm.startswith("../") or norm == "..":
        return ""
    return norm.lstrip("/")


def is_allowed_rel(path: str) -> bool:
    return path.startswith("docs/design/") or path.startswith("docs/stories/")


def list_artifacts(root: pathlib.Path, stage: str) -> List[Dict]:
    stage = (stage or "").lower()
    roots = ["docs/design", "docs/stories"]
    if "wireframe" in stage:
        roots = ["docs/design/wireframes", "docs/stories"]
    elif "visual-identity" in stage or "design-tokens" in stage:
        roots = ["docs/design/identity", "docs/stories"]
    elif "mockup" in stage:
        roots = ["docs/design/mockups", "docs/design/identity", "docs/stories"]

    items: List[Dict] = []
    for r in roots:
        base = root / r
        if not base.exists():
            continue
        for p in base.rglob("*"):
            if not p.is_file():
                continue
            rel = p.relative_to(root).as_posix()
            ext = p.suffix.lower()
            kind = "text"
            if ext in {".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg"}:
                kind = "image"
            elif ext in {".json", ".md", ".tsx", ".ts", ".css", ".html", ".yaml", ".yml"}:
                kind = "text"
            else:
                kind = "file"
            items.append(
                {
                    "path": rel,
                    "kind": kind,
                    "mtime": int(p.stat().st_mtime),
                }
            )
    items.sort(key=lambda x: x["mtime"], reverse=True)
    return items[:200]


def content_type_for(path: pathlib.Path) -> str:
    ext = path.suffix.lower()
    if ext == ".html":
        return "text/html; charset=utf-8"
    if ext == ".json":
        return "application/json; charset=utf-8"
    if ext in {".md", ".txt", ".tsx", ".ts", ".css", ".yml", ".yaml"}:
        return "text/plain; charset=utf-8"
    if ext == ".svg":
        return "image/svg+xml"
    if ext == ".png":
        return "image/png"
    if ext in {".jpg", ".jpeg"}:
        return "image/jpeg"
    if ext == ".gif":
        return "image/gif"
    if ext == ".webp":
        return "image/webp"
    return "application/octet-stream"


def render_index(token: str) -> str:
    token_js = json.dumps(token)
    return f"""<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>MAPLE Design Review</title>
  <style>
    :root {{
      --bg: #0f1117;
      --card: #171b22;
      --text: #e5e7eb;
      --muted: #9ca3af;
      --ok: #22c55e;
      --warn: #f59e0b;
      --bad: #ef4444;
      --line: #2b313b;
      --accent: #60a5fa;
    }}
    body {{ margin:0; font-family: Inter, system-ui, -apple-system, sans-serif; background:var(--bg); color:var(--text); }}
    .wrap {{ max-width:1200px; margin:0 auto; padding:24px; }}
    .h1 {{ font-size:24px; font-weight:700; margin:0 0 16px; }}
    .sub {{ color:var(--muted); font-size:13px; margin-bottom:16px; }}
    .row {{ display:grid; grid-template-columns: 1fr 1fr; gap:16px; }}
    .card {{ background:var(--card); border:1px solid var(--line); border-radius:12px; padding:16px; }}
    .label {{ color:var(--muted); font-size:12px; text-transform:uppercase; letter-spacing:.05em; }}
    .value {{ font-size:15px; margin-top:4px; }}
    .btns {{ display:flex; gap:10px; flex-wrap:wrap; margin-top:14px; }}
    button {{ border:0; border-radius:10px; padding:10px 14px; color:#fff; font-weight:600; cursor:pointer; }}
    .approve {{ background:var(--ok); }}
    .changes {{ background:var(--warn); color:#111827; }}
    .refresh {{ background:#374151; }}
    textarea {{ width:100%; min-height:92px; margin-top:10px; background:#0b0d12; color:var(--text); border:1px solid var(--line); border-radius:10px; padding:10px; }}
    .list {{ margin-top:12px; max-height:520px; overflow:auto; border:1px solid var(--line); border-radius:10px; }}
    .item {{ display:flex; justify-content:space-between; gap:12px; padding:10px 12px; border-bottom:1px solid var(--line); }}
    .item:last-child {{ border-bottom:none; }}
    a {{ color:var(--accent); text-decoration:none; }}
    .chip {{ font-size:11px; color:var(--muted); border:1px solid var(--line); border-radius:999px; padding:2px 8px; }}
    .status {{ font-weight:700; }}
  </style>
</head>
<body>
  <div class="wrap">
    <div class="h1">MAPLE Design Review Portal</div>
    <div class="sub">Companion to TUI pipeline approvals. TUI remains the primary control surface.</div>
    <div class="row">
      <div class="card">
        <div class="label">Workflow</div><div id="wf" class="value">-</div>
        <div class="label" style="margin-top:10px">Stage</div><div id="stage" class="value">-</div>
        <div class="label" style="margin-top:10px">Status</div><div id="status" class="value status">-</div>
        <div class="label" style="margin-top:10px">Pending approval</div><div id="pending" class="value">-</div>
        <div class="label" style="margin-top:10px">Updated</div><div id="updated" class="value">-</div>
        <textarea id="feedback" placeholder="Request changes (optional notes for agent/human)..."></textarea>
        <div class="btns">
          <button class="approve" onclick="approveStage()">Approve stage</button>
          <button class="changes" onclick="requestChanges()">Request changes</button>
          <button class="refresh" onclick="refreshAll()">Refresh</button>
        </div>
        <div id="msg" class="sub" style="margin-top:10px"></div>
      </div>
      <div class="card">
        <div class="label">Artifacts</div>
        <div id="artifacts" class="list"></div>
      </div>
    </div>
  </div>
  <script>
    const TOKEN = {token_js};

    function esc(s) {{
      return String(s || "")
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;");
    }}

    async function api(path, options={{}}) {{
      const headers = options.headers || {{}};
      headers["X-Maple-Token"] = TOKEN;
      if (!headers["Content-Type"] && options.body) headers["Content-Type"] = "application/json";
      const res = await fetch(path, {{...options, headers}});
      const txt = await res.text();
      let data = {{}};
      try {{ data = JSON.parse(txt || "{{}}"); }} catch (e) {{ data = {{ raw: txt }}; }}
      if (!res.ok) throw new Error(data.error || `HTTP ${{res.status}}`);
      return data;
    }}

    async function refreshAll() {{
      const s = await api("/api/state");
      document.getElementById("wf").textContent = s.taffy || "-";
      document.getElementById("stage").textContent = s.stage || "-";
      document.getElementById("status").textContent = s.status || "-";
      document.getElementById("pending").textContent = s.approval_pending || "-";
      document.getElementById("updated").textContent = s.updated_at || "-";
      if (s.feedback && s.feedback.message) {{
        document.getElementById("feedback").value = s.feedback.message;
      }}

      const a = await api("/api/artifacts");
      const box = document.getElementById("artifacts");
      box.innerHTML = "";
      if (!a.items || a.items.length === 0) {{
        box.innerHTML = '<div class="item"><span>No artifacts found for this stage.</span></div>';
        return;
      }}
      for (const item of a.items) {{
        const row = document.createElement("div");
        row.className = "item";
        const href = "/artifact/" + encodeURIComponent(item.path);
        row.innerHTML = `<span><a target="_blank" href="${{href}}">${{esc(item.path)}}</a></span><span class="chip">${{esc(item.kind)}}</span>`;
        box.appendChild(row);
      }}
    }}

    async function approveStage() {{
      try {{
        const r = await api("/api/approve", {{ method: "POST" }});
        document.getElementById("msg").textContent = `Approved: ${{r.stage || "stage"}}`;
        await refreshAll();
      }} catch (e) {{
        document.getElementById("msg").textContent = e.message;
      }}
    }}

    async function requestChanges() {{
      const message = document.getElementById("feedback").value.trim();
      if (!message) {{
        document.getElementById("msg").textContent = "Add feedback before requesting changes.";
        return;
      }}
      try {{
        const r = await api("/api/request-changes", {{
          method: "POST",
          body: JSON.stringify({{ message }})
        }});
        document.getElementById("msg").textContent = `Changes requested for: ${{r.stage || "stage"}}`;
        await refreshAll();
      }} catch (e) {{
        document.getElementById("msg").textContent = e.message;
      }}
    }}

    refreshAll().catch(err => {{
      document.getElementById("msg").textContent = err.message;
    }});
    setInterval(() => refreshAll().catch(() => {{}}), 4000);
  </script>
</body>
</html>"""


class Handler(BaseHTTPRequestHandler):
    server_version = "MapleDesignPortal/1.0"

    def _json(self, payload: Dict, code: int = 200) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(code)
        self.send_header("Content-Type", "application/json; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _text(self, data: bytes, ctype: str, code: int = 200) -> None:
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def _state(self) -> PortalState:
        return self.server.portal_state  # type: ignore[attr-defined]

    def _check_token(self) -> Optional[str]:
        token = self.headers.get("X-Maple-Token", "")
        expected = self._state().token()
        if not expected or token != expected:
            return "invalid token"
        return None

    def _read_json_body(self) -> Dict:
        try:
            length = int(self.headers.get("Content-Length", "0"))
        except Exception:
            length = 0
        if length <= 0:
            return {}
        raw = self.rfile.read(length)
        try:
            return json.loads(raw.decode("utf-8"))
        except Exception:
            return {}

    def do_GET(self) -> None:
        if self.path == "/health":
            self._json({"ok": True})
            return
        if self.path == "/" or self.path.startswith("/index.html"):
            token = self._state().token()
            html_doc = render_index(token).encode("utf-8")
            self._text(html_doc, "text/html; charset=utf-8")
            return
        if self.path.startswith("/api/state"):
            self._json(self._state().pipeline())
            return
        if self.path.startswith("/api/artifacts"):
            stage = self._state().pipeline().get("approval_pending") or self._state().pipeline().get("stage", "")
            items = list_artifacts(self._state().root, stage)
            self._json({"items": items, "stage": stage})
            return
        if self.path.startswith("/artifact/"):
            rel = urllib.parse.unquote(self.path[len("/artifact/"):])
            rel = normalize_rel(rel)
            if not rel or not is_allowed_rel(rel):
                self._json({"error": "forbidden path"}, 403)
                return
            file_path = self._state().root / rel
            if not file_path.exists() or not file_path.is_file():
                self._json({"error": "not found"}, 404)
                return
            try:
                data = file_path.read_bytes()
            except Exception:
                self._json({"error": "cannot read file"}, 500)
                return
            self._text(data, content_type_for(file_path))
            return
        self._json({"error": "not found"}, 404)

    def do_POST(self) -> None:
        if self.path == "/api/approve":
            err = self._check_token()
            if err:
                self._json({"error": err}, 403)
                return
            self._json(self._state().approve())
            return
        if self.path == "/api/request-changes":
            err = self._check_token()
            if err:
                self._json({"error": err}, 403)
                return
            body = self._read_json_body()
            message = str(body.get("message", "")).strip()
            if not message:
                self._json({"error": "message required"}, 400)
                return
            self._json(self._state().request_changes(message))
            return
        self._json({"error": "not found"}, 404)

    def log_message(self, fmt: str, *args) -> None:
        return


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="MAPLE design review portal")
    parser.add_argument("--root", required=True, help="repo root path")
    parser.add_argument("--port", type=int, default=4173, help="localhost port")
    parser.add_argument("--token-file", required=True, help="file containing auth token")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    root = pathlib.Path(args.root).resolve()
    token_file = pathlib.Path(args.token_file).resolve()
    state = PortalState(root=root, token_file=token_file)
    server = ThreadingHTTPServer(("127.0.0.1", args.port), Handler)
    server.portal_state = state  # type: ignore[attr-defined]
    print(f"[design-portal] listening on http://127.0.0.1:{args.port}", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    main()
