#!/usr/bin/env python3
"""未完成項的權威是 `docs/worklist.json`；WORKLIST.md 的那一節由本工具產生。

    tools/run_python_tool.sh tools/worklist.py verify    # 逐條核實
    tools/run_python_tool.sh tools/worklist.py render    # 回寫 WORKLIST.md
    tools/run_python_tool.sh tools/worklist.py selftest  # 機制的正反對照

markdown 的 `- [ ]` 會長出過期斷言：東西做好了而沒有人回頭改那一條，於是清單上
留著一句自信的「還沒接」，被拿去當「還剩多少」的依據。這裡把「這條還成立嗎」
寫成資料的一部分，讓機器每一輪都問。

`verify` 為真＝**這一條仍然未完成**；為假就是東西做好了而條目沒改，命令回非零。

`present`／`absent` 的 `paths` 是 **OR**：任一個檔案命中就算命中。條目如果要求
「每一個都要有」（例如每份規格都要連回說明書頁碼），用 `required_in_all`——
OR 會讓「做完其中一份」就把整條判成完成。
"""

import argparse
import fnmatch
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DATA_PATH = ROOT / "docs/worklist.json"
WORKLIST_MD = ROOT / "WORKLIST.md"

BEGIN = "<!-- worklist:begin 由 tools/worklist.py render 產生，不要手改 -->"
END = "<!-- worklist:end -->"

CONTEXT_MD = ROOT / "CONTEXT.md"
STATS_BEGIN = "<!-- worklist:stats:begin 由 tools/worklist.py render 產生，不要手改 -->"
STATS_END = "<!-- worklist:stats:end -->"

# 掃得動的副檔名。二進位與素材不看。
SCAN_SUFFIXES = {".go", ".py", ".sh", ".md", ".json", ".txt", ".work", ".mod"}


def is_test_file(path: Path) -> bool:
    """測試檔不算數。

    測試本來就會提到還沒接上的東西——為了釘住將來的行為，或為了測那個資料結構
    本身。把它們算進來，`absent` 會因為測試裡有一行呼叫就判成「已經做了」，
    真缺口就這樣被蓋掉。條目問的是「**產品程式碼**做了這件事沒有」。
    """
    name = path.name
    return (
        name.endswith("_test.go")
        or name.endswith("_test.py")
        or name.startswith("test_")
    )


def iter_files(root: Path, targets):
    """展開 `paths`：檔案就是它自己，目錄就遞迴，測試檔一律跳過。"""
    for target in targets:
        p = root / target
        if p.is_file():
            if not is_test_file(p):
                yield p
        elif p.is_dir():
            for f in sorted(p.rglob("*")):
                if f.is_file() and f.suffix in SCAN_SUFFIXES and not is_test_file(f):
                    yield f


def find_hit(root: Path, verify: dict):
    """回 pattern 命中的第一個檔案（相對路徑），沒有就回 None。

    **`re.MULTILINE` 是必要的**：`scan` 是逐行比對，`^` 在那裡是「行首」；這一支
    讀的是整個檔案，不加旗標的話同一個 `^` 會變成「檔案開頭」，於是綁章節標題那
    種 pattern 在 `scan` 認領得到、在這裡卻永遠找不到——**同一個 pattern 兩種語意**，
    而且兩邊都不會報錯。
    """
    expr = re.compile(verify["pattern"], re.MULTILINE)
    for f in iter_files(root, verify["paths"]):
        try:
            text = f.read_text(encoding="utf-8", errors="ignore")
        except OSError:
            continue
        if expr.search(text):
            return str(f.relative_to(root))
    return None


