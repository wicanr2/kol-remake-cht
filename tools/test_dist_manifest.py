"""交付清單的版號與完整性回歸測試；僅在 Docker 執行。"""
import json
import importlib.util
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest
import hashlib
import zipfile
import signal

TOOLS = Path(__file__).resolve().parent
# 只作隔離測試輸入，不是本專案的正式版號。
VERSION = "v.0.0.0-20000101"
INVALID_VERSIONS = ("0.2.2", "20261004", "v1.0.0-20261004",
                    "v.1.0.0-20261004-r2", "../keep", "v.1.0.0-20261004/../keep")


def executable(path, text):
    path.write_text(text)
    path.chmod(0o755)


class ReleaseEntryTest(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        (self.root / "tools").mkdir()
        self.bin = self.root / "bin"
        self.bin.mkdir()
        self.marker = self.root / "called"
        self.env = dict(os.environ, PATH=f"{self.bin}:{os.environ['PATH']}",
                        KOL_TEST_MARKER=str(self.marker), PYTHONDONTWRITEBYTECODE="1")
        for name in ("build_dist.sh", "package_dist.sh", "verify_dist.sh", "build_macos.sh", "make_promo.sh"):
            shutil.copyfile(TOOLS / name, self.root / "tools" / name)
        executable(self.root / "tools/check_repo_visibility.sh",
                   '#!/bin/bash\nprintf "visibility\\n" >> "$KOL_TEST_MARKER"\nexit 79\n')
        executable(self.bin / "git", "#!/bin/bash\nexit 0\n")
        executable(self.bin / "docker", "#!/bin/bash\nexit 0\n")

    def run_entry(self, name, *args, env=None):
        return subprocess.run(["bash", str(self.root / "tools" / name), *args],
                              env=env or self.env, capture_output=True, text=True, timeout=15)

    def test_invalid_versions_stop_before_visibility_or_work(self):
        for name in ("build_dist.sh", "package_dist.sh", "verify_dist.sh"):
            for version in INVALID_VERSIONS:
                with self.subTest(name=name, version=version):
                    result = self.run_entry(name, version)
                    self.assertEqual(result.returncode, 2, result.stderr)
                    self.assertFalse(self.marker.exists())
                    self.assertFalse((self.root / "dist-all").exists())
            good = self.run_entry(name, VERSION)
            self.assertEqual(good.returncode, 79, good.stderr)
            self.assertEqual(self.marker.read_text(), "visibility\n")
            self.marker.unlink()

    def test_container_entries_reject_invalid_versions(self):
        for name, mode in (("build_dist.sh", "--local-linux-in-container"),
                           ("build_dist.sh", "--stage-in-container"),
                           ("verify_dist.sh", "--content-in-container"),
                           ("package_dist.sh", "--in-container")):
            for version in INVALID_VERSIONS:
                with self.subTest(name=name, version=version):
                    result = self.run_entry(name, mode, version)
                    self.assertEqual(result.returncode, 2, result.stderr)
                    self.assertIn("版號必須使用", result.stderr)

    def test_content_wrapper_delegates_to_readonly_container(self):
        executable(self.root / "tools/check_repo_visibility.sh", "#!/bin/bash\nexit 0\n")
        (self.root / "dist-all" / VERSION).mkdir(parents=True)
        executable(self.bin / "docker", f'''#!{sys.executable}
import json, os, sys
with open(os.environ["KOL_TEST_MARKER"], "w") as log:
    json.dump(sys.argv[1:], log)
sys.exit(79)
''')
        result = self.run_entry("verify_dist.sh", VERSION)
        self.assertEqual(result.returncode, 79, result.stderr)
        args = json.loads(self.marker.read_text())
        self.assertIn(f"{self.root}:/src:ro", args)
        self.assertEqual(args[-4:], ["bash", "/src/tools/verify_dist.sh", "--content-in-container", VERSION])
        for option in ("--rm", "--memory", "--cpus", "--pids-limit", "-u"):
            self.assertIn(option, args)
        self.assertEqual(args[args.index("--network") + 1], "none")

    def test_smoke_version_gate_and_xvfb_cleanup(self):
        executable(self.root / "tools/check_repo_visibility.sh", "#!/bin/bash\nexit 0\n")
        executable(self.root / "tools/verify_macos.sh", "#!/bin/bash\nexit 0\n")
        for bundle in ("full", "release"):
            path = self.root / "dist-all" / VERSION / bundle / "kol-macos"
            path.parent.mkdir(parents=True)
            path.write_bytes(b"fixture")
        executable(self.bin / "docker", f'''#!{sys.executable}
import json, os, sys
path = os.environ["KOL_TEST_MARKER"]
with open(path, "a") as log:
    log.write(json.dumps(sys.argv[1:]) + "\\n")
with open(path) as log:
    count = len(log.readlines())
sys.exit(0 if count == 1 else 79)
''')
        result = self.run_entry("verify_dist.sh", VERSION)
        self.assertEqual(result.returncode, 79, result.stderr)
        calls = [json.loads(line) for line in self.marker.read_text().splitlines()]
        self.assertEqual(len(calls), 2)
        self.assertIn(f"{self.root}:/src:ro", calls[1])
        self.assertIn(f"KOL_EXPECTED_VERSION={VERSION}", calls[1])
        self.assertEqual(calls[1][-3:-1], ["bash", "-lc"])
        inner = calls[1][-1]
        temp = self.root / "smoke-tmp"
        temp.mkdir()
        (temp / ".X11-unix").mkdir()
        inner = inner.replace("/tmp/", f"{temp}/")
        executable(self.bin / "Xvfb", f'''#!{sys.executable}
from pathlib import Path
import os, signal, sys
base = Path({str(temp)!r})
def stop(*args):
    (base / "stopped").write_text("terminated\\n")
    sys.exit(0)
signal.signal(signal.SIGTERM, stop)
signal.signal(signal.SIGALRM, stop)
signal.alarm(10)
(base / "pid").write_text(str(os.getpid()))
(base / ".X11-unix/X93").touch()
while True:
    signal.pause()
''')
        executable(self.root / "kol", f'''#!{sys.executable}
from pathlib import Path
import os, sys
if sys.argv[1:] == ["-version"]:
    print(os.environ["KOL_TEST_ACTUAL_VERSION"])
    sys.exit(0)
if os.environ.get("KOL_TEST_SCENE_FAIL") == "1":
    sys.exit(17)
if "-shot" in sys.argv:
    Path(sys.argv[sys.argv.index("-shot") + 1]).write_bytes(b"frame fixture")
else:
    path = Path(sys.argv[sys.argv.index("-shot-dir") + 1])
    path.mkdir(exist_ok=True)
    (path / "panel-6.png").write_bytes(b"frame fixture")
''')
        for actual, scene_fail, expected in ((VERSION, "0", 0),
                                             ("development", "0", 3),
                                             (VERSION, "1", 17)):
            with self.subTest(actual=actual, scene_fail=scene_fail):
                (temp / "stopped").unlink(missing_ok=True)
                (temp / "pid").unlink(missing_ok=True)
                (temp / ".X11-unix/X93").unlink(missing_ok=True)
                try:
                    result = subprocess.run(["bash", "-c", inner], cwd=self.root,
                                            env=dict(self.env, KOL_EXPECTED_VERSION=VERSION,
                                                     KOL_TEST_ACTUAL_VERSION=actual,
                                                     KOL_TEST_SCENE_FAIL=scene_fail),
                                            capture_output=True, text=True, timeout=15)
                    self.assertEqual(result.returncode, expected, result.stdout + result.stderr)
                    self.assertEqual((temp / "stopped").read_text(), "terminated\n")
                    with self.assertRaises(ProcessLookupError):
                        os.kill(int((temp / "pid").read_text()), 0)
                finally:
                    if (temp / "pid").exists():
                        try:
                            os.kill(int((temp / "pid").read_text()), signal.SIGTERM)
                        except ProcessLookupError:
                            pass

    def test_build_wrapper_stages_only_in_bounded_container(self):
        executable(self.root / "tools/check_repo_visibility.sh", "#!/bin/bash\nexit 0\n")
        executable(self.bin / "docker", f'''#!{sys.executable}
import json, os, sys
with open(os.environ["KOL_TEST_MARKER"], "w") as log:
    json.dump(sys.argv[1:], log)
sys.exit(79)
''')
        result = self.run_entry("build_dist.sh", VERSION)
        self.assertEqual(result.returncode, 79, result.stderr)
        args = json.loads(self.marker.read_text())
        self.assertIn(f"{self.root}:/src:ro", args)
        self.assertIn(f"{self.root}/dist-all/{VERSION}:/out", args)
        self.assertEqual(args[-5:], ["bash", "/src/tools/build_dist.sh", "--stage-in-container", VERSION, "prepare"])
        self.assertEqual(args[args.index("--network") + 1], "none")
        for option in ("--rm", "--memory", "--cpus", "--pids-limit", "-u"):
            self.assertIn(option, args)

    def test_promo_source_copy_runs_in_container_after_success(self):
        executable(self.bin / "docker", f'''#!{sys.executable}
import json, os, sys
with open(os.environ["KOL_TEST_MARKER"], "w") as log:
    json.dump(sys.argv[1:], log)
sys.exit(79)
''')
        docs = self.root / "docs/promo"
        docs.mkdir(parents=True)
        (self.root / "assets").mkdir()
        executable(docs / "compose.sh", '#!/bin/bash\nexit "${KOL_TEST_COMPOSE_EXIT:-0}"\n')
        (docs / "theme.sh").write_bytes(b"theme fixture\n")
        output = self.root / "promo-output"
        result = self.run_entry("make_promo.sh", str(output))
        self.assertEqual(result.returncode, 79, result.stderr)
        args = json.loads(self.marker.read_text())
        self.assertIn(f"{self.root}/docs:/docs:ro", args)
        self.assertIn(f"{self.root}/assets:/assets:ro", args)
        self.assertIn(f"{output}:/out", args)
        self.assertEqual(args[-3:-1], ["bash", "-c"])
        inner = args[-1].replace("/docs/promo/", f"{docs}/").replace("/out/", f"{output}/")
        subprocess.run(["bash", "-c", inner], env=self.env, check=True, timeout=15)
        self.assertEqual((output / "theme.sh").read_bytes(), b"theme fixture\n")
        self.assertEqual((output / "compose.sh").read_bytes(), (docs / "compose.sh").read_bytes())
        for path in output.iterdir():
            path.unlink()
        result = subprocess.run(["bash", "-c", inner], env=dict(self.env, KOL_TEST_COMPOSE_EXIT="77"), timeout=15)
        self.assertEqual(result.returncode, 77)
        self.assertEqual(list(output.iterdir()), [])

    def test_build_preserves_existing_directory_and_dangling_link(self):
        executable(self.root / "tools/check_repo_visibility.sh", "#!/bin/bash\nexit 0\n")
        executable(self.root / "tools/go_docker.sh",
                   '#!/bin/bash\nprintf "compiler\\n" >> "$KOL_TEST_MARKER"\nexit 79\n')
        parent = self.root / "dist-all"
        parent.mkdir()
        out = parent / VERSION
        out.mkdir()
        sentinel = out / "keep"
        sentinel.write_bytes(b"existing delivery\x00")
        result = self.run_entry("build_dist.sh", VERSION)
        self.assertEqual(result.returncode, 2, result.stderr)
        self.assertIn("拒絕覆寫", result.stderr)
        self.assertEqual(sentinel.read_bytes(), b"existing delivery\x00")
        self.assertFalse(self.marker.exists())
        shutil.rmtree(out)
        out.symlink_to(parent / "missing")
        result = self.run_entry("build_dist.sh", VERSION)
        self.assertEqual(result.returncode, 2, result.stderr)
        self.assertTrue(out.is_symlink())
        self.assertFalse((parent / "missing").exists())
        self.assertFalse(self.marker.exists())

    def test_package_preserves_existing_directory_and_dangling_link(self):
        # 容器固定 /src；只在副本改掛載路徑，避免碰到真正的 dist-all。
        script = self.root / "tools/package_dist.sh"
        source = script.read_text()
        self.assertEqual(source.count('DIST="/src/dist-all/$VERSION"'), 1)
        script.write_text(source.replace('DIST="/src/dist-all/$VERSION"',
                                        f'DIST="{self.root}/dist-all/$VERSION"'))
        executable(self.bin / "convert",
                   '#!/bin/bash\nprintf "convert\\n" >> "$KOL_TEST_MARKER"\nexit 79\n')
        dist = self.root / "dist-all" / VERSION
        dist.mkdir(parents=True)
        out = dist / "packages"
        out.mkdir()
        sentinel = out / "keep"
        sentinel.write_bytes(b"existing packages\x00")
        result = self.run_entry("package_dist.sh", "--in-container", VERSION)
        self.assertEqual(result.returncode, 2, result.stderr)
        self.assertIn("拒絕覆寫", result.stderr)
        self.assertEqual(sentinel.read_bytes(), b"existing packages\x00")
        self.assertFalse(self.marker.exists())
        shutil.rmtree(out)
        out.symlink_to(dist / "missing")
        result = self.run_entry("package_dist.sh", "--in-container", VERSION)
        self.assertEqual(result.returncode, 2, result.stderr)
        self.assertTrue(out.is_symlink())
        self.assertFalse((dist / "missing").exists())
        self.assertFalse(self.marker.exists())
        out.unlink()
        result = self.run_entry("package_dist.sh", "--in-container", VERSION)
        self.assertEqual(result.returncode, 79, result.stderr)
        self.assertEqual(self.marker.read_text(), "convert\n")

    def test_full_build_forwards_one_version_to_all_platforms(self):
        executable(self.root / "tools/check_repo_visibility.sh", "#!/bin/bash\nexit 0\n")
        executable(self.root / "tools/go_docker.sh", f'''#!{sys.executable}
import json, os, sys
with open(os.environ["KOL_TEST_MARKER"], "a") as log:
    log.write(json.dumps([os.environ.get("KOL_GO_ENV", ""), sys.argv[1:]]) + "\\n")
''')
        executable(self.root / "tools/build_macos.sh", f'''#!{sys.executable}
import json, os, sys
with open(os.environ["KOL_TEST_MARKER"], "a") as log:
    log.write(json.dumps([os.environ.get("KOL_BUILD_VERSION"), sys.argv[1:]]) + "\\n")
sys.exit(79)
''')
        result = self.run_entry("build_dist.sh", VERSION)
        self.assertEqual(result.returncode, 79, result.stderr)
        calls = [json.loads(line) for line in self.marker.read_text().splitlines()]
        self.assertEqual(len(calls), 3)
        for platform, args in calls[:2]:
            self.assertEqual(args[args.index("-ldflags") + 1], f"-X main.buildVersion={VERSION}")
        self.assertEqual(calls[0][0], "")
        self.assertEqual(calls[1][0], "GOOS=windows GOARCH=amd64 CGO_ENABLED=0")
        self.assertEqual(calls[2], [VERSION, [f"dist-all/{VERSION}/full/kol-macos"]])

    def test_macos_forwards_version_and_links_both_architectures(self):
        executable(self.bin / "docker", f'''#!{sys.executable}
import json, os, sys
with open(os.environ["KOL_TEST_MARKER"], "w") as log:
    json.dump(sys.argv[1:], log)
sys.exit(79)
''')
        for version in INVALID_VERSIONS:
            result = self.run_entry("build_macos.sh", "build/kol", env=dict(self.env, KOL_BUILD_VERSION=version))
            self.assertEqual(result.returncode, 2, result.stderr)
            self.assertFalse(self.marker.exists())
            self.assertFalse((self.root / "build").exists())
        result = self.run_entry("build_macos.sh", "build/kol", env=dict(self.env, KOL_BUILD_VERSION=VERSION))
        self.assertEqual(result.returncode, 79, result.stderr)
        args = json.loads(self.marker.read_text())
        self.assertIn(f"KOL_BUILD_VERSION={VERSION}", args)
        inner = args[-1]
        self.assertEqual(args[-3:-1], ["bash", "-c"])
        # 實際執行容器命令，只替換來源／輸出掛載位置；編譯器與 lipo 是隔離替身。
        cache = self.root / "mod"
        cache.mkdir()
        inner = inner.replace("/go/pkg/mod /tmp/mod", f"{cache} {self.root}/copied-mod")
        inner = inner.replace('"/src/$KOL_OUT"', f'"{self.root}/$KOL_OUT"')
        executable(self.bin / "osxcross-conf", "#!/bin/bash\necho OSXCROSS_TARGET=darwin24.5\n")
        executable(self.bin / "x86_64-apple-darwin24.5-lipo", "#!/bin/bash\nexit 0\n")
        executable(self.bin / "go", f'''#!{sys.executable}
import json, os, sys
with open(os.environ["KOL_TEST_MARKER"], "a") as log:
    log.write(json.dumps([os.environ["GOARCH"], sys.argv[1:]]) + "\\n")
''')
        for version in (VERSION, ""):
            self.marker.unlink()
            shutil.rmtree(self.root / "copied-mod", ignore_errors=True)
            subprocess.run(["bash", "-c", inner], check=True, capture_output=True,
                           env=dict(self.env, KOL_BUILD_VERSION=version, KOL_MACOS_MIN="11.0", KOL_OUT="build/kol"), timeout=15)
            calls = [json.loads(line) for line in self.marker.read_text().splitlines()]
            self.assertEqual([call[0] for call in calls], ["arm64", "amd64"])
            for arch, args in calls:
                if version:
                    self.assertEqual(args[args.index("-ldflags") + 1], f"-X main.buildVersion={version}")
                else:
                    self.assertNotIn("-ldflags", args)


class StageTest(unittest.TestCase):
    """在Docker內跑真實組裝及內容分支，副本只改/src、/out掛載位置。"""

    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name)
        self.source = self.root / "source"
        self.source.mkdir()
        self.out = self.source / "dist-all" / VERSION
        self.out.mkdir(parents=True)
        inputs = {
            "LICENSE": b"license fixture\n", "README.md": b"readme fixture\n",
            "docs/manual/README.md": b"manual fixture\n",
            "docs/i18n/ui.json": b"{}\n", "docs/re/title-sequence.json": b"{}\n",
            "docs/re/town-inn-sample.json": b"{}\n", "docs/re/town-outdoor-sample.json": b"{}\n",
            "assets/original-derived-png/ega/sample.png": b"private image fixture\x00",
            "assets/original-derived-data/items.json": b"{}\n",
            "assets/ui-font/font.png": b"private font fixture\x00",
            "tools/extract.py": b"# extractor fixture\n",
            "tools/run_python_tool.sh": b"#!/bin/sh\n",
        }
        for name, content in inputs.items():
            path = self.source / name
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(content)
        self.inputs = inputs
        self.stage = self.root / "build_dist.sh"
        source = (TOOLS / "build_dist.sh").read_text()
        source = source.replace("REPO_ROOT=/src", f"REPO_ROOT={self.source}")
        source = source.replace("OUT=/out", f"OUT={self.out}")
        source = source.replace("[ -d /src ] && [ -d /out ]", f"[ -d {self.source} ] && [ -d {self.out} ]")
        source = source.replace("stat -c '%u:%g' /out", f"stat -c '%u:%g' {self.out}")
        source = source.replace("/src/tools/write_dist_manifest.py", str(TOOLS / "write_dist_manifest.py"))
        source = source.replace("--dist /out", f"--dist {self.out}")
        self.stage.write_text(source)
        self.content = self.root / "verify_dist.sh"
        source = (TOOLS / "verify_dist.sh").read_text()
        source = source.replace("REPO_ROOT=/src", f"REPO_ROOT={self.source}")
        source = source.replace("[ -d /src ]", f"[ -d {self.source} ]")
        source = source.replace("/src/tools/verify_packages.py", str(TOOLS / "verify_packages.py"))
        source = source.replace('"/src/$DIST"', f'"{self.source}/$DIST"')
        self.content.write_text(source)

    def run_stage(self, phase):
        return subprocess.run(["bash", str(self.stage), "--stage-in-container", VERSION, phase],
                              capture_output=True, text=True, timeout=15)

    def run_content(self):
        return subprocess.run(["bash", str(self.content), "--content-in-container", VERSION],
                              capture_output=True, text=True, timeout=15)

    def populate(self):
        result = self.run_stage("prepare")
        self.assertEqual(result.returncode, 0, result.stderr)
        for name, data in (("kol", b"\x7fELFfixture"), ("kol.exe", b"MZfixture"), ("kol-macos", b"mac fixture")):
            (self.out / "full" / name).write_bytes(data)
        result = self.run_stage("files")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_stage_copies_identical_inputs_and_separates_private_assets(self):
        self.populate()
        for name, content in self.inputs.items():
            self.assertEqual((self.source / name).read_bytes(), content)
            if name.startswith("tools/"):
                self.assertEqual((self.out / "release" / name).read_bytes(), content)
            elif name.startswith("assets/"):
                self.assertEqual((self.out / "full" / name).read_bytes(), content)
                self.assertFalse((self.out / "release" / name).exists())
            else:
                for bundle in ("full", "release"):
                    self.assertEqual((self.out / bundle / name).read_bytes(), content)
        self.assertTrue((self.out / "full/DO-NOT-DISTRIBUTE.txt").is_file())
        result = self.run_stage("manifest")
        self.assertEqual(result.returncode, 0, result.stderr)
        for bundle in ("full", "release", "promo"):
            manifest = json.loads((self.out / bundle / "manifest.json").read_text())
            self.assertEqual(manifest["version"], VERSION)
            self.assertEqual(manifest["bundle"], bundle)
            self.assertEqual(len(manifest["platform"]), 4)
            for row in manifest["files"]:
                path = self.out / bundle / row["path"]
                self.assertEqual(hashlib.sha256(path.read_bytes()).hexdigest(), row["sha256"])
        result = self.run_stage("prepare")
        self.assertEqual(result.returncode, 2, result.stderr)
        self.assertIn("全新空目錄", result.stderr)

    def test_large_deny_list_fails_closed(self):
        self.populate()
        # 大量違規檔案仍須拒絕；find的-print -quit只回傳第一個命中。
        for n in range(2000):
            (self.source / "docs/manual" / (f"{n:04}-" + "x" * 100 + ".png")).write_bytes(b"image fixture")
        result = self.run_stage("files")
        self.assertEqual(result.returncode, 3, result.stderr[-1000:])
        self.assertIn("release", result.stderr)
        result = self.run_content()
        # 先補清單及條款，確保內容驗收走到同一deny-list。
        self.run_stage("manifest").check_returncode()
        result = self.run_content()
        self.assertEqual(result.returncode, 3, result.stderr[-1000:])

    def test_content_checks_manifest_and_package_shapes(self):
        self.populate()
        (self.out / "promo/kol-promo.mp4").write_bytes(b"video fixture")
        (self.out / "promo/sources.json").write_bytes(b"{}\n")
        self.run_stage("manifest").check_returncode()
        packages = self.out / "packages"
        packages.mkdir()
        for bundle in ("full", "release"):
            prefix = f"kol-{VERSION}-{bundle}"
            appimage = packages / f"{prefix}-linux-x86_64.AppImage"
            appimage.write_bytes(b"\x7fELFfixture")
            appimage.chmod(0o755)
            for target, executable_name in (("windows-amd64", "kol.exe"),
                                            ("macos-universal", "app/Contents/MacOS/kol")):
                with zipfile.ZipFile(packages / f"{prefix}-{target}.zip", "w") as archive:
                    entry = zipfile.ZipInfo(f"{prefix}/{executable_name}")
                    entry.external_attr = 0o755 << 16
                    archive.writestr(entry, b"binary fixture")
                    archive.writestr(f"{prefix}/如何開始.txt", b"instructions fixture")
                    if target == "macos-universal":
                        archive.writestr(f"{prefix}/app/Contents/Info.plist", b"plist fixture")
                    if bundle == "full":
                        archive.writestr(f"{prefix}/assets/data.txt", b"private fixture")
        result = self.run_content()
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        path = self.out / "full/manifest.json"
        manifest = json.loads(path.read_text())
        path.write_text(json.dumps(dict(manifest, version="v.0.0.1-20000101")))
        result = self.run_content()
        self.assertEqual(result.returncode, 3, result.stdout + result.stderr)
        self.assertIn("版號", result.stdout)
        path.write_text(json.dumps(manifest))
        (self.out / "release/kol.exe").write_bytes(b"XXfixture")
        result = self.run_content()
        self.assertEqual(result.returncode, 3, result.stderr)
        self.assertIn("不是 PE", result.stderr)


