#!/usr/bin/env python3
"""正常 DOS 玩家路徑的 KOLBOOK 角色建立 UI oracle 擷取器。

在容器內執行：唯讀掛載固定原版 ZIP，於可寫暫存區重建遊戲目錄，用 Xvfb 執行
DOSBox-X，並以 XTEST 送出腳本化按鍵。按鍵不是盲打時序，而是「等畫面靜止到指定
秒數 → 擷取該畫面 → 送下一鍵」，因此原版磁碟解碼時間的變動不會讓序列走偏。

輸出：去重畫格 PNG、畫格時間軸、每個 step 的檢查點，以及
CHARDATA／CHARPICS／SHLDPICS 的前後 SHA-256 收據。

設計限制：
- 原版 ZIP 唯讀，遊戲目錄是每次重建的副本；SHA-256 不符即失敗關閉。
- 只送鍵盤事件與讀畫面，不注入座標、不使用除錯入口、不跳過任何選單層。
- 畫格時間戳是主機牆鐘，逐次不同；可重生的比對對象是檢查點畫面的 PNG SHA-256。
  屬性擲骰畫面由 BIOS tick 播種，逐次不同是預期行為，不得當成擷取失敗。
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import pathlib
import signal
import struct
import subprocess
import sys
import time
import zipfile
import zlib

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))
from x11_key_inject import X11Client  # noqa: E402

EXPECTED_ZIP_SHA256 = "2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46"
CHARACTER_FILES = ("CHARDATA", "CHARPICS", "SHLDPICS")

DOSBOX_CONF = """[sdl]
autolock=false
output=surface
windowposition=0,0
fullscreen=false
[dosbox]
machine=ega
memsize=4
captures={captures}
[cpu]
core=normal
cputype=386
cycles={cycles}
[render]
aspect=false
scaler=none
[autoexec]
mount c {gamedir}
c:
kol.exe
exit
"""


# --------------------------------------------------------------------------
# 情境
# --------------------------------------------------------------------------
# step 欄位：
#   label   檢查點名稱（進入受控狀態、送出按鍵之前的畫面）
#   key     要送出的 keysym 名稱；None 代表只擷取不送鍵
#   settle  畫面需連續多少秒沒有變化才算靜止
#   timeout 等待靜止的上限秒數
#   delay   改為固定等待秒數，用於本來就持續動畫、不會靜止的畫面（例如 TITLE 標題序列）

def step(label: str, key: str | None, settle: float = 2.0, timeout: float = 120.0,
         delay: float | None = None) -> dict:
    return {"label": label, "key": key, "settle": settle, "timeout": timeout, "delay": delay}


# 由 KOL.EXE 正常啟動走到「New Character」名字輸入層的共同前置。
# 依 Apple II 參考卡的開始步驟：標題序列按任意鍵離開 → 書本 → 任意鍵翻到目錄頁 →
# 按項目編號 → 角色選項選單以反白條加 Return 選擇。TITLE 標題序列本身是循環動畫，
# 不會靜止，因此第一步用固定等待。走完 PRELUDE 後停在 New Character 的名字輸入層。
PRELUDE = [
    step("TITLE 標題序列", "space", delay=8.0),
    step("書本封面", "space", settle=2.5),
    step("目錄頁", "5", settle=2.5),
    step("角色選項選單（反白 LIST CHARACTERS）", "Down", settle=2.5),
    step("角色選項選單（反白 CREATE CHARACTER）", "Return", settle=2.0),
]

NAME_LETTERS = "abcdefghijklmnopq"     # 17 個，超過手冊宣稱的 16 上限


def _name_length_steps() -> list[dict]:
    steps: list[dict] = []
    for i, ch in enumerate(NAME_LETTERS):
        typed = NAME_LETTERS[:i] or "（空）"
        steps.append(step(f"名字欄內容 {typed}", ch, settle=1.2, timeout=60.0))
    steps += [
        step(f"送出全部 {len(NAME_LETTERS)} 個字母之後的名字欄", "BackSpace", settle=1.5),
        step("第一次 Backspace 之後", "BackSpace", settle=1.5),
        step("第二次 Backspace 之後", "Escape", settle=1.5),
        step("名字層取消之後", None, settle=3.0),
    ]
    return steps


# 走到角色清單的前置：目錄頁按 5 之後，角色選項選單預設就反白 LIST CHARACTERS，
# 因此不必先按 Down。
LIST_PRELUDE = [
    step("TITLE 標題序列", "space", delay=8.0),
    step("書本封面", "space", settle=2.5),
    step("目錄頁", "5", settle=2.5),
]

SCENARIOS: dict[str, dict] = {
    "list-characters": {
        "description": "從角色選項選單選 LIST CHARACTERS，擷取角色清單畫面，"
                       "再取消回到選單；全程不建立、不刪除、不保存任何角色。",
        "steps": list(LIST_PRELUDE) + [
            step("角色選項選單（反白 LIST CHARACTERS）", "Return", settle=3.0),
            step("角色清單", "Escape", settle=2.5),
            step("角色清單取消之後", None, settle=3.0),
        ],
    },
    "accept-and-save-ranger": {
        "description": "同 accept-and-save，但職業層往下一項（RANGER）。用來把"
                       "「武器 selector → ITEM.LST 記錄索引」的對應再取兩個資料點，"
                       "避免只靠單一職業推廣。",
        "steps": list(PRELUDE) + [
            step("名字層", "r", settle=1.5),
            step("名字欄 r", "a", settle=1.2),
            step("名字欄 ra", "n", settle=1.2),
            step("名字欄 ran", "Return", settle=1.5),
            step("種族層（選 HUMAN）", "Return", settle=2.5),
            step("性別層（選 MALE）", "Return", settle=2.5),
            step("職業層（反白第一項）", "Down", settle=2.0),
            step("職業層（選第二項）", "Return", settle=3.0),
            step("屬性頁（反白 REROLL STATS）", "Down", settle=3.0),
            step("屬性頁（反白 ACCEPT STATS）", "Return", settle=3.0),
            step("圖像層（selector 0）", "Return", settle=3.0),
            step("接受圖像之後", None, settle=4.0),
        ],
    },
    "prelude": {
        "description": "正常玩家路徑走到 New Character 名字輸入層。",
        "steps": list(PRELUDE) + [step("新角色名字層", None, settle=3.0)],
    },
    "name-length": {
        "description": "在名字欄連續送 17 個可見 ASCII 字母，再逐次 Backspace，"
                       "量測此 DOS build 實際接受並顯示的字元數；最後取消不保存。",
        "steps": list(PRELUDE) + _name_length_steps(),
    },
    "blank-name": {
        "description": "名字欄不輸入任何字元直接按 Return，檢查是否前進到種族層；"
                       "隨後逐層取消，不保存任何角色。",
        "steps": list(PRELUDE) + [
            step("空白名字欄", "Return", settle=2.5),
            step("空名送出後的畫面", "Escape", settle=2.5),
            step("第一次取消之後", "Escape", settle=2.0),
            step("第二次取消之後", "Escape", settle=2.0),
            step("第三次取消之後", None, settle=3.0),
        ],
    },
    "escape-layers": {
        "description": "逐層驗證 ESC 的返回目的地：種族→名字、性別→種族、職業→性別、"
                       "屬性→職業，最後退出建立流程且不保存。名字層在取消後會清空，"
                       "因此回到名字層要重新輸入。",
        "steps": list(PRELUDE) + [
            step("名字層", "t", settle=1.5),
            step("名字欄 t", "e", settle=1.2),
            step("名字欄 te", "s", settle=1.2),
            step("名字欄 tes", "t", settle=1.2),
            step("名字欄 test", "Return", settle=1.5),
            step("種族層", "Escape", settle=2.5),
            step("種族層 ESC 之後", "t", settle=1.5),
            step("重新輸入 t", "e", settle=1.2),
            step("重新輸入 te", "s", settle=1.2),
            step("重新輸入 tes", "t", settle=1.2),
            step("重新輸入 test", "Return", settle=1.5),
            step("種族層（第二次）", "Return", settle=2.5),
            step("性別層", "Escape", settle=2.5),
            step("性別層 ESC 之後", "Return", settle=2.5),
            step("種族層（第三次）", "Return", settle=2.5),
            step("職業層", "Escape", settle=2.5),
            step("職業層 ESC 之後", "Return", settle=2.5),
            step("性別層（第三次）", "Return", settle=3.0),
            step("屬性頁", "Escape", settle=3.0),
            step("屬性頁 ESC 之後", "Escape", settle=2.5),
            step("再退一層", "Escape", settle=2.5),
            step("再退一層之二", "Escape", settle=2.5),
            step("再退一層之三", None, settle=3.0),
        ],
    },
    "accept-and-save": {
        "description": "完整走完建立流程：名字 zed、HUMAN、MALE、第一個職業、"
                       "ACCEPT STATS、接受第一張圖像，量測三個角色檔的實際寫入。",
        "steps": list(PRELUDE) + [
            step("名字層", "z", settle=1.5),
            step("名字欄 z", "e", settle=1.2),
            step("名字欄 ze", "d", settle=1.2),
            step("名字欄 zed", "Return", settle=1.5),
            step("種族層（選 HUMAN）", "Return", settle=2.5),
            step("性別層（選 MALE）", "Return", settle=2.5),
            step("職業層（選第一項）", "Return", settle=3.0),
            step("屬性頁（反白 REROLL STATS）", "Down", settle=3.0),
            step("屬性頁（反白 ACCEPT STATS）", "Return", settle=3.0),
            step("圖像層（selector 0）", "Return", settle=3.0),
            step("接受圖像之後", None, settle=4.0),
        ],
    },
    "picture-cycle": {
        "description": "在圖像層驗證 selector 的循環：先向左一次檢查由 0 回繞，"
                       "再連續向右走完整圈並回到起點；全程不接受圖像、不保存。",
        "steps": list(PRELUDE) + [
            step("名字層", "z", settle=1.5),
            step("名字欄 z", "e", settle=1.2),
            step("名字欄 ze", "d", settle=1.2),
            step("名字欄 zed", "Return", settle=1.5),
            step("種族層（選 HUMAN）", "Return", settle=2.5),
            step("性別層（選 MALE）", "Return", settle=2.5),
            step("職業層（選第一項）", "Return", settle=3.0),
            step("屬性頁（反白 REROLL STATS）", "Down", settle=3.0),
            step("屬性頁（反白 ACCEPT STATS）", "Return", settle=3.0),
            step("圖像層 selector 0", "Left", settle=1.0, timeout=60.0),
            step("向左一次之後（預期回繞到最後一張）", "Right", settle=1.0, timeout=60.0),
        ] + [
            step(f"向右第 {i} 次之前", "Right", settle=0.8, timeout=60.0)
            for i in range(1, 82)
        ] + [
            step("走完一圈之後", "Escape", settle=2.5),
            step("圖像層取消之後", None, settle=3.0),
        ],
    },
    "reroll": {
        "description": "走到屬性頁後連續兩次 REROLL STATS，比對屬性數值變化；"
                       "全程不接受屬性、不保存角色。",
        "steps": list(PRELUDE) + [
            step("名字層", "t", settle=1.5),
            step("名字欄 t", "e", settle=1.2),
            step("名字欄 te", "s", settle=1.2),
            step("名字欄 tes", "t", settle=1.2),
            step("名字欄 test", "Return", settle=1.5),
            step("種族層（選 HUMAN）", "Return", settle=2.5),
            step("性別層（選 MALE）", "Return", settle=2.5),
            step("職業層（選第一項）", "Return", settle=3.0),
            step("屬性頁初次擲骰（反白 REROLL STATS）", "Return", settle=3.0),
            step("第一次 REROLL 之後", "Return", settle=3.0),
            step("第二次 REROLL 之後", "Escape", settle=3.0),
            step("屬性頁取消之後", "Escape", settle=2.5),
            step("再取消一層", "Escape", settle=2.5),
            step("再取消一層之二", "Escape", settle=2.5),
            step("再取消一層之三", None, settle=3.0),
        ],
    },
}


def sha256_file(path: pathlib.Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as fh:
        for block in iter(lambda: fh.read(1 << 20), b""):
            h.update(block)
    return h.hexdigest()


def extract_game(zip_path: pathlib.Path, dest: pathlib.Path) -> dict:
    digest = sha256_file(zip_path)
    if digest != EXPECTED_ZIP_SHA256:
        raise SystemExit(f"ZIP SHA-256 不符：{digest}")
    dest.mkdir(parents=True, exist_ok=True)
    count = 0
    with zipfile.ZipFile(zip_path) as zf:
        for info in zf.infolist():
            if info.is_dir():
                continue
            (dest / info.filename.split("/")[-1]).write_bytes(zf.read(info))
            count += 1
    return {"zip_sha256": digest, "entry_count": count}


def character_hashes(gamedir: pathlib.Path) -> dict:
    return {name: {"bytes": (gamedir / name).stat().st_size,
                   "sha256": sha256_file(gamedir / name)}
            for name in CHARACTER_FILES}


# --------------------------------------------------------------------------
# XWD framebuffer → PNG
# --------------------------------------------------------------------------

def xwd_crop_rgb(data: bytes, rect: dict) -> bytes | None:
    """從 Xvfb 的 XWD framebuffer 取指定矩形的 RGB 像素。"""
    if len(data) < 100:
        return None
    f = struct.unpack(">25I", data[:100])
    hdr_size, version, depth = f[0], f[1], f[3]
    pm_w, pm_h = f[4], f[5]
    bpp, bytes_per_line, ncolors = f[11], f[12], f[19]
    if version != 7 or depth != 24 or bpp != 32:
        return None
    x0, y0, w, h = rect["x"], rect["y"], rect["width"], rect["height"]
    if x0 < 0 or y0 < 0 or x0 + w > pm_w or y0 + h > pm_h or w <= 0 or h <= 0:
        return None
    px_off = hdr_size + ncolors * 12
    rows = []
    for y in range(y0, y0 + h):
        base = px_off + y * bytes_per_line + x0 * 4
        line = data[base:base + w * 4]
        if len(line) < w * 4:
            return None
        out = bytearray(w * 3)
        for x in range(w):
            out[x * 3] = line[x * 4 + 2]
            out[x * 3 + 1] = line[x * 4 + 1]
            out[x * 3 + 2] = line[x * 4]
        rows.append(bytes(out))
    return b"".join(rows)


def write_png(path: pathlib.Path, w: int, h: int, rgb: bytes) -> str:
    raw = b"".join(b"\x00" + rgb[y * w * 3:(y + 1) * w * 3] for y in range(h))

    def chunk(tag: bytes, body: bytes) -> bytes:
        payload = tag + body
        return (struct.pack(">I", len(body)) + payload
                + struct.pack(">I", zlib.crc32(payload) & 0xFFFFFFFF))

    png = (b"\x89PNG\r\n\x1a\n"
           + chunk(b"IHDR", struct.pack(">IIBBBBB", w, h, 8, 2, 0, 0, 0))
           + chunk(b"IDAT", zlib.compress(raw, 9))
           + chunk(b"IEND", b""))
    path.write_bytes(png)
    return hashlib.sha256(png).hexdigest()


# --------------------------------------------------------------------------
# 擷取
# --------------------------------------------------------------------------

class Capture:
    def __init__(self, client: X11Client, window: int, fb_path: pathlib.Path,
                 framedir: pathlib.Path, start: float):
        self.client = client
        self.window = window
        self.fb_path = fb_path
        self.framedir = framedir
        self.start = start
        self.seen: dict[str, int] = {}
        self.timeline: list[dict] = []
        self.last_digest: str | None = None

    @staticmethod
    def diff_ratio(a: bytes, b: bytes) -> float:
        """兩張畫面的差異比例（不同的 bit 數／總 bit 數）；長度不同視為完全不同。

        這個比例只用來判斷畫面何時算靜止。跨輪比對用的是
        `tools/compare_character_ui_runs.py` 的 byte 比例，兩者不是同一個量。
        """
        if len(a) != len(b) or not a:
            return 1.0
        xor = int.from_bytes(a, "big") ^ int.from_bytes(b, "big")
        return xor.bit_count() / (len(a) * 8)

    def sample(self) -> tuple[str, int, bytes] | None:
        rect = self.client.get_geometry(self.window)
        try:
            data = self.fb_path.read_bytes()
        except OSError:
            return None
        rgb = xwd_crop_rgb(data, rect)
        if rgb is None:
            return None
        digest = hashlib.sha256(
            rgb + f"{rect['width']}x{rect['height']}".encode()).hexdigest()
        if digest != self.last_digest:
            self.last_digest = digest
            if digest in self.seen:
                index = self.seen[digest]
                self.timeline.append({"t": round(time.time() - self.start, 2),
                                      "frame": index, "repeat": True})
            else:
                index = len(self.seen)
                self.seen[digest] = index
                name = f"frame-{index:03d}.png"
                png_sha = write_png(self.framedir / name,
                                    rect["width"], rect["height"], rgb)
                self.timeline.append({
                    "t": round(time.time() - self.start, 2),
                    "frame": index,
                    "file": f"frames/{name}",
                    "size": [rect["width"], rect["height"]],
                    "png_sha256": png_sha,
                })
        return digest, self.seen[digest], rgb

    def wait_stable(self, settle: float, timeout: float, interval: float,
                    tolerance: float) -> dict:
        """等到畫面連續 settle 秒內的變化都小於 tolerance。

        名字欄的閃爍游標會讓嚴格逐 byte 比對永遠不穩定，因此以差異比例判斷；
        容差只影響「何時送下一鍵」，畫格本身仍以完整雜湊去重保存。
        """
        start_wait = time.time()
        deadline = start_wait + timeout
        last_change = start_wait
        ref = None
        current = None
        while time.time() < deadline:
            got = self.sample()
            if got is not None:
                current = got
                if ref is None or self.diff_ratio(ref, got[2]) > tolerance:
                    ref = got[2]
                    last_change = time.time()
                elif time.time() - last_change >= settle:
                    return {"stable": True, "frame": got[1],
                            "waited_s": round(time.time() - start_wait, 2)}
            time.sleep(interval)
        return {"stable": False, "frame": current[1] if current else None,
                "waited_s": round(timeout, 2)}


def resolve_steps(args) -> tuple[str, str, list[dict]]:
    if args.steps_json:
        spec = json.loads(pathlib.Path(args.steps_json).read_text())
        return (spec.get("name", "adhoc"), spec.get("description", ""),
                [step(s["label"], s.get("key"), s.get("settle", 2.0),
                      s.get("timeout", 120.0), s.get("delay"))
                 for s in spec["steps"]])
    if args.scenario not in SCENARIOS:
        raise SystemExit(f"未知情境：{args.scenario}")
    sc = SCENARIOS[args.scenario]
    return args.scenario, sc["description"], sc["steps"]


def run(args) -> int:
    work = pathlib.Path(args.work)
    gamedir = work / "game"
    fbdir = work / "fb"
    outdir = pathlib.Path(args.out)
    framedir = outdir / "frames"
    for d in (fbdir, framedir, work / "capture"):
        d.mkdir(parents=True, exist_ok=True)

    scenario_name, description, steps = resolve_steps(args)
    inventory = extract_game(pathlib.Path(args.zip), gamedir)
    pre = character_hashes(gamedir)

    conf_text = DOSBOX_CONF.format(captures=str(work / "capture"),
                                   gamedir=str(gamedir), cycles=args.cycles)
    conf_path = work / "dosbox-x.conf"
    conf_path.write_text(conf_text)

    ver = subprocess.run([args.dosbox, "-version"], capture_output=True, text=True)
    version_line = next((l for l in (ver.stdout + ver.stderr).splitlines() if l.strip()), "unknown")

    env = dict(os.environ, HOME=str(work))
    xvfb = subprocess.Popen(
        ["Xvfb", args.display, "-screen", "0", "800x600x24",
         "-fbdir", str(fbdir), "-nolisten", "tcp"],
        stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, env=env)
    fb_path = fbdir / "Xvfb_screen0"
    deadline = time.time() + 10
    while not fb_path.exists() and time.time() < deadline:
        time.sleep(0.1)
    if not fb_path.exists():
        xvfb.terminate()
        raise SystemExit("Xvfb framebuffer 未建立")

    checkpoints: list[dict] = []
    dosbox = None
    client = None
    try:
        with (outdir / "dosbox.log").open("wb") as logfh:
            dosbox = subprocess.Popen(
                [args.dosbox, "-conf", str(conf_path), "-nomenu"],
                stdout=logfh, stderr=subprocess.STDOUT,
                env=dict(env, DISPLAY=args.display, SDL_VIDEODRIVER="x11"))
            start = time.time()

            client = None
            window = None
            while time.time() - start < 30:
                time.sleep(0.5)
                try:
                    c = X11Client(args.display)
                except OSError:
                    continue
                kids = c.query_children()
                if kids:
                    client, window = c, kids[-1]
                    break
                c.close()
            if client is None:
                raise SystemExit("找不到 DOSBox-X 視窗")
            client.set_input_focus(window)

            cap = Capture(client, window, fb_path, framedir, start)
            if args.observe:
                end = time.time() + args.observe
                while time.time() < end:
                    cap.sample()
                    time.sleep(args.interval)
                steps = []
            for index, st in enumerate(steps):
                if st.get("delay") is not None:
                    end = time.time() + st["delay"]
                    while time.time() < end:
                        cap.sample()
                        time.sleep(args.interval)
                    got = cap.sample()
                    result = {"stable": True, "frame": got[1] if got else None,
                              "waited_s": st["delay"]}
                else:
                    result = cap.wait_stable(st["settle"], st["timeout"],
                                             args.interval, args.tolerance)
                checkpoints.append({
                    "index": index,
                    "label": st["label"],
                    "key": st["key"],
                    "settle_s": st["settle"],
                    "fixed_delay_s": st.get("delay"),
                    "reached_stable": result["stable"],
                    "t": round(time.time() - start, 2),
                    "frame": result["frame"],
                })
                if st["key"] is not None:
                    client.fake_key(st["key"])
                    time.sleep(args.post_key_delay)
    finally:
        if client is not None:
            client.close()
        if dosbox is not None and dosbox.poll() is None:
            dosbox.send_signal(signal.SIGTERM)
            try:
                dosbox.wait(timeout=15)
            except subprocess.TimeoutExpired:
                dosbox.kill()
        xvfb.terminate()
        try:
            xvfb.wait(timeout=10)
        except subprocess.TimeoutExpired:
            xvfb.kill()

    post = character_hashes(gamedir)
    changed = [n for n in CHARACTER_FILES if pre[n]["sha256"] != post[n]["sha256"]]
    if args.save_character_files:
        # 執行後的角色檔含原版模板 bytes，只作本機離線比對，不得提交。
        savedir = outdir / "character-files"
        savedir.mkdir(parents=True, exist_ok=True)
        for name in CHARACTER_FILES:
            (savedir / name).write_bytes((gamedir / name).read_bytes())
    frame_index = {f["frame"]: f for f in cap.timeline if "png_sha256" in f}

    receipt = {
        "tool": "tools/dosbox_character_ui_oracle.py",
        "scenario": scenario_name,
        "description": description,
        "generated_utc": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "dosbox_x_version": version_line,
        "dosbox_conf_sha256": hashlib.sha256(conf_text.encode()).hexdigest(),
        "dosbox_cycles": args.cycles,
        "input_method": "X11 XTEST FakeInput，畫面靜止後逐鍵送出",
        "sample_interval_s": args.interval,
        "stability_tolerance": args.tolerance,
        "original_zip": inventory,
        "character_files_before": pre,
        "character_files_after": post,
        "character_files_changed": changed,
        "character_files_saved": bool(args.save_character_files),
        "distinct_frames": len(cap.seen),
        "checkpoints": [
            dict(cp, file=frame_index.get(cp["frame"], {}).get("file"),
                 png_sha256=frame_index.get(cp["frame"], {}).get("png_sha256"))
            for cp in checkpoints
        ],
        "timeline": cap.timeline,
    }
    (outdir / "receipt.json").write_text(
        json.dumps(receipt, ensure_ascii=False, indent=2) + "\n")
    print(json.dumps({
        "scenario": scenario_name,
        "distinct_frames": len(cap.seen),
        "checkpoints": len(checkpoints),
        "unstable": [c["label"] for c in checkpoints if not c["reached_stable"]],
        "changed": changed,
    }, ensure_ascii=False))
    return 0


def main() -> int:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--zip", default="/in/kol.zip")
    p.add_argument("--work", default="/work")
    p.add_argument("--out", default="/out")
    p.add_argument("--scenario", default="explore")
    p.add_argument("--steps-json", default=None, help="外部 step 定義檔，覆蓋 --scenario")
    p.add_argument("--cycles", default="fixed 3000")
    p.add_argument("--interval", type=float, default=0.3)
    p.add_argument("--post-key-delay", type=float, default=0.4)
    p.add_argument("--save-character-files", action="store_true",
                   help="把執行後的 CHARDATA／CHARPICS／SHLDPICS 複製到輸出目錄（僅本機比對用）")
    p.add_argument("--tolerance", type=float, default=0.002,
                   help="判定畫面靜止時允許的差異比例（閃爍游標用）")
    p.add_argument("--observe", type=float, default=0.0,
                   help="只觀察並錄畫格 N 秒，不送任何按鍵")
    p.add_argument("--display", default=":77")
    p.add_argument("--dosbox", default="/opt/dosbox-x/bin/dosbox-x")
    return run(p.parse_args())


if __name__ == "__main__":
    sys.exit(main())