def still_open(item: dict, root: Path = ROOT):
    """回 `(仍未完成, 說明)`。

    「仍未完成」為 False 代表這一條的訊號已經消失——東西可能做好了而條目沒改，
    要回去看那一條。
    """
    v = item["verify"]
    kind = v["kind"]

    if kind == "manual":
        return True, "要人判（沒有機器訊號）"

    if kind == "json_len":
        target = root / v["path"]
        if not target.exists():
            return True, f'{v["path"]} 還不存在'
        payload = json.loads(target.read_text(encoding="utf-8"))
        for key in v["field"].split("."):
            payload = payload[key]
        n = len(payload)
        return n <= v["max"], f'{v["path"]} 的 {v["field"]} 有 {n} 項（門檻 {v["max"]}）'

    if kind == "required_in_all":
        # pattern 必須出現在**每一個** path。少一個就是還沒做完，並指出是哪一個。
        missing = []
        for target in v["paths"]:
            f = root / target
            if not f.exists():
                missing.append(f"{target}（不存在）")
                continue
            if not re.search(v["pattern"], f.read_text(encoding="utf-8", errors="ignore"), re.MULTILINE):
                missing.append(target)
        if missing:
            head = "、".join(missing[:3])
            more = f" 等 {len(missing)} 個" if len(missing) > 3 else ""
            return True, f"還缺 {head}{more}"
        return False, f'{len(v["paths"])} 個都有了'

    where = find_hit(root, v)
    if kind == "present":
        if where:
            return True, f"自承還在 {where}"
        return False, f'找不到 {v["pattern"]}'
    if kind == "absent":
        if where:
            return False, f"已經出現在 {where}"
        return True, "還沒出現"

    raise ValueError(f'不認得的 verify.kind：{kind}')


def load():
    return json.loads(DATA_PATH.read_text(encoding="utf-8"))


def check_schema(data: dict) -> list:
    """schema 自己也要被看著：欄位缺了、id 撞號、kind 寫錯都要當場失敗。"""
    problems = []
    seen = set()
    layers = set(data.get("layers", {}))
    for item in data["items"]:
        ident = item.get("id", "(無 id)")
        for field in ("id", "layer", "title", "body", "acceptance", "verify"):
            if field not in item:
                problems.append(f"{ident}：缺欄位 {field}")
        if ident in seen:
            problems.append(f"{ident}：id 重複")
        seen.add(ident)
        if item.get("layer") not in layers:
            problems.append(f'{ident}：layer「{item.get("layer")}」不在 layers 表裡')
        v = item.get("verify", {})
        kind = v.get("kind")
        if kind not in {"present", "absent", "json_len", "manual", "required_in_all"}:
            problems.append(f"{ident}：verify.kind = {kind}")
        elif kind in {"present", "absent", "required_in_all"}:
            if not v.get("paths") or not v.get("pattern"):
                problems.append(f"{ident}：{kind} 要有 paths 與 pattern")
        elif kind == "json_len":
            if not v.get("path") or not v.get("field") or "max" not in v:
                problems.append(f"{ident}：json_len 要有 path、field 與 max")
        elif kind == "manual" and not v.get("note"):
            problems.append(f"{ident}：manual 要在 note 寫明將來接上時該綁什麼")
        # 已完成項跑的是回歸檢查（訊號還在＝仍然完成），而 `present` 的語意是
        # 「找得到＝仍未完成」——兩者剛好相反，用上去每一輪都會誤報回歸。
        if item.get("status") == "done" and kind == "present":
            problems.append(
                f"{ident}：已完成項不能用 present（語意相反，會每輪誤報回歸）；"
                "改用 absent 或 required_in_all，讓「東西還在」等於「仍然完成」")
    return problems


# 產品碼裡「這件事還沒做」的自承說法。掃出來的每一行都要有人認領：
# 不是某個條目的 verify 綁著它，就是在 `acknowledged` 裡寫明為什麼它不是待辦。
# 這組字要寬。窄的 pattern 掃得再廣也沒有用——2026-09-10 量過一次：只抓
# 「還沒解／還沒實作／…」時，未認領的來源裡命中 47 行，換成這一組是 242 行，
# 中間那 195 行不是不存在，是**過濾器有洞**。寬了會抓到不是待辦的句子，
# 那正是 `acknowledged` 存在的理由：寧可多寫幾條理由，也不要靜靜漏掉。
ADMISSION = re.compile(
    r"還沒|尚未|仍待|待驗|待查|待確認|待決定|未完成|未實作|未接|未驗|未解|TODO|FIXME|XXX|HACK")

# 掃描範圍：程式碼（產品碼與工具）與文件。測試檔不掃——它們提到還沒接的東西是
# 正常的。掃得寬一點沒關係：掃到的每一行不是有條目綁著，就是要在 `acknowledged`
# 寫明為什麼不是待辦，所以「為什麼不看某一類」也會留在資料裡，而不是藏在這支
# 程式的範圍設定裡。
SCAN_CODE = (("cmd", ".go"), ("internal", ".go"), ("tools", ".py"), ("tools", ".sh"))
SCAN_DOCS = ("docs",)

