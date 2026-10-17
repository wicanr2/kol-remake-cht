#!/usr/bin/env python3
"""Docker 內重建純 EGA、EGA＋HD 的 Windows 完整版與引擎包。"""
import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import zipfile

ROOT = Path('/src')


def sha(path):
    with path.open('rb') as stream:
        return hashlib.file_digest(stream, 'sha256').hexdigest()


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def main():
    assert Path('/.dockerenv').exists(), '只允許在 Docker 執行'
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--version', required=True)
    parser.add_argument('--binary', required=True, type=Path)
    parser.add_argument('--out', required=True, type=Path)
    args = parser.parse_args()
    version, out = args.version, args.out
    assert re.fullmatch(r'v\.\d+\.\d+\.\d+-\d{8}', version)
    assert args.binary.is_file()
    assert not out.exists() or not any(out.iterdir()), '拒絕覆寫交付'
    out.mkdir(exist_ok=True)
    for name in ('full', 'release', 'smoke', 'promo'):
        (out / name).mkdir()
    # Previous accepted resource inventory pins all art and rules data. This
    # hotfix changes code and packaging only, never regenerates approved art.
    baseline = ROOT / 'docs/verification/release-v1.0.3-20261017/build-inputs.json'
    resources = json.loads(baseline.read_text())['resources']
    for row in resources:
        source = ROOT / row['path']
        assert source.is_file() and source.stat().st_size == row['bytes']
        assert sha(source) == row['sha256'], source
    program = {str(p.relative_to(ROOT)): sha(p) for d in ('cmd', 'internal')
               for p in sorted((ROOT / d).rglob('*.go'))}
    program.update({n: sha(ROOT / n) for n in ('go.mod', 'go.sum')})
    write(out / 'build-inputs.json', dict(version=version, program_inputs=program,
          binary_sha256=sha(args.binary), base_resources_sha256=sha(baseline), resources=resources))
    common = {}
    for name in ('LICENSE', 'docs/i18n', 'docs/re/title-sequence.json',
                 'docs/re/town-inn-sample.json', 'docs/re/town-outdoor-sample.json',
                 'docs/re/ida-fingerprints/KOLBOOK-character-writer-summary.json'):
        source = ROOT / name
        if source.is_dir():
            common.update({str(p.relative_to(ROOT)): p for p in sorted(source.rglob('*')) if p.is_file()})
        else:
            common[name] = source
    common['kol.exe'] = args.binary
    packages = []
    for theme in ('ega', 'ega-hd'):
        selected = [r for r in resources if theme == 'ega-hd' or not r['path'].startswith('assets/hd/')]
        for bundle in ('full', 'release'):
            stem = f'kol-{version}-{bundle}-{theme}-windows-amd64'
            paths = dict(common)
            if bundle == 'full':
                paths.update({r['path']: ROOT / r['path'] for r in selected})
                paths.update({str(p.relative_to(ROOT)): p for p in sorted((ROOT / 'docs/manual').rglob('*')) if p.is_file()})
            else:
                paths.update({str(p.relative_to(ROOT)): p for p in sorted((ROOT / 'tools').glob('*.py'))})
            content = {
                'VERSION': (version + '\n').encode(),
                'KNOWN-LIMITATIONS.txt': 'Windows 已用 Wine 驗證，尚無 Windows 真機驗收。\r\n原版完整 parity 及自然全任務通關未宣稱。\r\n'.encode(),
            }
            label = '純 EGA' if theme == 'ega' else 'EGA＋HD'
            instructions = f'''《武士傳說》{version} Windows {label}

解壓後開啟 kol.exe。完整版已含遊戲；引擎包須自行匯入合法抽取資料。
純 EGA 不含手繪圖像，F10 的 HD 選項停用。EGA＋HD 可在遊戲中以 F10 即時切換。
手繪圖片按需載入，久未使用的快取自動釋放，回到 EGA 時釋放 HD 圖片。
兩版均保留四首配樂與獨立 F9 開關；遊戲規則和 v3 存檔相同。
方向鍵移動，Enter 執行，Esc 返回，Ctrl+Q 離開。住宿會保存。
存檔：%AppData%\\kol-remake-cht\\save.dat
雙擊啟動的日誌：%AppData%\\kol-remake-cht\\kol.log，前次為 kol-previous.log。
引擎包匯入：python tools/import_game_data.py --source "合法資料目錄" --target "."
純 EGA 匯入不複製 HD 圖像。資料須與 import-manifest.json 全部相符。
完整版含原作衍生資料，只留本機，不得上傳。引擎授權見 LICENSE。
'''
            content['README.txt'] = b'\xef\xbb\xbf' + instructions.replace('\n', '\r\n').encode()
            if bundle == 'full':
                content['DO-NOT-DISTRIBUTE.txt'] = '含原作衍生資料，僅本機保存，不得上傳或散布。\n'.encode()
            else:
                content['import-manifest.json'] = (json.dumps(dict(schema='kol-required-data/1', version=version,
                    graphics=theme, source='使用者合法持有、按本專案流程抽出的遊戲資料',
                    rights='清單不包含素材位元組；匯入資料不得散布。', files=selected), ensure_ascii=False, indent=2)+'\n').encode()
                assert not any(n.startswith('assets/') or Path(n).suffix.lower() in ('.png', '.ogg', '.pdf', '.zip') for n in paths)
            rows = [dict(path=n, bytes=p.stat().st_size, sha256=sha(p)) for n,p in paths.items()]
            rows += [dict(path=n, bytes=len(b), sha256=hashlib.sha256(b).hexdigest()) for n,b in content.items()]
            manifest = dict(schema='kol-windows-variant/1', version=version, graphics=theme, bundle=bundle,
                            local_only=bundle=='full', files=sorted(rows,key=lambda r:r['path']))
            content['manifest.json'] = (json.dumps(manifest,ensure_ascii=False,indent=2)+'\n').encode()
            target = out / bundle / (stem+'.zip')
            with zipfile.ZipFile(target,'w',compression=zipfile.ZIP_DEFLATED,compresslevel=6,allowZip64=True) as z:
                for name in sorted(set(paths)|set(content)):
                    info=zipfile.ZipInfo(stem+'/'+name,(2026,10,17,0,0,0))
                    info.compress_type=zipfile.ZIP_DEFLATED
                    info.external_attr=0o100644 << 16
                    with z.open(info,'w',force_zip64=True) as dst:
                        if name in content: dst.write(content[name])
                        else:
                            with paths[name].open('rb') as src: shutil.copyfileobj(src,dst,1<<20)
            with zipfile.ZipFile(target) as z: assert z.testzip() is None
            packages.append(dict(path=str(target.relative_to(out)), bytes=target.stat().st_size, sha256=sha(target),
                                 graphics=theme, bundle=bundle, resource_files=len(selected)))
            print(target.name,target.stat().st_size,flush=True)
    write(out/'SHA256SUMS.json',dict(version=version,files=packages))
    (out/'release/SHA256SUMS.txt').write_text(''.join(f"{p['sha256']}  {Path(p['path']).name}\n" for p in packages if p['bundle']=='release'))


if __name__ == '__main__':
    main()
