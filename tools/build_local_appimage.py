"""從隔離快照建立僅本機 Linux AppImage；由 build_dist.sh 的容器入口呼叫。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def run(*args, **kwargs):
    subprocess.run(args, check=True, **kwargs)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--version", required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if not re.fullmatch(r"v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}", args.version):
        parser.error("版號格式錯誤")
    if not Path("/.dockerenv").exists():
        parser.error("只能在 Docker 內執行")
    out = args.output
    assert (out.stat().st_uid, out.stat().st_gid) == (os.getuid(), os.getgid())
    assert not any(p.name != ".keep" for p in out.iterdir()), "輸出必須為全新空目錄"
    source = Path("/src")
    stage = Path("/tmp/kol-source")
    app = Path("/tmp/kol.AppDir")
    stage.mkdir()
    app.mkdir()
    inputs = []

    def copy(relative, destination):
        original = source / relative
        paths = sorted(original.rglob("*")) if original.is_dir() else [original]
        for path in paths:
            assert not path.is_symlink(), f"輸入不允許符號連結：{path}"
            if not path.is_file():
                continue
            target = destination / path.relative_to(source)
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(path, target)
            sha = digest(target)
            assert sha == digest(path), f"輸入在複製時改變：{path}"
            inputs.append({"path": str(path.relative_to(source)), "sha256": sha})

    for relative in ("cmd", "internal", "go.mod", "go.sum"):
        copy(relative, stage)
    for relative in ("assets/original-derived-png", "assets/original-derived-data",
                     "assets/ui-font", "docs/i18n", "docs/manual", "LICENSE",
                     "docs/re/title-sequence.json", "docs/re/town-inn-sample.json",
                     "docs/re/town-outdoor-sample.json",
                     "docs/re/ida-fingerprints/KOLBOOK-character-writer-summary.json"):
        copy(relative, app)
    # 尚未驗收的音訊候選不接入正式 runtime。
    for name in ("full", "release", "promo"):
        (out / name).mkdir()
    for name in ("release", "promo"):
        (out / name / "尚未交付.txt").write_text(
            f"版本 {args.version} 僅建立本機 Linux 完整包；本目錄尚未交付。\n")
    (app / "VERSION").write_text(args.version + "\n")
    (app / "DO-NOT-DISTRIBUTE.txt").write_text("含原作衍生素材，只限本機使用，不得公開散布。\n")
    (app / "usr/bin").mkdir(parents=True)
    env = dict(os.environ, HOME="/tmp", GOWORK="off", GOPROXY="off", GOSUMDB="off",
               GOCACHE="/cache", GOMODCACHE="/tmp/mod")
    shutil.copytree("/go/pkg/mod", "/tmp/mod")
    with (out / "build.log").open("w") as log:
        run("/usr/local/go/bin/go", "build", "-trimpath", "-buildvcs=false",
            "-ldflags", f"-X main.buildVersion={args.version}", "-o",
            str(app / "usr/bin/kol"), "./cmd/kol", cwd=stage, env=env,
            stdout=log, stderr=subprocess.STDOUT)
    (app / "AppRun").write_text('''#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
cd "$HERE" || exit 1
exec "$HERE/usr/bin/kol" "$@"
''')
    (app / "AppRun").chmod(0o755)
    (app / "kol.desktop").write_text("[Desktop Entry]\nType=Application\nName=武士傳說\n"
        "Exec=AppRun\nIcon=kol\nCategories=Game;RolePlaying;\nTerminal=false\n")
    run("convert", "-size", "256x256", "xc:#0b1020", "-fill", "#c9a227", "-draw",
        "polygon 128,18 220,60 220,140 128,238 36,140 36,60", str(app / "kol.png"))
    shutil.copyfile(app / "kol.png", app / ".DirIcon")
    artifact = out / "full" / f"kol-{args.version}-full-linux-x86_64.AppImage"
    with (out / "package.log").open("w") as log:
        run("/src/workplace/tools/appimagetool", "--appimage-extract-and-run",
            "--runtime-file", "/src/workplace/tools/runtime-x86_64", str(app), str(artifact),
            env=dict(env, ARCH="x86_64"), stdout=log, stderr=subprocess.STDOUT)
    for item in inputs:
        assert digest(source / item["path"]) == item["sha256"], "建置期間輸入異動"
    receipt = dict(version=args.version, platform="linux/amd64", rights="僅本機",
                   validation="待從 AppImage 執行正常玩家路徑驗收",
                   binary_sha256=digest(app / "usr/bin/kol"),
                   appimage_sha256=digest(artifact), inputs=inputs,
                   tools={name: digest(source / "workplace/tools" / name)
                          for name in ("appimagetool", "runtime-x86_64")})
    (out / "build-inputs.json").write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+"\n")
    (out / "full/SHA256SUMS").write_text(f"{digest(artifact)}  {artifact.name}\n")
    run("python3", "/src/tools/write_dist_manifest.py", "--dist", str(out),
        "--version", args.version, "--platform", "linux/amd64")
    assert all((p.stat().st_uid, p.stat().st_gid) == (os.getuid(), os.getgid())
               for p in out.rglob("*")), "輸出擁有權錯誤"
    print(artifact)


if __name__ == "__main__":
    main()