# 這一支自己不掃：它的內容就是那組 pattern 的定義與測試用的假資料。
SELF = "tools/worklist.py"


def scan_admissions(root: Path = ROOT) -> list:
    """回所有自承還沒做的行：`(相對路徑, 行號, 內容)`。"""
    found = []
    targets = []
    for base, suffix in SCAN_CODE:
        d = root / base
        if d.is_dir():
            targets += [f for f in sorted(d.rglob("*" + suffix)) if not is_test_file(f)]
    for base in SCAN_DOCS:
        d = root / base
        if d.is_dir():
            targets += sorted(d.rglob("*.md"))
    targets += sorted(root.glob("*.md"))

    for f in targets:
        rel = str(f.relative_to(root))
        if rel == SELF:
            continue
        for n, line in enumerate(f.read_text(encoding="utf-8", errors="ignore").split("\n"), 1):
            if ADMISSION.search(line):
                found.append((rel, n, line.strip()))
    return found


def claimed_by(data: dict, path: str, line: str, root: Path = ROOT) -> str | None:
    """這一行有沒有被某個條目的 verify 綁著？回那個條目的 id。

    判定要**精確到這一行**：只看 `paths` 涵不涵蓋這個檔案是不夠的——一個綁
    `internal/` 的條目會把整棵樹都認領走，於是這支掃描永遠是乾淨的而什麼都沒看。
    """
    for item in data["items"]:
        v = item["verify"]
        if v["kind"] not in {"present", "absent", "required_in_all"}:
            continue
        covered = False
        for target in v["paths"]:
            if path == target or path.startswith(target.rstrip("/") + "/"):
                covered = True
                break
        if not covered:
            continue
        if re.search(v["pattern"], line):
            return item["id"]
    return None


def acknowledged_by(data: dict, path: str, line: str) -> str | None:
    """這一行有沒有在 `acknowledged` 裡被寫明「不是待辦」？回那條的理由。

    `file` 以 `/` 結尾就是整個目錄，否則是單一檔案。省略 `match` 代表整檔認領——
    用在流水帳與研究筆記那種「記的是當時狀態、不是現在的缺口」的檔案上。
    """
    for ack in data.get("acknowledged", []):
        target = ack["file"]
        if target.endswith("/"):
            if not path.startswith(target):
                continue
        elif "*" in target or "?" in target:
            if not fnmatch.fnmatch(path, target):
                continue
        elif target != path:
            continue
        if "match" not in ack:
            return ack["why"]
        if re.search(ack["match"], line):
            return ack["why"]
    return None


def orphan_admissions(data: dict, root: Path = ROOT) -> list:
    """回「沒有人認領」的自承行。"""
    out = []
    for path, n, line in scan_admissions(root):
        if claimed_by(data, path, line, root) or acknowledged_by(data, path, line):
            continue
        out.append((path, n, line))
    return out


def cmd_scan(_args) -> int:
    data = load()
    rows = scan_admissions()
    orphans = orphan_admissions(data)
    claimed = len(rows) - len(orphans)
    print(f"程式碼與文件裡的自承共 {len(rows)} 行；{claimed} 行有人認領，{len(orphans)} 行沒有。")
    if orphans:
        print()
        print("★ 下列自承沒有條目綁著，也沒有寫明為什麼不是待辦：")
        for path, n, line in orphans:
            print(f"  {path}:{n}")
            print(f"      {line[:110]}")
        print()
        print("處置：替它開一條 `items`（verify 綁這一行），")
        print("或在 `acknowledged` 寫明它為什麼不是待辦。")
        return 1
    print("每一行自承都有人認領。")
    return 0


