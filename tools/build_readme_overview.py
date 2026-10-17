#!/usr/bin/env python3
"""以原生SVG版面拼排已驗證截圖；入口為docs/screenshots/README.md。"""
from pathlib import Path
import argparse
import base64
import ctypes
import ctypes.util
import hashlib
import json
import os
from PIL import Image

ROOT = Path('/src')
VERSION = 'v.1.0.3-20261017'
WIDTH = HEIGHT = 768
FONT = ROOT / 'assets/hd/first-slice/interface/font-atlas.json'
ATLAS = FONT.with_suffix('.png')
RELEASE = 'docs/verification/release-v1.0.3-20261017/'
FULL = 'docs/verification/hd-full-runtime-20261017/'
SOURCES = [
    ('書本與翻頁', FULL + 'book-open-018-hd.png', '受控場景'),
    ('六人隊伍與旅店', RELEASE + 'inn-hd-controls.png', '本版實際AppImage'),
    ('商店與服務子頁', 'docs/verification/hd-native-ui-runtime-20261014/town-shop-hd.png', '既有正式HD驗收'),
    ('人物裝備與隊伍', 'docs/verification/hd-equipment-runtime-20261015/party-hd.png', '既有正式HD驗收'),
    ('城鎮連續地圖', RELEASE + 'town3-hd.png', '本版受控場景'),
    ('世界探索', RELEASE + 'world-hd.png', '本版受控場景'),
    ('戰場與操作介面', RELEASE + 'battle-hd.png', '本版受控場景'),
    ('人物與傷勢', FULL + 'body-1-1-hd.png', '受控場景'),
    ('怪物遭遇', FULL + 'monster-native-encounter-hd.png', '受控場景'),
    ('雪景與天候', FULL + 'weather-4-01-hd.png', '受控場景'),
    ('世界事件', 'docs/verification/hd-scene-events-runtime-20261017/world-05-hd.png', '既有正式HD驗收'),
    ('遊戲內畫面切換', RELEASE + 'game-menu-hd.png', '本版實際AppImage'),
]

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def render_svg(svg, output):
    # 固定image已含librsvg；直接使用其公開ABI，不依賴未安裝的rsvg-convert。
    # https://gnome.pages.gitlab.gnome.org/librsvg/Rsvg-2.0/ctor.Handle.new_from_data.html
    # https://gnome.pages.gitlab.gnome.org/librsvg/Rsvg-2.0/method.Handle.render_document.html
    pointer = ctypes.c_void_p
    rsvg = ctypes.CDLL(ctypes.util.find_library('rsvg-2'))
    cairo = ctypes.CDLL(ctypes.util.find_library('cairo'))
    gobject = ctypes.CDLL(ctypes.util.find_library('gobject-2.0'))
    glib = ctypes.CDLL(ctypes.util.find_library('glib-2.0'))
    class Rectangle(ctypes.Structure):
        _fields_ = [(name, ctypes.c_double) for name in ('x', 'y', 'width', 'height')]
    bindings = [
        (rsvg, 'rsvg_handle_new_from_data', [pointer, ctypes.c_size_t, ctypes.POINTER(pointer)], pointer),
        (rsvg, 'rsvg_handle_render_document', [pointer, pointer, ctypes.POINTER(Rectangle), ctypes.POINTER(pointer)], ctypes.c_int),
        (cairo, 'cairo_image_surface_create', [ctypes.c_int, ctypes.c_int, ctypes.c_int], pointer),
        (cairo, 'cairo_create', [pointer], pointer),
        (cairo, 'cairo_status', [pointer], ctypes.c_int),
        (cairo, 'cairo_surface_status', [pointer], ctypes.c_int),
        (cairo, 'cairo_surface_write_to_png', [pointer, ctypes.c_char_p], ctypes.c_int),
        (cairo, 'cairo_destroy', [pointer], None),
        (cairo, 'cairo_surface_destroy', [pointer], None),
        (cairo, 'cairo_version_string', [], ctypes.c_char_p),
        (gobject, 'g_object_unref', [pointer], None),
        (glib, 'g_error_free', [pointer], None),
    ]
    for library, name, arguments, result in bindings:
        function = getattr(library, name)
        function.argtypes, function.restype = arguments, result
    error = pointer()
    data = ctypes.create_string_buffer(svg)
    handle = rsvg.rsvg_handle_new_from_data(data, len(svg), ctypes.byref(error))
    assert handle and not error.value, 'librsvg rejected SVG'
    surface = cairo.cairo_image_surface_create(0, WIDTH, HEIGHT)
    context = cairo.cairo_create(surface)
    try:
        assert cairo.cairo_status(context) == cairo.cairo_surface_status(surface) == 0
        viewport = Rectangle(0, 0, WIDTH, HEIGHT)
        assert rsvg.rsvg_handle_render_document(handle, context, ctypes.byref(viewport), ctypes.byref(error))
        assert cairo.cairo_surface_write_to_png(surface, str(output).encode()) == 0
    finally:
        cairo.cairo_destroy(context)
        cairo.cairo_surface_destroy(surface)
        gobject.g_object_unref(handle)
        if error.value:
            glib.g_error_free(error)
    version = '.'.join(str(ctypes.c_uint.in_dll(rsvg, 'rsvg_' + suffix + '_version').value)
                       for suffix in ('major', 'minor', 'micro'))
    return 'librsvg ' + version + ' / Cairo ' + cairo.cairo_version_string().decode()

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    out = args.output
    assert Path('/.dockerenv').exists() and out.is_dir()
    assert (out.stat().st_uid, out.stat().st_gid) == (os.getuid(), os.getgid())
    font = json.loads(FONT.read_text())
    atlas = Image.open(ATLAS).convert('RGBA')
    alpha = atlas.getchannel('A')
    headlines = [
        '武士傳說 HD 總覽', VERSION,
        '手繪圖形／繁體中文／遊戲內切換',
        '本版封包及各批HD驗收畫格',
        '遊戲內 F10 切換 EGA／HD，詳細比較見 README',
    ]
    texts = headlines + [row[0] for row in SOURCES]
    chars = sorted(set(''.join(texts)) - {' '})
    missing = set(chars) - set(font['glyphs'])
    assert not missing, sorted(missing)
    # 標籤以既有中文字模的alpha輪廓產生SVG路徑，不重繪截圖內容。
    definitions = []
    for char in chars:
        glyph = font['glyphs'][char]
        pieces = []
        for y in range(font['cell']):
            x = 0
            while x < font['cell']:
                if alpha.getpixel((glyph['x'] + x, glyph['y'] + y)) < 96:
                    x += 1
                    continue
                start = x
                while x < font['cell'] and alpha.getpixel((glyph['x'] + x, glyph['y'] + y)) >= 96:
                    x += 1
                pieces.append(f'M{start},{y}h{x-start}v1h{start-x}z')
        definitions.append(f'<path id="g{ord(char):x}" d="{" ".join(pieces)}"/>')

    def text(value, x, baseline, size, color='#e2cfaa'):
        scale = size / font['size']
        cursor = float(x)
        uses = []
        for char in value:
            if char == ' ':
                cursor += size * .32
                continue
            glyph = font['glyphs'][char]
            uses.append(f'<use xlink:href="#g{ord(char):x}" fill="{color}" '
                        f'transform="translate({cursor-font["xpad"]*scale:.4f},'
                        f'{baseline-font["baseline"]*scale:.4f}) scale({scale:.6f})"/>')
            cursor += glyph['advance'] * scale
        assert cursor <= WIDTH - 14, (value, cursor)
        return ''.join(uses)

    elements = [f'<rect width="{WIDTH}" height="{HEIGHT}" fill="#1b1015"/>',
                '<path d="M18 59H750M18 735H750" stroke="#947146"/>',
                text(headlines[0], 18, 35, 26),
                text(headlines[1], 542, 33, 11, '#b9a891'),
                text(headlines[2], 20, 53, 12, '#c1b2a0')]
    records = []
    for index, (title, relative, scope) in enumerate(SOURCES):
        source = ROOT / relative
        proof = source.parent / 'manifest.json'
        digest = sha(source)
        assert digest in proof.read_text(), ('source is absent from approved manifest', relative)
        with Image.open(source) as image:
            dimensions = list(image.size)
        assert dimensions[0] * 5 == dimensions[1] * 8, dimensions
        col, row = index % 3, index // 3
        x, y = 18 + col * 248, 68 + row * 166
        data = base64.b64encode(source.read_bytes()).decode('ascii')
        elements.extend([
            f'<rect x="{x}" y="{y}" width="236" height="147.5" fill="#352027" stroke="#987b50"/>',
            f'<image x="{x+1}" y="{y+1}" width="234" height="146.25" '
            f'preserveAspectRatio="xMidYMid meet" xlink:href="data:image/png;base64,{data}"/>',
            text(title, x + 4, y + 160, 12),
        ])
        records.append(dict(panel=index + 1, title=title, source=relative, sha256=digest,
                            dimensions=dimensions, evidence_scope=scope,
                            receipt=str(proof.relative_to(ROOT)), receipt_sha256=sha(proof),
                            frame_rectangle=[x + 1, y + 1, 234, 146.25], crop=False))
    elements.extend([text(headlines[3], 20, 749, 11, '#c1b2a0'),
                     text(headlines[4], 20, 764, 11, '#c1b2a0')])
    svg = ('<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" '
           f'width="{WIDTH}" height="{HEIGHT}" viewBox="0 0 {WIDTH} {HEIGHT}">'
           '<defs>' + ''.join(definitions) + '</defs>' + ''.join(elements) + '</svg>')
    version = render_svg(svg.encode(), out / 'hd-overview.png')
    with Image.open(out / 'hd-overview.png') as image:
        assert image.size == (WIDTH, HEIGHT) and image.mode in ('RGB', 'RGBA')
    manifest = dict(schema='kol-readme-hd-overview/1', version=VERSION, status='PASS',
                    output='hd-overview.png', dimensions=[WIDTH, HEIGHT],
                    sha256=sha(out / 'hd-overview.png'), bytes=(out / 'hd-overview.png').stat().st_size,
                    panels=records, panel_count=len(records), missing_glyphs=[],
                    svg_sha256=hashlib.sha256(svg.encode()).hexdigest(),
                    font=dict(metadata=str(FONT.relative_to(ROOT)), sha256=sha(FONT), atlas_sha256=sha(ATLAS),
                              method='native SVG paths from approved glyph alpha; cutoff 96'),
                    tool=dict(path='tools/build_readme_overview.py', sha256=sha(Path(__file__)),
                              image='kol-package:1.26.7-ebiten2.9.9-20261011-r1',
                              image_id='sha256:0b933af1238ab5839ad7a64b6ae05f069f57cb85da5917dfbe59730f6d7ba53f',
                              renderer=version, command='python3 /src/tools/build_readme_overview.py --output /out',
                              svg_api='rsvg_handle_render_document'),
                    composition='12 complete screenshots, proportional SVG layout, no crop or artwork replacement',
                    visual_review='pending', normal_full_playthrough=False, original_parity_claimed=False,
                    rights='private GitHub 可保存；含原作衍生截圖及核准HD，無公開散布權')
    (out / 'hd-overview.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + '\n')
    print('OVERVIEW PASS', len(records), 'panels', WIDTH, HEIGHT, manifest['bytes'], flush=True)

if __name__ == '__main__':
    main()
