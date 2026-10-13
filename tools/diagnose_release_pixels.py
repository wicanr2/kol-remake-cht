#!/usr/bin/env python3
"""Docker內的渲染診斷；入口：docs/verification/release-v1.0.2-20261013.md。"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess

OUT = Path('/out')
BASE = OUT / 'smoke/pixel-diagnosis'


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def winpath(path):
    return 'Z:' + str(path).replace('/', '\\')


def capture():
    BASE.mkdir(exist_ok=True)
    binary = OUT / 'full/kol.exe'
    fixture = OUT / 'smoke/linux-gui/inputs/save.json'
    for backend in ('auto', 'directx', 'opengl'):
        dest = BASE / backend
        dest.mkdir()
        env = dict(os.environ, EBITENGINE_GRAPHICS_LIBRARY=backend)
        command = ['/usr/lib/wine/wine64', winpath(binary), '-scene', 'play',
                   '-lang', 'zh-Hant', '-graphics', 'hd', '-load', winpath(fixture),
                   '-seed', '1', '-play-script', 'L R RET SPACE',
                   '-shot-steps', '0,1,2,3,4', '-shot-dir', winpath(dest)]
        write(dest / 'command.json', dict(argv=command, environment={
            'EBITENGINE_GRAPHICS_LIBRARY': backend}, binary_sha256=sha(binary),
            fixture_sha256=sha(fixture)))
        with (dest / 'run.log').open('w') as log:
            result = subprocess.run(command, cwd=OUT / 'full', env=env,
                                    stdout=log, stderr=subprocess.STDOUT, timeout=120)
        write(dest / 'execution.json', dict(returncode=result.returncode))
        print(backend, result.returncode, flush=True)


def native():
    dest=BASE/'linux-native-same-image'
    dest.mkdir()
    binary=OUT/'full/kol'
    fixture=OUT/'smoke/linux-gui/inputs/save.json'
    command=[str(binary), '-scene', 'play', '-lang', 'zh-Hant', '-graphics', 'hd',
             '-load', str(fixture), '-seed', '1', '-play-script', 'L R RET SPACE',
             '-shot-steps', '0,1,2,3,4', '-shot-dir', str(dest)]
    env=dict(os.environ, EBITENGINE_GRAPHICS_LIBRARY='opengl')
    write(dest/'command.json',dict(argv=command,environment={
        'EBITENGINE_GRAPHICS_LIBRARY':'opengl'}, binary_sha256=sha(binary), fixture_sha256=sha(fixture)))
    with (dest/'run.log').open('w') as log:
        subprocess.run(command,cwd=OUT/'full',env=env,stdout=log,
                       stderr=subprocess.STDOUT,check=True,timeout=120)
    print('linux-native-same-image PASS',flush=True)


def packages():
    """同一Wine image內執行實際AppImage；保留原跨image收據。"""
    version=json.loads((OUT/'build-inputs.json').read_text())['version']
    binary=OUT/'full'/f'kol-{version}-full-linux-x86_64.AppImage'
    fixture=OUT/'smoke/linux-gui/inputs/save.json'
    for graphics in ('ega','hd'):
        dest=BASE/('linux-package-'+graphics)
        dest.mkdir()
        command=[str(binary),'--appimage-extract-and-run','-scene','play',
            '-lang','zh-Hant','-graphics',graphics,'-load',str(fixture),'-seed','1',
            '-play-script','L R RET SPACE','-shot-steps','0,1,2,3,4','-shot-dir',str(dest)]
        env=dict(os.environ)
        env.pop('EBITENGINE_GRAPHICS_LIBRARY',None)
        env.pop('EBITEN_GRAPHICS_LIBRARY',None)
        write(dest/'command.json',dict(argv=command,environment={'graphics_library':'auto'},
            binary_sha256=sha(binary),fixture_sha256=sha(fixture),
            image='eob-remake-wine-verify:ubuntu-noble-20261008-r2'))
        with (dest/'run.log').open('w') as log:
            subprocess.run(command,cwd=Path('/tmp'),env=env,stdout=log,
                           stderr=subprocess.STDOUT,check=True,timeout=120)
        print('actual AppImage',graphics,'PASS',flush=True)


def matched():
    import numpy as np
    from PIL import Image
    comparisons=[]
    for graphics in ('ega','hd'):
        for bundle in ('full','release'):
            for path in sorted((BASE/('linux-package-'+graphics)).glob('step-*.png')):
                reference=OUT/f'smoke/wine-{bundle}-{graphics}'/path.name
                a=np.asarray(Image.open(path).convert('RGBA')).astype(np.int16)
                b=np.asarray(Image.open(reference).convert('RGBA')).astype(np.int16)
                assert a.shape==b.shape
                diff=np.abs(a-b);maximum=int(diff.max())
                comparisons.append(dict(bundle=bundle,graphics=graphics,frame=path.name,
                    max_channel_delta=maximum,different_pixels=int(np.count_nonzero(diff.max(2))),
                    predeclared_numeric_bound_met=maximum <= (0 if graphics=='ega' else 1),
                    linux_frame=str(path.relative_to(OUT)),linux_sha256=sha(path),
                    windows_frame=str(reference.relative_to(OUT)),windows_sha256=sha(reference)))
    assert len(comparisons)==20
    status='PASS' if all(row['predeclared_numeric_bound_met'] for row in comparisons) else 'NUMERIC_BOUND_UNMET'
    receipt=dict(schema='kol-matched-renderer-pixels/1',status=status,numeric_bound={'ega':0,'hd':1},
        comparisons=comparisons,production_changed=False,graphics_library='auto on both executables',
        controlled_image='eob-remake-wine-verify:ubuntu-noble-20261008-r2',
        controlled_image_id='sha256:c047a6a259a1fc9a90babf7c937310fa9df4b3465bf75f3fcaad1bc640099c2f',
        baseline_failure=dict(path='smoke/windows-pixels.json',sha256=sha(OUT/'smoke/windows-pixels.json'),
                              status='NUMERIC_BOUND_UNMET'),
        original_parity=False,windows_real_machine=False,publication_authorized=False)
    write(OUT/'smoke/windows-pixels-matched-renderer.json',receipt)
    print(status,json.dumps([{k:v for k,v in row.items() if k not in (
        'linux_frame','linux_sha256','windows_frame','windows_sha256')} for row in comparisons]),flush=True)


def analyze():
    import numpy as np
    from PIL import Image
    comparisons = []
    for backend in ('auto', 'directx', 'opengl', 'linux-native-same-image'):
        for path in sorted((BASE / backend).glob('step-*.png')):
            target = np.asarray(Image.open(path).convert('RGBA')).astype(np.int16)
            for label, reference in (
                ('linux-opengl', OUT / 'smoke/linux-reference-hd' / path.name),
                ('original-wine', OUT / 'smoke/wine-full-hd' / path.name),
                ('wine-opengl-same-image',BASE/'opengl'/path.name)):
                source = np.asarray(Image.open(reference).convert('RGBA')).astype(np.int16)
                diff = np.abs(target-source)
                coordinates = np.argwhere(diff.max(2) > 1)
                examples = [dict(x=int(x), y=int(y), reference=source[y,x].tolist(),
                                 target=target[y,x].tolist()) for y,x in coordinates[:12]]
                comparisons.append(dict(backend=backend, reference=label, frame=path.name,
                    max_channel_delta=int(diff.max()), different_pixels=int(np.count_nonzero(diff.max(2))),
                    over_one_pixels=len(coordinates), examples=examples,
                    source_sha256=sha(reference), target_sha256=sha(path)))
    write(BASE / 'receipt.json', dict(schema='kol-renderer-diagnosis/1',
        production_changed=False, numeric_bound=1, original_parity=False,
        comparisons=comparisons))
    print(json.dumps([{k:v for k,v in row.items() if k not in (
        'examples','source_sha256','target_sha256')} for row in comparisons], indent=2))


if __name__ == '__main__':
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('phase', choices=('capture','native','packages','matched','analyze'))
    args=parser.parse_args()
    assert Path('/.dockerenv').exists()
    globals()[args.phase]()