def cmd_verify(_args) -> int:
    data = load()
    problems = check_schema(data)
    if problems:
        print("schema 有問題：")
        for p in problems:
            print(f"  ✗ {p}")
        return 2

    stale, manual, regressed = [], [], []
    print(f'{"id":<34} {"狀態":<14} 依據')
    print("-" * 100)
    for item in data["items"]:
        if item.get("status") == "done":
            continue
        open_, why = still_open(item)
        mark = "仍未完成" if open_ else "★ 可能已完成"
        print(f'{item["id"]:<34} {mark:<14} {why}')
        if not open_:
            stale.append((item["id"], why))
        if item["verify"]["kind"] == "manual":
            manual.append(item["id"])

    # 已完成的也要跑，判讀相反：訊號還在＝仍然完成，訊號不見＝回歸了。
    # 跳過它們等於在 note 裡寫「留著當回歸訊號」卻沒有任何機制在看——
    # 那比沒有 verify 更糟，它給人「有在看」的假象。
    done = [i for i in data["items"] if i.get("status") == "done"]
    if done:
        print()
        print("已核實完成（跑的是回歸檢查：訊號還在＝仍然完成）")
        print("-" * 100)
        for item in done:
            if item["verify"]["kind"] == "manual":
                print(f'{item["id"]:<34} {"要人判":<14} 完成後沒有訊號可綁')
                continue
            gone, why = still_open(item)
            mark = "★ 可能回歸" if gone else "仍然完成"
            print(f'{item["id"]:<34} {mark:<14} {why}')
            if gone:
                regressed.append((item["id"], why))

    total = len(data["items"]) - len(done)
    print()
    print(f"未完成 {total} 條；已核實完成 {len(done)} 條（都跑過回歸檢查）。")
    print(f"其中 {len(manual)} 條沒有機器訊號，要人判：{'、'.join(manual) if manual else '無'}")
    if stale:
        print()
        print("★ 下列條目的訊號不見了，回去看它是不是做完了：")
        for ident, why in stale:
            print(f"  - {ident}：{why}")
    if regressed:
        print()
        print("★ 下列已完成項的訊號不見了，可能被改回去了：")
        for ident, why in regressed:
            print(f"  - {ident}：{why}")

    orphans = orphan_admissions(data)
    if orphans:
        print()
        print(f"★ 有 {len(orphans)} 行自承沒有條目認領（跑 scan 看細節）：")
        for path, n, line in orphans[:5]:
            print(f"  - {path}:{n}　{line[:70]}")
        if len(orphans) > 5:
            print(f"  …另外 {len(orphans)-5} 行")

    desync = out_of_sync(data)
    if desync:
        print()
        print("★ 下列產生出來的內容跟 docs/worklist.json 不同步，跑一次 render：")
        for d_ in desync:
            print(f"  - {d_}")
    else:
        print()
        print("WORKLIST.md 與 CONTEXT.md 的產生區塊都與 JSON 同步。")
    return 1 if (stale or regressed or desync or orphans) else 0


def render_markdown(data: dict) -> str:
    lines = [
        BEGIN,
        "",
        "## 未完成項（權威是 `docs/worklist.json`）",
        "",
        "這一節由 `tools/worklist.py render` 產生，**不要手改**；改 `docs/worklist.json`。",
        "每一條掛一個 `verify`：跑起來為真＝這一條仍然未完成，為假＝訊號不見了，",
        "要回去看它是不是做完了。核實方式：",
        "",
        "```",
        "tools/run_python_tool.sh tools/worklist.py verify",
        "```",
        "",
    ]
    for key, label in data["layers"].items():
        items = [i for i in data["items"] if i["layer"] == key and i.get("status") != "done"]
        if not items:
            continue
        lines.append(f"### {label}")
        lines.append("")
        for item in items:
            lines.append(f'- **{item["id"]}** — {item["title"]}')
            for para in item["body"].split("\n"):
                lines.append(f"  {para}" if para else "")
            if item.get("blocked_by"):
                lines.append(f'  卡在：{item["blocked_by"]}')
            lines.append(f'  完成條件：{item["acceptance"]}')
            v = item["verify"]
            if v["kind"] == "manual":
                lines.append(f'  核實：要人判——{v["note"]}')
            elif v["kind"] == "json_len":
                lines.append(f'  核實：`{v["path"]}` 的 `{v["field"]}` ≤ {v["max"]} 就是還沒完')
            elif v["kind"] == "required_in_all":
                where = "、".join(f"`{p}`" for p in v["paths"])
                lines.append(f'  核實：{where} **每一個**都要有 `{v["pattern"]}`，少一個就是還沒完')
            else:
                where = "、".join(f"`{p}`" for p in v["paths"])
                verb = "找得到" if v["kind"] == "present" else "找不到"
                lines.append(f'  核實：{where} {verb} `{v["pattern"]}` 就是還沒完')
            lines.append("")
        lines.append("")

    done = [i for i in data["items"] if i.get("status") == "done"]
    if done:
        lines.append("### 本輪核實為已完成（原本掛在未完成清單上）")
        lines.append("")
        for item in done:
            lines.append(f'- **{item["id"]}** — {item["title"]}：{item.get("resolved", "已完成")}')
        lines.append("")
    lines.append(END)
    return "\n".join(lines)


