from pathlib import Path
import hashlib,json,os
root=Path('workplace/platform-smoke-20261001')
assert not root.exists()
assert (root.parent.stat().st_uid,root.parent.stat().st_gid)==(os.getuid(),os.getgid())
root.mkdir()
for name in ['build','linux','windows','macos']: (root/name).mkdir()
files=[p for base in ['cmd','internal'] for p in Path(base).rglob('*.go') if not p.name.endswith('_test.go')]
files += [Path('go.mod'),Path('go.sum'),Path('docs/i18n/newchar-zh-Hant.json'),Path('docs/i18n/ui-zh-Hant.json'),
          Path('docs/i18n/world-prompts-zh-Hant.tsv'),Path('assets/ui-font/manifest.json'),Path('assets/ui-font/zh-hant-16.png')]
files += list(Path('assets/original-derived-data').glob('*.json'))
digest=lambda p:hashlib.sha256(Path(p).read_bytes()).hexdigest()
doc={'schema':'kol-platform-smoke-input-v1','date':'2026-10-01','scope':'目前工作樹中間編譯／啟動驗證，未建立或覆寫正式dist-all交付',
     'current_source_sha256':{str(p):digest(p) for p in sorted(set(files))},'release_version':None,'build_version':'development',
     'reference_binary':{'path':'workplace/cjk-edge-20261001/kol-edge-fixed','sha256':digest('workplace/cjk-edge-20261001/kol-edge-fixed')},
     'platforms':['linux/amd64','windows/amd64','darwin/arm64','darwin/amd64'],
     'game_seed':1,'rights':'素材、保存及binary中間產物僅本機；目視PNG可依既有private GitHub特例保存'}
(root/'inputs.json').write_text(json.dumps(doc,ensure_ascii=False,indent=2)+'\n')
source=Path('docs/verification/inventory-empty-slots-20261001/normal-script.txt')
assert len(source.read_text().split())==198
(root/'purchase-script.txt').write_text(source.read_text())
print('建立中間驗證輸出，已固定',len(files),'份目前來源；正常購買腳本198步，正式版號未動。')
