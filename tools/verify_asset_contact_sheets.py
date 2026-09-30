#!/usr/bin/env python3
"""在 Docker 核實正式接觸圖集的完整母體、來源、目視狀態與逐像素合成。"""
import argparse
import json
import pathlib
import zipfile
from png_io import read_png
from build_asset_contact_sheets import read_image, sha, ZIP_HASH

ROOT=pathlib.Path(__file__).resolve().parents[1]
DEFAULT=ROOT/"assets/original-derived-png/contact-sheets"


def verify(folder, root=ROOT):
    data=json.loads((folder/"manifest.json").read_text())
    assert data["schema"]=="kol-contact-sheets-v1" and data["status"]=="passed", "圖集未核准"
    assert data["rights_classification"]=="private_repo_original_derivative", "權利分類不符"
    assert data["source_zip_sha256"]==ZIP_HASH and len(data["pages"])==77, "來源／頁數不符"
    assert sha(root/data["tool"])==data["tool_sha256"], "重生器版本不同"
    for name,digest in data["source_manifests"].items():
        assert sha(root/name)==digest, "上游 manifest 改變"
    expected={str(p.relative_to(root)) for p in (root/"assets/original-derived-png").rglob("*.png") if "contact-sheets" not in p.parts}
    seen=[]
    archive=root/"Knights-of-legend.zip"
    assert sha(archive)==ZIP_HASH, "原始 ZIP 不符"
    with zipfile.ZipFile(archive) as z:
        for page in data["pages"]:
            assert page["visual_status"]=="passed" and page["visual_review"]["scope"], "缺逐頁目視"
            path=folder/page["file"]
            assert path.resolve().parent==folder.resolve(), "圖集路徑越界"
            assert sha(path)==page["png_sha256"], "圖集 PNG 改變"
            image=read_png(path)
            assert (image.width,image.height)==(page["width"],page["height"]), "圖集尺寸錯誤"
            for entry in page["source_entries"]:
                import hashlib
                assert hashlib.sha256(z.read(entry["entry"])).hexdigest()==entry["sha256"], "原版 entry 改變"
            boxes=[]
            for item in page["items"]:
                seen.append(item["source_png"])
                source=root/item["source_png"]
                assert source.resolve().is_relative_to((root/"assets/original-derived-png").resolve()), "來源路徑越界"
                assert sha(source)==item["source_png_sha256"], "來源 PNG 改變"
                sprite=read_image(source)
                x,y,w,h=item["rectangle"]
                scale=item["scale"]
                assert scale in (1,2) and [sprite.width,sprite.height]==item["native_size"], "來源尺寸／倍率不同"
                assert (w,h)==(sprite.width*scale,sprite.height*scale), "縮放尺寸不同"
                assert 0<=x and 0<=y and x+w<=image.width and y+h<=image.height, "裁切越界"
                for ox,oy,ow,oh in boxes:
                    assert x+w<=ox or ox+ow<=x or y+h<=oy or oy+oh<=y, "圖塊重疊"
                boxes.append((x,y,w,h))
                for dy in range(h):
                    expected_row=bytearray()
                    for dx in range(w):
                        pos=((dy//scale)*sprite.width+dx//scale)*sprite.channels
                        alpha=sprite.pixels[pos+3] if sprite.channels==4 else 255
                        assert alpha in (0,255), "alpha 非二值"
                        if alpha:
                            expected_row+=sprite.pixels[pos:pos+3]
                        else:
                            expected_row+=bytes([112 if ((dx//8+dy//8)%2) else 160])*3
                    start=((y+dy)*image.width+x)*image.channels
                    assert image.pixels[start:start+w*3]==expected_row, "像素／透明極性／縮放不同"
    assert len(seen)==len(set(seen))==data["source_png_count"]==803, "重複或遺漏 PNG"
    assert set(seen)==expected, "未完整覆蓋正式母體"
    assert len({p["group"] for p in data["pages"]})==data["groups"]==44, "批次未齊"
    return data


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--sheets",type=pathlib.Path,default=DEFAULT)
    args=parser.parse_args()
    try:
        verify(args.sheets)
    except (AssertionError,KeyError,OSError,ValueError) as e:
        parser.exit(1,f"圖集核實失敗：{e}\n")
    print("77 頁／44 批／803 張來源核實通過：圖號完整、矩形不重疊、逐像素／透明度一致、來源與權利紀錄齊備。")


if __name__=="__main__":
    main()