def stats(data: dict) -> dict:
    """把清單的數字算出來。

    這些數字手寫過一次就會過期——`CONTEXT.md` 的「45 條未完成」在五條 `- [~]`
    併進來之後就不對了，而沒有任何東西會發現。算出來的數字不會。
    """
    items = data["items"]
    done = [i for i in items if i.get("status") == "done"]
    open_ = [i for i in items if i.get("status") != "done"]
    manual = [i for i in open_ if i["verify"]["kind"] == "manual"]
    kinds = {}
    for i in open_:
        kinds[i["verify"]["kind"]] = kinds.get(i["verify"]["kind"], 0) + 1
    return {"total": len(items), "open": len(open_), "done": len(done),
            "manual": len(manual), "kinds": kinds}


def render_stats(data: dict) -> str:
    st = stats(data)
    kinds = "、".join(f"`{k}` {n}" for k, n in sorted(st["kinds"].items(), key=lambda x: -x[1]))
    return "\n".join([
        STATS_BEGIN,
        f'- **未完成項的權威是 `docs/worklist.json`**，不是 `WORKLIST.md` 的 checkbox。',
        f'  核實用 `tools/run_python_tool.sh tools/worklist.py verify`；`WORKLIST.md` 的',
        f'  那一節與這幾行數字都由 `render` 產生。',
        f'- 目前 **{st["open"]} 條未完成**（{st["manual"]} 條沒有機器訊號、標 `manual`），',
        f'  另有 {st["done"]} 條已核實完成——後者跑的是回歸檢查，訊號不見了就開口。',
        f'  未完成項的 verify 種類：{kinds}。',
        STATS_END,
    ])


def out_of_sync(data: dict) -> list:
    """回報哪些產生出來的區塊與 `docs/worklist.json` 不同步。

    光是「數字由 script 產生」還不夠：改了 JSON 卻忘記跑 `render`，檔案裡留下的
    仍然是舊數字，而且看起來一樣理直氣壯。所以 `verify` 每次順便比一次。
    """
    stale = []
    md = WORKLIST_MD.read_text(encoding="utf-8")
    if BEGIN in md and END in md:
        current = BEGIN + md.split(BEGIN, 1)[1].split(END, 1)[0] + END
        if current.strip() != render_markdown(data).strip():
            stale.append("WORKLIST.md 的未完成項一節")
    else:
        stale.append("WORKLIST.md 沒有 worklist:begin 標記")

    ctx = CONTEXT_MD.read_text(encoding="utf-8")
    if STATS_BEGIN in ctx and STATS_END in ctx:
        current = STATS_BEGIN + ctx.split(STATS_BEGIN, 1)[1].split(STATS_END, 1)[0] + STATS_END
        if current.strip() != render_stats(data).strip():
            stale.append("CONTEXT.md 的統計區塊")
    else:
        stale.append("CONTEXT.md 沒有 worklist:stats 標記")
    return stale


