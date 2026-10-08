"""Docker 內獨立合成完整296×32圖示列；不裁掉圖示、框或間隙。"""
from pathlib import Path
import sys, json
sys.path.insert(0, '/src/tools')
from png_io import read_png

ROOT, OUT = Path('/src'), Path('/out')
ASSETS = ROOT / 'assets/original-derived-png/ega/KOLMPICS'
WEAPON = [0x0F, 0x2A, 0x2B, 0x2C, 0x2D, 0x2E]
GUARD = [0x0F, 0x2A, 0x30, 0x31, 0x32, 0x33, 0x34, 0x36]


def expected(codes, selected):
    bg = read_png(ASSETS / 'kolmpics-000.png')
    assert (bg.width, bg.height) == (320, 200)
    rgb = bytearray(bg.rgb())
    def setpixel(x, y, c):
        i = (y * 320 + x) * 3
        rgb[i:i+3] = c
    for slot, code in enumerate(codes):
        im = read_png(ASSETS / f'kolmpics-{code+3:03d}.png')
        assert (im.width, im.height) == (24, 20)
        data = im.rgb()
        x = 32 + 32 * slot
        for yy in range(20):
            for xx in range(24):
                i = (yy * 24 + xx) * 3
                setpixel(x+xx, 171+yy, data[i:i+3])
        c = bytes([255, 255, 255] if slot == selected else [85, 85, 255])
        for xx in range(x, x+24):
            setpixel(xx, 170, c)
            setpixel(xx, 191, c)
        for yy in range(170, 192):
            setpixel(x, yy, c)
            setpixel(x+23, yy, c)
    return bytes(v for y in range(164, 196) for x in range(8, 304)
                 for v in rgb[(y*320+x)*3:(y*320+x)*3+3])


def difference(path, want, scale):
    im = read_png(path)
    assert (im.width, im.height) == (320*scale, 200*scale)
    raw = im.crop(8*scale, 164*scale, 296*scale, 32*scale).rgb()
    diff = []
    for y in range(32*scale):
        for x in range(296*scale):
            i = (y*296*scale+x)*3
            j = ((y//scale)*296+x//scale)*3
            if raw[i:i+3] != want[j:j+3]:
                diff.append((x, y))
    bbox = ([min(x for x,y in diff), min(y for x,y in diff),
             max(x for x,y in diff), max(y for x,y in diff)] if diff else None)
    return len(diff), bbox

