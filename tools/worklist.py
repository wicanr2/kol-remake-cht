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
import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
DATA_PATH = ROOT / "docs/worklist.json"
WORKLIST_MD = ROOT / "WORKLIST.md"

BEGIN = "<!-- worklist:begin 由 tools/worklist.py render 產生，不要手改 -->"
END = "<!-- worklist:end -->"

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
    """回 pattern 命中的第一個檔案（相對路徑），沒有就回 None。"""
    expr = re.compile(verify["pattern"])
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
            if not re.search(v["pattern"], f.read_text(encoding="utf-8", errors="ignore")):
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
    return problems


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
    return 1 if (stale or regressed) else 0


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

    # --- schema 檢查自己也要有反例 ---------------------------------------
    bad = {"layers": {"a": "A"}, "items": [{"id": "x", "layer": "b", "verify": {"kind": "nope"}}]}
    if not check_schema(bad):
        failures.append("check_schema：壞掉的 schema 竟然通過")

    real = load()
    if check_schema(real):
        failures.append("check_schema：docs/worklist.json 自己沒過")

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
    sub.add_parser("selftest", help="機制的正反對照").set_defaults(func=cmd_selftest)
    args = ap.parse_args()
    return args.func(args)


if __name__ == "__main__":
    sys.exit(main())
