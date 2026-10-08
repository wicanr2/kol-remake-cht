#!/usr/bin/env python3
"""在隔離容器核對指定正式來源的跨平台建置及 RE-295 正常保存路徑。

入口及執行紀錄見 WORKLOG.md 的 2026-10-07 接手段落。
來源唯讀掛 /src，獨立工作區掛 /out；這不是正式封包或真機驗收。
"""
from pathlib import Path
import argparse
import hashlib
import json
import os
import shutil
import struct
import subprocess
import time

ROOT = Path('/src')
OUT = Path('/out')
BASE = ROOT / 'workplace/rideaway-normal-20261006'
PREVIOUS = ROOT / 'workplace/visibility-current-regression-20261006'
BASE_BINARY = '91d11101d4e8db605d4b143be1e46f14f7e878f7135033c37c9da272808a8995'


def expected_binary():
    return read(OUT / 'plan.json')['source_binary_sha256']


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read(path):
    return json.loads(path.read_text(encoding='utf-8'))


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')


def check_inputs():
    inputs = read(OUT / 'program-inputs.json')
    assert len(inputs) == 342 and sum(n.endswith('.go') for n in inputs) == 340
    actual = {str(p.relative_to(ROOT)) for d in ('cmd', 'internal')
              for p in (ROOT / d).rglob('*.go')}
    actual.update(('go.mod', 'go.sum'))
    assert actual == set(inputs), (sorted(actual - set(inputs)), sorted(set(inputs) - actual))
    assert all(sha(ROOT / name) == digest for name, digest in inputs.items())
    runtime = read(OUT / 'runtime-inputs.json')
    assert all(sha(ROOT / name) == digest for name, digest in runtime.items())
    assert sha(BASE / 'kol-current') == BASE_BINARY
    return inputs


def prepare():
    preserved = {'prepare-failure-v1', 'baseline-re295', 'run-platform.sh'}
    assert not [p for p in OUT.iterdir() if p.name not in preserved], '輸出目錄只能含既有基準與執行腳本'
    inputs = read(args.source_inputs)
    assert len(inputs) == 342 and all(sha(ROOT / n) == h for n, h in inputs.items())
    write(OUT / 'program-inputs.json', inputs)
    runtime = {str(p.relative_to(ROOT)): sha(p)
               for d in ('assets', 'docs/i18n') for p in sorted((ROOT / d).rglob('*'))
               if p.is_file() and p.suffix in ('.png', '.json', '.ogg', '.tsv')}
    for name in ('title-sequence.json', 'town-inn-sample.json', 'town-outdoor-sample.json'):
        path = ROOT / 'docs/re' / name
        runtime[str(path.relative_to(ROOT))] = sha(path)
    write(OUT / 'runtime-inputs.json', runtime)
    (OUT / 'platform').mkdir()
    build = (PREVIOUS / 'build-windows.sh').read_text()
    build = build.replace('go version > /out/platform/windows-go-version.txt',
                          'go version > /out/platform/windows-go-version.txt\n'
                          'go build -trimpath -o /out/platform/kol-linux-amd64 ./cmd/kol '
                          '> /out/platform/linux-build.log 2>&1')
    (OUT / 'build-native.sh').write_text(build)
    shutil.copy2(PREVIOUS / 'mac-build.sh', OUT / 'build-macos.sh')
    write(OUT / 'plan.json', dict(
        source=args.source_name, source_binary_sha256=sha(args.source_binary), seed=1,
        cases=['ride', 'rest', 'save', 'reload'], languages=['en', 'zh-Hant'],
        expected_states_per_platform=2310, expected_pngs_per_platform=404,
        method='Reuse unchanged RE-295 ordinary keys and actual saves; no reseed, '
               'coordinate/state substitution or retry to select results.',
        windows_real_machine=False, macos_execution=False, packages=False,
        baseline_inputs_sha256=sha(BASE / 'formal-inputs.json'),
        tests_path=str(args.test_log.relative_to(ROOT)),
        tests_sha256=sha(args.test_log),
        controller_sources={str(p.relative_to(ROOT)): sha(p)
                            for lang in ('en', 'zh-Hant')
                            for n in (f'current-ride-{lang}', f'formal-rest-{lang}',
                                      f'formal-save-{lang}', f'formal-save-{lang}/reload')
                            for p in (BASE / n / 'command.json', BASE / n / 'trace.json')},
        build_scripts={n: sha(OUT / n) for n in ('build-native.sh', 'build-macos.sh')},
        tool_sha256=sha(Path(__file__))))
    check_inputs()
    print('PREPARED current: 342 sources; fixed seed 1; 2310 states per platform', flush=True)


