"""從既有正常操作案例驗證 AppImage 未指定 -save 的保存；僅 Docker。"""
import hashlib
import argparse
import json
import os
from pathlib import Path
import subprocess

assert Path("/.dockerenv").exists()
parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("--home-fallback", action="store_true")
args = parser.parse_args()
config_home = "/out/home/.config" if args.home_fallback else "/out/config"
out = Path("/out")
assert (out.stat().st_uid, out.stat().st_gid) == (os.getuid(), os.getgid())
assert not any(p.name != ".keep" for p in out.iterdir()), "需要新的驗收目錄"
version = json.loads((Path("/dist") / "build-inputs.json").read_text())["version"]
artifact = Path(f"/dist/full/kol-{version}-full-linux-x86_64.AppImage")
sha = lambda p: hashlib.sha256(p.read_bytes()).hexdigest()
before = sha(artifact)
base = Path("/src/tools/verify_complete_save_gui.sh")
original = base.read_text()
marker = "# v1 相容性採明確測試輸入，不冒充原版存檔。"
assert original.count(marker) == 1
script = original.split(marker)[0]
assert script.count(" -save /out/created.json") == 3
script = script.replace(" -save /out/created.json", "")
script = script.replace("/out/created.json", config_home + "/kol-remake-cht/save.dat")
if args.home_fallback:
    assert script.count("export HOME=/tmp DISPLAY=:95") == 1
    script = script.replace("export HOME=/tmp DISPLAY=:95", "export HOME=/out/home DISPLAY=:95")
assert " -save " not in script
(out / "scenario.sh").write_text(script)
launcher = out / "kol"
launcher.write_text(f'#!/bin/sh\nexec "{artifact}" --appimage-extract-and-run "$@"\n')
launcher.chmod(0o755)
with (out / "verification.log").open("w") as log:
    environment = dict(os.environ)
    if args.home_fallback:
        environment.pop("XDG_CONFIG_HOME", None)
    else:
        environment["XDG_CONFIG_HOME"] = config_home
    subprocess.run(["bash", str(out / "scenario.sh")], check=True,
                   env=environment,
                   stdout=log, stderr=subprocess.STDOUT)
assert sha(artifact) == before
save = Path(config_home) / "kol-remake-cht/save.dat"
state = json.loads(save.read_text())
# v0.2.3 的封包寫 v2；SPEC-036 之後的建置寫 v3（角色多了 picture／shield）。
assert state["schema"] in ("kol-save-v2", "kol-save-v3")
assert state["characters"][14]["name"] == "NOVA2"
assert state["characters"][15]["name"] == "LUNA"
assert state["characters"][15]["gold"] == 1440
assert not (out / "created.json").exists()
for path in out.glob("*.log"):
    assert "繪製失敗" not in path.read_text() and "panic:" not in path.read_text(), path
files = [p for p in out.rglob("*") if p.is_file()]
assert all((p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid()) for p in files)
receipt = dict(version=version, appimage_sha256=before, source_script_sha256=sha(base),
               verifier_sha256=sha(Path(__file__)), seed=1,
               save_arguments=False, config_home=config_home, home_fallback=args.home_fallback,
               save_path=str(save),
               scope="正常建角、重啟明確 -load、住宿、刪角及空槽重用；未指定 -save",
               limitation="Linux AppImage 保存驗收；不代表其他作業系統或原版對拍",
               input_counts={name: sum(line.startswith("input scene=") for line in
                             (out / (name + ".log")).read_text().splitlines())
                             for name in ("creation", "reload", "delete")},
               outputs_sha256={str(p.relative_to(out)): sha(p) for p in files})
(out / "receipt.json").write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+"\n")
print("AppImage 未指定 -save 的正常保存與重啟通過")