def cmd_render(args) -> int:
    data = load()
    problems = check_schema(data)
    if problems:
        print("schema 有問題，先修好再 render：")
        for p in problems:
            print(f"  ✗ {p}")
        return 2

    block = render_markdown(data)
    if args.stdout:
        print(block)
        return 0

    text = WORKLIST_MD.read_text(encoding="utf-8")
    if BEGIN in text and END in text:
        head, rest = text.split(BEGIN, 1)
        _, tail = rest.split(END, 1)
        text = head + block + tail
    else:
        text = text.rstrip("\n") + "\n\n" + block + "\n"
    WORKLIST_MD.write_text(text, encoding="utf-8")
    print(f"已回寫 {WORKLIST_MD.relative_to(ROOT)}")

    # CONTEXT.md 的那幾個數字同一份來源；手寫過一次就會過期。
    ctx = CONTEXT_MD.read_text(encoding="utf-8")
    if STATS_BEGIN in ctx and STATS_END in ctx:
        head, rest = ctx.split(STATS_BEGIN, 1)
        _, tail = rest.split(STATS_END, 1)
        CONTEXT_MD.write_text(head + render_stats(data) + tail, encoding="utf-8")
        print(f"已回寫 {CONTEXT_MD.relative_to(ROOT)} 的統計區塊")
    else:
        print(f"⚠ {CONTEXT_MD.relative_to(ROOT)} 沒有 worklist:stats 標記，數字沒同步")
    return 0


