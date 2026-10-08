from pathlib import Path
import json,csv,hashlib,os,sys,shutil
root=Path('workplace/cjk-equipment-20261001')
assert (root.parent.stat().st_uid,root.parent.stat().st_gid)==(os.getuid(),os.getgid())
if root.exists():
    assert sys.argv[1:]==['--refresh']
    backup=root/'selection-before-consumer-review.json';assert not backup.exists()
    shutil.copyfile(root/'selection.json',backup)
else:root.mkdir()
names={r['en'].casefold():r['zh-Hant'] for r in csv.DictReader(Path('docs/i18n/display-names-zh-Hant.tsv').open(),delimiter='\t')}
font=json.loads(Path('assets/ui-font/manifest.json').read_text())['fonts'][0]
adv={g['character']:g['advance'] for g in font['glyphs']}
def entry(g,source):
    text=names[g['name'].casefold()]
    return {'name':g['name'],'zh':text,'width':sum(adv.get(c,font['advance']) for c in text),'source':source,'position':g['position']}
towns=json.loads(Path('assets/original-derived-data/town-buildings.json').read_text())['towns']
roster=json.loads(Path('assets/original-derived-data/character-roster.json').read_text())['slots']
population=[entry(g,f'城{t["town"]}/門{d["index"]}/商品{g["index"]}') for t in towns for d in t['doors'] for g in d.get('goods',[]) if g.get('kind')==2 and g['position'] in [0,1,2,6]]
population += [entry(g,f'初始槽{c["slot"]}/{c["name"]}/穿戴{i}') for c in roster if c and c['occupied'] for i,g in enumerate(c['equipment']) if g and i in [0,1,2,6]]
assert max(x['width'] for x in population)==96
selected=next(x for x in population if x['name']=='Brigandine L' and '/Moraine/' in x['source'])
source=Path('docs/verification/cjk-ui/natural-sign-20261001')
original=(source/'script.txt').read_text().split()
assert len(original)==3080
prefix=original[:166]
inputs=[source/'receipt.json',source/'script.txt',source/'en-trace.json',Path('assets/original-derived-data/character-roster.json'),Path('assets/original-derived-data/town-buildings.json'),Path('assets/ui-font/manifest.json'),Path('docs/i18n/display-names-zh-Hant.tsv'),Path('internal/ui/battle_cjk.go'),Path('internal/battle/equipment_panel.go'),Path('docs/i18n/newchar-zh-Hant.json'),Path('docs/i18n/ui-zh-Hant.json')]
selection={'schema':'kol-cjk-equipment-selection-v1','selected_before_execution':True,'scope':'七城可販售及14名占用初始角色的頭／軀幹／腿／盾護甲名稱；不含怪物、動態改名或戰利品','population':population,'population_count':len(population),'maximum_width':96,'selected':selected,'source_prefix_inputs':166,'seed':1,'load':None,'ambush_rolls':None,'monsters':None,'quest_sites':None,'inputs_sha256':{str(p):hashlib.sha256(p.read_bytes()).hexdigest() for p in inputs}}
(root/'selection.json').write_text(json.dumps(selection,ensure_ascii=False,indent=2)+'\n')
(root/'scout-script.txt').write_text(' '.join(prefix+['SPACE'])+'\n')
print('已預先固定護甲母體',len(population),'與96像素樣本；正常路徑首166步後切換裝備資訊。')
