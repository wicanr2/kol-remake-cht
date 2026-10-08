from pathlib import Path
import json,os,shutil,hashlib
root=Path('workplace/cjk-equipment-20261001')
uid,gid=os.getuid(),os.getgid()
def owner(p):
    s=Path(p).stat();assert (s.st_uid,s.st_gid)==(uid,gid),str(p)
for p in [root,Path('docs/i18n'),Path('docs/spec'),Path('internal/battle'),Path('internal/ui')]:owner(p)
for p in ['docs/i18n/newchar-zh-Hant.json','docs/i18n/ui-zh-Hant.json','docs/spec/019-battle-equipment-panel.md','internal/battle/equipment_panel.go','internal/ui/battle_cjk.go']:owner(p)
spec=Path('docs/spec/019-battle-equipment-panel.md')
section='''
## 2026-10-01 繁中第六列勘誤

子契約狀態：**DRAFT**。玩家可見第六列原譯「身體」，但既有本規格的六列位移
與 SPEC-010／RE-094 已確認 +0x190 是腰帶武器；重製 `EquipmentPanel` 也讀取
`BeltName`。因此依術語表「入鞘／切換」使用的腰帶名稱，改標「腰帶」。
不改原始 B: 鍵、來源記錄、座標、亂數、裝備規則或保存格式；原版語意已證實，
此為繁中譯文回填，不新增原版逆向結論。

審查依據：本規格的固定 EXE 雜湊與 IDA 線性定位仍有效；SPEC-010 的
0x30／0x190 全記錄交換及 RE-094、`internal/battle/equipment_panel.go` 的
`m.BeltName`、術語表已核准的「入鞘／切換」共同指向腰帶。
初始穿戴盾牌應讀 +0x160（`equipment[6]`），不是 `equipment[4]`；
驗證模型及護甲母體須按正式消費端位移取值。

驗收：最寬護甲實際母體預先選樣；從標題、六人招募及自然遭遇核對六人裝備欄。
英中狀態、切換前後復原、完整文字框與正式字模像素核對通過；
96 像素護甲必須在 95 像素反例中被拒絕。原版／其他畫面與平台驗收不擴張。
'''
assert '## 2026-10-01 繁中第六列勘誤' not in spec.read_text()
spec.write_text(spec.read_text()+'\n'+section)
# 已有原版已證實的契約，只回填既有顯示譯文；先審查再實作。
roster=json.loads(Path('assets/original-derived-data/character-roster.json').read_text())['slots']
leader=next(c for c in roster if c and c.get('name')=='Theodoric')
fighter=next(c for c in roster if c and c.get('name')=='Fighting Bull')
assert leader['equipment'][9]['offset']==0x190 and leader['equipment'][9]['name']=='Greatsword'
assert fighter['equipment'][6]['offset']==0x160 and fighter['equipment'][6]['position']==6
assert 'd.Rows[0], d.Rows[4], d.Rows[5] = m.WeaponName, m.Worn[4], m.BeltName' in Path('internal/battle/equipment_panel.go').read_text()
assert '+0x190 起 16 byte 整段交換' in Path('docs/spec/010-combat-commands.md').read_text()
spec.write_text(spec.read_text().replace('子契約狀態：**DRAFT**。','子契約狀態：**READY**（上述既有原版契約與顯示消費端逐項審查通過）。',1))
backup={}
for filename in ['docs/i18n/newchar-zh-Hant.json','docs/i18n/ui-zh-Hant.json','docs/verification/cjk-text-fit.json']:
    p=Path(filename);b=root/('before-'+p.name);assert not b.exists();shutil.copyfile(p,b);backup[filename]=hashlib.sha256(b.read_bytes()).hexdigest()
p=Path('docs/i18n/newchar-zh-Hant.json');d=json.loads(p.read_text());entry=next(x for x in d['ui'] if x['en']=='B:')
assert entry['zh']=='身體：';entry['zh']='腰帶：';entry['note']='戰鬥裝備欄第 6 列；SPEC-019／RE-094 的 +0x190 腰帶武器'
p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
p=Path('docs/i18n/ui-zh-Hant.json');d=json.loads(p.read_text());r=d['banks']['KOLDATA']['257'];assert r['zh'].endswith('身：@');r['zh']=r['zh'][:-3]+'腰：@'
p.write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
p=Path('internal/battle/equipment_panel.go');s=p.read_text();assert '盾、身體' in s;p.write_text(s.replace('盾、身體','盾、腰帶'))
p=Path('internal/ui/battle_cjk.go');s=p.read_text();s=s.replace('// drawBattleSheetCJK 是左上那一格的裝備欄。六列標籤走譯文，值照原文印\n// ——那是物品名，不在譯文表裡。','// drawBattleSheetCJK 是左上那一格的裝備欄。六列標籤走譯文，\n// 物品值由 displayName 查顯示譯名；第六列是腰帶武器（SPEC-019）。');p.write_text(s)
(root/'label-change.json').write_text(json.dumps({'schema':'kol-battle-belt-label-correction-v1','backup_sha256':backup,'scope':'只改繁中譯文與註解；正式Go控制流、裝備規則、保存與英文畫面不变','old':'身體：','new':'腰帶：','KOLDATA_257':'身：→腰：'},ensure_ascii=False,indent=2)+'\n')
print('SPEC-019 子契約審查READY；B:改腰帶，盾牌驗證索引確認為equipment[6]。')
