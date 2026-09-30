#!/usr/bin/env python3
"""在 Docker 中由既有已驗證 PNG 重生審查圖集；不修改 runtime 素材。"""
import argparse
import hashlib
import json
import math
import pathlib
import struct
import zlib
import zipfile
from png_io import Image, read_png, write_png_rgb

ROOT = pathlib.Path(__file__).resolve().parents[1]
ASSETS = ROOT / "assets/original-derived-png"
ZIP_HASH = "2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46"


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def read_image(path):
    # 地圖索引是 type 0 灰階資料；不把灰階當作 EGA 調色盤。
    raw = path.read_bytes()
    width, height, depth, kind = struct.unpack(">IIBB", raw[16:26])
    if kind != 0:
        return read_png(path)
    if depth != 8:
        raise ValueError("只支援已存在的 8-bit 灰階資料")
    pos, packed = 8, bytearray()
    while pos < len(raw):
        size = struct.unpack_from(">I", raw, pos)[0]
        tag, data = raw[pos+4:pos+8], raw[pos+8:pos+8+size]
        if tag == b"IDAT":
            packed += data
        pos += size+12
    decoded = zlib.decompress(packed)
    if len(decoded) != height*(width+1):
        raise ValueError("灰階長度不符")
    pixels = bytearray()
    # 正式灰階抽取器固定 filter 0；其他值拒絕，避免猜測格式。
    for y in range(height):
        row = decoded[y*(width+1):(y+1)*(width+1)]
        if row[0] != 0:
            raise ValueError("灰階 filter 不符既有抽取契約")
        for value in row[1:]:
            pixels += bytes([value])*3
    return Image(width, height, 3, bytes(pixels))


def inputs():
    manifests = {}
    groups = []
    def load(relative):
        path = ASSETS / relative
        manifests[str(path.relative_to(ROOT))] = sha(path)
        return json.loads(path.read_text())
    main = load("ega/manifest.json")
    if main["source_zip"]["sha256"] != ZIP_HASH:
        raise ValueError("素材來源 ZIP 不符")
    for archive in main["archives"]:
        groups.append({"name": archive["archive"], "kind": archive["composition"],
                       "evidence": archive["evidence"], "archive_root": archive["archive_root"],
                       "items": [dict(i, path="ega/"+i["file"]) for i in archive["images"]]})
    battle = load("ega/BATTLE/manifest.json")
    for key in ["images", "maps"]:
        groups.append({"name": "BATTLE-"+key.upper(), "kind": "battle_tiles" if key=="images" else "map_cell_array",
                       "evidence": battle["evidence"], "archive_root": battle["source_archive"] if key=="images" else "Knights-of-legend/TOWN0",
                       "items": [dict(i, unit_index=n, path="ega/BATTLE/"+i["file"]) for n,i in enumerate(battle[key])]})
    shield = load("bookdata/manifest.json")
    groups.append({"name": "DEFAULT-SHIELD", "kind": "default_shield", "evidence": shield["evidence"],
                   "archive_root": shield["source_entry"], "items": [dict(i,unit_index=n,path="bookdata/"+i["file"]) for n,i in enumerate(shield["images"])]})
    fonts = load("fonts/manifest.json")
    groups.append({"name": "FONTS", "kind": "font_ink", "evidence": fonts["encoding"], "archive_root": "Knights-of-legend/BOOKDATA",
                   "items": [dict(i,unit_index=n,path="fonts/"+i["file"],width=i["sheet"]["width"],height=i["sheet"]["height"]) for n,i in enumerate(fonts["fonts"])]})
    expected = {str((ASSETS/i["path"]).relative_to(ROOT)) for g in groups for i in g["items"]}
    actual = {str(p.relative_to(ROOT)) for p in ASSETS.rglob("*.png") if "contact-sheets" not in p.parts}
    if expected != actual or len(expected) != 803:
        raise ValueError(f"PNG 母體不符：缺 {actual-expected}，多 {expected-actual}")
    return groups, manifests, fonts["fonts"][1]["glyphs"]


