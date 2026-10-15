"""在空的/out建立重跑輸入；已解開的歷史證據唯讀掛/evidence。"""
from pathlib import Path
import json,shutil
source=Path('/evidence');out=Path('/out')
assert source.is_dir() and out.is_dir() and not any(out.iterdir())
assert out.stat().st_uid==out.stat().st_gid==1000
names=['monster-all-gallery_test.go.txt','monster-gui_test.go.txt','final-source-inputs-v2.json',
       'generation-plan.json','plan.json','baseline-source.tar','formal-asset-inputs.json']
for name in names:shutil.copyfile(source/name,out/name)
generation=json.loads((out/'generation-plan.json').read_text())
for row in generation['entries']:shutil.copyfile(source/row['file'],out/row['file'])
# 候選只作唯讀比較基準；新正式畫格寫入自己的版本目錄。
for name in ['all-monster-gallery-v1','all-monster-gallery-v1.json']:
    (out/name).symlink_to(source/name,target_is_directory=(source/name).is_dir())
tools=Path('/src/docs/verification/hd-monsters-runtime-20261015')
for target,name in [('monitor-gallery.py','monitor-gallery.py'),('prepare-regression.py','prepare-regression.py'),
                    ('check-gallery.py','gallery-check.py'),('compare-formal.py','compare-formal.py'),
                    ('check-regression.py','check-regression.py')]:shutil.copyfile(tools/name,out/target)
(out/'bin').mkdir()
print('Fresh replay inputs ready;402-source index and candidate evidence remain fixed')
