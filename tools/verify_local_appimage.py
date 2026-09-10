"""核對本機 AppImage 的封包內容及完整保存操作輸出；僅限 Docker。"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


assert Path("/.dockerenv").exists()
dist = Path("/dist")
out = Path("/out")
build = json.loads((dist / "build-inputs.json").read_text())
version = build["version"]
artifact = dist / "full" / f"kol-{version}-full-linux-x86_64.AppImage"
assert sha(artifact) == build["appimage_sha256"]
with tempfile.TemporaryDirectory() as temporary:
    subprocess.run([str(artifact), "--appimage-extract"], cwd=temporary,
                   stdout=subprocess.DEVNULL, check=True)
    app = Path(temporary) / "squashfs-root"
    assert (app / "VERSION").read_text().strip() == version
    binary = app / "usr/bin/kol"
    assert sha(binary) == build["binary_sha256"]
    header = binary.read_bytes()[:20]
    assert header[:4] == b"\x7fELF" and header[4:6] == b"\x02\x01"
    assert int.from_bytes(header[18:20], "little") == 62
    actual = subprocess.check_output([str(app / "AppRun"), "-version"], text=True).strip()
    assert actual == version, actual
    bundled = [item for item in build["inputs"] if
               item["path"].startswith(("assets/", "docs/")) or item["path"] == "LICENSE"]
    for item in bundled:
        assert sha(app / item["path"]) == item["sha256"], item["path"]
    for path in app.rglob("*"):
        assert path.suffix.lower() not in (".exe", ".zip", ".pdf", ".wav", ".i64"), path
    for required in ("AppRun", "kol.desktop", "kol.png", "DO-NOT-DISTRIBUTE.txt"):
        assert (app / required).is_file()

log = (out / "verification.log").read_text()
for message in ("正常建角兩人、重啟選角與忠實住宿保存通過",
                "刪除不搬槽、重用空槽產生新身分及滿槽流程通過",
                "v1 匯入保留原檔、零值、傷勢與未知設定通過",
                "受控世界時間已重載並走過正常走路／露營"):
    assert message in log, message
for path in out.glob("*.log"):
    text = path.read_text()
    assert "繪製失敗" not in text and "panic:" not in text, path
assert (out / "legacy.json").read_bytes() == (out / "legacy-before.json").read_bytes()
files = sorted(p for p in out.iterdir() if p.is_file() and p.name != "receipt.json")
assert all((p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid()) for p in files)
receipt = dict(version=version, appimage_sha256=sha(artifact),
               binary_sha256=build["binary_sha256"], platform="linux/amd64",
               execution="AppImage --appimage-extract-and-run，Docker Xvfb",
               bundled_inputs_checked=len(bundled), seed=1,
               limitations=["非原版同狀態對拍", "世界時間使用受控 v2 輸入",
                            "未驗主機 FUSE 掛載、Windows 或 macOS", "未完成全遊戲通關"],
               outputs_sha256={p.name: sha(p) for p in files},
               input_counts={name: sum(line.startswith("input scene=") for line in
                             (out / (name + ".log")).read_text().splitlines())
                             for name in ("creation", "reload", "delete", "import-v1", "clock-reload")},
               verifier_sha256=sha(Path(__file__)))
(out / "receipt.json").write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+"\n")
print(json.dumps({k:v for k,v in receipt.items() if k != "outputs_sha256"}, ensure_ascii=False))