def generate(output):
    output = output.resolve()
    if output.exists():
        raise ValueError("輸出已存在，拒絕覆寫")
    if not any(output.is_relative_to(ROOT / p) for p in ["workplace", "assets/original-derived-png"]):
        raise ValueError("輸出必須在研究區或原版衍生 PNG 目錄")
    if output.parent.stat().st_uid != __import__("os").getuid():
        raise ValueError("輸出目錄擁有權不符")
    groups, manifests, glyphs = inputs()
    font = {i["codepoint"]: bytes.fromhex(i["rows_hex"]) for i in glyphs}
    loaded = {}
    stats = {}
    # 所有來源在建立輸出前先完成雜湊、形狀、alpha 與色盤檢查。
    palette = {(0,0,0),(0,0,170),(0,170,0),(0,170,170),(170,0,0),(170,0,170),(170,85,0),(170,170,170),
               (85,85,85),(85,85,255),(85,255,85),(85,255,255),(255,85,85),(255,85,255),(255,255,85),(255,255,255)}
    for group in groups:
        for item in group["items"]:
            path = ASSETS/item["path"]
            if sha(path) != item["png_sha256"]:
                raise ValueError(f"來源 PNG 雜湊不符：{path}")
            image = read_image(path)
            if (image.width,image.height) != (item.get("width",64),item.get("height",64)):
                raise ValueError(f"來源尺寸不符：{path}")
            colors = set()
            alphas = set()
            for pos in range(0,len(image.pixels),image.channels):
                colors.add(tuple(image.pixels[pos:pos+3]))
                alphas.add(image.pixels[pos+3] if image.channels==4 else 255)
            gray = struct.unpack(">IIBB",path.read_bytes()[16:26])[3]==0
            if not gray and not colors<=palette:
                raise ValueError(f"來源色盘越界：{path}".replace("盘","盤"))
            if not alphas<={0,255}:
                raise ValueError(f"來源非二值 alpha：{path}")
            if "has_transparency" in item and item["has_transparency"] != (0 in alphas):
                raise ValueError(f"透明旗標不符：{path}")
            loaded[item["path"]] = image
            stats[item["path"]] = {"color_count":len(colors),"alpha_values":sorted(alphas),"png_kind":"gray_data" if gray else "ega_rgba"}
    archive = ROOT / "Knights-of-legend.zip"
    if sha(archive) != ZIP_HASH:
        raise ValueError("原始 ZIP 雜湊不符")
    with zipfile.ZipFile(archive) as source_zip:
        for group in groups:
            roots = [group["archive_root"]]
            if group["kind"] == "font_ink":
                roots.append("Knights-of-legend/KOLDATA")
            members = source_zip.namelist()
            # 有 .0 時是索引 archive；同名無副檔名角色記錄表不是它的來源。
            names = [n for n in members if any((n == r and r+".0" not in members) or
                     (n.startswith(r+".") and n[len(r)+1:].isdigit()) for r in roots)]
            if not names:
                raise ValueError("原版 archive 定位不存在")
            group["source_entries"] = [{"entry":n,"sha256":hashlib.sha256(source_zip.read(n)).hexdigest()} for n in sorted(names)]
    output.mkdir()
    pages=[]
    def pixel(buf,w,x,y,color):
        start=(y*w+x)*3
        buf[start:start+3]=bytes(color)
    def text(buf,w,x,y,value):
        for n,ch in enumerate(value):
            for dy,bits in enumerate(font.get(ord(ch),b"")):
                for dx in range(8):
                    if bits & (128>>dx):
                        pixel(buf,w,x+n*8+dx,y+dy,(240,240,240))
    for group in groups:
        # 尺寸大圖獨立分頁，保留小圖的可讀比例，不把大圖裁切成縮圖。
        subsets = [[i for i in group["items"] if loaded[i["path"]].width>=224 or loaded[i["path"]].height>=160],
                   [i for i in group["items"] if loaded[i["path"]].width<224 and loaded[i["path"]].height<160]]
        page_number=0
        for items in subsets:
            if not items:
                continue
            maxw=max(loaded[i["path"]].width for i in items)
            maxh=max(loaded[i["path"]].height for i in items)
            scale=2 if maxw<=112 and maxh<=148 else 1
            cellw=max(96,maxw*scale+16)
            cellh=maxh*scale+28
            columns=max(1,960//cellw)
            per_page=columns*max(1,760//cellh)
            for offset in range(0,len(items),per_page):
                batch=items[offset:offset+per_page]
                width=columns*cellw+16
                height=math.ceil(len(batch)/columns)*cellh+36
                pixels=bytearray(bytes((32,32,32))*width*height)
                text(pixels,width,8,8,group["name"]+" "+str(page_number))
                placements=[]
                for n,item in enumerate(batch):
                    image=loaded[item["path"]]
                    x=8+(n%columns)*cellw
                    y=28+(n//columns)*cellh
                    text(pixels,width,x,y,f'{item["unit_index"]:03d} {image.width}x{image.height}')
                    y+=12
                    for dy in range(image.height*scale):
                        for dx in range(image.width*scale):
                            src=((dy//scale)*image.width+dx//scale)*image.channels
                            alpha=image.pixels[src+3] if image.channels==4 else 255
                            bg=112 if ((dx//8+dy//8)%2) else 160
                            color=tuple(image.pixels[src+c] if alpha else bg for c in range(3))
                            pixel(pixels,width,x+dx,y+dy,color)
                    placements.append({"source_png":str((ASSETS/item["path"]).relative_to(ROOT)),"source_png_sha256":item["png_sha256"],
                                       "source_payloads":item.get("source_payloads",[]),"source_unit":item["unit_index"],
                                       "rectangle":[x,y,image.width*scale,image.height*scale],"native_size":[image.width,image.height],
                                       "scale":scale,"checks":stats[item["path"]]})
                filename=group["name"].lower()+f"-{page_number:02d}.png"
                digest=write_png_rgb(output/filename,width,height,bytes(pixels))
                pages.append({"file":filename,"png_sha256":digest,"width":width,"height":height,
                              "group":group["name"],"kind":group["kind"],"archive_root":group["archive_root"],
                              "evidence":group["evidence"],"source_entries":group["source_entries"],
                              "items":placements,"visual_status":"pending"})
                page_number+=1
    receipt={"schema":"kol-contact-sheets-v1","tool":"tools/build_asset_contact_sheets.py",
             "tool_sha256":sha(pathlib.Path(__file__)),"source_zip_sha256":ZIP_HASH,"source_manifests":manifests,
             "groups":len(groups),"source_png_count":803,"pages":pages,"status":"pending-visual-review",
             "render_policy":"原尺寸或整數最近鄰 2 倍，不裁切；二值 alpha 以灰色棋盤顯示。灰階地圖為索引資料，遮罩／raw-planar 保留資料意義。",
             "sequence_scope":"依原 manifest 的 unit_index，展示所有來源圖號；不宣稱動畫時長、caller 動態裁切或自然遊玩全部可達。",
             "rights":"private_repo_original_derivative；已驗證來源 PNG 的可重生審查合成圖，目視核准前只留本機；不得公開發布"}
    (output/"manifest.json").write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+"\n")
    print(f"已重生 {len(pages)} 張審查圖集，涵蓋 {len(groups)} 批、803 張 PNG；等待逐頁目視。")


if __name__=="__main__":
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output",type=pathlib.Path,required=True)
    args=parser.parse_args()
    generate(args.output)
