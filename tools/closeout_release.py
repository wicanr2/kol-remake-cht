#!/usr/bin/env python3
"""v1 收尾流水線。容器執行方法與限制見 docs/verification/release-v1.0.2-20261013.md。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import plistlib
import re
import shutil
import struct
import subprocess
import tarfile
import tempfile
import zipfile

SRC, OUT = Path('/src'), Path('/out')
VERSION = None  # 必須由 --version 指定，拒絕覆寫既有版本。
LIMITS = '''本版包含已核准 HD 介面、頭像、四首原創配樂及遊戲內 EGA／HD 切換。
這是有限場景驗收，不宣稱完整原版 parity。
完整戰鬥與全程通關的原版同狀態仍未驗完。
已登記場景按 SPEC-040 通過中文與畫面驗收，其他動態分支仍有未驗項。
HD 涵蓋旅店、六人肖像、旅店頭像、中文字、面板及已核准操作列；其他美術沿 EGA。
配樂預設開啟，F9獨立切換；換EGA／HD持續播放，開關只在本次遊玩有效。
三項加速細則、戰鬥中途保存與後續 HD 美術仍未定版。
Windows 僅容器 Wine 驗證；macOS 僅交叉編譯與結構驗證，均無真機驗收。
Linux 容器以 AppImage 解包執行驗證，主機 FUSE 未驗。
原版 C matching 工作暫停，未納入本次封包。
完整版、遊戲資料與含原作畫面的推廣影片僅本機保存。
GitHub Release 只提供需自行匯入合法資料的引擎包；不改儲存庫可見度。
'''


def sha(p):
    with p.open('rb') as f:
        return hashlib.file_digest(f, 'sha256').hexdigest()


def read(p):
    return json.loads(p.read_text())


def write(p, d):
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(json.dumps(d, ensure_ascii=False, indent=2) + '\n')


def run(cmd, *, cwd=SRC, env=None, log=None, timeout=600):
    if log:
        with log.open('w') as f:
            subprocess.run(cmd, cwd=cwd, env=env, stdout=f, stderr=subprocess.STDOUT,
                           check=True, timeout=timeout)
    else:
        subprocess.run(cmd, cwd=cwd, env=env, check=True, timeout=timeout)


def copy(src, dst):
    if src.is_dir():
        shutil.copytree(src, dst)
    else:
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, dst)


def prepare():
    assert not any(OUT.iterdir()), '拒絕覆寫既有交付'
    program = {str(p.relative_to(SRC)): sha(p) for n in ('cmd', 'internal')
               for p in sorted((SRC / n).rglob('*.go'))}
    program.update({n: sha(SRC / n) for n in ('go.mod', 'go.sum')})
    accepted = SRC / 'docs/verification/hd-music-runtime-20261013'
    acceptance = read(accepted / 'receipt.json')
    assert acceptance['status'] == 'PASS'
    matches_music = program == acceptance['program_inputs']
    precision_path = SRC / 'docs/verification' / ('release-' + VERSION.replace('v.', 'v', 1)) / 'precision-fix.json'
    precision = None
    if not matches_music:
        precision = read(precision_path)
        assert precision['status'] == 'PASS' and precision['production'] is True
        assert precision['base_music_receipt_sha256'] == sha(accepted / 'receipt.json')
        assert precision['program_inputs'] == program
        delta = {n for n in set(program) | set(acceptance['program_inputs'])
                 if program.get(n) != acceptance['program_inputs'].get(n)}
        assert delta == {'internal/ui/hd.go', 'internal/ui/hd_interface.go',
                         'internal/ui/hd_sampling.go', 'internal/ui/hd_sampling_test.go'}, delta
    assert sha(SRC / 'assets/audio/hd/manifest.json') == acceptance['catalog_sha256']
    for n in ('full', 'release', 'promo', 'smoke'):
        (OUT / n).mkdir()
    resources = []
    for bundle in ('full', 'release'):
        base = OUT / bundle
        for n in ('LICENSE', 'docs/i18n', 'docs/re/title-sequence.json',
                  'docs/re/town-inn-sample.json', 'docs/re/town-outdoor-sample.json',
                  'docs/re/ida-fingerprints/KOLBOOK-character-writer-summary.json'):
            copy(SRC / n, base / n)
        (base / 'VERSION').write_text(VERSION + '\n')
        (base / 'KNOWN-LIMITATIONS.txt').write_text(LIMITS)
    # 說明書完整譯文只留本機完整版，不延伸為 Release 散布授權。
    copy(SRC / 'docs/manual', OUT / 'full/docs/manual')
    for n in ('original-derived-png', 'original-derived-data', 'ui-font', 'audio', 'hd'):
        copy(SRC / 'assets' / n, OUT / 'full/assets' / n)
    # runtime 不帶研究 README、反組譯或原版可執行檔；保存既有來源 manifest。
    for p in sorted((OUT / 'full/assets').rglob('*')):
        if p.is_file():
            assert p.suffix.lower() not in ('.exe', '.zip', '.pdf', '.asm', '.i64')
            resources.append(dict(path=str(p.relative_to(OUT / 'full')),
                                  bytes=p.stat().st_size, sha256=sha(p)))
    write(OUT / 'release/import-manifest.json', dict(version=VERSION,
          schema='kol-required-data/1', source='使用者合法持有、按本專案流程抽出的遊戲資料',
          rights='清單不包含素材位元組；匯入資料不得散布。', files=resources))
    (OUT / 'release/tools').mkdir()
    for p in sorted((SRC / 'tools').glob('*.py')):
        if p.name != 'closeout_release.py':
            copy(p, OUT / 'release/tools' / p.name)
    copy(SRC / 'LICENSE', OUT / 'LICENSE')
    copy(SRC / 'LICENSE', OUT / 'promo/LICENSE')
    (OUT / 'full/DO-NOT-DISTRIBUTE.txt').write_text('含原作衍生資料，僅本機保存，不得上傳或散布。\n')
    instructions = '''《武士傳說》繁體中文重製 {version}

完整版已含遊戲資料。Windows 解壓後開啟 kol.exe；macOS 開啟武士傳說.app。
Linux 執行 chmod +x *.AppImage，再執行 AppImage。
若沒有 FUSE，可加 --appimage-extract-and-run。

release 是引擎包，需自備合法持有、已抽取的相符遊戲資料。
它不含原版 ZIP、EXE、圖片或資料表。不是解壓就能直接玩。
資料抽取器在 tools/，與開源程式碼一起保留；完整匯入接受含 assets/ 的抽取結果目錄。
使用 Python 3.11 或更新版本：
  python tools/import_game_data.py --source "合法資料目錄" --target "."
macOS 的 tools 在 app/Contents/Resources，--target 指向該 Resources。
Linux AppImage 請先 --appimage-extract，從 squashfs-root/tools 匯入至新的可寫資料目錄，
例如 $HOME/.local/share/kol-remake-cht，再設定 KOL_DATA_DIR 指向該目錄執行 AppImage。
必須全部檔案與 SHA-256 相符才匯入；不接受原版 ZIP 直接當 runtime 資料。

方向鍵移動；A S Z X 斜走；逗號、句號選下方圖示；Enter 執行；Esc 返回。
滑鼠點選、右鍵返回、滾輪移動游標。Ctrl+Q 離開。
介面預設繁中，-lang en 切英文。住宿會保存。
按 F10 或下緣「畫面」開啟設定，即時切換 EGA／HD；Esc 返回原場景。
HD 為 1920×1200，含已核准介面與人物頭像，其他場景沿原畫；切換不改存檔。
四首原創室內樂預設開啟。按F9或下緣「配樂」獨立切換，換EGA／HD持續播放。
Linux：~/.config/kol-remake-cht/save.dat
Windows：%AppData%\\kol-remake-cht\\save.dat
macOS：~/Library/Application Support/kol-remake-cht/save.dat
以 -load "完整存檔路徑" 重載。戰鬥中途保存尚未提供。

macOS 未經 Apple 公證，首次開啟可能須於系統設定的隱私權與安全性選仍要打開。
平台及忠實度限制見 KNOWN-LIMITATIONS.txt。
程式授權見 LICENSE；原版素材及翻譯不因此取得公開散布權。
完整版與影片只留本機；GitHub Release 僅提供需匯入資料的引擎包。
'''.format(version=VERSION)
    for bundle in ('full', 'release'):
        (OUT / bundle / 'README.txt').write_bytes(b'\xef\xbb\xbf' + instructions.replace('\n', '\r\n').encode())
    baseline = SRC / 'docs/verification/local/quest-rng-formal-20261007'
    retained = read(baseline / 'retention.json')
    assert sha(baseline / 'evidence.tar.xz') == retained['archive']['sha256']
    # 唯一一次正常抽樣及新程序重載，在執行前固定 seed、鍵序、比較與失敗判準。
    with tarfile.open(baseline / 'evidence.tar.xz') as archive:
        for folder in ('formal-save-zh-Hant', 'formal-save-zh-Hant/reload'):
            for n in ('command.json', 'trace.json'):
                dest = OUT / 'smoke/baseline' / folder / n
                dest.parent.mkdir(parents=True, exist_ok=True)
                dest.write_bytes(archive.extractfile(folder + '/' + n).read())
        dest = OUT / 'smoke/baseline/formal-save-zh-Hant/complete.json'
        dest.write_bytes(archive.extractfile('formal-save-zh-Hant/complete.json').read())
    write(OUT / 'build-inputs.json', dict(version=VERSION, git_head=args.git_head,
          program_inputs=program, resources=resources, program_matches_accepted_receipt=matches_music,
          precision_source_receipt_sha256=sha(precision_path) if precision else None,
          program_matches_precision_receipt=precision is not None,
          packaging_tools={str(p.relative_to(SRC)): sha(p) for p in
            [SRC / 'tools/closeout_release.py', SRC / 'tools/import_game_data.py', SRC / 'docker/Dockerfile.package']},
          visibility='private; verified by host gh',
          accepted_source_receipt_sha256=sha(accepted / 'receipt.json'),
          accepted_source_archive_sha256=sha(SRC / 'docs/verification/local/hd-music-runtime-20261013/evidence.tar.xz'),
          previous_receipt_sha256=sha(baseline / 'receipt.json'),
          previous_archive_sha256=sha(baseline / 'evidence.tar.xz')))
    write(OUT / 'smoke/predeclared.json', dict(version=VERSION, scenario_count=1,
          seed=1, seed_setup='程序啟動前 -seed 1，與保留原版前綴同種子；不重擲',
          route='標題、書本、選隊、任務5戰鬥、逃離、自然ORC遭遇、回城住宿保存、新程序重載',
          comparator='前版正常路徑語意 trace、完整 v3 除 saved_at；僅 battle.exec_text 與 battle.encounter_text 允許中文修正；不新增原版覆蓋',
          failure='崩潰、卡住、狀態差異、保存損壞任一項即停止交付',
          baseline_receipt='docs/verification/local/quest-rng-formal-20261007/receipt.json',
          no_forced_win=True, no_state_injection=True, no_ambush_override=True,
          original_execution=False, macos_execution=False, real_machine=False))
    print('PREPARE PASS:', len(program), 'accepted program inputs; one scenario predeclared', flush=True)


def native():
    stage = Path('/tmp/kol-source')
    stage.mkdir()
    for n in ('cmd', 'internal', 'go.mod', 'go.sum'):
        copy(SRC / n, stage / n)
    shutil.copytree('/go/pkg/mod', '/tmp/mod')
    env = dict(os.environ, PATH='/usr/local/go/bin:' + os.environ['PATH'], GOWORK='off', GOPROXY='off', GOSUMDB='off',
               GOMODCACHE='/tmp/mod', GOCACHE='/tmp/gocache', GOPATH='/tmp/gopath')
    for name, more in (('kol', {}), ('kol.exe', dict(GOOS='windows', GOARCH='amd64', CGO_ENABLED='0'))):
        run(['go', 'build', '-trimpath', '-buildvcs=false', '-ldflags',
             '-X main.buildVersion=' + VERSION, '-o', str(OUT / 'full' / name), './cmd/kol'],
            cwd=stage, env=dict(env, **more), log=OUT / 'smoke' / (name + '-build.log'))
        print('BUILD', name, sha(OUT / 'full' / name), flush=True)
    formal_checks(env)


def formal_checks(env):
    run(['xvfb-run', '-a', 'go', 'test', './cmd/...', './internal/...'], env=env,
        log=OUT / 'smoke/tests.log')
    result = subprocess.run(['xvfb-run', '-a', 'go', 'vet', './cmd/...', './internal/...'],
                            cwd=SRC, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                            timeout=600)
    (OUT / 'smoke/vet.log').write_bytes(result.stdout)
    warnings = [line for line in result.stdout.decode().splitlines() if not line.startswith('# ')]
    expected = ['internal/ui/bitmap_editor.go:56:4: github.com/wicanr2/kol-remake-cht/internal/book.Label struct literal uses unkeyed fields',
                'internal/ui/bitmap_editor.go:57:4: github.com/wicanr2/kol-remake-cht/internal/book.Label struct literal uses unkeyed fields']
    assert result.returncode == 0 or (result.returncode == 1 and warnings == expected), warnings
    write(OUT / 'smoke/checks.json', dict(tests='PASS', vet_exit=result.returncode,
          vet_warnings=warnings, classification='既有未具名欄位警示；保留正式來源，不宣稱 vet 全綠'))
    print('TEST PASS; VET', result.returncode, warnings, flush=True)


def tests():
    shutil.copytree('/go/pkg/mod', '/tmp/mod')
    env = dict(os.environ, PATH='/usr/local/go/bin:' + os.environ['PATH'], GOWORK='off',
               GOPROXY='off', GOSUMDB='off', GOMODCACHE='/tmp/mod', GOCACHE='/tmp/gocache',
               GOPATH='/tmp/gopath')
    if (OUT / 'smoke/tests.log').exists():
        (OUT / 'smoke/tests.log').rename(OUT / 'smoke/tests-environment-failure.log')
    formal_checks(env)


def assemble():
    for name in ('kol', 'kol.exe', 'kol-macos'):
        copy(OUT / 'full' / name, OUT / 'release' / name)
    icon = Path('/tmp/kol.png')
    run(['convert', '-size', '256x256', 'xc:#140606', '-fill', '#c8992e', '-draw',
         'polygon 128,18 220,60 220,140 128,238 36,140 36,60', str(icon)])
    for bundle in ('full', 'release'):
        base = OUT / bundle
        app = Path('/tmp') / (bundle + '.AppDir')
        shutil.copytree(base, app)
        for name in ('kol.exe', 'kol-macos'):
            (app / name).unlink()
        (app / 'usr/bin').mkdir(parents=True)
        (app / 'kol').rename(app / 'usr/bin/kol')
        (app / 'AppRun').write_text('''#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
if [ -n "${KOL_DATA_DIR:-}" ]; then
  cd "$KOL_DATA_DIR" || exit 1
else
  cd "$HERE" || exit 1
fi
exec "$HERE/usr/bin/kol" "$@"
''')
        (app / 'AppRun').chmod(0o755)
        (app / 'kol.desktop').write_text('[Desktop Entry]\nType=Application\nName=武士傳說\nExec=AppRun\nIcon=kol\nCategories=Game;RolePlaying;\n')
        copy(icon, app / 'kol.png')
        copy(icon, app / '.DirIcon')
        squash = Path('/tmp') / (bundle + '.squashfs')
        run(['mksquashfs', str(app), str(squash), '-noappend', '-comp', 'gzip', '-processors', '2'],
            log=OUT / 'smoke' / (bundle + '-squashfs.log'))
        image = base / f'kol-{VERSION}-{bundle}-linux-x86_64.AppImage'
        with image.open('wb') as f:
            for path in (Path('/opt/runtime-x86_64'), squash):
                with path.open('rb') as source:
                    shutil.copyfileobj(source, f)
        image.chmod(0o755)
        # 封包不再攜帶另一平台的程式；每包附條款、說明與限制。
        for platform in ('windows-amd64', 'macos-universal'):
            root = Path('/tmp') / f'{bundle}-{platform}'
            root.mkdir()
            name = f'kol-{VERSION}-{bundle}-{platform}'
            staged = root / name
            if platform.startswith('windows'):
                shutil.copytree(base, staged, ignore=shutil.ignore_patterns('kol', 'kol-macos', '*.AppImage', '*.zip'))
            else:
                staged.mkdir()
                mac = staged / '武士傳說.app/Contents'
                (mac / 'MacOS').mkdir(parents=True)
                res = mac / 'Resources'
                shutil.copytree(base, res, ignore=shutil.ignore_patterns('kol', 'kol.exe', 'kol-macos', '*.AppImage', '*.zip'))
                copy(base / 'kol-macos', mac / 'MacOS/kol')
                (mac / 'MacOS/kol').chmod(0o755)
                run(['convert', str(icon), str(res / 'kol.icns')])
                (mac / 'Info.plist').write_bytes(plistlib.dumps(dict(
                    CFBundleName='武士傳說', CFBundleDisplayName='武士傳說', CFBundleExecutable='kol',
                    CFBundleIdentifier='tw.wicanr2.kol-remake-cht', CFBundleVersion=VERSION[2:].split('-')[0],
                    CFBundleShortVersionString=VERSION[2:].split('-')[0], KOLReleaseVersion=VERSION,
                    CFBundlePackageType='APPL', CFBundleIconFile='kol',
                    LSMinimumSystemVersion='11.0', NSHighResolutionCapable=True)))
                copy(base / 'README.txt', staged / 'README.txt')
                copy(base / 'LICENSE', staged / 'LICENSE')
            run(['python3', str(SRC / 'tools/make_zip.py'), str(base / (name + '.zip')), str(staged), name])
        print('PACKAGED', bundle, flush=True)


def sample():
    smoke = OUT / 'smoke'
    base = smoke / 'baseline/formal-save-zh-Hant'
    image = OUT / 'full' / f'kol-{VERSION}-full-linux-x86_64.AppImage'
    report = []
    # 這一個情境只執行一次；重載屬同一保存垂直鏈。
    for purpose, reference in (('normal', base), ('reload', base / 'reload')):
        dest = smoke / purpose
        dest.mkdir()
        cmd = read(reference / 'command.json')
        assert cmd[cmd.index('-seed') + 1] == '1'
        assert '-ambush-rolls' not in cmd and '-world-pos' not in cmd
        cmd[0:1] = [str(image), '--appimage-extract-and-run']
        for flag, path in (('-trace', dest / 'trace.json'), ('-save', smoke / 'save.dat'),
                           ('-shot-dir', dest), ('-load', smoke / 'save.dat')):
            if flag in cmd:
                cmd[cmd.index(flag) + 1] = str(path)
        # 增加同一次重播的畫格，供本版影片使用，不增加情境或重試。
        if purpose == 'normal':
            current = set(cmd[cmd.index('-shot-steps') + 1].split(','))
            current.update(str(n) for n in (1,2,3,4,5,15,30,60,90,110,120,125,130,135,150,200,300,400,500,600,700,800,900,950,1000))
            cmd[cmd.index('-shot-steps') + 1] = ','.join(sorted(current, key=int))
        write(dest / 'command.json', cmd)
        before = sha(smoke / 'save.dat') if purpose == 'reload' else None
        run(cmd, cwd=Path('/tmp'), log=dest / 'run.log', timeout=300)
        actual, expected = read(dest / 'trace.json'), read(reference / 'trace.json')
        exact = actual == expected
        def semantic(trace):
            trace = json.loads(json.dumps(trace))
            for step in trace['steps']:
                for field in ('exec_text', 'encounter_text'):
                    if 'battle' in step:
                        step['battle'].pop(field, None)
            return trace
        assert semantic(actual) == semantic(expected), purpose + ' semantic trace changed'
        if purpose == 'normal':
            a, b = read(smoke / 'save.dat'), read(base / 'complete.json')
            assert a['schema'] == 'kol-save-v3'
            a.pop('saved_at'); b.pop('saved_at')
            assert a == b, 'save changed'
        else:
            assert sha(smoke / 'save.dat') == before, 'reload modified input save'
        report.append(dict(purpose=purpose, states=len(actual['steps']),
                           pngs=len(list(dest.glob('*.png'))), trace_equal=exact, semantic_trace_equal=True))
    write(smoke / 'sampling.json', dict(status='PASS', scenario_count=1, seed=1, runs=report,
          appimage_sha256=sha(image), save_sha256=sha(smoke / 'save.dat'),
          full_v3_equal_except_saved_at=True, original_new_execution=False,
          whole_game_parity=False))
    print('ONE SAMPLE PASS', report, flush=True)


def verify():
    report = []
    required = read(OUT / 'release/import-manifest.json')['files']
    for bundle in ('full', 'release'):
        base = OUT / bundle
        image = base / f'kol-{VERSION}-{bundle}-linux-x86_64.AppImage'
        with tempfile.TemporaryDirectory(prefix='kol-package-') as tmp:
            root = Path(tmp)
            run([str(image), '--appimage-extract'], cwd=root,
                log=OUT / 'smoke' / (bundle + '-extract-verification.log'))
            extracted = root / 'squashfs-root'
            assert sha(extracted / 'usr/bin/kol') == sha(base / 'kol')
            assert (extracted / 'AppRun').stat().st_mode & 0o111
            assert (extracted / 'VERSION').read_text().strip() == VERSION
            assert (extracted / 'LICENSE').read_bytes() == (SRC / 'LICENSE').read_bytes()
            if bundle == 'full':
                for row in required:
                    assert sha(extracted / row['path']) == row['sha256']
            else:
                assert not (extracted / 'assets').exists()
                assert not (extracted / 'docs/manual').exists()
                forbidden = {'.png', '.ogg', '.pdf', '.jpg', '.jpeg', '.icns', '.zip'}
                assert not [p for p in extracted.rglob('*') if p.is_file()
                            and p.suffix.lower() in forbidden
                            and p.name not in ('kol.png', '.DirIcon', 'kol.icns')]
                assert (extracted / 'tools/import_game_data.py').is_file()
            report.append(dict(bundle=bundle, platform='linux-x86_64', status='PASS',
                               full_payload_hashes_checked=bundle == 'full'))
        for platform, suffix, executable in (
                ('windows-amd64', '.zip', 'kol.exe'),
                ('macos-universal', '.zip', 'Contents/MacOS/kol')):
            path = base / f'kol-{VERSION}-{bundle}-{platform}{suffix}'
            with zipfile.ZipFile(path) as z:
                assert z.testzip() is None
                names = z.namelist()
                exe = next(n for n in names if n.endswith('/' + executable))
                reference = base / ('kol.exe' if platform.startswith('windows') else 'kol-macos')
                assert hashlib.sha256(z.read(exe)).hexdigest() == sha(reference)
                assert any(n.endswith('/LICENSE') for n in names)
                for info in z.infolist():
                    if not info.filename.isascii():
                        assert info.flag_bits & 0x800
                    if info.filename.endswith('/README.txt'):
                        content = z.read(info)
                        assert content.startswith(b'\xef\xbb\xbf') and b'\r\n' in content
                if platform.startswith('macos'):
                    assert (z.getinfo(exe).external_attr >> 16) & 0o111
                    plist = plistlib.loads(z.read(next(n for n in names if n.endswith('/Info.plist'))))
                    assert plist['KOLReleaseVersion'] == VERSION
                if bundle == 'release':
                    assert not any('/assets/' in n for n in names), 'release leaked assets'
                    assert not any('/docs/manual/' in n for n in names), 'release leaked manual'
                    assert not [n for n in names if Path(n).suffix.lower() in
                                ('.png', '.ogg', '.pdf', '.jpg', '.jpeg', '.icns', '.zip')
                                and Path(n).name not in ('kol.png', '.DirIcon', 'kol.icns')]
                else:
                    for row in required:
                        n = next(n for n in names if n.endswith('/' + row['path']))
                        assert hashlib.sha256(z.read(n)).hexdigest() == row['sha256']
                report.append(dict(bundle=bundle, platform=platform, entries=len(names), status='PASS'))
    # 完整封包重算逐檔 manifest；收據／影片由各 phase 另外驗證。
    for bundle in ('full', 'release', 'promo'):
        files = [dict(path=str(p.relative_to(OUT / bundle)), bytes=p.stat().st_size, sha256=sha(p))
                 for p in sorted((OUT / bundle).rglob('*'))
                 if p.is_file() and p != OUT / bundle / 'manifest.json']
        write(OUT / bundle / 'manifest.json', dict(schema='kol-dist-manifest-v1', version=VERSION,
              bundle=bundle, rights=('引擎包；可上傳目前private GitHub；公開散布另行審查'
                                     if bundle == 'release' else '僅本機；不得上傳或公開散布'),
              platform=['linux/amd64','windows/amd64','darwin/amd64','darwin/arm64'],
              file_count=len(files), files=files))
    for name in ('sampling', 'structures', 'release-smoke', 'windows-smoke'):
        assert read(OUT / 'smoke' / (name + '.json'))['status'] == 'PASS'
    assert read(OUT / 'smoke/linux-controls.json')['status'] == 'PASS'
    pixel_check=read(OUT / 'smoke/windows-pixels.json')
    assert pixel_check['status'] == 'PASS', 'HD一色階閘門未通過，拒絕完成交付驗證'
    full_image=OUT/'full'/f'kol-{VERSION}-full-linux-x86_64.AppImage'
    assert pixel_check['reference_appimage_sha256']==sha(full_image)
    assert read(OUT/'smoke/sampling.json')['appimage_sha256']==sha(full_image)
    for check in read(OUT/'smoke/windows-smoke.json')['checks']:
        package=OUT/check['bundle']/f"kol-{VERSION}-{check['bundle']}-windows-amd64.zip"
        assert check['package_sha256']==sha(package)
    # 使用者維持一色階標準；目視接受不能覆蓋數值失敗。
    numeric_gate_met=pixel_check['status']=='PASS'
    assert read(OUT / 'smoke/checks.json')['tests'] == 'PASS'
    assert read(OUT / 'promo/sources.json')['visual_review'] == 'PASS'
    inputs = read(OUT / 'build-inputs.json')
    assert all(sha(SRC / n) == h for n,h in inputs['program_inputs'].items())
    inputs['final_packaging_tools'] = {str(p.relative_to(SRC)): sha(p) for p in
        [SRC/'tools/closeout_release.py', SRC/'tools/import_game_data.py',
         SRC/'tools/compose_release_promo.py', SRC/'docker/Dockerfile.package']}
    inputs['runtime_verification_tools'] = {str(p.relative_to(SRC)): sha(p) for p in
        [SRC/'tools'/n for n in ('verify_release_runtime.py','verify_package_controls.py','verify_music_runtime.py',
         'closeout_macos.sh','run_release_linux.sh','run_release_wine.sh','run_release_sample.sh','capture_wine_audio.sh',
         'diagnose_release_pixels.py')]}
    inputs['staged_release_tool_snapshots']={str(p.relative_to(OUT/'release')):sha(p) for p in sorted((OUT/'release/tools').glob('*.py'))}
    write(OUT / 'build-inputs.json', inputs)
    write(OUT / 'smoke/packages.json', dict(status='PASS', packages=report,
          windows_real_machine=False, macos_execution=False, host_fuse=False,
          cross_platform_numeric_status=pixel_check['status'],numeric_gate_met=numeric_gate_met,
          technical_gate_met=numeric_gate_met,publication_authorized=False,ready_for_publication=False))
    allfiles = [dict(path=str(p.relative_to(OUT)), bytes=p.stat().st_size, sha256=sha(p))
                for p in sorted(OUT.rglob('*')) if p.is_file() and p.name != 'SHA256SUMS.json']
    write(OUT / 'SHA256SUMS.json', dict(version=VERSION, files=allfiles))
    assert all(p.stat().st_uid == os.getuid() and p.stat().st_gid == os.getgid() for p in OUT.rglob('*'))
    print('VERIFY PASS', report, flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('phase', choices=('prepare', 'native', 'tests', 'assemble', 'sample', 'verify'))
    parser.add_argument('--git-head', default='')
    parser.add_argument('--version', required=True)
    args = parser.parse_args()
    assert re.fullmatch(r'v\.[0-9]+\.[0-9]+\.[0-9]+-[0-9]{8}', args.version), '版號格式錯誤'
    VERSION = args.version
    assert Path('/.dockerenv').exists() and OUT.is_dir()
    assert (OUT.stat().st_uid, OUT.stat().st_gid) == (os.getuid(), os.getgid())
    globals()[args.phase]()
