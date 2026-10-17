"""在發行容器保存精簡證據；不改動封包或正式遊戲來源。"""
from pathlib import Path
import hashlib
import json
import shutil
import sys
from PIL import Image

ROOT = Path('/src')
VERSION = 'v.1.0.3-20261017'
OUT = ROOT / 'dist-all' / VERSION
DOC = ROOT / 'docs/verification' / ('release-' + VERSION.replace('v.', 'v', 1))
WORK = ROOT / 'workplace' / ('release-' + VERSION.replace('v.', 'v', 1))

def sha(path):
    h = hashlib.sha256()
    with path.open('rb') as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b''):
            h.update(chunk)
    return h.hexdigest()

def read(path):
    return json.loads(path.read_text())

def write(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')

def copy(source, target):
    assert source.is_file(), source
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(source, target)
    assert sha(source) == sha(target)

def prepare():
    assert ROOT.stat().st_uid == OUT.stat().st_uid == 1000
    DOC.mkdir(exist_ok=True)
    for name in ('checks', 'sampling', 'structures', 'release-smoke', 'linux-controls',
                 'hd-scenes-comparison'):
        copy(OUT / 'smoke' / (name + '.json'), DOC / (name + '.json'))
    for name in ('sources.json', 'ffprobe.json', 'visual-review.json'):
        copy(OUT / 'promo' / name, DOC / ('promo-' + name))
    previews = {
        'inn-hd-controls.png': 'smoke/linux-full-controls/inn-hd.png',
        'game-menu-ega.png': 'smoke/linux-full-controls/menu-ega.png',
        'game-menu-hd.png': 'smoke/linux-full-controls/menu-hd.png',
        'battle-ega.png': 'smoke/hd-scenes-linux/motion-runtime/battle-08-map-16-first-ega.png',
        'battle-hd.png': 'smoke/hd-scenes-linux/motion-runtime/battle-08-map-16-first-hd.png',
        'world-hd.png': 'smoke/hd-scenes-linux/motion-runtime/map-world-first-hd.png',
        'town3-hd.png': 'smoke/hd-scenes-linux/motion-runtime/map-town3-first-hd.png',
        'promo-poster.png': 'promo/frame-02.png',
    }
    rows = []
    for name, relative in previews.items():
        source = OUT / relative
        image = Image.open(source).convert('RGB')
        original = list(image.size)
        image.thumbnail((768, 768), Image.Resampling.LANCZOS)
        target = DOC / name
        image.save(target, optimize=True)
        rows.append(dict(path=name, source=str(source.relative_to(ROOT)),
                         source_sha256=sha(source), source_dimensions=original,
                         dimensions=list(image.size), sha256=sha(target),
                         tool='Pillow 9.4.0 / LANCZOS',
                         evidence='actual package' if '/linux-full-controls/' in relative else 'controlled renderer' if '/motion-runtime/' in relative else 'decoded promo frame',
                         rights='private GitHub 可保存；無公開散布權'))
    write(DOC / 'screenshots.json', dict(version=VERSION, files=rows))
    tools = ['closeout_release.py', 'compose_release_promo.py', 'closeout_macos.sh',
             'verify_release_runtime.py', 'verify_package_controls.py',
             'verify_music_runtime.py', 'run_release_linux.sh', 'run_release_wine.sh',
             'run_release_sample.sh', 'capture_wine_audio.sh', 'import_game_data.py', 'png_io.py']
    for name in tools:
        copy(ROOT / 'tools' / name, DOC / 'recipe' / name)
    copy(ROOT / 'docker/Dockerfile.package', DOC / 'recipe/Dockerfile.package')
    for source in sorted(WORK.iterdir()):
        if source.is_file():
            copy(source, DOC / 'recipe/workspace' / source.name)
    packages = sorted((OUT / 'release').glob('kol-' + VERSION + '-release-*'))
    assert len(packages) == 3 and all(p.suffix in ('.zip', '.AppImage') for p in packages)
    sums = OUT / 'release/SHA256SUMS.txt'
    sums.write_text(''.join(sha(p) + '  ' + p.name + '\n' for p in packages))
    write(OUT / 'smoke/publication-authorization.json', dict(
        version=VERSION, approved=True, destination='wicanr2/kol-remake-cht',
        visibility='private', attachments={p.name: sha(p) for p in packages + [sums]},
        basis='沿使用者已明確核准之 private Release 四附件範圍，及本輪完成後打包 release 指示；完整版與影片留本機，舊版保持。'))
    print('PREPARED', len(rows), 'previews; exact four publication attachments')

def finish():
    packages = read(OUT / 'smoke/packages.json')
    assert packages['status'] == 'PASS' and packages['ready_for_publication'] is True
    for name in ('windows-smoke', 'windows-pixels', 'packages', 'publication-authorization', 'tool-images'):
        copy(OUT / 'smoke' / (name + '.json'), DOC / (name + '.json'))
    copy(OUT / 'build-inputs.json', DOC / 'build-inputs.json')
    for name in ('tests.log', 'vet.log', 'wine-memory-before.txt', 'wine-memory-after.txt', 'wine-memory-peak.txt'):
        copy(OUT / 'smoke' / name, DOC / name)
    index = read(OUT / 'SHA256SUMS.json')
    rows = [x for x in index['files'] if x['path'].startswith(('full/kol-' + VERSION + '-full-', 'release/kol-' + VERSION + '-release-')) and Path(x['path']).suffix in ('.AppImage', '.zip')]
    assert len(rows) == 6
    movie_path = 'promo/kol-' + VERSION + '-promo.mp4'
    movie = next(x for x in index['files'] if x['path'] == movie_path)
    pixels = read(DOC / 'windows-pixels.json')
    assert pixels['status'] == 'PASS'
    write(DOC / 'receipt.json', dict(
        schema='kol-release-receipt/3', version=VERSION, technical_status='PASS',
        publication_status='READY', packages=rows, movie=movie,
        source_base=read(DOC / 'build-inputs.json')['git_head'],
        program_files=430, program_sha256='c1b5969e75604926a9dead27d65de6db1923b2923d2b81f211254e5e9a7634ff',
        hd_asset_files=832, imported_resource_files=1744,
        controlled_scenes=13, frames_per_platform=52,
        cross_platform_ega_bound=0, cross_platform_hd_bound=1,
        actual_package_pixel_comparisons=len(pixels['comparisons']),
        normal_sample_states=1046, new_process_reload_states=13, normal_sample_count=1,
        original_oracle=False, whole_game_parity=False, natural_all_quests_to_ending=False,
        windows_real_machine=False, macos_execution=False, host_fuse=False,
        local_output=str(OUT.relative_to(ROOT)),
        release_url='https://github.com/wicanr2/kol-remake-cht/releases/tag/' + VERSION,
        rights=dict(full='僅本機', promo='僅本機', release='只上傳目前 private repo 的三平台引擎包及 SHA256SUMS.txt'),
        verification=dict(complete='packages.json', controlled_hd='hd-scenes-comparison.json',
                          actual_package_pixels='windows-pixels.json', linux_controls='linux-controls.json',
                          windows_controls='windows-smoke.json', normal_save_reload='sampling.json',
                          promo_sources='promo-sources.json', promo_visual='promo-visual-review.json',
                          inputs='build-inputs.json', images='screenshots.json', toolchain='tool-images.json'),
        environment_limits=['首次HD載入在限3CPU、軟體顯示容器約194秒；非真機效能或最低需求。',
                            'Windows首次12GiB容器的F10冷載入等待逾時，保留原紀錄；16GiB無swap重新驗證。',
                            'vet兩則既有book.Label未具名欄位警示保留；完整cmd/internal測試通過。']))
    manifest()
    print('FORMAL PASS', len(rows), 'packages; movie', movie['sha256'])

def manifest():
    rows = [dict(path=str(p.relative_to(DOC)), bytes=p.stat().st_size, sha256=sha(p))
            for p in sorted(DOC.rglob('*')) if p.is_file() and p.name != 'manifest.json']
    write(DOC / 'manifest.json', dict(schema='kol-verification-manifest/1', version=VERSION,
          rights='private GitHub 可保存；截圖與原作衍生內容無公開散布權', files=rows))

if __name__ == '__main__':
    assert Path('/.dockerenv').exists()
    {'prepare': prepare, 'finish': finish, 'manifest': manifest}[sys.argv[1]]()