class ManifestTest(unittest.TestCase):
    def test_version_and_nested_manifests(self):
        tool = Path(__file__).with_name("write_dist_manifest.py")
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            nested = root / "full/assets/manifest.json"
            nested.parent.mkdir(parents=True)
            nested.write_text('{"asset": true}\n')
            command = [sys.executable, str(tool), "--dist", str(root),
                       "--platform", "linux/amd64", "--version"]
            bad = subprocess.run(command + ["0.2.2"], capture_output=True)
            self.assertNotEqual(bad.returncode, 0)
            self.assertFalse((root / "full/manifest.json").exists())
            for _ in range(2):
                subprocess.run(command + ["v.0.2.2-20260908"],
                               check=True, capture_output=True)
                manifest = json.loads((root / "full/manifest.json").read_text())
                self.assertEqual(manifest["platform"], ["linux/amd64"])
                self.assertEqual(manifest["file_count"], 1)
                self.assertEqual(manifest["files"][0]["path"], "assets/manifest.json")

    def test_manifest_rejects_wrong_metadata_with_valid_payload(self):
        spec = importlib.util.spec_from_file_location("verify_packages", TOOLS / "verify_packages.py")
        verifier = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(verifier)
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            base = root / "full"
            base.mkdir()
            (base / "keep").write_bytes(b"payload")
            subprocess.run([sys.executable, str(TOOLS / "write_dist_manifest.py"),
                            "--dist", str(root), "--version", VERSION, "--platform", "linux/amd64"],
                           check=True, capture_output=True)
            path = base / "manifest.json"
            good = json.loads(path.read_text())
            verifier.FAILURES = 0
            verifier.check_manifest(str(root), "full", VERSION)
            self.assertEqual(verifier.FAILURES, 0)
            for field, value in (("schema", "wrong"), ("bundle", "release"),
                                 ("version", "0.2.2"), ("version", None),
                                 ("version", "v.0.0.1-20000101")):
                with self.subTest(field=field, value=value):
                    path.write_text(json.dumps(dict(good, **{field: value})))
                    verifier.FAILURES = 0
                    verifier.check_manifest(str(root), "full", VERSION)
                    self.assertGreater(verifier.FAILURES, 0)
            path.write_text(json.dumps(good))
            for mode in ("changed", "extra", "missing"):
                with self.subTest(mode=mode):
                    (base / "keep").write_bytes(b"payload")
                    (base / "extra").unlink(missing_ok=True)
                    if mode == "changed":
                        (base / "keep").write_bytes(b"payloae")
                    elif mode == "extra":
                        (base / "extra").write_bytes(b"extra")
                    else:
                        (base / "keep").unlink()
                    verifier.FAILURES = 0
                    verifier.check_manifest(str(root), "full", VERSION)
                    self.assertGreater(verifier.FAILURES, 0)

    def test_verify_cli_rejects_invalid_version_before_reading_files(self):
        for version in INVALID_VERSIONS:
            result = subprocess.run([sys.executable, str(TOOLS / "verify_packages.py"),
                                     "/does-not-exist", version], capture_output=True, timeout=15)
            self.assertEqual(result.returncode, 2)
            self.assertIn("用法".encode(), result.stderr)


if __name__ == "__main__":
    unittest.main()