def run(platform):
    check_inputs()
    assert sha(OUT / 'platform/kol-linux-amd64') == expected_binary()
    binary = OUT / 'platform' / ('kol-linux-amd64' if platform == 'linux' else 'kol-windows-amd64.exe')
    prefix = [] if platform == 'linux' else ['/usr/lib/wine/wine64']
    to_path = str if platform == 'linux' else lambda p: 'Z:' + str(p).replace('/', '\\')
    reports = []
    for lang in ('en', 'zh-Hant'):
        save = None
        for purpose, source_name in (('ride', f'current-ride-{lang}'),
                                     ('rest', f'formal-rest-{lang}'),
                                     ('save', f'formal-save-{lang}'),
                                     ('reload', f'formal-save-{lang}/reload')):
            baseline = BASE / source_name
            case = OUT / f'{platform}-{purpose}-{lang}'
            case.mkdir()
            cmd = read(baseline / 'command.json')
            assert cmd[cmd.index('-seed') + 1] == '1'
            assert not any(f in cmd for f in ('-ambush-rolls', '-world-x', '-world-y'))
            cmd[0] = to_path(binary)
            for flag, name in (('-trace', 'trace.json'), ('-shot-dir', ''),
                               ('-save', 'complete.json'), ('-text-log', 'text.jsonl')):
                if flag in cmd:
                    cmd[cmd.index(flag) + 1] = str(case / name)
            before = None
            if purpose == 'reload':
                assert save is not None
                before = sha(save)
                cmd[cmd.index('-load') + 1] = str(save)
            if platform == 'wine':
                for flag in ('-trace', '-shot-dir', '-save', '-text-log', '-load'):
                    if flag in cmd:
                        cmd[cmd.index(flag) + 1] = to_path(cmd[cmd.index(flag) + 1])
            cmd = prefix + cmd
            write(case / 'command.json', cmd)
            started = time.monotonic()
            with (case / 'run.log').open('w') as log:
                result = subprocess.run(cmd, cwd=ROOT, stdout=log, stderr=subprocess.STDOUT, timeout=300)
            report = dict(platform=platform, language=lang, purpose=purpose,
                          exit=result.returncode, elapsed_seconds=time.monotonic() - started)
            write(case / 'process.json', report)
            assert result.returncode == 0, (case / 'run.log').read_text()[-2000:]
            trace = read(case / 'trace.json')
            assert trace == read(baseline / 'trace.json'), str(case)
            images = {p.name: sha(p) for p in case.glob('*.png')}
            assert images == {p.name: sha(p) for p in baseline.glob('*.png')}, str(case)
            report.update(states=len(trace['steps']), pngs=len(images),
                          trace_equal_to_RE295=True, pngs_equal_to_RE295=True,
                          baseline=str(baseline.relative_to(ROOT)))
            if purpose == 'save':
                save = case / 'complete.json'
                current, reference = read(save), read(baseline / 'complete.json')
                assert current['schema'] == 'kol-save-v3'
                del current['saved_at'], reference['saved_at']
                assert current == reference
                report.update(save_sha256=sha(save), full_v3_equal_except_saved_at=True)
            if purpose == 'reload':
                assert sha(save) == before
                report.update(save_unchanged=True, save_sha256=before)
            if purpose == 'ride':
                assert not (case / 'complete.json').exists()
            reports.append(report)
            print(platform, lang, purpose, report['states'], 'states', report['pngs'], 'PNGs PASS', flush=True)
    assert sum(row['states'] for row in reports) == 2310
    assert sum(row['pngs'] for row in reports) == 404
    write(OUT / f'{platform}-review.json', dict(status='PASS finite normal player paths',
          states=2310, pngs=404, runs=reports, binary_sha256=sha(binary), seed=1,
          windows_real_machine=False, macos_execution=False))


