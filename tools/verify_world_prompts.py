#!/usr/bin/env python3
"""核對世界提示 TSV 的固定來源、完整覆蓋、字模與原框文字安全矩形。

必須在 Docker 內執行：python3 tools/verify_world_prompts.py。
正式 renderer 的行數與實際字寬另由 internal/ui 的真實 catalog 測試核對。
"""
from pathlib import Path
import csv, json, re

ROOT = Path(__file__).resolve().parents[1]
norm = lambda lines: ' '.join(' '.join(lines).split())

def expected_sources():
    data = ROOT / 'assets/original-derived-data'
    expected = {}
    for x in json.loads((data/'quest-sites.json').read_text())['sites']:
        expected[f'quest.{x["index"]:02}'] = norm(x['narrative'])
    for x in json.loads((data/'monsters.json').read_text())['monsters']:
        expected[f'monster.{x["index"]:02}'] = norm(x['encounter'])
    for i, lines in enumerate(json.loads((data/'wilderness-scenes.json').read_text())['scenes']):
        expected[f'healer.{i:02}'] = norm(lines)
    ui=json.loads((ROOT/'docs/i18n/ui-zh-Hant.json').read_text())['banks']['KOLDATA']
    for i in [227,228,229,230,234]:
        expected[f'ui.{i}']=norm(re.sub(r'~[VHCvhc][0-9a-fA-F]{2}','',ui[str(i)]['en']).split('@'))
    return expected

def main():
    path=ROOT/'docs/i18n/world-prompts-zh-Hant.tsv'
    with path.open(encoding='utf-8',newline='') as f:
        reader=csv.DictReader(f,delimiter='\t')
        assert reader.fieldnames==['key','en','zh-Hant','status','source']
        rows=list(reader)
    expected=expected_sources()
    assert len(rows)==len(expected)==67
    assert len({row['key'] for row in rows})==len(rows)
    assert len({row['en'] for row in rows})==len(rows)
    assert {row['key'] for row in rows}==set(expected)
    font=json.loads((ROOT/'assets/ui-font/manifest.json').read_text())['fonts'][0]
    glyphs={g['codepoint']:g['advance'] for g in font['glyphs']}
    for row in rows:
        assert row['en']==expected[row['key']],row['key']
        assert row['zh-Hant'].strip() and row['source'] and row['status'] in ['核准','暫譯'],row['key']
        assert '\n' not in row['zh-Hant'] and '\t' not in row['zh-Hant']
        assert not re.search(r'[|~%{}]',row['en']+row['zh-Hant']),row['key']
        missing={c for c in row['zh-Hant'] if not c.isspace() and ord(c) not in glyphs}
        assert not missing,(row['key'],missing)
        # 安全矩形直接使用 Go renderer 的原框契約，由真實 catalog 測試核對；
        # 這裡不另建一份座標或換行演算法。
    print(f'世界提示 {len(rows)} 筆：来源、唯一鍵、完整覆蓋、控制碼與字模通過'.replace('来源','來源'))
    return 0

if __name__=='__main__':
    raise SystemExit(main())