def cmd_selftest(_args) -> int:
    """正反對照：只驗「它印出仍未完成」證明不了機制有在看。

    空的檔案、寫錯的路徑、永遠為真的 pattern 都會印出一樣的好消息。所以每一種
    kind 都先確認**訊號在時報未完成**，再把訊號拿掉、確認它真的開口。
    """
    import shutil
    import tempfile

    failures = []

    def expect(name, got, want):
        if got != want:
            failures.append(f"{name}：得到 {got}，預期 {want}")

    with tempfile.TemporaryDirectory() as tmp:
        sandbox = Path(tmp)
        (sandbox / "internal").mkdir()
        code = sandbox / "internal/thing.go"
        test = sandbox / "internal/thing_test.go"

        # --- present：自承註解在＝仍未完成 -------------------------------
        code.write_text("// 這一段還沒做\nfunc A() {}\n", encoding="utf-8")
        v = {"kind": "present", "paths": ["internal"], "pattern": "還沒做"}
        expect("present／訊號在", still_open({"verify": v}, sandbox)[0], True)
        code.write_text("func A() {}\n", encoding="utf-8")
        expect("present／訊號拿掉", still_open({"verify": v}, sandbox)[0], False)

        # --- absent：東西還沒出現＝仍未完成 ------------------------------
        v = {"kind": "absent", "paths": ["internal"], "pattern": r"func Wanted\("}
        expect("absent／還沒出現", still_open({"verify": v}, sandbox)[0], True)
        code.write_text("func Wanted() {}\n", encoding="utf-8")
        expect("absent／出現了", still_open({"verify": v}, sandbox)[0], False)

        # --- 測試檔不算數：只有 _test.go 提到，仍然是「還沒做」 ---------
        code.write_text("package thing\n", encoding="utf-8")
        test.write_text("func TestWanted() { Wanted() }\n", encoding="utf-8")
        expect("absent／只有測試檔提到", still_open({"verify": v}, sandbox)[0], True)

        # --- json_len：項數過門檻就閉嘴 ----------------------------------
        payload = sandbox / "count.json"
        payload.write_text(json.dumps({"a": {"b": [1, 2]}}), encoding="utf-8")
        v = {"kind": "json_len", "path": "count.json", "field": "a.b", "max": 2}
        expect("json_len／還沒過門檻", still_open({"verify": v}, sandbox)[0], True)
        payload.write_text(json.dumps({"a": {"b": [1, 2, 3]}}), encoding="utf-8")
        expect("json_len／過門檻", still_open({"verify": v}, sandbox)[0], False)

        # --- 路徑寫錯不能是好消息：present 找不到就是「可能已完成」 ------
        v = {"kind": "present", "paths": ["nonexistent"], "pattern": "還沒做"}
        expect("present／路徑寫錯", still_open({"verify": v}, sandbox)[0], False)

        # --- manual 一律回仍未完成 ---------------------------------------
        expect("manual", still_open({"verify": {"kind": "manual", "note": "x"}}, sandbox)[0], True)

        # --- required_in_all：OR 會讓「做完一份」誤判成整條完成 -----------
        a = sandbox / "internal/a.md"
        b = sandbox / "internal/b.md"
        a.write_text("有連結 docs/manual/x\n", encoding="utf-8")
        b.write_text("沒有\n", encoding="utf-8")
        v = {"kind": "required_in_all", "paths": ["internal/a.md", "internal/b.md"],
             "pattern": "docs/manual/"}
        expect("required_in_all／只做了一份", still_open({"verify": v}, sandbox)[0], True)
        # 同一組檔案用 absent 會誤判成完成——這正是要避免的
        v_or = {"kind": "absent", "paths": ["internal"], "pattern": "docs/manual/"}
        expect("absent／OR 的誤判（對照組）", still_open({"verify": v_or}, sandbox)[0], False)
        b.write_text("也有 docs/manual/y\n", encoding="utf-8")
        expect("required_in_all／兩份都做了", still_open({"verify": v}, sandbox)[0], False)
        a.unlink(); b.unlink()

        # --- 已完成項的回歸檢查：東西還在＝仍然完成，被拿掉＝要開口 -------
        code.write_text("func Wanted() {}\n", encoding="utf-8")
        done_v = {"kind": "absent", "paths": ["internal"], "pattern": r"func Wanted\("}
        expect("done／東西還在", still_open({"verify": done_v}, sandbox)[0], False)
        code.write_text("package thing\n", encoding="utf-8")
        expect("done／東西被拿掉", still_open({"verify": done_v}, sandbox)[0], True)

        shutil.rmtree(sandbox, ignore_errors=True)

    # --- 統計數字要跟著資料走，不是寫死 -----------------------------------
    probe = {"layers": {"a": "A"}, "items": [
        {"id": "x", "layer": "a", "title": "t", "body": "b", "acceptance": "a",
         "verify": {"kind": "manual", "note": "n"}},
        {"id": "y", "layer": "a", "title": "t", "body": "b", "acceptance": "a",
         "status": "done", "verify": {"kind": "manual", "note": "n"}},
    ]}
    st = stats(probe)
    if (st["total"], st["open"], st["done"], st["manual"]) != (2, 1, 1, 1):
        failures.append(f"stats：算錯了 {st}")
    block = render_stats(probe)
    if "1 條未完成" not in block or "1 條已核實完成" not in block:
        failures.append("render_stats：數字沒有進到輸出")
    probe["items"].append({"id": "z", "layer": "a", "title": "t", "body": "b",
                           "acceptance": "a", "verify": {"kind": "manual", "note": "n"}})
    if "2 條未完成" not in render_stats(probe):
        failures.append("render_stats：加一條之後數字沒有跟著動——那就是寫死的")

    # --- `^` 在 scan（逐行）與 verify（整檔）要是同一個意思 ---------------
    with tempfile.TemporaryDirectory() as tmp:
        box = Path(tmp)
        (box / "docs").mkdir()
        (box / "docs" / "s.md").write_text("開頭一行\n## 尚未關閉的來源\n內容\n", encoding="utf-8")
        v = {"kind": "present", "paths": ["docs"], "pattern": "^## .*尚未"}
        if not find_hit(box, v):
            failures.append("find_hit：`^` 應該是行首（少了 re.MULTILINE 就只匹配檔案開頭）")
        item = {"verify": v}
        if not still_open(item, box)[0]:
            failures.append("still_open：綁章節標題的 present 應該回仍未完成")

    # --- scan：孤兒要被抓到，認領之後要閉嘴 -------------------------------
    with tempfile.TemporaryDirectory() as tmp:
        box = Path(tmp)
        (box / "internal").mkdir()
        (box / "internal" / "x.go").write_text("// 這一段還沒實作\n", encoding="utf-8")
        (box / "internal" / "x_test.go").write_text("// 測試裡的還沒實作不算\n", encoding="utf-8")
        rows = scan_admissions(box)
        if len(rows) != 1:
            failures.append(f"scan_admissions：掃到 {len(rows)} 行，應該只有產品碼那一行")
        empty = {"layers": {}, "items": [], "acknowledged": []}
        if len(orphan_admissions(empty, box)) != 1:
            failures.append("orphan_admissions：沒有人認領時應該回報一行")
        claimed = {"layers": {}, "acknowledged": [], "items": [{
            "id": "c", "layer": "a", "title": "t", "body": "b", "acceptance": "a",
            "verify": {"kind": "present", "paths": ["internal/x.go"], "pattern": "還沒實作"}}]}
        if orphan_admissions(claimed, box):
            failures.append("orphan_admissions：條目綁著了還說沒人認領")
        # 綁到別的檔案不算認領——不然一條就能把整棵樹認領走。
        wrong = json.loads(json.dumps(claimed))
        wrong["items"][0]["verify"]["paths"] = ["internal/other.go"]
        if not orphan_admissions(wrong, box):
            failures.append("orphan_admissions：綁到別的檔案竟然也算認領")
        acked = {"layers": {}, "items": [], "acknowledged": [
            {"file": "internal/x.go", "match": "還沒實作", "why": "說明用"}]}
        if orphan_admissions(acked, box):
            failures.append("orphan_admissions：acknowledged 之後不該再報")
        # 整檔認領（省略 match）、目錄前綴、萬用字元三種寫法都要成立。
        for name, ack in (
            ("整檔", {"file": "internal/x.go", "why": "整檔"}),
            ("目錄", {"file": "internal/", "why": "整個目錄"}),
            ("萬用字元", {"file": "internal/x*.go", "why": "一批檔案"}),
        ):
            probe = {"layers": {}, "items": [], "acknowledged": [ack]}
            if orphan_admissions(probe, box):
                failures.append(f"acknowledged／{name}：這種寫法沒有生效")
        # 反面：不相干的目錄與不相符的萬用字元都不該認領走。
        for name, ack in (
            ("別的目錄", {"file": "cmd/", "why": "x"}),
            ("不符的萬用字元", {"file": "internal/z*.go", "why": "x"}),
        ):
            probe = {"layers": {}, "items": [], "acknowledged": [ack]}
            if not orphan_admissions(probe, box):
                failures.append(f"acknowledged／{name}：不該認領卻認領了")

    # --- schema 檢查自己也要有反例 ---------------------------------------
    bad = {"layers": {"a": "A"}, "items": [{"id": "x", "layer": "b", "verify": {"kind": "nope"}}]}
    if not check_schema(bad):
        failures.append("check_schema：壞掉的 schema 竟然通過")

    # done ＋ present 的語意是反的，schema 要擋下來。
    flipped = {"layers": {"a": "A"}, "items": [{
        "id": "x", "layer": "a", "title": "t", "body": "b", "acceptance": "a",
        "status": "done",
        "verify": {"kind": "present", "paths": ["x"], "pattern": "y"}}]}
    if not any("語意相反" in m for m in check_schema(flipped)):
        failures.append("check_schema：已完成項用 present 竟然過了")
    flipped["items"][0]["verify"]["kind"] = "absent"
    if check_schema(flipped):
        failures.append("check_schema：已完成項用 absent 應該要過")

    real = load()
    if check_schema(real):
        failures.append("check_schema：docs/worklist.json 自己沒過")

    # 同步檢查：現況要是同步的；資料改一條之後必須報不同步。
    if out_of_sync(real):
        failures.append("out_of_sync：現在應該是同步的")
    moved = json.loads(json.dumps(real))
    moved["items"].append({"id": "selftest-probe", "layer": next(iter(moved["layers"])),
                           "title": "只在記憶體裡", "body": "b", "acceptance": "a",
                           "verify": {"kind": "manual", "note": "n"}})
    if not out_of_sync(moved):
        failures.append("out_of_sync：資料多一條之後竟然還說同步——那它沒有真的在比")

    if failures:
        print("正反對照失敗：")
        for f in failures:
            print(f"  ✗ {f}")
        return 1
    print("正反對照通過：每一種 kind 都驗過「訊號在」與「訊號拿掉」兩個方向。")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    sub = ap.add_subparsers(dest="cmd", required=True)
    sub.add_parser("verify", help="逐條核實，抓過期斷言").set_defaults(func=cmd_verify)
    r = sub.add_parser("render", help="產生 WORKLIST.md 的未完成項一節")
    r.add_argument("--stdout", action="store_true", help="只印出來，不回寫")
    r.set_defaults(func=cmd_render)
    sub.add_parser("scan", help="掃產品碼的自承，列出沒有條目認領的").set_defaults(func=cmd_scan)
    sub.add_parser("selftest", help="機制的正反對照").set_defaults(func=cmd_selftest)
    args = ap.parse_args()
    return args.func(args)


if __name__ == "__main__":
    sys.exit(main())