def structures():
    check_inputs()
    p = OUT / 'platform'
    assert sha(p / 'kol-linux-amd64') == expected_binary()
    data = (p / 'kol-windows-amd64.exe').read_bytes()
    assert data[:2] == b'MZ'
    offset = struct.unpack_from('<I', data, 0x3c)[0]
    assert data[offset:offset + 4] == b'PE\0\0'
    assert struct.unpack_from('<H', data, offset + 4)[0] == 0x8664
    assert struct.unpack_from('<H', data, offset + 24)[0] == 0x20b
    fat = (p / 'kol-macos-universal').read_bytes()
    assert struct.unpack_from('>II', fat) == (0xcafebabe, 2)
    arches = []
    for index in range(2):
        cpu, sub, offset, size, alignment = struct.unpack_from('>IIIII', fat, 8 + 20 * index)
        name = {0x1000007: 'amd64', 0x100000c: 'arm64'}[cpu]
        thin = (p / ('kol-macos-' + name)).read_bytes()
        assert fat[offset:offset + size] == thin
        assert thin[:4] == b'\xcf\xfa\xed\xfe' and struct.unpack_from('<I', thin, 4)[0] == cpu
        arches.append(name)
    assert sorted(arches) == ['amd64', 'arm64']
    for arch in ('arm64', 'x86_64'):
        load = (p / (arch + '-load.txt')).read_text()
        assert 'minos 11.0' in load or 'version 11.0' in load
        if arch == 'arm64':
            assert 'LC_CODE_SIGNATURE' in load
        libraries = (p / (arch + '-libraries.txt')).read_text().splitlines()[1:]
        assert libraries and all(line.strip().startswith(('/System/Library/', '/usr/lib/')) for line in libraries)
    modules = {}
    for name in ('kol-linux-amd64', 'kol-windows-amd64.exe', 'kol-macos-arm64', 'kol-macos-amd64'):
        info = subprocess.run(['go', 'version', '-m', str(p / name)], check=True,
                              capture_output=True, text=True, timeout=20).stdout
        for marker in ('go1.26.7', '\tdep\tgithub.com/hajimehoshi/ebiten/v2\tv2.9.9\t',
                       '\tdep\tgithub.com/ebitengine/oto/v3\tv3.4.0\t'):
            assert marker in info, (name, marker)
        modules[name] = info
    write(OUT / 'build-info.json', modules)
    write(OUT / 'structure-review.json', dict(status='PASS builds and binary structure',
          source=read(OUT / 'plan.json')['source'], linux_rebuilt=True, macos_minimum='11.0', macos_arches=sorted(arches),
          arm64_signature=True, windows_real_machine=False, macos_execution=False,
          binaries={f.name: sha(f) for f in p.glob('kol-*')}))
    print('PASS rebuilt Linux identity, Windows PE32+, macOS universal structure and dependencies')


def seal():
    check_inputs()
    assert read(OUT / 'structure-review.json')['status'].startswith('PASS')
    assert all(read(OUT / f'{p}-review.json')['states'] == 2310 for p in ('linux', 'wine'))
    artifacts = {str(p.relative_to(OUT)): sha(p) for p in sorted(OUT.rglob('*'))
                 if p.is_file() and p.name != 'receipt.json' and 'wineprefix' not in p.parts}
    write(OUT / 'receipt.json', dict(status='PASS finite current cross-platform verification',
          source=read(OUT / 'plan.json')['source'], source_binary_sha256=expected_binary(), source_inputs=342,
          states=4620, pngs=808, tests=read(OUT / 'plan.json')['tests_path'],
          windows_real_machine=False, macos_execution=False, full_remake_complete=False,
          rights='local-only original-derived; no distribution authorization', artifacts=artifacts))
    print('SEALED 4620 states / 808 PNGs; no production or package change')


def verify():
    check_inputs()
    receipt = read(OUT / 'receipt.json')
    assert receipt['source_binary_sha256'] == expected_binary()
    assert all(sha(OUT / n) == h for n, h in receipt['artifacts'].items())
    assert sha(Path(__file__)) == read(OUT / 'plan.json')['tool_sha256']
    plan = read(OUT / 'plan.json')
    assert sha(ROOT / plan['tests_path']) == plan['tests_sha256']
    for platform in ('linux', 'wine'):
        for row in read(OUT / f'{platform}-review.json')['runs']:
            case = OUT / f"{platform}-{row['purpose']}-{row['language']}"
            baseline = ROOT / row['baseline']
            assert read(case / 'trace.json') == read(baseline / 'trace.json')
            assert {p.name: sha(p) for p in case.glob('*.png')} == {p.name: sha(p) for p in baseline.glob('*.png')}
    print('PASS immutable receipt, current source, normal traces, PNGs and reused test evidence')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('stage', choices=('prepare', 'linux', 'wine', 'structures', 'seal', 'verify'))
    parser.add_argument('--source-inputs', type=Path, default=BASE/'formal-inputs.json')
    parser.add_argument('--source-binary', type=Path, default=BASE/'kol-current')
    parser.add_argument('--test-log', type=Path, default=BASE/'tests.log')
    parser.add_argument('--source-name', default='RE-295')
    args = parser.parse_args()
    assert Path('/.dockerenv').exists()
    assert os.getuid() == os.getgid() == 1000
    assert OUT.stat().st_uid == OUT.stat().st_gid == 1000
    if args.stage in ('linux', 'wine'):
        run(args.stage)
    else:
        globals()[args.stage]()
